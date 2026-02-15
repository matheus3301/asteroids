package ai

import "github.com/matheus3301/asteroids/internal/game"

// NeuralAgent implements game.AIAgent using a feedforward neural network.
type NeuralAgent struct {
	Net       *Network
	Threshold float64
}

// NewNeuralAgent creates an agent with the default threshold of 0.5.
func NewNeuralAgent(net *Network) *NeuralAgent {
	return &NeuralAgent{Net: net, Threshold: 0.5}
}

// Act runs the neural network on the observation and returns an AIAction.
func (a *NeuralAgent) Act(observation []float64) game.AIAction {
	out := a.Net.Forward(observation)
	return game.AIAction{
		RotateLeft:  out[0] > a.Threshold,
		RotateRight: out[1] > a.Threshold,
		Thrust:      out[2] > a.Threshold,
		Shoot:       out[3] > a.Threshold,
		Hyperspace:  out[4] > a.Threshold,
	}
}
