"use strict";

const THEMES = {
  classic: {
    bg: "#080a0f", ghost: "#526579", clearing: "#ffffff", uiBg: "#080a12",
    pieces: ["", "#3ad0e8", "#f7d038", "#a85cd8", "#46c64a", "#e0556a", "#4a78e0", "#e08a3c"],
  },
  neon: {
    bg: "#030305", ghost: "#3e4770", clearing: "#ffffff", uiBg: "#040308",
    pieces: ["", "#00ffff", "#ffff00", "#ff00ff", "#00ff66", "#ff174f", "#4771ff", "#ff8c1a"],
  },
  retro: {
    bg: "#001006", ghost: "#17682f", clearing: "#caffda", uiBg: "#001006",
    pieces: ["", "#00ff66", "#00cc55", "#00aa44", "#33ff77", "#00dd55", "#22cc66", "#55ff88"],
  },
  mono: {
    bg: "#070707", ghost: "#555555", clearing: "#ffffff", uiBg: "#080808",
    pieces: ["", "#ffffff", "#dddddd", "#bbbbbb", "#999999", "#888888", "#cccccc", "#aaaaaa"],
  },
};

const MODES = {
  marathon: {
    label: "MARATHON",
    description: "Classic endless play. Build your score as gravity increases.",
    resultMetric: "score",
  },
  sprint: {
    label: "SPRINT",
    description: "Clear 40 lines as quickly as possible.",
    resultMetric: "time",
  },
  ultra: {
    label: "ULTRA",
    description: "Score as much as possible in two minutes.",
    resultMetric: "score",
  },
  zen: {
    label: "ZEN",
    description: "Relaxed endless play. The board resets instead of topping out.",
    resultMetric: "score",
  },
};

const PIECES = [
  { name: "I", cells: [[1, 0], [1, 1], [1, 2], [1, 3]] },
  { name: "O", cells: [[1, 1], [1, 2], [2, 1], [2, 2]] },
  { name: "T", cells: [[0, 1], [1, 0], [1, 1], [1, 2]] },
  { name: "S", cells: [[0, 1], [0, 2], [1, 0], [1, 1]] },
  { name: "Z", cells: [[0, 0], [0, 1], [1, 1], [1, 2]] },
  { name: "J", cells: [[0, 0], [1, 0], [1, 1], [1, 2]] },
  { name: "L", cells: [[0, 2], [1, 0], [1, 1], [1, 2]] },
];

const FIELD_W = 10;
const FIELD_H = 20;
const DAS = 130;
const ARR = 33;
const ULTRA_DURATION_MS = 120000;
const SPRINT_LINES = 40;

const field = document.getElementById("field");
const fctx = field.getContext("2d");
const holdCanvas = document.getElementById("hold");
const hctx = holdCanvas.getContext("2d");
const nextCanvas = document.getElementById("next");
const nctx = nextCanvas.getContext("2d");
const overlay = document.getElementById("overlay");
const overlayCard = document.getElementById("overlay-card");
const pauseControl = document.getElementById("pause-control");
const modeLabel = document.getElementById("mode-label");
const callout = document.getElementById("callout");
const calloutTitle = document.getElementById("callout-title");
const calloutDetail = document.getElementById("callout-detail");
const liveStatus = document.getElementById("live-status");

const SETTINGS_KEY = "tetris.settings.v2";
const LEGACY_SETTINGS_KEY = "tetris.settings.v1";
const STATS_KEY = "tetris.stats.v2";
const LEGACY_STATS_KEY = "tetris.stats.v1";
const RECORDS_KEY = "tetris.records.v2";
const LEGACY_SCORES_KEY = "tetris.highscores.v1";

const DEFAULT_SETTINGS = {
  startingLevel: 1,
  ghostEnabled: true,
  holdEnabled: true,
  rotate180Enabled: true,
  theme: "classic",
  gestureSensitivity: 2,
  hapticsEnabled: true,
  soundEnabled: true,
  volume: 6,
  reducedMotion: window.matchMedia("(prefers-reduced-motion: reduce)").matches,
  highContrast: false,
  gestureTutorialSeen: false,
  lastMode: "marathon",
  playerName: "Player",
};

const MENU_ITEMS = [
  { id: "play", label: "Play" },
  { id: "leaderboards", label: "Leaderboards" },
  { id: "statistics", label: "Statistics" },
  { id: "howto", label: "How to Play" },
  { id: "settings", label: "Settings" },
  { id: "credits", label: "Credits" },
];

const SETTING_ROWS = [
  { key: "startingLevel", label: "Starting Level", type: "number", min: 1, max: 20 },
  { key: "theme", label: "Theme", type: "choice", choices: ["classic", "neon", "retro", "mono"] },
  { key: "gestureSensitivity", label: "Gesture Sensitivity", type: "choice", choices: [1, 2, 3], names: ["Low", "Medium", "High"] },
  { key: "ghostEnabled", label: "Ghost Piece", type: "boolean" },
  { key: "holdEnabled", label: "Hold Piece", type: "boolean" },
  { key: "rotate180Enabled", label: "180 Rotation", type: "boolean" },
  { key: "hapticsEnabled", label: "Haptics", type: "boolean" },
  { key: "soundEnabled", label: "Sound Effects", type: "boolean" },
  { key: "volume", label: "Sound Volume", type: "number", min: 0, max: 10, suffix: "0" },
  { key: "reducedMotion", label: "Reduced Motion", type: "boolean" },
  { key: "highContrast", label: "High Contrast", type: "boolean" },
];

function storageRead(key, fallback) {
  try {
    const value = JSON.parse(localStorage.getItem(key));
    return value == null ? fallback : value;
  } catch {
    return fallback;
  }
}

function storageWrite(key, value) {
  try {
    localStorage.setItem(key, JSON.stringify(value));
    return true;
  } catch (error) {
    console.error("Unable to save Tetris data", error);
    announce("Your browser could not save this change.");
    return false;
  }
}

function loadSettings() {
  const current = storageRead(SETTINGS_KEY, null);
  const legacy = storageRead(LEGACY_SETTINGS_KEY, {});
  const saved = current || legacy;
  const loaded = Object.assign({}, DEFAULT_SETTINGS, saved || {});
  loaded.startingLevel = clamp(Number(loaded.startingLevel) || 1, 1, 20);
  loaded.gestureSensitivity = clamp(Number(loaded.gestureSensitivity) || 2, 1, 3);
  if (!current && Number(loaded.volume) <= 1 && legacy && Object.prototype.hasOwnProperty.call(legacy, "volume")) {
    loaded.volume = Math.round(Number(loaded.volume) * 10);
  }
  loaded.volume = clamp(Number(loaded.volume) || 0, 0, 10);
  if (!THEMES[loaded.theme]) loaded.theme = "classic";
  if (!MODES[loaded.lastMode]) loaded.lastMode = "marathon";
  return loaded;
}

let settings = loadSettings();

function saveSettings() {
  storageWrite(SETTINGS_KEY, settings);
}

function emptyRecords() {
  return { marathon: [], sprint: [], ultra: [], zen: [] };
}

function loadRecords() {
  const records = Object.assign(emptyRecords(), storageRead(RECORDS_KEY, {}));
  for (const mode of Object.keys(MODES)) {
    if (!Array.isArray(records[mode])) records[mode] = [];
  }
  const legacy = storageRead(LEGACY_SCORES_KEY, []);
  if (records.marathon.length === 0 && Array.isArray(legacy) && legacy.length > 0) {
    records.marathon = legacy.map((entry) => Object.assign({ mode: "marathon", elapsedMs: 0 }, entry));
    storageWrite(RECORDS_KEY, records);
  }
  return records;
}

function scoreValue(mode, entry) {
  return mode === "sprint" ? Number(entry.elapsedMs) || Infinity : Number(entry.score) || 0;
}

function sortRecords(mode, entries) {
  return entries.sort((a, b) => {
    if (mode === "sprint") return scoreValue(mode, a) - scoreValue(mode, b);
    return scoreValue(mode, b) - scoreValue(mode, a);
  });
}

function isBetter(mode, value, comparison) {
  return mode === "sprint" ? value < comparison : value > comparison;
}

function eligibleForRecords(mode, outcome, snap) {
  if (mode === "sprint") return outcome === "complete" && snap.elapsedMs > 0;
  if (mode === "ultra") return (outcome === "complete" || outcome === "gameover") && snap.score > 0;
  if (mode === "zen") return outcome === "ended" && snap.score > 0;
  return outcome === "gameover" && snap.score > 0;
}

function runIsPersonalBest(mode, snap, outcome) {
  if (!eligibleForRecords(mode, outcome, snap)) return false;
  const entries = sortRecords(mode, loadRecords()[mode].slice());
  if (entries.length === 0) return true;
  return isBetter(mode, scoreValue(mode, snap), scoreValue(mode, entries[0]));
}

function runQualifies(mode, snap, outcome) {
  if (!eligibleForRecords(mode, outcome, snap)) return false;
  const entries = sortRecords(mode, loadRecords()[mode].slice());
  if (entries.length < 10) return true;
  return isBetter(mode, scoreValue(mode, snap), scoreValue(mode, entries[entries.length - 1]));
}

function saveRecord(name, mode, snap) {
  const records = loadRecords();
  records[mode].push({
    name,
    mode,
    score: snap.score,
    level: snap.level,
    lines: snap.lines,
    elapsedMs: snap.elapsedMs,
    date: new Date().toISOString().slice(0, 10),
  });
  records[mode] = sortRecords(mode, records[mode]).slice(0, 10);
  storageWrite(RECORDS_KEY, records);
}

function loadStats() {
  const saved = storageRead(STATS_KEY, null);
  if (saved) return saved;
  const legacy = storageRead(LEGACY_STATS_KEY, {});
  return Object.assign({ modes: {} }, legacy);
}

function recordStats(snap, mode, outcome) {
  const stats = loadStats();
  const notable = snap.stats || {};
  stats.gamesPlayed = (stats.gamesPlayed || 0) + 1;
  stats.completedGames = (stats.completedGames || 0) + (outcome === "complete" || outcome === "ended" ? 1 : 0);
  stats.highScore = Math.max(stats.highScore || 0, snap.score || 0);
  stats.highLevel = Math.max(stats.highLevel || 0, snap.level || 0);
  stats.totalLines = (stats.totalLines || 0) + (snap.lines || 0);
  stats.totalScore = (stats.totalScore || 0) + (snap.score || 0);
  stats.totalPlayTimeMs = (stats.totalPlayTimeMs || 0) + (snap.elapsedMs || 0);
  stats.longestGameMs = Math.max(stats.longestGameMs || 0, snap.elapsedMs || 0);
  stats.longestCombo = Math.max(stats.longestCombo || 0, notable.longestCombo || 0);
  stats.totalTetrises = (stats.totalTetrises || 0) + (notable.tetrises || 0);
  stats.totalTSpins = (stats.totalTSpins || 0) + (notable.tSpins || 0);
  stats.totalMiniTSpins = (stats.totalMiniTSpins || 0) + (notable.miniTSpins || 0);
  stats.totalPerfectClears = (stats.totalPerfectClears || 0) + (notable.perfectClears || 0);
  stats.totalPieces = (stats.totalPieces || 0) + (notable.piecesPlaced || 0);
  stats.zenResets = (stats.zenResets || 0) + (notable.zenResets || 0);
  stats.modes = stats.modes || {};
  const modeStats = stats.modes[mode] || {};
  modeStats.plays = (modeStats.plays || 0) + 1;
  modeStats.completions = (modeStats.completions || 0) + (outcome === "complete" || outcome === "ended" ? 1 : 0);
  modeStats.bestScore = Math.max(modeStats.bestScore || 0, snap.score || 0);
  modeStats.mostLines = Math.max(modeStats.mostLines || 0, snap.lines || 0);
  if (mode === "sprint" && outcome === "complete") {
    modeStats.bestTimeMs = modeStats.bestTimeMs
      ? Math.min(modeStats.bestTimeMs, snap.elapsedMs)
      : snap.elapsedMs;
  }
  stats.modes[mode] = modeStats;
  storageWrite(STATS_KEY, stats);
}

let started = false;
let screen = "loading";
let currentMode = settings.lastMode;
let lastSnap = null;
let previousSnap = null;
let finalSnap = null;
let finalOutcome = "gameover";
let runRecorded = false;
let menuIndex = 0;
let modeIndex = Math.max(0, Object.keys(MODES).indexOf(currentMode));
let settingsIndex = 0;
let lastTime = 0;
let animationFrame = 0;
let dasDir = null;
let dasTimer = 0;
let dasCharged = false;
let gesture = null;
let recentTap = null;
let calloutTimer = 0;
let previewCache = "";
let audioContext = null;
let restoreFocus = null;
let accumulatedRunMs = 0;
let activeRunStartedAt = 0;

async function loadWasm() {
  try {
    const go = new Go();
    const response = await fetch("main.wasm");
    if (!response.ok) throw new Error(`WASM request failed with ${response.status}`);
    let result;
    try {
      result = await WebAssembly.instantiateStreaming(response.clone(), go.importObject);
    } catch {
      result = await WebAssembly.instantiate(await response.arrayBuffer(), go.importObject);
    }
    go.run(result.instance).catch((error) => console.error("WASM engine stopped", error));
    const deadline = performance.now() + 5000;
    while (typeof window.Tetris === "undefined" && performance.now() < deadline) {
      await new Promise((resolve) => setTimeout(resolve, 10));
    }
    if (typeof window.Tetris === "undefined") throw new Error("Tetris engine did not initialize");
    init();
  } catch (error) {
    console.error(error);
    showLoadError(error);
  }
}

function init() {
  applyPreferences();
  resizeCanvases();
  renderBlank();
  bindEvents();
  showMenu(false);
}

function bindEvents() {
  window.addEventListener("keydown", onKeyDown);
  window.addEventListener("keyup", onKeyUp);
  window.addEventListener("blur", () => {
    releaseInputs();
    if (screen === "playing") pauseGame("Game paused while the window was inactive.");
  });
  document.addEventListener("visibilitychange", () => {
    releaseInputs();
    if (document.hidden && screen === "playing") pauseGame("Game paused while the app was hidden.");
  });
  field.addEventListener("pointerdown", onGestureStart);
  field.addEventListener("pointermove", onGestureMove);
  field.addEventListener("pointerup", onGestureEnd);
  field.addEventListener("pointercancel", cancelGesture);
  pauseControl.addEventListener("click", () => pauseGame());
  const observer = new ResizeObserver(() => {
    if (!resizeCanvases()) return;
    if (lastSnap) render(lastSnap, true);
    else renderBlank();
  });
  observer.observe(field);
  observer.observe(holdCanvas);
  observer.observe(nextCanvas);
}

function startLoop() {
  stopLoop();
  lastTime = performance.now();
  animationFrame = requestAnimationFrame(frame);
}

function stopLoop() {
  if (animationFrame) cancelAnimationFrame(animationFrame);
  animationFrame = 0;
}

function currentRunTime(now = performance.now()) {
  const activeMs = activeRunStartedAt ? now - activeRunStartedAt : 0;
  return Math.max(0, Math.round(accumulatedRunMs + activeMs));
}

function snapshotForRun() {
  if (typeof window.Tetris === "undefined") return null;
  const snap = Tetris.snapshot();
  if (snap) snap.elapsedMs = currentRunTime();
  return snap;
}

function stopRunTimer() {
  if (!activeRunStartedAt) return;
  accumulatedRunMs = currentRunTime();
  activeRunStartedAt = 0;
}

function frame(now) {
  if (!started || screen !== "playing") {
    animationFrame = 0;
    return;
  }
  const elapsed = now - lastTime;
  if (elapsed < 14) {
    animationFrame = requestAnimationFrame(frame);
    return;
  }
  const dt = Math.min(elapsed, 100);
  lastTime = now;
  Tetris.tick(dt);
  updateKeyboardInput(dt);
  const snap = Tetris.snapshot();
  if (snap) {
    snap.elapsedMs = currentRunTime(now);
    processSnapshot(snap);
  }
  if (started && screen === "playing") animationFrame = requestAnimationFrame(frame);
  else animationFrame = 0;
}

function processSnapshot(snap) {
  lastSnap = snap;
  render(snap);
  detectGameEvents(previousSnap, snap);
  previousSnap = snap;
  updateModeLabel(snap);

  if (currentMode === "sprint" && snap.lines >= SPRINT_LINES && !snap.clearing) {
    finishRun(snap, "complete");
    return;
  }
  if (currentMode === "ultra" && snap.elapsedMs >= ULTRA_DURATION_MS && !snap.clearing) {
    snap.elapsedMs = ULTRA_DURATION_MS;
    finishRun(snap, "complete");
    return;
  }
  if (snap.state === "gameover") {
    finishRun(snap, "gameover");
  }
}

function syncCanvas(canvas) {
  const dpr = Math.min(window.devicePixelRatio || 1, 2);
  const width = Math.max(1, Math.round(canvas.clientWidth * dpr));
  const height = Math.max(1, Math.round(canvas.clientHeight * dpr));
  if (canvas.width === width && canvas.height === height) return false;
  canvas.width = width;
  canvas.height = height;
  return true;
}

function resizeCanvases() {
  const changed = syncCanvas(field) | syncCanvas(holdCanvas) | syncCanvas(nextCanvas);
  if (changed) previewCache = "";
  return Boolean(changed);
}

function drawCell(ctx, x, y, size, color, pieceIndex) {
  const pad = Math.max(1, Math.round(size * 0.045));
  ctx.fillStyle = color;
  ctx.fillRect(x + pad, y + pad, size - pad * 2, size - pad * 2);
  if (!settings.reducedMotion) {
    ctx.fillStyle = "rgba(255,255,255,0.2)";
    ctx.fillRect(x + pad, y + pad, size - pad * 2, Math.max(2, size * 0.15));
  }
  if (settings.highContrast) {
    ctx.strokeStyle = pieceIndex % 2 === 0 ? "#000" : "#fff";
    ctx.lineWidth = Math.max(1, size * 0.045);
    ctx.strokeRect(x + pad * 1.5, y + pad * 1.5, size - pad * 3, size - pad * 3);
  }
}

function drawGrid(ctx, width, height) {
  const cell = width / FIELD_W;
  ctx.strokeStyle = "rgba(255,255,255,0.045)";
  ctx.lineWidth = Math.max(1, width / 600);
  for (let column = 1; column < FIELD_W; column++) {
    ctx.beginPath();
    ctx.moveTo(column * cell, 0);
    ctx.lineTo(column * cell, height);
    ctx.stroke();
  }
  for (let row = 1; row < FIELD_H; row++) {
    ctx.beginPath();
    ctx.moveTo(0, row * cell);
    ctx.lineTo(width, row * cell);
    ctx.stroke();
  }
}

function render(snap, forcePreviews = false) {
  const theme = THEMES[settings.theme];
  const cell = field.width / FIELD_W;
  fctx.fillStyle = theme.bg;
  fctx.fillRect(0, 0, field.width, field.height);
  drawGrid(fctx, field.width, field.height);

  for (let row = 0; row < FIELD_H; row++) {
    for (let column = 0; column < FIELD_W; column++) {
      const value = snap.grid[row * FIELD_W + column];
      if (value > 0) drawCell(fctx, column * cell, row * cell, cell, theme.pieces[value], value);
    }
  }

  if (snap.clearing) {
    fctx.fillStyle = theme.clearing;
    for (const row of snap.clearRows) fctx.fillRect(0, row * cell, field.width, cell);
  }

  if (snap.ghost) {
    fctx.globalAlpha = 0.62;
    fctx.fillStyle = theme.ghost;
    for (const [row, column] of snap.ghost.cells) {
      if (row < 0 || row >= FIELD_H || column < 0 || column >= FIELD_W) continue;
      const inset = Math.max(2, cell * 0.09);
      fctx.fillRect(column * cell + inset, row * cell + inset, cell - inset * 2, cell - inset * 2);
    }
    fctx.globalAlpha = 1;
  }

  if (snap.active) {
    for (const [row, column] of snap.active.cells) {
      if (row < 0 || row >= FIELD_H || column < 0 || column >= FIELD_W) continue;
      drawCell(fctx, column * cell, row * cell, cell, theme.pieces[snap.active.color], snap.active.color);
    }
  }

  updateStats(snap);
  const holdKey = snap.hold ? `${snap.hold.type}:${snap.hold.color}` : "none";
  const nextKey = (snap.next || []).map((piece) => `${piece.type}:${piece.color}`).join(",");
  const key = `${settings.theme}:${settings.highContrast}:${holdKey}:${nextKey}:${holdCanvas.width}:${nextCanvas.width}:${nextCanvas.height}`;
  if (forcePreviews || key !== previewCache) {
    drawHold(snap.hold, theme);
    drawNext(snap.next || [], theme);
    previewCache = key;
  }
}

function renderBlank() {
  const theme = THEMES[settings.theme];
  fctx.fillStyle = theme.bg;
  fctx.fillRect(0, 0, field.width, field.height);
  drawGrid(fctx, field.width, field.height);
  drawHold(null, theme);
  drawNext([], theme);
  previewCache = "";
  setText("score", "0");
  setText("level", String(settings.startingLevel));
  setText("lines", "0");
  setText("combo", "-");
  setText("b2b", "-");
  setText("time", "00:00");
}

function pieceBounds(piece) {
  let minRow = Infinity;
  let maxRow = -Infinity;
  let minColumn = Infinity;
  let maxColumn = -Infinity;
  for (const [row, column] of piece.cells) {
    minRow = Math.min(minRow, row);
    maxRow = Math.max(maxRow, row);
    minColumn = Math.min(minColumn, column);
    maxColumn = Math.max(maxColumn, column);
  }
  return { minRow, maxRow, minColumn, maxColumn };
}

function drawPieceInSlot(ctx, pieceInfo, theme, x, y, width, height) {
  const piece = PIECES[pieceInfo.type];
  if (!piece) return;
  const bounds = pieceBounds(piece);
  const columns = bounds.maxColumn - bounds.minColumn + 1;
  const rows = bounds.maxRow - bounds.minRow + 1;
  const cell = Math.min(width / (columns + 0.8), height / (rows + 0.8));
  const pieceWidth = columns * cell;
  const pieceHeight = rows * cell;
  const originX = x + (width - pieceWidth) / 2 - bounds.minColumn * cell;
  const originY = y + (height - pieceHeight) / 2 - bounds.minRow * cell;
  for (const [row, column] of piece.cells) {
    drawCell(ctx, originX + column * cell, originY + row * cell, cell, theme.pieces[pieceInfo.color], pieceInfo.color);
  }
}

function drawHold(piece, theme) {
  hctx.fillStyle = theme.bg;
  hctx.fillRect(0, 0, holdCanvas.width, holdCanvas.height);
  if (piece) {
    drawPieceInSlot(hctx, piece, theme, 0, 0, holdCanvas.width, holdCanvas.height);
    holdCanvas.setAttribute("aria-label", `Held ${PIECES[piece.type].name} piece`);
  } else {
    holdCanvas.setAttribute("aria-label", "No held piece");
  }
}

function drawNext(queue, theme) {
  nctx.fillStyle = theme.bg;
  nctx.fillRect(0, 0, nextCanvas.width, nextCanvas.height);
  const horizontal = nextCanvas.width > nextCanvas.height * 1.35;
  const count = Math.min(queue.length, horizontal ? 3 : 5);
  for (let index = 0; index < count; index++) {
    const x = horizontal ? index * nextCanvas.width / count : 0;
    const y = horizontal ? 0 : index * nextCanvas.height / count;
    const width = horizontal ? nextCanvas.width / count : nextCanvas.width;
    const height = horizontal ? nextCanvas.height : nextCanvas.height / count;
    drawPieceInSlot(nctx, queue[index], theme, x, y, width, height);
  }
  const names = queue.slice(0, count).map((piece) => PIECES[piece.type].name).join(", ");
  nextCanvas.setAttribute("aria-label", names ? `Next pieces: ${names}` : "No upcoming pieces");
}

function updateStats(snap) {
  setText("score", Number(snap.score).toLocaleString());
  setText("level", snap.level);
  setText("lines", snap.lines);
  setText("combo", snap.combo > 0 ? snap.combo : "-");
  setText("b2b", snap.b2b ? "YES" : "-");
  const shownTime = currentMode === "ultra"
    ? Math.max(0, ULTRA_DURATION_MS - snap.elapsedMs)
    : snap.elapsedMs;
  setText("time", currentMode === "ultra" ? formatCountdown(shownTime) : formatTime(shownTime));
  field.setAttribute(
    "aria-label",
    `Tetris playfield. ${snap.lines} lines, score ${snap.score}, level ${snap.level}. Use touch gestures or keyboard controls.`,
  );
}

function detectGameEvents(before, snap) {
  if (!before) return;
  const oldStats = before.stats || {};
  const stats = snap.stats || {};
  const lineDelta = snap.lines - before.lines;
  const fullSpin = (stats.tSpins || 0) > (oldStats.tSpins || 0);
  const miniSpin = (stats.miniTSpins || 0) > (oldStats.miniTSpins || 0);
  const perfectClear = (stats.perfectClears || 0) > (oldStats.perfectClears || 0);
  const zenReset = (stats.zenResets || 0) > (oldStats.zenResets || 0);

  if (perfectClear) {
    showCallout("PERFECT CLEAR", "The board is clean");
    playCue("perfect");
    vibrate([18, 35, 28]);
    return;
  }
  if (zenReset) {
    showCallout("BREATHE", "Fresh board");
    playCue("zen");
    vibrate(20);
    return;
  }
  if (lineDelta > 0 || fullSpin || miniSpin) {
    let title = `${lineDelta} LINE${lineDelta === 1 ? "" : "S"}`;
    if (lineDelta === 4) title = "TETRIS";
    if (fullSpin) title = `T-SPIN${lineDelta ? ` ${lineName(lineDelta)}` : ""}`;
    if (miniSpin) title = `MINI T-SPIN${lineDelta ? ` ${lineName(lineDelta)}` : ""}`;
    const details = [];
    if (before.b2b && (lineDelta === 4 || fullSpin || miniSpin)) details.push("Back-to-back");
    if (snap.combo > 0) details.push(`${snap.combo} combo`);
    if (snap.level > before.level) details.push(`Level ${snap.level}`);
    showCallout(title, details.join(" / "));
    playCue(fullSpin || miniSpin ? "tspin" : lineDelta === 4 ? "tetris" : "clear");
    vibrate(lineDelta === 4 || fullSpin ? [18, 28, 28] : 16);
  } else if (snap.level > before.level) {
    showCallout(`LEVEL ${snap.level}`, "Gravity increased");
    playCue("level");
  } else if ((stats.piecesPlaced || 0) > (oldStats.piecesPlaced || 0)) {
    playCue("lock");
  }
}

function lineName(lines) {
  return ["", "SINGLE", "DOUBLE", "TRIPLE"][lines] || "";
}

function showCallout(title, detail = "") {
  calloutTitle.textContent = title;
  calloutDetail.textContent = detail;
  callout.classList.remove("show");
  void callout.offsetWidth;
  callout.classList.add("show");
  callout.setAttribute("aria-hidden", "false");
  clearTimeout(calloutTimer);
  calloutTimer = setTimeout(() => callout.setAttribute("aria-hidden", "true"), 1000);
  announce(detail ? `${title}. ${detail}.` : title);
}

function updateModeLabel(snap) {
  let detail = "";
  if (snap) {
    if (currentMode === "marathon") detail = `L${snap.level}`;
    else if (currentMode === "sprint") detail = `${Math.min(snap.lines, SPRINT_LINES)}/${SPRINT_LINES}`;
    else if (currentMode === "ultra") detail = formatCountdown(Math.max(0, ULTRA_DURATION_MS - snap.elapsedMs));
    else if (currentMode === "zen") detail = `${(snap.stats && snap.stats.zenResets) || 0} RESETS`;
  }
  modeLabel.textContent = detail ? `${MODES[currentMode].label} / ${detail}` : MODES[currentMode].label;
}

function selectMode(mode) {
  if (!MODES[mode]) return;
  currentMode = mode;
  settings.lastMode = mode;
  saveSettings();
  if (hasCoarsePointer() && !settings.gestureTutorialSeen) showGestureTutorial(mode);
  else beginGame(mode);
}

function beginGame(mode = currentMode) {
  if (started && !runRecorded) abandonRun();
  currentMode = mode;
  settings.lastMode = mode;
  saveSettings();
  Tetris.new({
    startingLevel: settings.startingLevel,
    ghost: settings.ghostEnabled,
    hold: settings.holdEnabled,
    rotate180: settings.rotate180Enabled,
    zen: mode === "zen",
  });
  started = true;
  runRecorded = false;
  screen = "playing";
  finalSnap = null;
  releaseInputs();
  accumulatedRunMs = 0;
  activeRunStartedAt = performance.now();
  previousSnap = snapshotForRun();
  lastSnap = previousSnap;
  document.body.classList.add("game-active", "is-playing");
  hideOverlay();
  resizeCanvases();
  render(lastSnap, true);
  updateModeLabel(lastSnap);
  announce(`${MODES[mode].label} started.`);
  field.focus({ preventScroll: true });
  startLoop();
}

function abandonRun() {
  if (!started || runRecorded) return;
  const snap = snapshotForRun() || lastSnap;
  stopRunTimer();
  if (snap && ((snap.stats && snap.stats.piecesPlaced > 0) || snap.elapsedMs > 1000)) {
    recordStats(snap, currentMode, "abandoned");
  }
  runRecorded = true;
  started = false;
  stopLoop();
  releaseInputs();
  document.body.classList.remove("game-active", "is-playing");
}

function finishRun(snap, outcome) {
  if (runRecorded) return;
  runRecorded = true;
  stopRunTimer();
  started = false;
  stopLoop();
  releaseInputs();
  document.body.classList.remove("game-active", "is-playing");
  finalSnap = snap;
  finalOutcome = outcome;
  const personalBest = runIsPersonalBest(currentMode, snap, outcome);
  recordStats(snap, currentMode, outcome);
  screen = "results";
  playCue(outcome === "complete" ? "complete" : "gameover");
  vibrate(outcome === "complete" ? [24, 35, 24, 35, 36] : [50, 45, 70]);
  if (runQualifies(currentMode, snap, outcome)) showNamePrompt(snap, outcome, personalBest);
  else showResults(snap, outcome, personalBest);
}

function pauseGame(message = "") {
  if (screen !== "playing") return;
  Tetris.act("pause");
  stopRunTimer();
  screen = "paused";
  stopLoop();
  releaseInputs();
  document.body.classList.remove("is-playing");
  showPause(message);
  announce("Game paused.");
}

function resumeGame() {
  if (screen !== "paused") return;
  Tetris.act("pause");
  screen = "playing";
  activeRunStartedAt = performance.now();
  document.body.classList.add("is-playing");
  hideOverlay();
  field.focus({ preventScroll: true });
  announce("Game resumed.");
  startLoop();
}

function restartGame() {
  abandonRun();
  beginGame(currentMode);
}

function endZenRun() {
  if (currentMode !== "zen") return;
  const snap = snapshotForRun() || lastSnap;
  if (snap) finishRun(snap, "ended");
}

function showOverlay(html, focusSelector = "button, input") {
  if (!overlay.classList.contains("show")) restoreFocus = document.activeElement;
  overlayCard.innerHTML = html;
  overlay.classList.add("show");
  requestAnimationFrame(() => {
    const focusTarget = focusSelector ? overlayCard.querySelector(focusSelector) : null;
    if (focusTarget) focusTarget.focus({ preventScroll: true });
  });
}

function hideOverlay() {
  overlay.classList.remove("show");
  if (restoreFocus && typeof restoreFocus.focus === "function" && restoreFocus.isConnected) {
    restoreFocus.focus({ preventScroll: true });
  }
  restoreFocus = null;
}

function showMenu(abandon = true) {
  if (abandon) abandonRun();
  screen = "menu";
  modeLabel.textContent = "CHOOSE YOUR GAME";
  const items = MENU_ITEMS.map((item, index) => (
    `<button class="menu-button${index === menuIndex ? " sel" : ""}" type="button" data-id="${item.id}" data-index="${index}">${item.label}</button>`
  )).join("");
  showOverlay(
    `<h2 id="overlay-title">TETRIS</h2><p>Precision play, anywhere.</p><div class="menu">${items}</div>` +
      '<p class="muted">Use arrow keys and Enter, or tap a choice.</p>',
    `[data-index="${menuIndex}"]`,
  );
  overlayCard.querySelectorAll(".menu-button").forEach((button) => {
    button.addEventListener("click", () => selectMenu(button.dataset.id));
  });
}

function selectMenu(id) {
  if (id === "play") showModeSelect();
  else if (id === "leaderboards") showLeaderboards(settings.lastMode);
  else if (id === "statistics") showStatistics();
  else if (id === "howto") showHowToPlay();
  else if (id === "settings") showSettings(true);
  else if (id === "credits") showCredits();
}

function showModeSelect() {
  screen = "modes";
  const modes = Object.entries(MODES).map(([id, mode], index) => (
    `<button class="mode-card" type="button" data-mode="${id}" data-index="${index}">` +
      `<strong>${mode.label}</strong><span>${mode.description}</span></button>`
  )).join("");
  showOverlay(
    `<h2 id="overlay-title">SELECT MODE</h2><div class="mode-grid">${modes}</div>` +
      '<button class="secondary" id="back-button" type="button">Back</button>',
    `[data-index="${modeIndex}"]`,
  );
  overlayCard.querySelectorAll(".mode-card").forEach((button) => {
    button.addEventListener("click", () => selectMode(button.dataset.mode));
  });
  document.getElementById("back-button").addEventListener("click", () => showMenu(false));
}

function showGestureTutorial(mode) {
  screen = "tutorial";
  showOverlay(
    '<h2 id="overlay-title">YOUR BOARD IS THE CONTROLLER</h2>' +
      '<p>No gameplay buttons. Every move happens directly on the playfield.</p>' +
      '<div class="gesture-guide">' +
        '<div><strong>&harr;</strong><b>Drag sideways</b><span>Move piece</span></div>' +
        '<div><strong>&#8634;</strong><b>Tap either half</b><span>Left CCW / right CW</span></div>' +
        '<div><strong>&uarr;</strong><b>Swipe up</b><span>Hold piece</span></div>' +
        '<div><strong>&darr;</strong><b>Drag / flick down</b><span>Soft / hard drop</span></div>' +
      '</div>' +
      '<p class="muted">Double-tap either half for a fast 180 rotation. The pause control stays in the corner.</p>' +
      '<div class="button-row"><button class="primary" id="tutorial-start" type="button">Start Playing</button>' +
      '<button class="secondary" id="tutorial-back" type="button">Back</button></div>',
    "#tutorial-start",
  );
  document.getElementById("tutorial-start").addEventListener("click", () => {
    settings.gestureTutorialSeen = true;
    saveSettings();
    beginGame(mode);
  });
  document.getElementById("tutorial-back").addEventListener("click", showModeSelect);
}

function showHowToPlay() {
  screen = "howto";
  showOverlay(
    '<h2 id="overlay-title">HOW TO PLAY</h2>' +
      '<h3>TOUCH</h3><div class="gesture-guide">' +
        '<div><strong>&harr;</strong><b>Drag sideways</b><span>Move piece</span></div>' +
        '<div><strong>&#8634;</strong><b>Tap halves</b><span>Rotate</span></div>' +
        '<div><strong>&uarr;</strong><b>Swipe up</b><span>Hold</span></div>' +
        '<div><strong>&darr;</strong><b>Flick down</b><span>Hard drop</span></div>' +
      '</div>' +
      '<h3>KEYBOARD</h3><p>Arrows move and rotate. Z rotates counter-clockwise, A rotates 180, Space hard drops, C holds, and P pauses.</p>' +
      '<p class="muted">Clear lines, build combos, and keep difficult clears going for a back-to-back bonus.</p>' +
      '<button class="primary" id="back-button" type="button">Back</button>',
  );
  document.getElementById("back-button").addEventListener("click", () => showMenu(false));
}

function showPause(message = "") {
  const endButton = currentMode === "zen"
    ? '<button class="secondary" id="end-button" type="button">End Run</button>'
    : "";
  showOverlay(
    '<h2 id="overlay-title">PAUSED</h2>' +
      (message ? `<p>${escapeHtml(message)}</p>` : `<p>${MODES[currentMode].label} is waiting for you.</p>`) +
      '<div class="button-row"><button class="primary" id="resume-button" type="button">Resume</button>' +
      '<button class="secondary" id="restart-button" type="button">Restart</button>' + endButton +
      '<button class="secondary danger" id="menu-button" type="button">Main Menu</button></div>',
    "#resume-button",
  );
  document.getElementById("resume-button").addEventListener("click", resumeGame);
  document.getElementById("restart-button").addEventListener("click", restartGame);
  document.getElementById("menu-button").addEventListener("click", () => showMenu(true));
  const end = document.getElementById("end-button");
  if (end) end.addEventListener("click", endZenRun);
}

function settingValue(row) {
  const value = settings[row.key];
  if (row.type === "boolean") return value ? "ON" : "OFF";
  if (row.names) return row.names[row.choices.indexOf(value)] || String(value);
  if (row.key === "theme") return String(value).toUpperCase();
  if (row.suffix) return `${value}${row.suffix}%`;
  return String(value);
}

function showSettings(initialFocus = false) {
  screen = "settings";
  const rows = SETTING_ROWS.map((row, index) => {
    let controls;
    if (row.type === "boolean") {
      controls = `<span class="controls"><b>${settingValue(row)}</b><button type="button" data-action="toggle" data-index="${index}" aria-label="Toggle ${row.label}">&#8635;</button></span>`;
    } else {
      controls = '<span class="controls">' +
        `<button type="button" data-action="decrement" data-index="${index}" aria-label="Decrease ${row.label}">-</button>` +
        `<b>${settingValue(row)}</b>` +
        `<button type="button" data-action="increment" data-index="${index}" aria-label="Increase ${row.label}">+</button></span>`;
    }
    return `<div class="setting-row${index === settingsIndex ? " sel" : ""}" data-row="${index}"><span>${row.label}</span>${controls}</div>`;
  }).join("");
  showOverlay(
    `<h2 id="overlay-title">SETTINGS</h2><div class="setting-list">${rows}</div>` +
      '<div class="button-row"><button class="secondary" id="replay-guide" type="button">Gesture Guide</button>' +
      '<button class="primary" id="back-button" type="button">Back</button></div>',
    initialFocus ? `[data-index="${settingsIndex}"]` : null,
  );
  overlayCard.querySelectorAll("[data-action]").forEach((button) => {
    button.addEventListener("click", (event) => {
      event.stopPropagation();
      settingsIndex = Number(button.dataset.index);
      const delta = button.dataset.action === "decrement" ? -1 : button.dataset.action === "increment" ? 1 : 0;
      adjustSetting(SETTING_ROWS[settingsIndex].key, delta);
    });
  });
  overlayCard.querySelectorAll("[data-row]").forEach((row) => {
    row.addEventListener("click", () => {
      settingsIndex = Number(row.dataset.row);
      overlayCard.querySelectorAll(".setting-row").forEach((item) => item.classList.remove("sel"));
      row.classList.add("sel");
    });
  });
  document.getElementById("replay-guide").addEventListener("click", () => {
    settings.gestureTutorialSeen = false;
    saveSettings();
    showHowToPlay();
  });
  document.getElementById("back-button").addEventListener("click", () => showMenu(false));
}

function adjustSetting(key, delta) {
  const row = SETTING_ROWS.find((item) => item.key === key);
  if (!row) return;
  if (row.type === "boolean") {
    settings[key] = !settings[key];
  } else if (row.type === "number") {
    settings[key] = clamp(Number(settings[key]) + delta, row.min, row.max);
  } else if (row.type === "choice") {
    const current = Math.max(0, row.choices.indexOf(settings[key]));
    const direction = delta === 0 ? 1 : delta;
    settings[key] = row.choices[(current + direction + row.choices.length) % row.choices.length];
  }
  saveSettings();
  applyPreferences();
  if (lastSnap) render(lastSnap, true);
  else renderBlank();
  showSettings(false);
  requestAnimationFrame(() => {
    const target = overlayCard.querySelector(`[data-index="${settingsIndex}"]`);
    if (target) target.focus({ preventScroll: true });
  });
}

function applyPreferences() {
  document.body.dataset.theme = settings.theme;
  document.body.dataset.highContrast = String(settings.highContrast);
  document.body.dataset.reducedMotion = String(settings.reducedMotion);
  const themeMeta = document.querySelector('meta[name="theme-color"]');
  if (themeMeta) themeMeta.content = THEMES[settings.theme].uiBg;
  previewCache = "";
}

function showLeaderboards(mode) {
  if (!MODES[mode]) mode = "marathon";
  screen = "leaderboards";
  const records = loadRecords();
  const tabs = Object.entries(MODES).map(([id, item]) => (
    `<button class="tab-button${id === mode ? " active" : ""}" type="button" data-mode="${id}">${item.label}</button>`
  )).join("");
  const entries = sortRecords(mode, records[mode].slice());
  const rows = entries.length
    ? entries.map((entry, index) => (
      `<li><span class="rank">${index + 1}</span><span>${escapeHtml(entry.name)}</span><b>${formatRecordMetric(mode, entry)}</b></li>`
    )).join("")
    : '<li><span></span><span class="muted">No runs recorded yet</span><span></span></li>';
  showOverlay(
    `<h2 id="overlay-title">LEADERBOARDS</h2><div class="tabs">${tabs}</div>` +
      `<ol class="score-list">${rows}</ol><button class="primary" id="back-button" type="button">Back</button>`,
    `[data-mode="${mode}"]`,
  );
  overlayCard.querySelectorAll(".tab-button").forEach((button) => {
    button.addEventListener("click", () => showLeaderboards(button.dataset.mode));
  });
  document.getElementById("back-button").addEventListener("click", () => showMenu(false));
}

function showStatistics() {
  screen = "statistics";
  const stats = loadStats();
  const averageScore = stats.gamesPlayed ? Math.round((stats.totalScore || 0) / stats.gamesPlayed) : 0;
  const rows = [
    ["Games", stats.gamesPlayed || 0],
    ["High Score", Number(stats.highScore || 0).toLocaleString()],
    ["Average Score", Number(averageScore).toLocaleString()],
    ["Total Lines", Number(stats.totalLines || 0).toLocaleString()],
    ["Play Time", formatLongTime(stats.totalPlayTimeMs || 0)],
    ["Longest Game", formatTime(stats.longestGameMs || 0)],
    ["Pieces", Number(stats.totalPieces || 0).toLocaleString()],
    ["Longest Combo", stats.longestCombo || 0],
    ["Tetrises", stats.totalTetrises || 0],
    ["T-Spins", stats.totalTSpins || 0],
    ["Mini T-Spins", stats.totalMiniTSpins || 0],
    ["Perfect Clears", stats.totalPerfectClears || 0],
  ];
  const modeRows = Object.entries(MODES).map(([id, mode]) => {
    const item = (stats.modes && stats.modes[id]) || {};
    const best = id === "sprint" && item.bestTimeMs ? formatPreciseTime(item.bestTimeMs) : Number(item.bestScore || 0).toLocaleString();
    return `<div class="stat2"><span>${mode.label} / ${item.plays || 0} RUNS</span><b>${best}</b></div>`;
  }).join("");
  showOverlay(
    '<h2 id="overlay-title">STATISTICS</h2>' +
      `<div class="stat-grid">${rows.map(([label, value]) => `<div class="stat2"><span>${label}</span><b>${value}</b></div>`).join("")}</div>` +
      `<h3>BEST BY MODE</h3><div class="stat-grid">${modeRows}</div>` +
      '<button class="primary" id="back-button" type="button">Back</button>',
  );
  document.getElementById("back-button").addEventListener("click", () => showMenu(false));
}

function showCredits() {
  screen = "credits";
  showOverlay(
    '<h2 id="overlay-title">CREDITS</h2>' +
      '<p>Tetris browser edition</p>' +
      '<p class="muted">Go/WASM game engine, Canvas rendering, and a gesture-first mobile interface.</p>' +
      '<p class="muted">Rules include a 7-bag randomizer, SRS kicks, T-spins, combos, back-to-back clears, hold, and perfect clears.</p>' +
      '<p class="muted">Tetris is a registered trademark of Tetris Holding, LLC. This fan-made clone is not affiliated with The Tetris Company.</p>' +
      '<button class="primary" id="back-button" type="button">Back</button>',
  );
  document.getElementById("back-button").addEventListener("click", () => showMenu(false));
}

function resultTitle(mode, outcome) {
  if (outcome === "complete" && mode === "sprint") return "SPRINT COMPLETE";
  if (outcome === "complete" && mode === "ultra") return "TIME";
  if (outcome === "ended" && mode === "zen") return "ZEN RUN";
  return "GAME OVER";
}

function resultMetric(mode, snap) {
  return mode === "sprint" ? formatPreciseTime(snap.elapsedMs) : Number(snap.score).toLocaleString();
}

function resultDetails(snap) {
  const stats = snap.stats || {};
  const rows = [
    ["Score", Number(snap.score).toLocaleString()],
    ["Time", formatPreciseTime(snap.elapsedMs)],
    ["Lines", snap.lines],
    ["Level", snap.level],
    ["Tetrises", stats.tetrises || 0],
    ["T-Spins", (stats.tSpins || 0) + (stats.miniTSpins || 0)],
  ];
  return `<div class="result-grid">${rows.map(([label, value]) => `<div class="result-stat"><span>${label}</span><b>${value}</b></div>`).join("")}</div>`;
}

function showNamePrompt(snap, outcome, personalBest) {
  screen = "results";
  showOverlay(
    `<h2 id="overlay-title">${resultTitle(currentMode, outcome)}</h2>` +
      `<div class="big">${resultMetric(currentMode, snap)}</div>` +
      (personalBest ? '<div class="record-banner">NEW PERSONAL BEST</div>' : "") +
      resultDetails(snap) +
      '<p>Enter a name for the leaderboard.</p>' +
      `<input id="record-name" maxlength="16" autocomplete="nickname" value="${escapeHtml(settings.playerName)}" aria-label="Leaderboard name" />` +
      '<div class="button-row"><button class="primary" id="save-button" type="button">Save Run</button>' +
      '<button class="secondary" id="skip-button" type="button">Skip</button></div>',
    "#record-name",
  );
  const input = document.getElementById("record-name");
  const save = () => {
    const name = (input.value.trim() || "Player").slice(0, 16);
    settings.playerName = name;
    saveSettings();
    saveRecord(name, currentMode, snap);
    showResults(snap, outcome, personalBest);
  };
  input.select();
  input.addEventListener("keydown", (event) => {
    if (event.key === "Enter") save();
  });
  document.getElementById("save-button").addEventListener("click", save);
  document.getElementById("skip-button").addEventListener("click", () => showResults(snap, outcome, personalBest));
}

function showResults(snap, outcome, personalBest) {
  screen = "results";
  const entries = sortRecords(currentMode, loadRecords()[currentMode].slice()).slice(0, 5);
  const rows = entries.length
    ? entries.map((entry, index) => `<li><span class="rank">${index + 1}</span><span>${escapeHtml(entry.name)}</span><b>${formatRecordMetric(currentMode, entry)}</b></li>`).join("")
    : '<li><span></span><span class="muted">No saved runs yet</span><span></span></li>';
  showOverlay(
    `<h2 id="overlay-title">${resultTitle(currentMode, outcome)}</h2>` +
      `<div class="big">${resultMetric(currentMode, snap)}</div>` +
      (personalBest ? '<div class="record-banner">NEW PERSONAL BEST</div>' : "") +
      resultDetails(snap) +
      `<h3>${MODES[currentMode].label} LEADERS</h3><ol class="score-list">${rows}</ol>` +
      '<div class="button-row"><button class="primary" id="again-button" type="button">Play Again</button>' +
      '<button class="secondary" id="modes-button" type="button">Change Mode</button>' +
      '<button class="secondary" id="menu-button" type="button">Menu</button></div>',
    "#again-button",
  );
  document.getElementById("again-button").addEventListener("click", () => beginGame(currentMode));
  document.getElementById("modes-button").addEventListener("click", showModeSelect);
  document.getElementById("menu-button").addEventListener("click", () => showMenu(false));
}

function showLoadError(error) {
  screen = "error";
  showOverlay(
    '<h2 id="overlay-title">FAILED TO LOAD</h2>' +
      `<p>${escapeHtml(String(error))}</p>` +
      '<p class="muted">Serve the web folder over HTTP and make sure main.wasm is present.</p>',
    null,
  );
}

const keyMap = {
  ArrowLeft: "moveLeft",
  ArrowRight: "moveRight",
  ArrowDown: "softDrop",
  ArrowUp: "rotateCW",
  " ": "hardDrop",
  z: "rotateCCW",
  Z: "rotateCCW",
  a: "rotate180",
  A: "rotate180",
  c: "hold",
  C: "hold",
  p: "pause",
  P: "pause",
  r: "restart",
  R: "restart",
};

function onKeyDown(event) {
  const key = event.key;
  if (event.target && event.target.matches("input, select, textarea")) return;

  if (screen === "menu") {
    if (key === "ArrowUp" || key === "k") menuIndex = (menuIndex - 1 + MENU_ITEMS.length) % MENU_ITEMS.length;
    else if (key === "ArrowDown" || key === "j") menuIndex = (menuIndex + 1) % MENU_ITEMS.length;
    else if (key === "Enter") selectMenu(MENU_ITEMS[menuIndex].id);
    else return;
    event.preventDefault();
    if (key !== "Enter") showMenu(false);
    return;
  }

  if (screen === "modes") {
    const modeIds = Object.keys(MODES);
    if (key === "ArrowLeft" || key === "ArrowUp") modeIndex = (modeIndex - 1 + modeIds.length) % modeIds.length;
    else if (key === "ArrowRight" || key === "ArrowDown") modeIndex = (modeIndex + 1) % modeIds.length;
    else if (key === "Enter") selectMode(modeIds[modeIndex]);
    else if (key === "Escape") showMenu(false);
    else return;
    event.preventDefault();
    if (key.startsWith("Arrow")) showModeSelect();
    return;
  }

  if (screen === "settings") {
    if (key === "ArrowUp" || key === "k") settingsIndex = Math.max(0, settingsIndex - 1);
    else if (key === "ArrowDown" || key === "j") settingsIndex = Math.min(SETTING_ROWS.length - 1, settingsIndex + 1);
    else if (key === "ArrowLeft") adjustSetting(SETTING_ROWS[settingsIndex].key, -1);
    else if (key === "ArrowRight" || key === "Enter") adjustSetting(SETTING_ROWS[settingsIndex].key, 1);
    else if (key === "Escape") showMenu(false);
    else return;
    event.preventDefault();
    if (key === "ArrowUp" || key === "ArrowDown" || key === "k" || key === "j") showSettings(false);
    return;
  }

  if (key === "Escape") {
    if (screen === "paused") showMenu(true);
    else if (screen !== "playing" && screen !== "results" && screen !== "loading") showMenu(false);
    return;
  }

  if (screen === "paused") {
    if ((key === "p" || key === "P") && !event.repeat) resumeGame();
    else if ((key === "r" || key === "R") && !event.repeat) restartGame();
    return;
  }

  if (event.target && event.target.matches("button")) return;

  if (screen !== "playing") return;
  const action = keyMap[key];
  if (!action) return;
  event.preventDefault();
  ensureAudio();
  if (action === "softDrop") {
    if (!event.repeat) Tetris.setSoftDrop(true);
    return;
  }
  if (action === "moveLeft" || action === "moveRight") {
    if (!event.repeat) {
      Tetris.act(action);
      dasDir = action;
      dasTimer = 0;
      dasCharged = false;
    }
    return;
  }
  if (action === "pause" && !event.repeat) pauseGame();
  else if (action === "restart" && !event.repeat) restartGame();
  else if (!event.repeat) performAction(action);
}

function onKeyUp(event) {
  const action = keyMap[event.key];
  if (action === "softDrop" && typeof window.Tetris !== "undefined") Tetris.setSoftDrop(false);
  if ((action === "moveLeft" || action === "moveRight") && dasDir === action) dasDir = null;
}

function updateKeyboardInput(dt) {
  if (!dasDir || screen !== "playing") return;
  dasTimer += dt;
  if (!dasCharged) {
    if (dasTimer >= DAS) {
      dasCharged = true;
      dasTimer = 0;
      Tetris.act(dasDir);
    }
  } else {
    while (dasTimer >= ARR) {
      Tetris.act(dasDir);
      dasTimer -= ARR;
    }
  }
}

function releaseInputs() {
  dasDir = null;
  dasTimer = 0;
  dasCharged = false;
  gesture = null;
  recentTap = null;
  if (typeof window.Tetris !== "undefined") Tetris.setSoftDrop(false);
}

function performAction(action) {
  if (screen !== "playing") return;
  Tetris.act(action);
  if (action === "rotateCW" || action === "rotateCCW" || action === "rotate180") {
    playCue("rotate");
    vibrate(8);
  } else if (action === "hardDrop") {
    playCue("drop");
    vibrate(16);
  } else if (action === "hold") {
    playCue("hold");
    vibrate(10);
  }
}

function onGestureStart(event) {
  if (screen !== "playing" || event.pointerType === "mouse") return;
  event.preventDefault();
  ensureAudio();
  field.setPointerCapture(event.pointerId);
  const rect = field.getBoundingClientRect();
  gesture = {
    id: event.pointerId,
    startX: event.clientX,
    startY: event.clientY,
    lastX: event.clientX,
    lastY: event.clientY,
    startedAt: performance.now(),
    rect,
    axis: null,
    horizontalSteps: 0,
    softSteps: 0,
  };
}

function onGestureMove(event) {
  if (!gesture || gesture.id !== event.pointerId || screen !== "playing") return;
  event.preventDefault();
  const dx = event.clientX - gesture.startX;
  const dy = event.clientY - gesture.startY;
  const absX = Math.abs(dx);
  const absY = Math.abs(dy);
  const deadZone = 7;
  if (!gesture.axis && Math.max(absX, absY) > deadZone) {
    gesture.axis = absX > absY * 1.1 ? "horizontal" : "vertical";
  }

  const sensitivityFactor = [1.25, 0.95, 0.72][settings.gestureSensitivity - 1];
  if (gesture.axis === "horizontal") {
    const stepSize = Math.max(12, gesture.rect.width / FIELD_W * sensitivityFactor);
    const steps = Math.trunc(dx / stepSize);
    while (gesture.horizontalSteps < steps) {
      Tetris.act("moveRight");
      gesture.horizontalSteps++;
    }
    while (gesture.horizontalSteps > steps) {
      Tetris.act("moveLeft");
      gesture.horizontalSteps--;
    }
  } else if (gesture.axis === "vertical" && dy > 0) {
    const stepSize = Math.max(10, gesture.rect.height / FIELD_H * sensitivityFactor);
    const steps = Math.floor(dy / stepSize);
    while (gesture.softSteps < steps) {
      Tetris.act("softDrop");
      gesture.softSteps++;
    }
  }
  gesture.lastX = event.clientX;
  gesture.lastY = event.clientY;
}

function onGestureEnd(event) {
  if (!gesture || gesture.id !== event.pointerId) return;
  event.preventDefault();
  const activeGesture = gesture;
  gesture = null;
  const dx = event.clientX - activeGesture.startX;
  const dy = event.clientY - activeGesture.startY;
  const duration = Math.max(1, performance.now() - activeGesture.startedAt);
  const cellHeight = activeGesture.rect.height / FIELD_H;

  if (!activeGesture.axis || Math.max(Math.abs(dx), Math.abs(dy)) <= 7) {
    const localX = event.clientX - activeGesture.rect.left;
    handleRotationTap(localX < activeGesture.rect.width / 2 ? "left" : "right");
    return;
  }
  if (activeGesture.axis !== "vertical") return;

  if (dy < -Math.max(38, cellHeight * 2.1)) {
    performAction("hold");
    return;
  }
  const velocity = dy / duration;
  const velocityThreshold = [1.05, 0.82, 0.62][settings.gestureSensitivity - 1];
  if (dy > Math.max(42, cellHeight * 2.8) && velocity >= velocityThreshold) {
    performAction("hardDrop");
  }
}

function handleRotationTap(side) {
  const now = performance.now();
  const isDoubleTap = settings.rotate180Enabled && recentTap && recentTap.side === side && now - recentTap.time <= 300;
  recentTap = isDoubleTap ? null : { side, time: now };
  const action = isDoubleTap
    ? side === "left" ? "gesture180CCW" : "gesture180CW"
    : side === "left" ? "gestureRotateCCW" : "gestureRotateCW";
  Tetris.act(action);
  playCue("rotate");
  vibrate(isDoubleTap ? 12 : 8);
}

function cancelGesture(event) {
  if (gesture && gesture.id === event.pointerId) gesture = null;
}

function ensureAudio() {
  if (!settings.soundEnabled) return null;
  if (navigator.userActivation && !navigator.userActivation.hasBeenActive) return null;
  const AudioContextClass = window.AudioContext || window.webkitAudioContext;
  if (!AudioContextClass) return null;
  if (!audioContext) audioContext = new AudioContextClass();
  if (audioContext.state === "suspended") audioContext.resume().catch(() => {});
  return audioContext;
}

function playCue(name) {
  if (!settings.soundEnabled || settings.volume <= 0) return;
  const context = ensureAudio();
  if (!context) return;
  const cues = {
    rotate: [[250, 0, 0.045]],
    hold: [[330, 0, 0.065]],
    drop: [[105, 0, 0.08]],
    lock: [[145, 0, 0.045]],
    clear: [[420, 0, 0.07], [570, 0.06, 0.08]],
    tetris: [[330, 0, 0.06], [495, 0.055, 0.07], [660, 0.115, 0.11]],
    tspin: [[520, 0, 0.06], [780, 0.065, 0.11]],
    perfect: [[440, 0, 0.07], [660, 0.07, 0.07], [880, 0.14, 0.16]],
    level: [[390, 0, 0.06], [520, 0.06, 0.09]],
    complete: [[440, 0, 0.08], [554, 0.08, 0.08], [659, 0.16, 0.18]],
    gameover: [[240, 0, 0.12], [180, 0.11, 0.2]],
    zen: [[330, 0, 0.13], [440, 0.12, 0.16]],
  };
  for (const [frequency, delay, duration] of cues[name] || []) {
    const oscillator = context.createOscillator();
    const gain = context.createGain();
    const start = context.currentTime + delay;
    oscillator.type = name === "drop" || name === "gameover" ? "triangle" : "sine";
    oscillator.frequency.setValueAtTime(frequency, start);
    gain.gain.setValueAtTime(0.0001, start);
    gain.gain.exponentialRampToValueAtTime(Math.max(0.001, settings.volume / 80), start + 0.008);
    gain.gain.exponentialRampToValueAtTime(0.0001, start + duration);
    oscillator.connect(gain).connect(context.destination);
    oscillator.start(start);
    oscillator.stop(start + duration + 0.02);
  }
}

function vibrate(pattern) {
  if (!settings.hapticsEnabled || !navigator.vibrate) return;
  navigator.vibrate(pattern);
}

function formatRecordMetric(mode, entry) {
  return mode === "sprint" ? formatPreciseTime(entry.elapsedMs) : Number(entry.score || 0).toLocaleString();
}

function formatTime(ms) {
  const totalSeconds = Math.max(0, Math.floor(Number(ms || 0) / 1000));
  const minutes = Math.floor(totalSeconds / 60);
  const seconds = totalSeconds % 60;
  return `${String(minutes).padStart(2, "0")}:${String(seconds).padStart(2, "0")}`;
}

function formatCountdown(ms) {
  const totalSeconds = Math.max(0, Math.ceil(Number(ms || 0) / 1000));
  const minutes = Math.floor(totalSeconds / 60);
  const seconds = totalSeconds % 60;
  return `${String(minutes).padStart(2, "0")}:${String(seconds).padStart(2, "0")}`;
}

function formatPreciseTime(ms) {
  const centiseconds = Math.max(0, Math.floor(Number(ms || 0) / 10));
  const minutes = Math.floor(centiseconds / 6000);
  const seconds = Math.floor(centiseconds / 100) % 60;
  const fraction = centiseconds % 100;
  return `${String(minutes).padStart(2, "0")}:${String(seconds).padStart(2, "0")}.${String(fraction).padStart(2, "0")}`;
}

function formatLongTime(ms) {
  const totalMinutes = Math.floor(Number(ms || 0) / 60000);
  const hours = Math.floor(totalMinutes / 60);
  const minutes = totalMinutes % 60;
  return hours > 0 ? `${hours}h ${minutes}m` : `${minutes}m`;
}

function setText(id, value) {
  const element = document.getElementById(id);
  const text = String(value);
  if (element && element.textContent !== text) element.textContent = text;
}

function announce(message) {
  liveStatus.textContent = "";
  requestAnimationFrame(() => { liveStatus.textContent = message; });
}

function hasCoarsePointer() {
  return window.matchMedia("(any-pointer: coarse)").matches;
}

function clamp(value, minimum, maximum) {
  return Math.min(maximum, Math.max(minimum, value));
}

function escapeHtml(value) {
  return String(value).replace(/[&<>"']/g, (character) => ({
    "&": "&amp;",
    "<": "&lt;",
    ">": "&gt;",
    '"': "&quot;",
    "'": "&#39;",
  })[character]);
}

loadWasm();
