package game

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type menuItem struct {
	label string
}

var mainMenuItems = []menuItem{
	{label: "START GAME"},
	{label: "SETTINGS"},
	{label: "QUIT"},
}

var pauseMenuItems = []menuItem{
	{label: "RESUME"},
	{label: "QUIT TO MENU"},
}

var settingsLabels = []string{
	"RESOLUTION",
	"FULLSCREEN",
	"BACK",
}

// --- Main Menu ---

func (g *Game) updateMenu() {
	if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
		g.menuCursor--
		if g.menuCursor < 0 {
			g.menuCursor = len(mainMenuItems) - 1
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
		g.menuCursor++
		if g.menuCursor >= len(mainMenuItems) {
			g.menuCursor = 0
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		g.menuSelect()
	}
}

func (g *Game) menuSelect() {
	switch g.menuCursor {
	case 0: // Start Game
		g.reset()
	case 1: // Settings
		g.state = stateSettings
		g.settingsCursor = 0
	case 2: // Quit
		g.quit = true
	}
}

func (g *Game) drawMenu(screen *ebiten.Image) {
	screen.Fill(color.Black)

	// Title
	titleScale := 6.0
	titleText := "ASTEROIDS"
	titleW := TextWidth(titleText, titleScale)
	titleX := (ScreenWidth - titleW) / 2
	DrawText(screen, titleText, titleX, 120, titleScale, color.RGBA{255, 255, 255, 255})

	// Menu items
	itemScale := 3.0
	startY := 300.0
	spacing := 50.0

	for i, item := range mainMenuItems {
		clr := color.RGBA{255, 255, 255, 255}
		if i == g.menuCursor {
			clr = color.RGBA{0, 255, 0, 255}
		}
		w := TextWidth(item.label, itemScale)
		x := (ScreenWidth - w) / 2
		y := startY + float64(i)*spacing
		DrawText(screen, item.label, x, y, itemScale, clr)
	}
}

// --- Settings ---

func (g *Game) updateSettings() {
	if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
		g.settingsCursor--
		if g.settingsCursor < 0 {
			g.settingsCursor = len(settingsLabels) - 1
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
		g.settingsCursor++
		if g.settingsCursor >= len(settingsLabels) {
			g.settingsCursor = 0
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		g.settingsSelect()
		if g.settingsCursor < 2 {
			g.settings.apply()
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
		g.settingsLeft()
		g.settings.apply()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyRight) {
		g.settingsRight()
		g.settings.apply()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.state = stateMenu
	}
}

func (g *Game) settingsSelect() {
	switch g.settingsCursor {
	case 0: // Resolution — cycle forward
		g.settings.resolutionIndex = (g.settings.resolutionIndex + 1) % len(resolutions)
	case 1: // Fullscreen — toggle
		g.settings.fullscreen = !g.settings.fullscreen
	case 2: // Back
		g.state = stateMenu
	}
}

func (g *Game) settingsLeft() {
	switch g.settingsCursor {
	case 0:
		g.settings.resolutionIndex--
		if g.settings.resolutionIndex < 0 {
			g.settings.resolutionIndex = len(resolutions) - 1
		}
	case 1:
		g.settings.fullscreen = !g.settings.fullscreen
	}
}

func (g *Game) settingsRight() {
	switch g.settingsCursor {
	case 0:
		g.settings.resolutionIndex = (g.settings.resolutionIndex + 1) % len(resolutions)
	case 1:
		g.settings.fullscreen = !g.settings.fullscreen
	}
}

func (g *Game) drawSettings(screen *ebiten.Image) {
	screen.Fill(color.Black)

	// Title
	titleScale := 4.0
	titleText := "SETTINGS"
	titleW := TextWidth(titleText, titleScale)
	titleX := (ScreenWidth - titleW) / 2
	DrawText(screen, titleText, titleX, 100, titleScale, color.RGBA{255, 255, 255, 255})

	itemScale := 2.5
	startY := 230.0
	spacing := 60.0

	for i, label := range settingsLabels {
		clr := color.RGBA{255, 255, 255, 255}
		if i == g.settingsCursor {
			clr = color.RGBA{0, 255, 0, 255}
		}

		var text string
		switch i {
		case 0:
			res := resolutions[g.settings.resolutionIndex]
			text = fmt.Sprintf("%s: %s", label, res.Label)
		case 1:
			val := "OFF"
			if g.settings.fullscreen {
				val = "ON"
			}
			text = fmt.Sprintf("%s: %s", label, val)
		default:
			text = label
		}

		w := TextWidth(text, itemScale)
		x := (ScreenWidth - w) / 2
		y := startY + float64(i)*spacing
		DrawText(screen, text, x, y, itemScale, clr)
	}

	// Hint
	hintScale := 1.5
	hint := "LEFT-RIGHT TO CHANGE . ENTER TO TOGGLE . ESC TO GO BACK"
	hintW := TextWidth(hint, hintScale)
	hintX := (ScreenWidth - hintW) / 2
	DrawText(screen, hint, hintX, 500, hintScale, color.RGBA{100, 100, 100, 255})
}

// --- Pause ---

func (g *Game) updatePaused() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.state = statePlaying
		return
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
		g.pauseCursor--
		if g.pauseCursor < 0 {
			g.pauseCursor = len(pauseMenuItems) - 1
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
		g.pauseCursor++
		if g.pauseCursor >= len(pauseMenuItems) {
			g.pauseCursor = 0
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		g.pauseSelect()
	}
}

func (g *Game) pauseSelect() {
	switch g.pauseCursor {
	case 0: // Resume
		g.state = statePlaying
	case 1: // Quit to Menu
		g.state = stateMenu
	}
}

func (g *Game) drawPaused(screen *ebiten.Image) {
	screen.Fill(color.Black)

	// Draw the frozen game world
	RenderSystem(g.world, screen)
	DrawThrust(g.world, screen)
	g.drawHUD(screen)

	// Dark overlay
	vector.FillRect(screen, 0, 0, ScreenWidth, ScreenHeight, color.RGBA{0, 0, 0, 150}, false)

	// Title
	titleScale := 5.0
	titleText := "PAUSED"
	titleW := TextWidth(titleText, titleScale)
	titleX := (ScreenWidth - titleW) / 2
	DrawText(screen, titleText, titleX, 180, titleScale, color.RGBA{255, 255, 255, 255})

	// Pause menu items
	itemScale := 3.0
	startY := 300.0
	spacing := 50.0

	for i, item := range pauseMenuItems {
		clr := color.RGBA{255, 255, 255, 255}
		if i == g.pauseCursor {
			clr = color.RGBA{0, 255, 0, 255}
		}
		w := TextWidth(item.label, itemScale)
		x := (ScreenWidth - w) / 2
		y := startY + float64(i)*spacing
		DrawText(screen, item.label, x, y, itemScale, clr)
	}
}
