package audio

import (
	"testing"
)

func TestNewVAD(t *testing.T) {
	config := DefaultVADConfig()
	vad := NewVAD(config, func() {}, func(data []int16) {})

	if vad == nil {
		t.Fatal("NewVAD() returned nil")
	}
}

func TestDefaultVADConfig(t *testing.T) {
	config := DefaultVADConfig()

	if config.EnergyThreshold <= 0 {
		t.Error("EnergyThreshold should be > 0")
	}
	if config.SampleRate <= 0 {
		t.Error("SampleRate should be > 0")
	}
	if config.FrameSize <= 0 {
		t.Error("FrameSize should be > 0")
	}
}
