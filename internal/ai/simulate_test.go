package ai

import (
	"testing"

	"github.com/matheus3301/asteroids/internal/game"
)

// staticAgent always returns the same action.
type staticAgent struct {
	action game.AIAction
}

func (a *staticAgent) Act(_ []float64) game.AIAction { return a.action }

func TestRunSimulation_Completes(t *testing.T) {
	agent := &staticAgent{}
	cfg := DefaultSimConfig()
	cfg.MaxTicks = 60 // short run

	result := RunSimulation(agent, cfg)

	if result.Ticks < 0 {
		t.Error("ticks should be non-negative")
	}
	if result.Fitness < 0 {
		t.Error("fitness should be non-negative")
	}
}

func TestRunSimulation_ThrustingAgentSurvives(t *testing.T) {
	// An agent that thrusts and shoots should survive some ticks
	agent := &staticAgent{action: game.AIAction{Thrust: true, Shoot: true}}
	cfg := DefaultSimConfig()
	cfg.MaxTicks = 120

	result := RunSimulation(agent, cfg)

	if result.Ticks == 0 {
		t.Error("agent should survive at least a few ticks (invulnerability)")
	}
}

func TestRunSimulation_FitnessIncludesComponents(t *testing.T) {
	agent := &staticAgent{}
	cfg := DefaultSimConfig()
	cfg.MaxTicks = 60
	cfg.NumRuns = 1

	result := RunSimulation(agent, cfg)

	// Fitness should be at least survival bonus * ticks (aim reward adds more)
	minFitness := float64(result.Ticks) * cfg.SurvivalBonus
	if result.Fitness < minFitness {
		t.Errorf("fitness %v should be >= survival component %v", result.Fitness, minFitness)
	}
}

func TestRunSimulation_MultiRun(t *testing.T) {
	agent := &staticAgent{action: game.AIAction{Thrust: true}}
	cfg := DefaultSimConfig()
	cfg.MaxTicks = 60
	cfg.NumRuns = 3

	result := RunSimulation(agent, cfg)
	// With 3 runs averaged, result should still be reasonable
	if result.Fitness <= 0 {
		t.Error("multi-run fitness should be positive")
	}
}

func TestNeuralAgent_ImplementsAIAgent(t *testing.T) {
	net := NewNetwork(game.ObservationSize, 16, 5)
	agent := NewNeuralAgent(net)

	// Verify it implements the interface
	var _ game.AIAgent = agent

	obs := make([]float64, game.ObservationSize)
	action := agent.Act(obs)

	// With zero weights, all outputs should be sigmoid(0)=0.5, which equals
	// the threshold, so all actions should be false.
	if action.RotateLeft || action.RotateRight || action.Thrust || action.Shoot || action.Hyperspace {
		t.Error("zero-weight network at threshold 0.5 should produce all-false actions")
	}
}

func TestNeuralAgent_ThresholdAffectsOutput(t *testing.T) {
	net := NewNetwork(game.ObservationSize, 16, 5)
	// With zero weights, output is 0.5. Setting threshold below 0.5 should activate all.
	agent := &NeuralAgent{Net: net, Threshold: 0.49}

	obs := make([]float64, game.ObservationSize)
	action := agent.Act(obs)

	if !action.RotateLeft || !action.RotateRight || !action.Thrust || !action.Shoot || !action.Hyperspace {
		t.Error("with threshold 0.49 and output 0.5, all actions should be true")
	}
}

func TestRunSimulation_WithNeuralAgent(t *testing.T) {
	net := NewNetwork(game.ObservationSize, 16, 5)
	agent := NewNeuralAgent(net)
	cfg := DefaultSimConfig()
	cfg.MaxTicks = 60

	result := RunSimulation(agent, cfg)

	if result.Ticks == 0 && result.Score == 0 && result.Fitness == 0 {
		t.Error("simulation should produce some result")
	}
}
