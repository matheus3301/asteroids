package game

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	ScreenWidth  = 800
	ScreenHeight = 600

	saucerInitialDelay = 600
	saucerRespawnDelay = 600

	hudIconScale = 0.52
)

var shipIconVerts = [][2]float64{
	{playerRadius * hudIconScale, 0},
	{-playerRadius * 0.8 * hudIconScale, -playerRadius * 0.6 * hudIconScale},
	{-playerRadius * 0.8 * hudIconScale, playerRadius * 0.6 * hudIconScale},
}

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
	sound *SoundManager

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
	g.settings.volume = 10
	return g
}

func (g *Game) ensureSound() {
	if g.sound == nil {
		g.sound = NewSoundManager()
		g.sound.SetMasterVolume(float64(g.settings.volume) / 10.0)
	}
}

func (g *Game) reset() {
	g.ensureSound()
	g.sound.Reset()
	g.sound.SetMasterVolume(float64(g.settings.volume) / 10.0)
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
			g.sound.PlayConfirm()
			g.state = stateMenu
		}
	}
	return nil
}

func (g *Game) updatePlaying() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.sound.PauseAll()
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

	SoundSystem(g.sound, w)

	if w.Lives <= 0 {
		g.sound.StopAll()
		g.state = stateGameOver
	}
}

func (g *Game) drawHUD(screen *ebiten.Image) {
	hudScale := 2.0
	hudColor := color.RGBA{255, 255, 255, 255}

	DrawText(screen, fmt.Sprintf("SCORE: %d", g.world.Score), 10, 10, hudScale, hudColor)

	// Lives as "LIVES:" label followed by ship icons, aligned with numbers
	livesLabel := "LIVES: "
	DrawText(screen, livesLabel, 10, 32, hudScale, hudColor)
	iconWing := playerRadius * 0.6 * hudIconScale
	iconStartX := 10.0 + TextWidth(livesLabel, hudScale) + iconWing
	iconY := 32.0 + 7.0*hudScale/2 // vertically center with text
	count := g.world.Lives - 1
	if count < 0 {
		count = 0
	}
	for i := 0; i < count; i++ {
		iconX := iconStartX + float64(i)*(iconWing*2+6)
		drawPolygon(screen, &Position{X: iconX, Y: iconY}, -math.Pi/2, shipIconVerts, hudColor)
	}

	DrawText(screen, fmt.Sprintf("LEVEL: %d", g.world.Level), 10, 54, hudScale, hudColor)
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
