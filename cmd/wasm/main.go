//go:build js && wasm

// Command wasm is the browser entry point for the Tetris engine. It compiles the
// I/O-free domain package to WebAssembly and bridges it to JavaScript: the page
// drives the simulation by calling tick(dt) every animation frame and reads a
// render snapshot via snapshot(). Input, rendering, and high-score persistence
// live in the browser (web/main.js).
package main

import (
	"math/rand"
	"syscall/js"
	"time"

	"github.com/mahir/tetris/internal/domain"
)

var (
	game    *domain.Game
	rng     *rand.Rand
	lastCfg domain.GameSettings
)

func stateString(s domain.GameState) string {
	switch s {
	case domain.StateReady:
		return "ready"
	case domain.StatePlaying:
		return "playing"
	case domain.StatePaused:
		return "paused"
	case domain.StateGameOver:
		return "gameover"
	}
	return "ready"
}

func pieceInfo(t domain.PieceType) map[string]interface{} {
	tm := domain.GetTetromino(t)
	return map[string]interface{}{"type": int(t), "color": tm.Color}
}

// cellArray converts absolute board cells {row, col} to visible-field coordinates
// (row - HiddenRows). Cells above the visible area yield negative rows, which the
// renderer clips.
func cellArray(cells [][2]int) []interface{} {
	out := make([]interface{}, 0, len(cells))
	for _, c := range cells {
		out = append(out, []interface{}{c[0] - domain.HiddenRows, c[1]})
	}
	return out
}

func spawnGame() {
	if rng == nil {
		rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	}
	game = domain.NewGame(rng, lastCfg)
}

func newGame(this js.Value, args []js.Value) interface{} {
	cfg := domain.GameSettings{
		GhostEnabled:  true,
		HoldEnabled:   true,
		Rotate180:     true,
		StartingLevel: 1,
	}
	if len(args) > 0 && args[0].Type() == js.TypeObject {
		opts := args[0]
		if v := opts.Get("ghost"); v.Type() == js.TypeBoolean {
			cfg.GhostEnabled = v.Bool()
		}
		if v := opts.Get("hold"); v.Type() == js.TypeBoolean {
			cfg.HoldEnabled = v.Bool()
		}
		if v := opts.Get("rotate180"); v.Type() == js.TypeBoolean {
			cfg.Rotate180 = v.Bool()
		}
		if v := opts.Get("zen"); v.Type() == js.TypeBoolean {
			cfg.ZenMode = v.Bool()
		}
		if v := opts.Get("startingLevel"); v.Type() == js.TypeNumber {
			lvl := v.Int()
			if lvl < 1 {
				lvl = 1
			}
			cfg.StartingLevel = lvl
		}
	}
	lastCfg = cfg
	spawnGame()
	return nil
}

func tick(this js.Value, args []js.Value) interface{} {
	if game == nil {
		return nil
	}
	dt := int64(16)
	if len(args) > 0 && args[0].Type() == js.TypeNumber {
		dt = int64(args[0].Float())
		if dt < 0 {
			dt = 0
		}
		if dt > 100 {
			dt = 100 // clamp after stalls (mirrors the CLI engine)
		}
	}
	game.Tick(dt)
	return nil
}

func act(this js.Value, args []js.Value) interface{} {
	if game == nil {
		return nil
	}
	if len(args) == 0 {
		return nil
	}
	switch args[0].String() {
	case "moveLeft":
		game.MoveLeft()
	case "moveRight":
		game.MoveRight()
	case "softDrop":
		game.SoftDrop()
	case "rotateCW":
		game.RotateCW()
	case "rotateCCW":
		game.RotateCCW()
	case "rotate180":
		game.Rotate180()
	case "gestureRotateCW":
		game.RotateGesture(domain.RotateCW)
	case "gestureRotateCCW":
		game.RotateGesture(domain.RotateCCW)
	case "gesture180CW":
		game.CompleteGesture180(domain.RotateCW)
	case "gesture180CCW":
		game.CompleteGesture180(domain.RotateCCW)
	case "hardDrop":
		game.HardDrop()
	case "hold":
		game.Hold()
	case "pause":
		game.TogglePause()
	case "restart":
		spawnGame()
	}
	return nil
}

func setSoftDrop(this js.Value, args []js.Value) interface{} {
	if game == nil {
		return nil
	}
	on := false
	if len(args) > 0 && args[0].Type() == js.TypeBoolean {
		on = args[0].Bool()
	}
	game.SetSoftDrop(on)
	return nil
}

func snapshot(this js.Value, args []js.Value) interface{} {
	if game == nil {
		return nil
	}

	grid := game.Board().Grid()
	flat := make([]interface{}, 0, domain.VisibleHeight*domain.BoardWidth)
	for y := domain.HiddenRows; y < domain.BoardHeight; y++ {
		for x := 0; x < domain.BoardWidth; x++ {
			flat = append(flat, grid[y][x])
		}
	}

	snap := map[string]interface{}{
		"state":     stateString(game.State()),
		"score":     game.Score(),
		"level":     game.Level(),
		"lines":     game.Lines(),
		"combo":     game.Combo(),
		"b2b":       game.BackToBack(),
		"elapsedMs": game.ElapsedMs(),
		"grid":      flat,
	}

	if game.HasActive() {
		tm := game.Active().Tetromino()
		snap["active"] = map[string]interface{}{
			"cells": cellArray(game.Active().Cells()),
			"color": tm.Color,
		}
	} else {
		snap["active"] = nil
	}

	if gy := game.GhostY(); gy >= 0 && game.HasActive() {
		tm := game.Active().Tetromino()
		cells := tm.Cells(game.Active().State, game.Active().X, gy)
		snap["ghost"] = map[string]interface{}{
			"cells": cellArray(cells),
			"color": tm.Color,
		}
	} else {
		snap["ghost"] = nil
	}

	if pt, ok := game.HeldPiece(); ok {
		snap["hold"] = pieceInfo(pt)
	} else {
		snap["hold"] = nil
	}

	next := game.NextQueue(5)
	nextArr := make([]interface{}, 0, len(next))
	for _, t := range next {
		nextArr = append(nextArr, pieceInfo(t))
	}
	snap["next"] = nextArr

	snap["clearing"] = game.IsClearing()
	clearRows := make([]interface{}, 0, len(game.ClearRows()))
	for _, y := range game.ClearRows() {
		clearRows = append(clearRows, y-domain.HiddenRows)
	}
	snap["clearRows"] = clearRows

	st := game.Stats()
	snap["stats"] = map[string]interface{}{
		"piecesPlaced":  st.PiecesPlaced,
		"tetrises":      st.Tetrises,
		"tSpins":        st.TSpins,
		"miniTSpins":    st.MiniTSpins,
		"perfectClears": st.PerfectClears,
		"longestCombo":  st.LongestCombo,
		"zenResets":     st.ZenResets,
	}

	return snap
}

func main() {
	obj := js.Global().Get("Object").New()
	obj.Set("new", js.FuncOf(newGame))
	obj.Set("tick", js.FuncOf(tick))
	obj.Set("act", js.FuncOf(act))
	obj.Set("setSoftDrop", js.FuncOf(setSoftDrop))
	obj.Set("snapshot", js.FuncOf(snapshot))
	js.Global().Set("Tetris", obj)

	// Keep the Go instance alive so the registered callbacks remain valid.
	select {}
}
