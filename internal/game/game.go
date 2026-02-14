package game

import (
	"fmt"
	"image/color"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	ScreenWidth  = 800
	ScreenHeight = 600

	saucerInitialDelay = 600
	saucerRespawnDelay = 600
)

type state int

const (
	stateMenu state = iota
	stateSettings
	statePlaying
	statePaused
	stateGameOver
)

// Game implements ebiten.Game and orchestrates the ECS world.
type Game struct {
	world *World
	state state

	menuCursor     int
	settingsCursor int
	pauseCursor    int
	settings       settings
	quit           bool
}

func New() *Game {
	g := &Game{
		state: stateMenu,
	}
	return g
}

func (g *Game) reset() {
	g.world = NewWorld()
	g.state = statePlaying
	w := g.world
	w.Score = 0
	w.Lives = 3
	w.NextExtraLifeAt = 10_000
	w.Level = 1
	w.SaucerSpawnTimer = saucerInitialDelay
	w.SaucerActive = 0
	w.Player = SpawnPlayer(w, ScreenWidth/2, ScreenHeight/2)
	spawnWave(w)
}

func (g *Game) Update() error {
	if g.quit {
		return ebiten.Termination
	}
	switch g.state {
	case stateMenu:
		g.updateMenu()
	case stateSettings:
		g.updateSettings()
	case statePlaying:
		g.updatePlaying()
	case statePaused:
		g.updatePaused()
	case stateGameOver:
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			g.state = stateMenu
		}
	}
	return nil
}

func (g *Game) updatePlaying() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.state = statePaused
		g.pauseCursor = 0
		return
	}
	w := g.world

	InputSystem(w)
	PhysicsSystem(w)
	WrapSystem(w)
	InvulnerabilitySystem(w)
	LifetimeSystem(w)
	SaucerSpawnSystem(w)
	SaucerAISystem(w)
	SaucerBulletLifetimeSystem(w)
	SaucerDespawnSystem(w)
	HyperspaceSystem(w, rand.Float64())
	ShootingSystem(w)
	events := CollisionSystem(w)
	CollisionResponseSystem(w, events)
	WaveClearSystem(w)

	if w.Lives <= 0 {
		g.state = stateGameOver
	}
}

func (g *Game) drawHUD(screen *ebiten.Image) {
	hudScale := 2.0
	hudColor := color.RGBA{255, 255, 255, 255}
	DrawText(screen, fmt.Sprintf("SCORE: %d", g.world.Score), 10, 10, hudScale, hudColor)
	DrawText(screen, fmt.Sprintf("LIVES: %d", g.world.Lives), 10, 28, hudScale, hudColor)
	DrawText(screen, fmt.Sprintf("LEVEL: %d", g.world.Level), 10, 46, hudScale, hudColor)
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.Black)

	switch g.state {
	case stateMenu:
		g.drawMenu(screen)
	case stateSettings:
		g.drawSettings(screen)
	case statePlaying:
		RenderSystem(g.world, screen)
		DrawThrust(g.world, screen)
		DrawSaucerDetail(g.world, screen)
		g.drawHUD(screen)
	case statePaused:
		g.drawPaused(screen)
	case stateGameOver:
		RenderSystem(g.world, screen)
		DrawThrust(g.world, screen)
		DrawSaucerDetail(g.world, screen)
		g.drawHUD(screen)

		titleScale := 5.0
		titleText := "GAME OVER"
		titleW := TextWidth(titleText, titleScale)
		titleX := (ScreenWidth - titleW) / 2
		DrawText(screen, titleText, titleX, float64(ScreenHeight)/2-60, titleScale, color.RGBA{255, 0, 0, 255})

		scoreScale := 2.5
		scoreText := fmt.Sprintf("FINAL SCORE: %d", g.world.Score)
		scoreW := TextWidth(scoreText, scoreScale)
		scoreX := (ScreenWidth - scoreW) / 2
		DrawText(screen, scoreText, scoreX, float64(ScreenHeight)/2+10, scoreScale, color.RGBA{255, 255, 255, 255})

		hintScale := 2.0
		hintText := "PRESS ENTER"
		hintW := TextWidth(hintText, hintScale)
		hintX := (ScreenWidth - hintW) / 2
		DrawText(screen, hintText, hintX, float64(ScreenHeight)/2+55, hintScale, color.RGBA{150, 150, 150, 255})
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}
