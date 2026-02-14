package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/matheus3301/asteroids/internal/game"
)

func main() {
	ebiten.SetWindowSize(game.ScreenWidth, game.ScreenHeight)
	ebiten.SetWindowTitle("Asteroids")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	g := game.New()
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
