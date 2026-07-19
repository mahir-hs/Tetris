"use strict";

/* ------------------------------------------------------------------ *
 *  Tetris — browser frontend (WASM engine + Canvas renderer)
 *  The game logic lives in main.wasm (internal/domain). This file
 *  drives the simulation, renders to <canvas>, handles input, and
 *  persists settings + high scores + statistics to localStorage.
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

/* ---- Persisted settings ---- */
const SETTINGS_KEY = "tetris.settings.v1";
const DEFAULT_SETTINGS = { startingLevel:1, ghostEnabled:true, holdEnabled:true, rotate180Enabled:true, theme:"classic" };
function loadSettings() {
  try { return Object.assign({}, DEFAULT_SETTINGS, JSON.parse(localStorage.getItem(SETTINGS_KEY)) || {}); }
  catch { return Object.assign({}, DEFAULT_SETTINGS); }
}
function saveSettings() { localStorage.setItem(SETTINGS_KEY, JSON.stringify(settings)); }
let settings = loadSettings();
let currentTheme = settings.theme || "classic";

/* ---- Persisted lifetime statistics ---- */
const STATS_KEY = "tetris.stats.v1";
function loadStats() { try { return JSON.parse(localStorage.getItem(STATS_KEY)) || {}; } catch { return {}; } }
function saveStats(s) { localStorage.setItem(STATS_KEY, JSON.stringify(s)); }
function recordStats(snap) {
  const s = loadStats();
  s.gamesPlayed = (s.gamesPlayed || 0) + 1;
  s.highScore = Math.max(s.highScore || 0, snap.score);
  s.highLevel = Math.max(s.highLevel || 0, snap.level);
  s.totalLines = (s.totalLines || 0) + snap.lines;
  s.totalScore = (s.totalScore || 0) + snap.score;
  s.totalPlayTimeMs = (s.totalPlayTimeMs || 0) + (snap.elapsedMs || 0);
  const st = snap.stats || {};
  s.longestCombo = Math.max(s.longestCombo || 0, st.longestCombo || 0);
  s.totalTetrises = (s.totalTetrises || 0) + (st.tetrises || 0);
  s.totalTSpins = (s.totalTSpins || 0) + (st.tSpins || 0);
  s.totalMiniTSpins = (s.totalMiniTSpins || 0) + (st.miniTSpins || 0);
  s.totalPerfectClears = (s.totalPerfectClears || 0) + (st.perfectClears || 0);
  s.totalPieces = (s.totalPieces || 0) + (st.piecesPlaced || 0);
  saveStats(s);
}

/* ---- High scores ---- */
const HS_KEY = "tetris.highscores.v1";
function loadScores() { try { return JSON.parse(localStorage.getItem(HS_KEY)) || []; } catch { return []; } }
function saveScore(name, score, level, lines) {
  const scores = loadScores();
  scores.push({ name, score, level, lines, date: new Date().toISOString().slice(0, 10) });
  scores.sort((a, b) => b.score - a.score);
  localStorage.setItem(HS_KEY, JSON.stringify(scores.slice(0, 10)));
}

/* ---- Runtime state ---- */
let started = false;
let screen = "menu";            // menu | highscores | statistics | settings | credits | playing | paused | gameover
let gameOverHandled = false;
let lastTime = 0;
let lastSnap = null;

let dasDir = null, dasTimer = 0, dasCharged = false;

/* ---- Menu model ---- */
const MENU_ITEMS = [
  { id:"play", label:"Play" },
  { id:"highscores", label:"High Scores" },
  { id:"statistics", label:"Statistics" },
  { id:"settings", label:"Settings" },
  { id:"credits", label:"Credits" },
];
let menuIndex = 0;

const SETTING_ROWS = [
  { key:"startingLevel", label:"Starting Level", type:"num", min:1, max:20 },
  { key:"ghostEnabled", label:"Ghost Piece", type:"bool" },
  { key:"holdEnabled", label:"Hold Piece", type:"bool" },
  { key:"rotate180Enabled", label:"180° Rotation", type:"bool" },
];
let settingsIndex = 0;

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
  themeSel.value = currentTheme;
  showMenu();
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
      lastSnap = snap;
      if (screen === "playing" || screen === "paused") {
        render(snap);
        syncScreen(snap);
        if (snap.state === "gameover" && !gameOverHandled) handleGameOver(snap);
      } else {
        render(snap); // frozen backdrop behind a menu screen
      }
    }
  } else if (lastSnap) {
    render(lastSnap);
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
 *  Screens
 * ------------------------------------------------------------------ */
function startGame() {
  if (typeof Tetris === "undefined") return;
  Tetris.new({
    startingLevel: settings.startingLevel,
    ghost: settings.ghostEnabled,
    hold: settings.holdEnabled,
    rotate180: settings.rotate180Enabled,
  });
  started = true;
  screen = "playing";
  gameOverHandled = false;
  hideOverlay();
}

function showMenu() {
  screen = "menu";
  overlayCard.innerHTML =
    '<h2>TETRIS</h2><ul class="menu">' +
    MENU_ITEMS.map((it, i) => `<li class="${i === menuIndex ? "sel" : ""}" data-id="${it.id}">${it.label}</li>`).join("") +
    '</ul><p class="muted">↑ ↓ move · Enter select</p>';
  overlay.classList.add("show");
  overlayCard.querySelectorAll(".menu li").forEach((li) => { li.onclick = () => selectMenu(li.dataset.id); });
}

function selectMenu(id) {
  if (id === "play") startGame();
  else if (id === "highscores") showHighScores();
  else if (id === "statistics") showStatistics();
  else if (id === "settings") showSettings();
  else if (id === "credits") showCredits();
}

function showHighScores() {
  screen = "highscores";
  const scores = loadScores();
  const rows = scores.length
    ? scores.slice(0, 10).map((s) => `<li>${escapeHtml(s.name)} — ${s.score} <span class="muted">(L${s.level})</span></li>`).join("")
    : '<li class="muted">No scores yet</li>';
  overlayCard.innerHTML = `<h2>HIGH SCORES</h2><ol>${rows}</ol><button class="primary" id="back-btn">Back</button>`;
  overlay.classList.add("show");
  document.getElementById("back-btn").onclick = showMenu;
}

function showStatistics() {
  screen = "statistics";
  const s = loadStats();
  const fmtTime = (ms) => {
    const t = Math.floor((ms || 0) / 1000);
    return String(Math.floor(t / 60)).padStart(2, "0") + ":" + String(t % 60).padStart(2, "0");
  };
  const rows = [
    ["Games Played", s.gamesPlayed || 0],
    ["High Score", s.highScore || 0],
    ["High Level", s.highLevel || 0],
    ["Total Lines", s.totalLines || 0],
    ["Total Score", s.totalScore || 0],
    ["Total Play Time", fmtTime(s.totalPlayTimeMs)],
    ["Longest Combo", s.longestCombo || 0],
    ["Tetrises", s.totalTetrises || 0],
    ["T-Spins", s.totalTSpins || 0],
    ["Mini T-Spins", s.totalMiniTSpins || 0],
    ["Perfect Clears", s.totalPerfectClears || 0],
    ["Pieces Placed", s.totalPieces || 0],
  ];
  overlayCard.innerHTML =
    "<h2>STATISTICS</h2>" +
    rows.map(([k, v]) => `<div class="stat2"><span>${k}</span><b>${v}</b></div>`).join("") +
    '<button class="primary" id="back-btn">Back</button>';
  overlay.classList.add("show");
  document.getElementById("back-btn").onclick = showMenu;
}

function showSettings() {
  screen = "settings";
  const rows = SETTING_ROWS.map((r, i) => {
    const val = r.type === "num" ? String(settings[r.key]) : (settings[r.key] ? "ON" : "OFF");
    return `<div class="setting-row${i === settingsIndex ? " sel" : ""}" data-i="${i}">` +
      `<span>${r.label}</span><span class="controls">` +
      `<button data-act="dec" data-i="${i}">−</button><b>${val}</b><button data-act="inc" data-i="${i}">+</button>` +
      `</span></div>`;
  }).join("");
  overlayCard.innerHTML =
    "<h2>SETTINGS</h2>" + rows +
    '<p class="muted">↑ ↓ select · ← → adjust · Esc back</p>' +
    '<button class="primary" id="back-btn">Back</button>';
  overlay.classList.add("show");
  overlayCard.querySelectorAll(".setting-row").forEach((row) => {
    row.onclick = () => { settingsIndex = +row.dataset.i; showSettings(); };
  });
  overlayCard.querySelectorAll("button[data-act]").forEach((b) => {
    b.onclick = (e) => {
      e.stopPropagation();
      settingsIndex = +b.dataset.i;
      adjustSetting(SETTING_ROWS[settingsIndex].key, b.dataset.act === "dec" ? -1 : 1);
    };
  });
  document.getElementById("back-btn").onclick = showMenu;
}

function adjustSetting(key, delta) {
  const row = SETTING_ROWS.find((r) => r.key === key);
  if (!row) return;
  if (row.type === "num") settings[key] = Math.min(row.max, Math.max(row.min, settings[key] + delta));
  else settings[key] = !settings[key];
  saveSettings();
  showSettings();
}

function showCredits() {
  screen = "credits";
  overlayCard.innerHTML =
    "<h2>CREDITS</h2>" +
    "<p>Tetris — browser edition</p>" +
    '<p class="muted">Engine: Go <code>internal/domain</code> compiled to WebAssembly.</p>' +
    '<p class="muted">Rendering: HTML5 Canvas. A port of the CLI Tetris.</p>' +
    '<p class="muted">Guideline rules: 7-bag, SRS, T-spin, combos, B2B, perfect clear.</p>' +
    '<button class="primary" id="back-btn">Back</button>';
  overlay.classList.add("show");
  document.getElementById("back-btn").onclick = showMenu;
}

function showPause() {
  overlayCard.innerHTML =
    "<h2>PAUSED</h2><p>Press P or tap ⏸ to resume</p>" +
    '<button class="primary" id="resume-btn">Resume</button>';
  overlay.classList.add("show");
  document.getElementById("resume-btn").onclick = () => { if (typeof Tetris !== "undefined") Tetris.act("pause"); };
}

function handleGameOver(snap) {
  gameOverHandled = true;
  recordStats(snap);
  const scores = loadScores();
  const qualifies = snap.score > 0 && (scores.length < 10 || snap.score > scores[scores.length - 1].score);

  if (qualifies) {
    overlayCard.innerHTML =
      "<h2>GAME OVER</h2>" +
      `<div class="big">${snap.score}</div>` +
      `<p>Level ${snap.level} · ${snap.lines} lines</p>` +
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
    ? scores.slice(0, 10).map((s) => `<li>${escapeHtml(s.name)} — ${s.score} <span class="muted">(L${s.level})</span></li>`).join("")
    : '<li class="muted">No scores yet</li>';
  overlayCard.innerHTML =
    "<h2>GAME OVER</h2>" +
    `<div class="big">${snap.score}</div>` +
    `<p>Level ${snap.level} · ${snap.lines} lines</p>` +
    '<p class="muted">High Scores</p><ol>' + rows + "</ol>" +
    '<button class="primary" id="again-btn">Play Again</button>' +
    '<button class="secondary" id="menu-btn">Menu</button>';
  overlay.classList.add("show");
  document.getElementById("again-btn").onclick = startGame;
  document.getElementById("menu-btn").onclick = showMenu;
}

function showLoadError(err) {
  overlayCard.innerHTML =
    "<h2>Failed to load</h2>" +
    `<p class="muted">${escapeHtml(String(err))}</p>` +
    '<p class="muted">Serve this folder over HTTP (not file://) and make sure main.wasm is present.</p>';
  overlay.classList.add("show");
}

function hideOverlay() { overlay.classList.remove("show"); }

/* ------------------------------------------------------------------ *
 *  Input: keyboard
 * ------------------------------------------------------------------ */
const keyMap = {
  ArrowLeft: "moveLeft", ArrowRight: "moveRight", ArrowDown: "softDrop", ArrowUp: "rotateCW",
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
  const k = e.key;
  if (["ArrowLeft", "ArrowRight", "ArrowUp", "ArrowDown", " "].includes(k)) e.preventDefault();

  if (screen === "menu") {
    if (k === "ArrowUp" || k === "k") menuIndex = (menuIndex - 1 + MENU_ITEMS.length) % MENU_ITEMS.length;
    else if (k === "ArrowDown" || k === "j") menuIndex = (menuIndex + 1) % MENU_ITEMS.length;
    else if (k === "Enter") selectMenu(MENU_ITEMS[menuIndex].id);
    else return;
    if (k === "ArrowUp" || k === "ArrowDown" || k === "k" || k === "j") showMenu();
    return;
  }

  if (screen === "settings") {
    if (k === "ArrowUp" || k === "k") settingsIndex = Math.max(0, settingsIndex - 1);
    else if (k === "ArrowDown" || k === "j") settingsIndex = Math.min(SETTING_ROWS.length - 1, settingsIndex + 1);
    else if (k === "ArrowLeft") adjustSetting(SETTING_ROWS[settingsIndex].key, -1);
    else if (k === "ArrowRight") adjustSetting(SETTING_ROWS[settingsIndex].key, 1);
    else if (k === "Enter") adjustSetting(SETTING_ROWS[settingsIndex].key, 0);
    else if (k === "Escape") showMenu();
    else return;
    if (k !== "Escape") showSettings();
    return;
  }

  if (screen === "highscores" || screen === "statistics" || screen === "credits") {
    if (k === "Escape" || k === "Enter" || k === "Backspace") showMenu();
    return;
  }

  if (screen === "gameover") {
    if (k === "Enter" || k === " " || k === "r" || k === "R") startGame();
    else if (k === "Escape" || k === "m" || k === "M") showMenu();
    return;
  }

  if (screen === "paused") {
    if ((k === "p" || k === "P") && !e.repeat) Tetris.act("pause");
    else if ((k === "r" || k === "R") && !e.repeat) startGame();
    return;
  }

  // screen === "playing" — gameplay
  const action = keyMap[k];
  if (!action) return;
  switch (action) {
    case "confirm": return;
    case "softDrop": if (!e.repeat) Tetris.setSoftDrop(true); return;
    case "moveLeft":
    case "moveRight":
      if (!e.repeat) { Tetris.act(action); dasDir = action; dasTimer = 0; dasCharged = false; }
      return;
    case "pause": if (!e.repeat) Tetris.act("pause"); return;
    case "restart": if (!e.repeat) startGame(); return;
    default: if (!e.repeat) Tetris.act(action);
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
      if (screen === "menu" || screen === "gameover") {
        if (action === "hardDrop" || action === "pause") startGame();
        return;
      }
      if (screen !== "playing") return;
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
function setText(id, val) { const el = document.getElementById(id); if (el) el.textContent = val; }

function formatTime(ms) {
  const total = Math.floor(ms / 1000);
  return String(Math.floor(total / 60)).padStart(2, "0") + ":" + String(total % 60).padStart(2, "0");
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
  settings.theme = currentTheme;
  saveSettings();
  if (!started) renderBlank();
});
window.addEventListener("keydown", onKeyDown);
window.addEventListener("keyup", onKeyUp);
bindTouch();
loadWasm();
