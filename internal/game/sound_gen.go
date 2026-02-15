package game

import (
	"encoding/binary"
	"math"
	"math/rand"
)

const sampleRate = 44100

// clampF clamps v between min and max.
func clampF(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

// writeStereoSample writes a mono float64 sample as stereo 16-bit LE PCM to buf at offset.
func writeStereoSample(buf []byte, offset int, sample float64) {
	s := int16(clampF(sample, -1, 1) * 32767)
	binary.LittleEndian.PutUint16(buf[offset:], uint16(s))
	binary.LittleEndian.PutUint16(buf[offset+2:], uint16(s))
}

// generateFire returns an 80ms sine sweep from 1000→200 Hz with linear fade.
func generateFire(sr int) []byte {
	dur := 0.08
	frames := int(float64(sr) * dur)
	buf := make([]byte, frames*4)
	phase := 0.0
	for i := 0; i < frames; i++ {
		t := float64(i) / float64(frames)
		freq := 1000.0 - 800.0*t
		phase += 2 * math.Pi * freq / float64(sr)
		envelope := (1.0 - t) * 0.4
		sample := math.Sin(phase) * envelope
		writeStereoSample(buf, i*4, sample)
	}
	return buf
}

// generateExplosion returns a noise burst + low sine, sized by asteroid size.
func generateExplosion(sr int, size AsteroidSize) []byte {
	var dur, lowMix float64
	switch size {
	case SizeLarge:
		dur, lowMix = 0.4, 0.6
	case SizeMedium:
		dur, lowMix = 0.25, 0.3
	default:
		dur, lowMix = 0.15, 0.1
	}
	frames := int(float64(sr) * dur)
	buf := make([]byte, frames*4)
	for i := 0; i < frames; i++ {
		t := float64(i) / float64(frames)
		envelope := math.Exp(-t * 6)
		noise := (rand.Float64()*2 - 1) * envelope * 0.5
		low := math.Sin(2*math.Pi*60*float64(i)/float64(sr)) * envelope * lowMix
		sample := clampF(noise+low, -1, 1)
		writeStereoSample(buf, i*4, sample)
	}
	return buf
}

// generateDeath returns an 800ms noise + 40 Hz rumble.
func generateDeath(sr int) []byte {
	dur := 0.8
	frames := int(float64(sr) * dur)
	buf := make([]byte, frames*4)
	for i := 0; i < frames; i++ {
		t := float64(i) / float64(frames)
		envelope := math.Exp(-t * 3)
		noise := (rand.Float64()*2 - 1) * envelope * 0.5
		rumble := math.Sin(2*math.Pi*40*float64(i)/float64(sr)) * envelope * 0.5
		sample := clampF(noise+rumble, -1, 1)
		writeStereoSample(buf, i*4, sample)
	}
	return buf
}

// generateThrustLoop returns 200ms of low-pass filtered noise for thrust.
func generateThrustLoop(sr int) []byte {
	dur := 0.2
	frames := int(float64(sr) * dur)
	buf := make([]byte, frames*4)
	alpha := 0.15
	prev := 0.0
	for i := 0; i < frames; i++ {
		noise := rand.Float64()*2 - 1
		prev = prev + alpha*(noise-prev)
		sample := prev * 0.3
		writeStereoSample(buf, i*4, sample)
	}
	return buf
}

// generateBeatTone returns a 60ms sine burst at the given frequency.
func generateBeatTone(sr int, freq float64) []byte {
	dur := 0.06
	frames := int(float64(sr) * dur)
	buf := make([]byte, frames*4)
	for i := 0; i < frames; i++ {
		t := float64(i) / float64(frames)
		envelope := math.Exp(-t * 8)
		sample := math.Sin(2*math.Pi*freq*float64(i)/float64(sr)) * envelope * 0.6
		writeStereoSample(buf, i*4, sample)
	}
	return buf
}

// generateBlip returns a 30ms 800 Hz sine with fast decay for menu navigation.
func generateBlip(sr int) []byte {
	dur := 0.03
	frames := int(float64(sr) * dur)
	buf := make([]byte, frames*4)
	for i := 0; i < frames; i++ {
		t := float64(i) / float64(frames)
		envelope := math.Exp(-t * 12)
		sample := math.Sin(2*math.Pi*800*float64(i)/float64(sr)) * envelope * 0.3
		writeStereoSample(buf, i*4, sample)
	}
	return buf
}

// generateConfirm returns a 60ms 600 Hz sine with moderate decay for menu selection.
func generateConfirm(sr int) []byte {
	dur := 0.06
	frames := int(float64(sr) * dur)
	buf := make([]byte, frames*4)
	for i := 0; i < frames; i++ {
		t := float64(i) / float64(frames)
		envelope := math.Exp(-t * 6)
		sample := math.Sin(2*math.Pi*600*float64(i)/float64(sr)) * envelope * 0.4
		writeStereoSample(buf, i*4, sample)
	}
	return buf
}

// beatIntervalFromAsteroidCount returns the beat interval in ticks.
// Fewer asteroids → faster heartbeat.
func beatIntervalFromAsteroidCount(count int) int {
	interval := 15 + count*4
	if interval > 60 {
		interval = 60
	}
	if interval < 15 {
		interval = 15
	}
	return interval
}
