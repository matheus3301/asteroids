package main

import (
	"fmt"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/matheus3301/asteroids/internal/ai"
	"github.com/matheus3301/asteroids/internal/game"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <weights.gob>\n", os.Args[0])
		os.Exit(1)
	}

	weightsPath := os.Args[1]
	f, err := os.Open(weightsPath)
	if err != nil {
		log.Fatalf("failed to open weights file: %v", err)
	}
	defer f.Close()

	net, err := ai.Load(f)
	if err != nil {
		log.Fatalf("failed to load network: %v", err)
	}

	fmt.Printf("Loaded network: layers=%v, params=%d\n", net.Layers, net.WeightCount())

	agent := ai.NewNeuralAgent(net)

	ebiten.SetWindowSize(game.ScreenWidth, game.ScreenHeight)
	ebiten.SetWindowTitle("Asteroids - AI Watch")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	g := game.New()
	g.SetAIAgent(agent)

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
