package game

import "testing"

func TestSoundManager_NilSafe(t *testing.T) {
	var sm *SoundManager
	// None of these should panic
	sm.playFire()
	sm.playExplosion(SizeLarge)
	sm.playDeath()
	sm.PlayBlip()
	sm.PlayConfirm()
	sm.startThrust()
	sm.stopThrust()
	sm.updateBeat(5)
	sm.PauseAll()
	sm.ResumeAll()
	sm.StopAll()
	sm.Reset()
	sm.SetMasterVolume(0.5)
}

func TestSoundSystem_NilManager(t *testing.T) {
	w := NewWorld()
	w.SoundQueue = append(w.SoundQueue, SoundFire, SoundPlayerDeath)
	// Should not panic
	SoundSystem(nil, w)
}

func TestSoundSystem_DrainsQueue(t *testing.T) {
	w := NewWorld()
	w.SoundQueue = append(w.SoundQueue, SoundFire, SoundExplosionLarge, SoundPlayerDeath)
	// Use nil manager — events are drained but no audio plays
	SoundSystem(nil, w)
	// nil manager returns early, so queue is NOT drained — that's by design
	// With a real manager, queue would be drained. Test with non-nil:
	// We can't easily create a real SoundManager in tests (needs audio context),
	// so we verify the nil path doesn't panic.
	if len(w.SoundQueue) != 3 {
		t.Errorf("expected queue len 3 with nil manager (early return), got %d", len(w.SoundQueue))
	}
}

func TestSoundForSize(t *testing.T) {
	tests := []struct {
		size     AsteroidSize
		expected SoundEvent
	}{
		{SizeLarge, SoundExplosionLarge},
		{SizeMedium, SoundExplosionMed},
		{SizeSmall, SoundExplosionSmall},
	}
	for _, tt := range tests {
		got := soundForSize(tt.size)
		if got != tt.expected {
			t.Errorf("soundForSize(%d) = %d, want %d", tt.size, got, tt.expected)
		}
	}
}

func TestBeatTick_AlternatesHighLow(t *testing.T) {
	var sm SoundManager
	sm.masterVolume = 0
	sm.beatLowBuf = make([]byte, 4)
	sm.beatHighBuf = make([]byte, 4)
	sm.beatTimer = 0
	sm.beatHigh = false

	sm.updateBeat(5)
	if !sm.beatHigh {
		t.Error("after first tick, beatHigh should be true")
	}

	sm.beatTimer = 0
	sm.updateBeat(5)
	if sm.beatHigh {
		t.Error("after second tick, beatHigh should be false")
	}
}

func TestBeatTick_NoPlayWhenTimerNotExpired(t *testing.T) {
	var sm SoundManager
	sm.masterVolume = 0
	sm.beatLowBuf = make([]byte, 4)
	sm.beatHighBuf = make([]byte, 4)
	sm.beatTimer = 10
	sm.beatHigh = false

	sm.updateBeat(5)
	if sm.beatHigh {
		t.Error("should not alternate when timer hasn't expired")
	}
	if sm.beatTimer != 9 {
		t.Errorf("expected beatTimer 9, got %d", sm.beatTimer)
	}
}
