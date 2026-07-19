"use strict";

/* ------------------------------------------------------------------ *
 *  Tetris — browser frontend (WASM engine + Canvas renderer)
 *  The game logic lives in main.wasm (internal/domain). This file
 *  drives the simulation, renders to <canvas>, handles input, and
 *  persists high scores to localStorage.
 * ------------------------------------------------------------------ */

/* ---- Theme color tables (mirrors internal/renderer/theme.go) ---- */
const THEMES = {
  classic: { bg:"#101010", ghost:"#555555", clearing:"#ffffff",
    pieces:["", "#3ad0e8","#f7d038","#a85cd8","#46c64a","#e0556a","#4a78e0","#e08a3c"] },
  neon:    { bg:"#050507", ghost:"#335577", clearing:"#ffffff",
    pieces:["", "#00ffff","#ffff00","#ff00ff","#00ff66","#ff0040","#3060ff","#ff8000"] },
  retro:   { bg:"#001100", ghost:"#0a5522", clearing:"#aaffcc",
    pieces:["", "#00ff66","#00cc55","#00aa44","#33ff77","#00dd55","#22cc66","#55ff88"] },
  mono:    { bg:"#0a0a0a", ghost:"#444444", clearing:"#ffffff",
    pieces:["", "#ffffff","#dddddd","#bbbbbb","#999999","#888888","#cccccc","#aaaaaa"] },
};

/* ---- Piece shapes (spawn state), indexed by PieceType 0..6 ---- */
const PIECES = [
  { cells:[[1,0],[1,1],[1,2],[1,3]] }, // I
  { cells:[[1,1],[1,2],[2,1],[2,2]] }, // O
  { cells:[[0,1],[1,0],[1,1],[1,2]] }, // T
  { cells:[[0,1],[0,2],[1,0],[1,1]] }, // S
  { cells:[[0,0],[0,1],[1,1],[1,2]] }, // Z
  { cells:[[0,0],[1,0],[1,1],[1,2]] }, // J
  { cells:[[0,2],[1,0],[1,1],[1,2]] }, // L
];

const FIELD_W = 10, FIELD_H = 20;
const DAS = 130;   // ms before auto-shift kicks in
const ARR = 33;    // ms between auto-shifts

/* ---- DOM + canvas handles ---- */
const field = document.getElementById("field");
const fctx = field.getContext("2d");
const CELL = field.width / FIELD_W;

const holdCanvas = document.getElementById("hold");
const hctx = holdCanvas.getContext("2d");
const nextCanvas = document.getElementById("next");
const nctx = nextCanvas.getContext("2d");

const overlay = document.getElementById("overlay");
const overlayCard = document.getElementById("overlay-card");
const themeSel = document.getElementById("theme");

let currentTheme = themeSel.value || "classic";

/* ---- Runtime state ---- */
let started = false;
let screen = "start";            // start | playing | paused | gameover
let gameOverHandled = false;
let lastTime = 0;

let dasDir = null;
let dasTimer = 0;
let dasCharged = false;

/* ------------------------------------------------------------------ *
 *  Boot
 * ------------------------------------------------------------------ */
async function loadWasm() {
  try {
    const go = new Go();
    const res = await fetch("main.wasm");
    const bytes = await res.arrayBuffer();
    const result = await WebAssembly.instantiate(bytes, go.importObject);
    go.run(result.instance);
    init();
  } catch (err) {
    console.error(err);
    showLoadError(err);
  }
}

function init() {
  showStart();
  lastTime = performance.now();
  requestAnimationFrame(frame);
}

/* ------------------------------------------------------------------ *
 *  Main loop
 * ------------------------------------------------------------------ */
function frame(now) {
  const dt = Math.min(now - lastTime, 100);
  lastTime = now;

  if (started && typeof Tetris !== "undefined") {
    Tetris.tick(dt);
    updateInput(dt);
    const snap = Tetris.snapshot();
    if (snap) {
      render(snap);
      syncScreen(snap);
      if (snap.state === "gameover" && !gameOverHandled) handleGameOver(snap);
    }
  } else {
    renderBlank();
  }
  requestAnimationFrame(frame);
}

function updateInput(dt) {
  if (!dasDir || screen !== "playing") return;
  dasTimer += dt;
  if (!dasCharged) {
    if (dasTimer >= DAS) { dasCharged = true; dasTimer = 0; Tetris.act(dasDir); }
  } else {
    while (dasTimer >= ARR) { Tetris.act(dasDir); dasTimer -= ARR; }
  }
}

function syncScreen(snap) {
  if (snap.state === "paused") {
    if (screen !== "paused") { screen = "paused"; showPause(); }
  } else if (snap.state === "playing") {
    if (screen === "paused") { screen = "playing"; hideOverlay(); }
  } else if (snap.state === "gameover") {
    screen = "gameover";
  }
}

/* ------------------------------------------------------------------ *
 *  Rendering
 * ------------------------------------------------------------------ */
function drawCell(ctx, x, y, size, color) {
  const pad = 1;
  ctx.fillStyle = color;
  ctx.fillRect(x + pad, y + pad, size - 2 * pad, size - 2 * pad);
  ctx.fillStyle = "rgba(255,255,255,0.18)";
  ctx.fillRect(x + pad, y + pad, size - 2 * pad, Math.max(2, size * 0.16));
}

function render(snap) {
  const theme = THEMES[currentTheme];

  fctx.fillStyle = theme.bg;
  fctx.fillRect(0, 0, field.width, field.height);

  for (let r = 0; r < FIELD_H; r++) {
    for (let c = 0; c < FIELD_W; c++) {
      const v = snap.grid[r * FIELD_W + c];
      if (v > 0) drawCell(fctx, c * CELL, r * CELL, CELL, theme.pieces[v]);
    }
  }

  if (snap.clearing) {
    fctx.fillStyle = theme.clearing;
    for (const r of snap.clearRows) fctx.fillRect(0, r * CELL, field.width, CELL);
  }

  if (snap.ghost) {
    fctx.globalAlpha = 0.55;
    fctx.fillStyle = theme.ghost;
    for (const [r, c] of snap.ghost.cells) {
      if (r < 0 || r >= FIELD_H || c < 0 || c >= FIELD_W) continue;
      fctx.fillRect(c * CELL + 2, r * CELL + 2, CELL - 4, CELL - 4);
    }
    fctx.globalAlpha = 1;
  }

  if (snap.active) {
    for (const [r, c] of snap.active.cells) {
      if (r < 0 || r >= FIELD_H || c < 0 || c >= FIELD_W) continue;
      drawCell(fctx, c * CELL, r * CELL, CELL, theme.pieces[snap.active.color]);
    }
  }

  updateStats(snap);
  drawHold(snap.hold, theme);
  drawNext(snap.next, theme);
}

function renderBlank() {
  const theme = THEMES[currentTheme];
  fctx.fillStyle = theme.bg;
  fctx.fillRect(0, 0, field.width, field.height);
  fctx.strokeStyle = "rgba(255,255,255,0.05)";
  fctx.lineWidth = 1;
  for (let c = 1; c < FIELD_W; c++) {
    fctx.beginPath(); fctx.moveTo(c * CELL, 0); fctx.lineTo(c * CELL, field.height); fctx.stroke();
  }
  for (let r = 1; r < FIELD_H; r++) {
    fctx.beginPath(); fctx.moveTo(0, r * CELL); fctx.lineTo(field.width, r * CELL); fctx.stroke();
  }
  drawHold(null, theme);
  drawNext([], theme);
  setText("score", 0); setText("level", 1); setText("lines", 0);
  setText("combo", "—"); setText("b2b", "—"); setText("time", "00:00");
}

function drawPieceCentered(ctx, type, color, areaW, areaH, cell, offsetY) {
  offsetY = offsetY || 0;
  const cells = PIECES[type].cells;
  let minR = Infinity, maxR = -Infinity, minC = Infinity, maxC = -Infinity;
  for (const [r, c] of cells) {
    if (r < minR) minR = r; if (r > maxR) maxR = r;
    if (c < minC) minC = c; if (c > maxC) maxC = c;
  }
  const w = (maxC - minC + 1) * cell;
  const h = (maxR - minR + 1) * cell;
  const ox = (areaW - w) / 2 - minC * cell;
  const oy = offsetY + (areaH - h) / 2 - minR * cell;
  for (const [r, c] of cells) drawCell(ctx, ox + c * cell, oy + r * cell, cell, color);
}

function drawHold(piece, theme) {
  hctx.fillStyle = theme.bg;
  hctx.fillRect(0, 0, holdCanvas.width, holdCanvas.height);
  if (piece) drawPieceCentered(hctx, piece.type, theme.pieces[piece.color], holdCanvas.width, holdCanvas.height, 18);
}

function drawNext(queue, theme) {
  nctx.fillStyle = theme.bg;
  nctx.fillRect(0, 0, nextCanvas.width, nextCanvas.height);
  const slotH = nextCanvas.height / 5;
  queue.forEach((p, i) => drawPieceCentered(nctx, p.type, theme.pieces[p.color], nextCanvas.width, slotH, 16, i * slotH));
}

function updateStats(snap) {
  setText("score", snap.score);
  setText("level", snap.level);
  setText("lines", snap.lines);
  setText("combo", snap.combo > 0 ? snap.combo : "—");
  setText("b2b", snap.b2b ? "YES" : "—");
  setText("time", formatTime(snap.elapsedMs));
}

/* ------------------------------------------------------------------ *
 *  Overlays + game lifecycle
 * ------------------------------------------------------------------ */
function startGame() {
  if (typeof Tetris === "undefined") return;
  Tetris.new({});
  started = true;
  screen = "playing";
  gameOverHandled = false;
  hideOverlay();
}

function showStart() {
  overlayCard.innerHTML =
    '<h2>TETRIS</h2>' +
    '<p>Stack the tetrominoes, clear lines, and don’t top out.</p>' +
    '<p class="muted">Press Enter or tap Play to begin</p>' +
    '<button class="primary" id="play-btn">Play</button>';
  overlay.classList.add("show");
  document.getElementById("play-btn").onclick = startGame;
}

function showPause() {
  overlayCard.innerHTML =
    '<h2>PAUSED</h2>' +
    '<p>Press P or tap ⏸ to resume</p>' +
    '<button class="primary" id="resume-btn">Resume</button>';
  overlay.classList.add("show");
  document.getElementById("resume-btn").onclick = () => { if (typeof Tetris !== "undefined") Tetris.act("pause"); };
}

function handleGameOver(snap) {
  gameOverHandled = true;
  const scores = loadScores();
  const qualifies = snap.score > 0 && (scores.length < 10 || snap.score > scores[scores.length - 1].score);

  if (qualifies) {
    overlayCard.innerHTML =
      '<h2>GAME OVER</h2>' +
      '<div class="big">' + snap.score + '</div>' +
      '<p>Level ' + snap.level + ' · ' + snap.lines + ' lines</p>' +
      '<p class="muted">New high score! Enter your name:</p>' +
      '<input id="hs-name" maxlength="12" placeholder="Player" />' +
      '<button class="primary" id="save-btn">Save Score</button>';
    overlay.classList.add("show");
    const input = document.getElementById("hs-name");
    input.focus();
    input.addEventListener("keydown", (e) => { if (e.key === "Enter") document.getElementById("save-btn").click(); });
    document.getElementById("save-btn").onclick = () => {
      const name = (input.value.trim() || "Player").slice(0, 12);
      saveScore(name, snap.score, snap.level, snap.lines);
      showGameOverBoard(snap);
    };
  } else {
    showGameOverBoard(snap);
  }
}

function showGameOverBoard(snap) {
  const scores = loadScores();
  const rows = scores.length
    ? scores.slice(0, 10).map((s) => '<li>' + escapeHtml(s.name) + ' — ' + s.score + ' <span class="muted">(L' + s.level + ')</span></li>').join("")
    : '<li class="muted">No scores yet</li>';
  overlayCard.innerHTML =
    '<h2>GAME OVER</h2>' +
    '<div class="big">' + snap.score + '</div>' +
    '<p>Level ' + snap.level + ' · ' + snap.lines + ' lines</p>' +
    '<p class="muted">High Scores</p>' +
    '<ol>' + rows + '</ol>' +
    '<button class="primary" id="again-btn">Play Again</button>';
  overlay.classList.add("show");
  document.getElementById("again-btn").onclick = startGame;
}

function showLoadError(err) {
  overlayCard.innerHTML =
    '<h2>Failed to load</h2>' +
    '<p class="muted">' + escapeHtml(String(err)) + '</p>' +
    '<p class="muted">Serve this folder over HTTP (not file://) and make sure main.wasm is present.</p>';
  overlay.classList.add("show");
}

function hideOverlay() { overlay.classList.remove("show"); }

/* ------------------------------------------------------------------ *
 *  High scores (localStorage)
 * ------------------------------------------------------------------ */
const HS_KEY = "tetris.highscores.v1";

function loadScores() {
  try { return JSON.parse(localStorage.getItem(HS_KEY)) || []; }
  catch { return []; }
}

function saveScore(name, score, level, lines) {
  const scores = loadScores();
  scores.push({ name, score, level, lines, date: new Date().toISOString().slice(0, 10) });
  scores.sort((a, b) => b.score - a.score);
  localStorage.setItem(HS_KEY, JSON.stringify(scores.slice(0, 10)));
}

/* ------------------------------------------------------------------ *
 *  Input: keyboard
 * ------------------------------------------------------------------ */
const keyMap = {
  ArrowLeft: "moveLeft",
  ArrowRight: "moveRight",
  ArrowDown: "softDrop",
  ArrowUp: "rotateCW",
  " ": "hardDrop",
  z: "rotateCCW", Z: "rotateCCW",
  a: "rotate180", A: "rotate180",
  c: "hold", C: "hold",
  p: "pause", P: "pause",
  r: "restart", R: "restart",
  Enter: "confirm",
};

function onKeyDown(e) {
  if (e.target && e.target.tagName === "INPUT") return; // typing a name
  const action = keyMap[e.key];
  if (!action) return;
  if (["ArrowLeft", "ArrowRight", "ArrowUp", "ArrowDown", " "].includes(e.key)) e.preventDefault();

  if (screen === "start") {
    if (action === "confirm" || action === "hardDrop") startGame();
    return;
  }
  if (screen === "gameover") {
    if (action === "confirm" || action === "restart") startGame();
    return;
  }
  if (screen === "paused") {
    if (action === "pause" && !e.repeat) Tetris.act("pause");
    else if (action === "restart" && !e.repeat) startGame();
    return;
  }

  switch (action) {
    case "confirm": return;
    case "softDrop":
      if (!e.repeat) Tetris.setSoftDrop(true);
      return;
    case "moveLeft":
    case "moveRight":
      if (!e.repeat) {
        Tetris.act(action);
        dasDir = action; dasTimer = 0; dasCharged = false;
      }
      return;
    case "pause":
      if (!e.repeat) Tetris.act("pause");
      return;
    case "restart":
      if (!e.repeat) startGame();
      return;
    default:
      if (!e.repeat) Tetris.act(action);
  }
}

function onKeyUp(e) {
  const action = keyMap[e.key];
  if (action === "softDrop") Tetris.setSoftDrop(false);
  if ((action === "moveLeft" || action === "moveRight") && dasDir === action) dasDir = null;
}

/* ------------------------------------------------------------------ *
 *  Input: touch / pointer
 * ------------------------------------------------------------------ */
function bindTouch() {
  document.querySelectorAll("#touch .btn").forEach((btn) => {
    const action = btn.dataset.action;
    const repeat = btn.dataset.repeat;
    let timer = null;

    const start = (e) => {
      e.preventDefault();
      if (screen === "start" || screen === "gameover") {
        if (action === "hardDrop" || action === "pause") startGame();
        return;
      }
      if (action === "softDrop") { Tetris.setSoftDrop(true); return; }
      Tetris.act(action);
      if (repeat === "true" || repeat === "hold") {
        timer = setInterval(() => {
          if (action === "softDrop") Tetris.setSoftDrop(true);
          else Tetris.act(action);
        }, repeat === "hold" ? 40 : 90);
      }
    };
    const end = (e) => {
      e.preventDefault();
      if (action === "softDrop") Tetris.setSoftDrop(false);
      if (timer) { clearInterval(timer); timer = null; }
    };

    btn.addEventListener("pointerdown", start);
    btn.addEventListener("pointerup", end);
    btn.addEventListener("pointercancel", end);
    btn.addEventListener("pointerleave", end);
  });
}

/* ------------------------------------------------------------------ *
 *  Helpers
 * ------------------------------------------------------------------ */
function setText(id, val) {
  const el = document.getElementById(id);
  if (el) el.textContent = val;
}

function formatTime(ms) {
  const total = Math.floor(ms / 1000);
  const m = String(Math.floor(total / 60)).padStart(2, "0");
  const s = String(total % 60).padStart(2, "0");
  return m + ":" + s;
}

function escapeHtml(s) {
  return String(s).replace(/[&<>"']/g, (c) =>
    ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c]));
}

/* ------------------------------------------------------------------ *
 *  Wire up
 * ------------------------------------------------------------------ */
themeSel.addEventListener("change", () => {
  currentTheme = themeSel.value;
  if (!started) renderBlank();
});
window.addEventListener("keydown", onKeyDown);
window.addEventListener("keyup", onKeyUp);
bindTouch();
loadWasm();
