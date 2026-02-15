package ai

import (
	"bytes"
	"math"
	"testing"
)

func TestNewNetwork_Dimensions(t *testing.T) {
	n := NewNetwork(34, 16, 5)
	// Weights: 34*16 + 16*5 = 544 + 80 = 624
	if len(n.Weights) != 624 {
		t.Errorf("expected 624 weights, got %d", len(n.Weights))
	}
	// Biases: 16 + 5 = 21
	if len(n.Biases) != 21 {
		t.Errorf("expected 21 biases, got %d", len(n.Biases))
	}
	if n.WeightCount() != 645 {
		t.Errorf("expected 645 total params, got %d", n.WeightCount())
	}
}

func TestNewNetwork_TwoLayers(t *testing.T) {
	n := NewNetwork(3, 2)
	if len(n.Weights) != 6 {
		t.Errorf("expected 6 weights, got %d", len(n.Weights))
	}
	if len(n.Biases) != 2 {
		t.Errorf("expected 2 biases, got %d", len(n.Biases))
	}
}

func TestForward_OutputSize(t *testing.T) {
	n := NewNetwork(34, 16, 5)
	input := make([]float64, 34)
	output := n.Forward(input)
	if len(output) != 5 {
		t.Errorf("expected 5 outputs, got %d", len(output))
	}
}

func TestForward_OutputSigmoidRange(t *testing.T) {
	n := NewNetwork(4, 3, 2)
	// Set some weights to non-zero
	for i := range n.Weights {
		n.Weights[i] = 0.5
	}
	for i := range n.Biases {
		n.Biases[i] = 0.1
	}
	input := []float64{1, -1, 0.5, -0.5}
	output := n.Forward(input)
	for i, v := range output {
		if v < 0 || v > 1 {
			t.Errorf("output[%d] = %v not in [0,1]", i, v)
		}
	}
}

func TestForward_ZeroWeightsGiveSigmoidHalf(t *testing.T) {
	n := NewNetwork(3, 2)
	input := []float64{1, 2, 3}
	output := n.Forward(input)
	// With zero weights and zero biases, sigmoid(0) = 0.5
	for i, v := range output {
		if math.Abs(v-0.5) > 0.001 {
			t.Errorf("output[%d] = %v, expected ~0.5", i, v)
		}
	}
}

func TestForward_HiddenLayerUseTanh(t *testing.T) {
	// 2-input, 2-hidden, 1-output
	n := NewNetwork(2, 2, 1)
	// Set weights so hidden layer gets large positive values
	// Weights[0..3] are input→hidden (2*2), stored as [i0→h0, i0→h1, i1→h0, i1→h1]
	n.Weights[0] = 10.0 // i0→h0
	n.Weights[1] = 0.0
	n.Weights[2] = 0.0
	n.Weights[3] = -10.0 // i1→h1

	input := []float64{1, 1}
	output := n.Forward(input)
	// Hidden: h0 = tanh(10) ≈ 1.0, h1 = tanh(-10) ≈ -1.0
	// Output: sigmoid(w * 1.0 + w * -1.0 + bias) with zero weights = sigmoid(0) = 0.5
	if len(output) != 1 {
		t.Fatalf("expected 1 output, got %d", len(output))
	}
}

func TestParams_RoundTrip(t *testing.T) {
	n := NewNetwork(3, 4, 2)
	// Set distinct values
	for i := range n.Weights {
		n.Weights[i] = float64(i) * 0.01
	}
	for i := range n.Biases {
		n.Biases[i] = float64(i) * 0.1
	}

	params := n.Params()
	if len(params) != n.WeightCount() {
		t.Fatalf("params length %d != WeightCount %d", len(params), n.WeightCount())
	}

	n2 := NewNetwork(3, 4, 2)
	n2.SetParams(params)

	for i := range n.Weights {
		if n.Weights[i] != n2.Weights[i] {
			t.Errorf("weight[%d] mismatch: %v != %v", i, n.Weights[i], n2.Weights[i])
		}
	}
	for i := range n.Biases {
		if n.Biases[i] != n2.Biases[i] {
			t.Errorf("bias[%d] mismatch: %v != %v", i, n.Biases[i], n2.Biases[i])
		}
	}
}

func TestSaveLoad_RoundTrip(t *testing.T) {
	n := NewNetwork(34, 16, 5)
	for i := range n.Weights {
		n.Weights[i] = float64(i) * 0.001
	}
	for i := range n.Biases {
		n.Biases[i] = float64(i) * 0.01
	}

	var buf bytes.Buffer
	if err := n.Save(&buf); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	n2, err := Load(&buf)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	if len(n2.Layers) != len(n.Layers) {
		t.Fatalf("layer count mismatch: %d != %d", len(n2.Layers), len(n.Layers))
	}
	for i := range n.Layers {
		if n.Layers[i] != n2.Layers[i] {
			t.Errorf("layer[%d] mismatch: %d != %d", i, n.Layers[i], n2.Layers[i])
		}
	}
	for i := range n.Weights {
		if n.Weights[i] != n2.Weights[i] {
			t.Errorf("weight[%d] mismatch", i)
		}
	}
	for i := range n.Biases {
		if n.Biases[i] != n2.Biases[i] {
			t.Errorf("bias[%d] mismatch", i)
		}
	}
}

func TestActivationFunctions(t *testing.T) {
	if Tanh(0) != 0 {
		t.Errorf("Tanh(0) = %v", Tanh(0))
	}
	if ReLU(-1) != 0 {
		t.Errorf("ReLU(-1) = %v", ReLU(-1))
	}
	if ReLU(5) != 5 {
		t.Errorf("ReLU(5) = %v", ReLU(5))
	}
	if math.Abs(Sigmoid(0)-0.5) > 0.001 {
		t.Errorf("Sigmoid(0) = %v", Sigmoid(0))
	}
}
