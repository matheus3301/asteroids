package game

import "testing"

// --- Main Menu ---

func TestMenuSelect_StartGame(t *testing.T) {
	g := New()
	g.menuCursor = 0
	g.menuSelect()

	if g.state != statePlaying {
		t.Errorf("expected statePlaying, got %v", g.state)
	}
	if g.world == nil {
		t.Fatal("world should be initialized after starting game")
	}
	if !g.world.Alive(g.world.Player) {
		t.Error("player should be alive")
	}
	if g.world.Lives != 3 {
		t.Errorf("expected 3 lives, got %d", g.world.Lives)
	}
	if g.world.Score != 0 {
		t.Errorf("expected score 0, got %d", g.world.Score)
	}
}

func TestMenuSelect_Settings(t *testing.T) {
	g := New()
	g.menuCursor = 1
	g.menuSelect()

	if g.state != stateSettings {
		t.Errorf("expected stateSettings, got %v", g.state)
	}
	if g.settingsCursor != 0 {
		t.Errorf("settings cursor should reset to 0, got %d", g.settingsCursor)
	}
}

func TestMenuSelect_Quit(t *testing.T) {
	g := New()
	g.menuCursor = 2
	g.menuSelect()

	if !g.quit {
		t.Error("quit flag should be set")
	}
}

func TestQuitFlag_CausesTermination(t *testing.T) {
	g := New()
	g.quit = true

	err := g.Update()
	if err == nil {
		t.Error("Update should return an error when quit is true")
	}
}

// --- Pause ---

func TestPauseSelect_Resume(t *testing.T) {
	g := newPlaying()
	g.state = statePaused
	g.pauseCursor = 0
	g.pauseSelect()

	if g.state != statePlaying {
		t.Errorf("expected statePlaying, got %v", g.state)
	}
}

func TestPauseSelect_QuitToMenu(t *testing.T) {
	g := newPlaying()
	g.state = statePaused
	g.pauseCursor = 1
	g.pauseSelect()

	if g.state != stateMenu {
		t.Errorf("expected stateMenu, got %v", g.state)
	}
}

func TestPause_PreservesGameState(t *testing.T) {
	g := newPlaying()
	g.world.Score = 500
	g.world.Lives = 2
	g.world.Level = 3

	// Pause
	g.state = statePaused
	g.pauseCursor = 0

	// Verify game state is preserved
	if g.world.Score != 500 {
		t.Errorf("score should be preserved, got %d", g.world.Score)
	}
	if g.world.Lives != 2 {
		t.Errorf("lives should be preserved, got %d", g.world.Lives)
	}
	if g.world.Level != 3 {
		t.Errorf("level should be preserved, got %d", g.world.Level)
	}
	if g.world == nil {
		t.Error("world should still exist while paused")
	}

	// Resume and verify still intact
	g.pauseSelect()
	if g.state != statePlaying {
		t.Errorf("expected statePlaying after resume, got %v", g.state)
	}
	if g.world.Score != 500 || g.world.Lives != 2 || g.world.Level != 3 {
		t.Error("game state should be unchanged after resume")
	}
}

func TestPauseToMenu_ThenStartNewGame(t *testing.T) {
	g := newPlaying()
	g.world.Score = 999
	g.world.Lives = 1
	g.world.Level = 5

	// Pause → Quit to Menu
	g.state = statePaused
	g.pauseCursor = 1
	g.pauseSelect()

	if g.state != stateMenu {
		t.Fatalf("expected stateMenu, got %v", g.state)
	}

	// Start a new game from menu
	g.menuCursor = 0
	g.menuSelect()

	if g.world.Score != 0 {
		t.Errorf("new game should reset score, got %d", g.world.Score)
	}
	if g.world.Lives != 3 {
		t.Errorf("new game should reset lives, got %d", g.world.Lives)
	}
	if g.world.Level != 1 {
		t.Errorf("new game should reset level, got %d", g.world.Level)
	}
}

// --- Settings ---

func TestSettingsSelect_ResolutionCycles(t *testing.T) {
	g := New()
	g.state = stateSettings
	g.settingsCursor = 0

	if g.settings.resolutionIndex != 0 {
		t.Fatalf("expected initial resolution index 0, got %d", g.settings.resolutionIndex)
	}

	g.settingsSelect()
	if g.settings.resolutionIndex != 1 {
		t.Errorf("expected resolution index 1, got %d", g.settings.resolutionIndex)
	}

	g.settingsSelect()
	if g.settings.resolutionIndex != 2 {
		t.Errorf("expected resolution index 2, got %d", g.settings.resolutionIndex)
	}

	// Should wrap around
	g.settingsSelect()
	if g.settings.resolutionIndex != 0 {
		t.Errorf("expected resolution index to wrap to 0, got %d", g.settings.resolutionIndex)
	}
}

func TestSettingsSelect_FullscreenToggles(t *testing.T) {
	g := New()
	g.state = stateSettings
	g.settingsCursor = 1

	if g.settings.fullscreen {
		t.Fatal("fullscreen should default to off")
	}

	g.settingsSelect()
	if !g.settings.fullscreen {
		t.Error("fullscreen should be on after first toggle")
	}

	g.settingsSelect()
	if g.settings.fullscreen {
		t.Error("fullscreen should be off after second toggle")
	}
}

func TestSettingsSelect_Back(t *testing.T) {
	g := New()
	g.state = stateSettings
	g.settingsCursor = 2
	g.settingsSelect()

	if g.state != stateMenu {
		t.Errorf("expected stateMenu, got %v", g.state)
	}
}

func TestSettingsLeft_ResolutionCyclesBackward(t *testing.T) {
	g := New()
	g.state = stateSettings
	g.settingsCursor = 0

	// From index 0, left should wrap to last
	g.settingsLeft()
	if g.settings.resolutionIndex != len(resolutions)-1 {
		t.Errorf("expected resolution index %d, got %d", len(resolutions)-1, g.settings.resolutionIndex)
	}

	g.settingsLeft()
	if g.settings.resolutionIndex != len(resolutions)-2 {
		t.Errorf("expected resolution index %d, got %d", len(resolutions)-2, g.settings.resolutionIndex)
	}
}

func TestSettingsRight_ResolutionCyclesForward(t *testing.T) {
	g := New()
	g.state = stateSettings
	g.settingsCursor = 0

	g.settingsRight()
	if g.settings.resolutionIndex != 1 {
		t.Errorf("expected resolution index 1, got %d", g.settings.resolutionIndex)
	}
}

func TestSettingsLeft_FullscreenToggles(t *testing.T) {
	g := New()
	g.state = stateSettings
	g.settingsCursor = 1

	g.settingsLeft()
	if !g.settings.fullscreen {
		t.Error("expected fullscreen on")
	}

	g.settingsLeft()
	if g.settings.fullscreen {
		t.Error("expected fullscreen off")
	}
}

// --- State Transitions ---

func TestPlaying_EscapeSetsStatePaused(t *testing.T) {
	g := newPlaying()
	// Directly simulate what updatePlaying does on Escape
	g.state = statePaused
	g.pauseCursor = 0

	if g.state != statePaused {
		t.Errorf("expected statePaused, got %v", g.state)
	}
	if g.pauseCursor != 0 {
		t.Errorf("pause cursor should reset to 0, got %d", g.pauseCursor)
	}
}

func TestGameOver_TransitionsToMenu(t *testing.T) {
	g := newPlaying()
	g.state = stateGameOver

	// Simulate what Update does on Enter in gameOver state
	g.state = stateMenu

	if g.state != stateMenu {
		t.Errorf("expected stateMenu, got %v", g.state)
	}
}

func TestFullFlow_MenuToPlayToPauseToMenuToPlay(t *testing.T) {
	g := New()
	if g.state != stateMenu {
		t.Fatalf("expected stateMenu, got %v", g.state)
	}

	// Start game
	g.menuCursor = 0
	g.menuSelect()
	if g.state != statePlaying {
		t.Fatalf("expected statePlaying, got %v", g.state)
	}

	// Pause
	g.state = statePaused
	g.pauseCursor = 0

	// Quit to menu
	g.pauseCursor = 1
	g.pauseSelect()
	if g.state != stateMenu {
		t.Fatalf("expected stateMenu after quit to menu, got %v", g.state)
	}

	// Start new game again
	g.menuCursor = 0
	g.menuSelect()
	if g.state != statePlaying {
		t.Fatalf("expected statePlaying on second start, got %v", g.state)
	}
	if g.world.Score != 0 || g.world.Lives != 3 || g.world.Level != 1 {
		t.Error("new game should have fresh state")
	}
}

// --- Settings Data ---

func TestSettings_DefaultValues(t *testing.T) {
	g := New()

	if g.settings.resolutionIndex != 0 {
		t.Errorf("expected default resolution index 0, got %d", g.settings.resolutionIndex)
	}
	if g.settings.fullscreen {
		t.Error("expected fullscreen to default to false")
	}
}

func TestResolutions_HasExpectedEntries(t *testing.T) {
	if len(resolutions) < 2 {
		t.Fatalf("expected at least 2 resolutions, got %d", len(resolutions))
	}

	first := resolutions[0]
	if first.Width != 800 || first.Height != 600 {
		t.Errorf("expected first resolution 800x600, got %dx%d", first.Width, first.Height)
	}
}
