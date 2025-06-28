package main

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/romshark/ebiten-experiment/game"
)

func main() {
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("A Gopher's Experiment")

	g := game.NewGame(32, 32, game.DefaultConfig())
	if err := ebiten.RunGame(g); err != nil {
		panic(err)
	}
}
