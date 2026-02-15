package game

import (
	"bytes"

	"github.com/hajimehoshi/ebiten/v2/audio"
)

// SoundEvent represents a one-shot sound to be played.
type SoundEvent int

const (
	SoundFire SoundEvent = iota
	SoundExplosionSmall
	SoundExplosionMed
	SoundExplosionLarge
	SoundPlayerDeath
	SoundExtraLife
)

// soundForSize maps an AsteroidSize to the corresponding SoundEvent.
func soundForSize(size AsteroidSize) SoundEvent {
	switch size {
	case SizeLarge:
		return SoundExplosionLarge
	case SizeMedium:
		return SoundExplosionMed
	default:
		return SoundExplosionSmall
	}
}

// SoundManager handles all audio playback for the game.
type SoundManager struct {
	ctx               *audio.Context
	fireBuf           []byte
	explosionSmallBuf []byte
	explosionMedBuf   []byte
	explosionLargeBuf []byte
	deathBuf          []byte
	thrustPlayer      *audio.Player
	thrustPlaying     bool
	beatLowBuf        []byte
	beatHighBuf       []byte
	beatHigh          bool
	beatTimer         int
	beatInterval      int
	masterVolume      float64
	blipBuf           []byte
	confirmBuf        []byte
}

// NewSoundManager creates a SoundManager and pre-generates all audio buffers.
func NewSoundManager() *SoundManager {
	ctx := audio.CurrentContext()
	if ctx == nil {
		ctx = audio.NewContext(sampleRate)
	}
	sm := &SoundManager{
		ctx:               ctx,
		fireBuf:           generateFire(sampleRate),
		explosionSmallBuf: generateExplosion(sampleRate, SizeSmall),
		explosionMedBuf:   generateExplosion(sampleRate, SizeMedium),
		explosionLargeBuf: generateExplosion(sampleRate, SizeLarge),
		deathBuf:          generateDeath(sampleRate),
		beatLowBuf:        generateBeatTone(sampleRate, 55),
		beatHighBuf:       generateBeatTone(sampleRate, 70),
		masterVolume:      1.0,
		beatInterval:      60,
		blipBuf:           generateBlip(sampleRate),
		confirmBuf:        generateConfirm(sampleRate),
	}

	thrustBuf := generateThrustLoop(sampleRate)
	loop := audio.NewInfiniteLoop(bytes.NewReader(thrustBuf), int64(len(thrustBuf)))
	p, err := ctx.NewPlayer(loop)
	if err == nil {
		sm.thrustPlayer = p
		sm.thrustPlayer.SetVolume(sm.masterVolume)
	}

	return sm
}

func (sm *SoundManager) playOneShot(buf []byte) {
	if sm == nil || sm.ctx == nil {
		return
	}
	p, err := sm.ctx.NewPlayer(bytes.NewReader(buf))
	if err != nil {
		return
	}
	p.SetVolume(sm.masterVolume)
	p.Play()
}

func (sm *SoundManager) playFire() {
	if sm == nil {
		return
	}
	sm.playOneShot(sm.fireBuf)
}

func (sm *SoundManager) playExplosion(size AsteroidSize) {
	if sm == nil {
		return
	}
	switch size {
	case SizeLarge:
		sm.playOneShot(sm.explosionLargeBuf)
	case SizeMedium:
		sm.playOneShot(sm.explosionMedBuf)
	default:
		sm.playOneShot(sm.explosionSmallBuf)
	}
}

func (sm *SoundManager) playDeath() {
	if sm == nil {
		return
	}
	sm.playOneShot(sm.deathBuf)
}

// PlayBlip plays a short navigation blip for menu cursor movement.
func (sm *SoundManager) PlayBlip() {
	if sm == nil {
		return
	}
	sm.playOneShot(sm.blipBuf)
}

// PlayConfirm plays a confirmation tone for menu selection.
func (sm *SoundManager) PlayConfirm() {
	if sm == nil {
		return
	}
	sm.playOneShot(sm.confirmBuf)
}

func (sm *SoundManager) startThrust() {
	if sm == nil || sm.thrustPlayer == nil {
		return
	}
	if sm.thrustPlaying {
		return
	}
	sm.thrustPlayer.Play()
	sm.thrustPlaying = true
}

func (sm *SoundManager) stopThrust() {
	if sm == nil || sm.thrustPlayer == nil {
		return
	}
	if !sm.thrustPlaying {
		return
	}
	sm.thrustPlayer.Pause()
	_ = sm.thrustPlayer.Rewind()
	sm.thrustPlaying = false
}

func (sm *SoundManager) updateBeat(asteroidCount int) {
	if sm == nil {
		return
	}
	sm.beatInterval = beatIntervalFromAsteroidCount(asteroidCount)
	sm.beatTimer--
	if sm.beatTimer <= 0 {
		if sm.beatHigh {
			sm.playOneShot(sm.beatHighBuf)
		} else {
			sm.playOneShot(sm.beatLowBuf)
		}
		sm.beatHigh = !sm.beatHigh
		sm.beatTimer = sm.beatInterval
	}
}

// PauseAll pauses continuous sounds.
func (sm *SoundManager) PauseAll() {
	if sm == nil {
		return
	}
	if sm.thrustPlayer != nil && sm.thrustPlaying {
		sm.thrustPlayer.Pause()
	}
}

// ResumeAll resumes continuous sounds that were playing.
func (sm *SoundManager) ResumeAll() {
	if sm == nil {
		return
	}
	if sm.thrustPlayer != nil && sm.thrustPlaying {
		sm.thrustPlayer.Play()
	}
}

// StopAll stops all continuous sounds and resets state.
func (sm *SoundManager) StopAll() {
	if sm == nil {
		return
	}
	sm.stopThrust()
}

// Reset stops everything and resets beat timers.
func (sm *SoundManager) Reset() {
	if sm == nil {
		return
	}
	sm.StopAll()
	sm.beatTimer = 0
	sm.beatHigh = false
}

// SetMasterVolume sets the master volume (0.0 to 1.0) and updates active players.
func (sm *SoundManager) SetMasterVolume(v float64) {
	if sm == nil {
		return
	}
	sm.masterVolume = clampF(v, 0, 1)
	if sm.thrustPlayer != nil {
		sm.thrustPlayer.SetVolume(sm.masterVolume)
	}
}

// SoundSystem drains the sound event queue and manages continuous sounds.
func SoundSystem(sm *SoundManager, w *World) {
	if sm == nil {
		return
	}

	for _, event := range w.SoundQueue {
		switch event {
		case SoundFire:
			sm.playFire()
		case SoundExplosionSmall:
			sm.playExplosion(SizeSmall)
		case SoundExplosionMed:
			sm.playExplosion(SizeMedium)
		case SoundExplosionLarge:
			sm.playExplosion(SizeLarge)
		case SoundPlayerDeath:
			sm.playDeath()
			sm.stopThrust()
		}
	}
	w.SoundQueue = w.SoundQueue[:0]

	if pc, ok := w.players[w.Player]; ok {
		if pc.Thrusting {
			sm.startThrust()
		} else {
			sm.stopThrust()
		}
	} else {
		sm.stopThrust()
	}

	sm.updateBeat(len(w.asteroids))
}
