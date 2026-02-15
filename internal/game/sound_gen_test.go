package game

import (
	"encoding/binary"
	"math"
	"testing"
)

func readSample(buf []byte, frame int) (left, right int16) {
	off := frame * 4
	left = int16(binary.LittleEndian.Uint16(buf[off:]))
	right = int16(binary.LittleEndian.Uint16(buf[off+2:]))
	return
}

func TestGenerateFire_Length(t *testing.T) {
	buf := generateFire(sampleRate)
	expectedFrames := int(float64(sampleRate) * 0.08)
	if len(buf) != expectedFrames*4 {
		t.Errorf("expected %d bytes, got %d", expectedFrames*4, len(buf))
	}
}

func TestGenerateFire_SampleRange(t *testing.T) {
	buf := generateFire(sampleRate)
	frames := len(buf) / 4
	for i := 0; i < frames; i++ {
		l, r := readSample(buf, i)
		if l < -32767 || r < -32767 {
			t.Fatalf("sample out of range at frame %d: L=%d R=%d", i, l, r)
		}
	}
}

func TestGenerateFire_NotSilent(t *testing.T) {
	buf := generateFire(sampleRate)
	frames := len(buf) / 4
	hasLoud := false
	for i := 0; i < frames; i++ {
		l, _ := readSample(buf, i)
		if l > 100 || l < -100 {
			hasLoud = true
			break
		}
	}
	if !hasLoud {
		t.Error("fire sound is silent")
	}
}

func TestGenerateFire_StereoSymmetry(t *testing.T) {
	buf := generateFire(sampleRate)
	frames := len(buf) / 4
	for i := 0; i < frames; i++ {
		l, r := readSample(buf, i)
		if l != r {
			t.Fatalf("stereo mismatch at frame %d: L=%d R=%d", i, l, r)
		}
	}
}

func TestGenerateFire_EnvelopeDecay(t *testing.T) {
	buf := generateFire(sampleRate)
	frames := len(buf) / 4
	tenPct := frames / 10

	var earlyPeak, latePeak float64
	for i := 0; i < tenPct; i++ {
		l, _ := readSample(buf, i)
		if math.Abs(float64(l)) > earlyPeak {
			earlyPeak = math.Abs(float64(l))
		}
	}
	for i := frames - tenPct; i < frames; i++ {
		l, _ := readSample(buf, i)
		if math.Abs(float64(l)) > latePeak {
			latePeak = math.Abs(float64(l))
		}
	}
	if earlyPeak <= latePeak {
		t.Errorf("envelope should decay: early peak %.0f, late peak %.0f", earlyPeak, latePeak)
	}
}

func TestGenerateExplosion_SizeOrdering(t *testing.T) {
	small := generateExplosion(sampleRate, SizeSmall)
	med := generateExplosion(sampleRate, SizeMedium)
	large := generateExplosion(sampleRate, SizeLarge)

	if len(large) <= len(med) {
		t.Errorf("large (%d) should be longer than med (%d)", len(large), len(med))
	}
	if len(med) <= len(small) {
		t.Errorf("med (%d) should be longer than small (%d)", len(med), len(small))
	}
}

func TestGenerateExplosion_NotSilent(t *testing.T) {
	for _, size := range []AsteroidSize{SizeLarge, SizeMedium, SizeSmall} {
		buf := generateExplosion(sampleRate, size)
		frames := len(buf) / 4
		hasLoud := false
		for i := 0; i < frames; i++ {
			l, _ := readSample(buf, i)
			if l > 100 || l < -100 {
				hasLoud = true
				break
			}
		}
		if !hasLoud {
			t.Errorf("explosion size %d is silent", size)
		}
	}
}

func TestGenerateExplosion_StereoSymmetry(t *testing.T) {
	buf := generateExplosion(sampleRate, SizeLarge)
	frames := len(buf) / 4
	for i := 0; i < frames; i++ {
		l, r := readSample(buf, i)
		if l != r {
			t.Fatalf("stereo mismatch at frame %d: L=%d R=%d", i, l, r)
		}
	}
}

func TestGenerateExplosion_EnvelopeDecay(t *testing.T) {
	buf := generateExplosion(sampleRate, SizeLarge)
	frames := len(buf) / 4
	tenPct := frames / 10

	var earlyRMS, lateRMS float64
	for i := 0; i < tenPct; i++ {
		l, _ := readSample(buf, i)
		earlyRMS += float64(l) * float64(l)
	}
	earlyRMS = math.Sqrt(earlyRMS / float64(tenPct))

	for i := frames - tenPct; i < frames; i++ {
		l, _ := readSample(buf, i)
		lateRMS += float64(l) * float64(l)
	}
	lateRMS = math.Sqrt(lateRMS / float64(tenPct))

	if earlyRMS <= lateRMS {
		t.Errorf("envelope should decay: early RMS %.0f, late RMS %.0f", earlyRMS, lateRMS)
	}
}

func TestGenerateDeath_Length(t *testing.T) {
	buf := generateDeath(sampleRate)
	expectedFrames := int(float64(sampleRate) * 0.8)
	if len(buf) != expectedFrames*4 {
		t.Errorf("expected %d bytes, got %d", expectedFrames*4, len(buf))
	}
}

func TestGenerateDeath_NotSilent(t *testing.T) {
	buf := generateDeath(sampleRate)
	frames := len(buf) / 4
	hasLoud := false
	for i := 0; i < frames; i++ {
		l, _ := readSample(buf, i)
		if l > 100 || l < -100 {
			hasLoud = true
			break
		}
	}
	if !hasLoud {
		t.Error("death sound is silent")
	}
}

func TestGenerateDeath_StereoSymmetry(t *testing.T) {
	buf := generateDeath(sampleRate)
	frames := len(buf) / 4
	for i := 0; i < frames; i++ {
		l, r := readSample(buf, i)
		if l != r {
			t.Fatalf("stereo mismatch at frame %d: L=%d R=%d", i, l, r)
		}
	}
}

func TestGenerateThrustLoop_Length(t *testing.T) {
	buf := generateThrustLoop(sampleRate)
	expectedFrames := int(float64(sampleRate) * 0.2)
	if len(buf) != expectedFrames*4 {
		t.Errorf("expected %d bytes, got %d", expectedFrames*4, len(buf))
	}
}

func TestGenerateThrustLoop_NotSilent(t *testing.T) {
	buf := generateThrustLoop(sampleRate)
	frames := len(buf) / 4
	hasLoud := false
	for i := 0; i < frames; i++ {
		l, _ := readSample(buf, i)
		if l > 100 || l < -100 {
			hasLoud = true
			break
		}
	}
	if !hasLoud {
		t.Error("thrust loop is silent")
	}
}

func TestGenerateBeatTone_Length(t *testing.T) {
	buf := generateBeatTone(sampleRate, 55)
	expectedFrames := int(float64(sampleRate) * 0.06)
	if len(buf) != expectedFrames*4 {
		t.Errorf("expected %d bytes, got %d", expectedFrames*4, len(buf))
	}
}

func TestGenerateBeatTone_NotSilent(t *testing.T) {
	buf := generateBeatTone(sampleRate, 55)
	frames := len(buf) / 4
	hasLoud := false
	for i := 0; i < frames; i++ {
		l, _ := readSample(buf, i)
		if l > 100 || l < -100 {
			hasLoud = true
			break
		}
	}
	if !hasLoud {
		t.Error("beat tone is silent")
	}
}

func TestGenerateBeatTone_EnvelopeDecay(t *testing.T) {
	buf := generateBeatTone(sampleRate, 55)
	frames := len(buf) / 4
	tenPct := frames / 10

	var earlyPeak, latePeak float64
	for i := 0; i < tenPct; i++ {
		l, _ := readSample(buf, i)
		if math.Abs(float64(l)) > earlyPeak {
			earlyPeak = math.Abs(float64(l))
		}
	}
	for i := frames - tenPct; i < frames; i++ {
		l, _ := readSample(buf, i)
		if math.Abs(float64(l)) > latePeak {
			latePeak = math.Abs(float64(l))
		}
	}
	if earlyPeak <= latePeak {
		t.Errorf("envelope should decay: early peak %.0f, late peak %.0f", earlyPeak, latePeak)
	}
}

func TestGenerateBlip_Length(t *testing.T) {
	buf := generateBlip(sampleRate)
	expectedFrames := int(float64(sampleRate) * 0.03)
	if len(buf) != expectedFrames*4 {
		t.Errorf("expected %d bytes, got %d", expectedFrames*4, len(buf))
	}
}

func TestGenerateBlip_NotSilent(t *testing.T) {
	buf := generateBlip(sampleRate)
	frames := len(buf) / 4
	hasLoud := false
	for i := 0; i < frames; i++ {
		l, _ := readSample(buf, i)
		if l > 100 || l < -100 {
			hasLoud = true
			break
		}
	}
	if !hasLoud {
		t.Error("blip is silent")
	}
}

func TestGenerateConfirm_Length(t *testing.T) {
	buf := generateConfirm(sampleRate)
	expectedFrames := int(float64(sampleRate) * 0.06)
	if len(buf) != expectedFrames*4 {
		t.Errorf("expected %d bytes, got %d", expectedFrames*4, len(buf))
	}
}

func TestGenerateConfirm_NotSilent(t *testing.T) {
	buf := generateConfirm(sampleRate)
	frames := len(buf) / 4
	hasLoud := false
	for i := 0; i < frames; i++ {
		l, _ := readSample(buf, i)
		if l > 100 || l < -100 {
			hasLoud = true
			break
		}
	}
	if !hasLoud {
		t.Error("confirm is silent")
	}
}

func TestGenerateConfirm_LongerThanBlip(t *testing.T) {
	blip := generateBlip(sampleRate)
	confirm := generateConfirm(sampleRate)
	if len(confirm) <= len(blip) {
		t.Errorf("confirm (%d) should be longer than blip (%d)", len(confirm), len(blip))
	}
}

func TestBeatIntervalFromAsteroidCount(t *testing.T) {
	tests := []struct {
		count    int
		expected int
	}{
		{0, 15},
		{1, 19},
		{5, 35},
		{10, 55},
		{11, 59},
		{12, 60},
		{20, 60},
	}
	for _, tt := range tests {
		got := beatIntervalFromAsteroidCount(tt.count)
		if got != tt.expected {
			t.Errorf("beatIntervalFromAsteroidCount(%d) = %d, want %d", tt.count, got, tt.expected)
		}
	}
}
