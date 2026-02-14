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
	world  *World
	state  state
	score  int
	lives  int
	level  int
	player Entity

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
	g.score = 0
	g.lives = 3
	g.level = 1
	g.player = SpawnPlayer(g.world, ScreenWidth/2, ScreenHeight/2)
	g.spawnWave()
}

func (g *Game) spawnWave() {
	count := 3 + g.level
	playerPos := g.world.positions[g.player]

	for i := 0; i < count; i++ {
		var x, y float64
		for {
			x = rand.Float64() * ScreenWidth
			y = rand.Float64() * ScreenHeight
			if playerPos != nil {
				dx := x - playerPos.X
				dy := y - playerPos.Y
				if math.Sqrt(dx*dx+dy*dy) > 150 {
					break
				}
			} else {
				break
			}
		}
		SpawnAsteroid(g.world, x, y, SizeLarge)
	}
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

	// Run systems
	InputSystem(w)
	PhysicsSystem(w)
	WrapSystem(w)
	InvulnerabilitySystem(w)
	LifetimeSystem(w)

	// Handle player shooting
	if pc, ok := w.players[g.player]; ok && pc.ShootPressed {
		SpawnBullet(w, g.player)
	}

	// Collision
	events := CollisionSystem(w)

	// Process bullet hits
	destroyed := make(map[Entity]bool)
	for _, hit := range events.BulletHits {
		if destroyed[hit.Asteroid] {
			continue
		}
		destroyed[hit.Asteroid] = true

		ast := w.asteroids[hit.Asteroid]
		apos := w.positions[hit.Asteroid]
		if ast == nil || apos == nil {
			continue
		}

		// Score
		switch ast.Size {
		case SizeLarge:
			g.score += 20
		case SizeMedium:
			g.score += 50
		case SizeSmall:
			g.score += 100
		}

		// Particles
		for i := 0; i < 8; i++ {
			SpawnParticle(w, apos.X, apos.Y)
		}

		// Split
		if ast.Size != SizeSmall {
			nextSize := ast.Size + 1
			SpawnAsteroid(w, apos.X, apos.Y, nextSize)
			SpawnAsteroid(w, apos.X, apos.Y, nextSize)
		}

		w.Destroy(hit.Bullet)
		w.Destroy(hit.Asteroid)
	}

	// Process player hit
	if events.PlayerHit {
		g.lives--
		ppos := w.positions[g.player]
		if ppos != nil {
			for i := 0; i < 15; i++ {
				SpawnParticle(w, ppos.X, ppos.Y)
			}
		}
		if g.lives <= 0 {
			g.state = stateGameOver
			w.Destroy(g.player)
		} else {
			// Respawn
			ppos.X = ScreenWidth / 2
			ppos.Y = ScreenHeight / 2
			vel := w.velocities[g.player]
			if vel != nil {
				vel.X = 0
				vel.Y = 0
			}
			rot := w.rotations[g.player]
			if rot != nil {
				rot.Angle = -math.Pi / 2
			}
			pc := w.players[g.player]
			if pc != nil {
				pc.Invulnerable = true
				pc.InvulnerableTimer = 120
				pc.BlinkTimer = 0
			}
		}
	}

	// Check if wave cleared
	if len(w.asteroids) == 0 {
		g.level++
		g.spawnWave()
	}
}

func (g *Game) drawHUD(screen *ebiten.Image) {
	hudScale := 2.0
	hudColor := color.RGBA{255, 255, 255, 255}
	DrawText(screen, fmt.Sprintf("SCORE: %d", g.score), 10, 10, hudScale, hudColor)
	DrawText(screen, fmt.Sprintf("LIVES: %d", g.lives), 10, 28, hudScale, hudColor)
	DrawText(screen, fmt.Sprintf("LEVEL: %d", g.level), 10, 46, hudScale, hudColor)
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
		g.drawHUD(screen)
	case statePaused:
		g.drawPaused(screen)
	case stateGameOver:
		RenderSystem(g.world, screen)
		DrawThrust(g.world, screen)
		g.drawHUD(screen)

		titleScale := 5.0
		titleText := "GAME OVER"
		titleW := TextWidth(titleText, titleScale)
		titleX := (ScreenWidth - titleW) / 2
		DrawText(screen, titleText, titleX, float64(ScreenHeight)/2-60, titleScale, color.RGBA{255, 0, 0, 255})

		scoreScale := 2.5
		scoreText := fmt.Sprintf("FINAL SCORE: %d", g.score)
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
