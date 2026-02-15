package ai

import (
	"encoding/gob"
	"io"
	"math"
)

// ActivationFunc is a function applied element-wise to layer outputs.
type ActivationFunc func(float64) float64

// Tanh activation function.
func Tanh(x float64) float64 { return math.Tanh(x) }

// ReLU activation function.
func ReLU(x float64) float64 {
	if x > 0 {
		return x
	}
	return 0
}

// Sigmoid activation function.
func Sigmoid(x float64) float64 { return 1.0 / (1.0 + math.Exp(-x)) }

// Network is a feedforward neural network with flat weight/bias storage.
type Network struct {
	Layers  []int     // layer sizes, e.g. [34, 16, 5]
	Weights []float64 // flat storage: layer 0→1, then 1→2, etc.
	Biases  []float64 // flat storage: layer 1, then 2, etc.
}

// NewNetwork creates a zero-initialized network with the given layer sizes.
func NewNetwork(layers ...int) *Network {
	if len(layers) < 2 {
		panic("need at least 2 layers (input + output)")
	}
	nw, nb := paramCounts(layers)
	return &Network{
		Layers:  append([]int{}, layers...),
		Weights: make([]float64, nw),
		Biases:  make([]float64, nb),
	}
}

// paramCounts returns total weights and total biases for the given layer sizes.
func paramCounts(layers []int) (int, int) {
	w, b := 0, 0
	for i := 0; i < len(layers)-1; i++ {
		w += layers[i] * layers[i+1]
		b += layers[i+1]
	}
	return w, b
}

// WeightCount returns the total number of tunable parameters (weights + biases).
func (n *Network) WeightCount() int {
	return len(n.Weights) + len(n.Biases)
}

// Forward runs a forward pass through the network.
// Hidden layers use tanh activation; the output layer uses sigmoid.
func (n *Network) Forward(input []float64) []float64 {
	current := input
	wOff := 0
	bOff := 0

	for l := 0; l < len(n.Layers)-1; l++ {
		inSize := n.Layers[l]
		outSize := n.Layers[l+1]
		next := make([]float64, outSize)

		for j := range outSize {
			sum := n.Biases[bOff+j]
			for i := range inSize {
				sum += current[i] * n.Weights[wOff+i*outSize+j]
			}
			// Output layer: sigmoid. Hidden layers: tanh.
			if l == len(n.Layers)-2 {
				next[j] = Sigmoid(sum)
			} else {
				next[j] = Tanh(sum)
			}
		}

		wOff += inSize * outSize
		bOff += outSize
		current = next
	}
	return current
}

// Params exports all weights and biases as a single flat genome slice.
func (n *Network) Params() []float64 {
	out := make([]float64, len(n.Weights)+len(n.Biases))
	copy(out, n.Weights)
	copy(out[len(n.Weights):], n.Biases)
	return out
}

// SetParams loads a flat genome into the network's weights and biases.
func (n *Network) SetParams(params []float64) {
	copy(n.Weights, params[:len(n.Weights)])
	copy(n.Biases, params[len(n.Weights):])
}

// networkData is the serialization format for gob encoding.
type networkData struct {
	Layers  []int
	Weights []float64
	Biases  []float64
}

// Save encodes the network to a writer using gob.
func (n *Network) Save(w io.Writer) error {
	enc := gob.NewEncoder(w)
	return enc.Encode(networkData{
		Layers:  n.Layers,
		Weights: n.Weights,
		Biases:  n.Biases,
	})
}

// Load decodes a network from a reader using gob.
func Load(r io.Reader) (*Network, error) {
	var data networkData
	dec := gob.NewDecoder(r)
	if err := dec.Decode(&data); err != nil {
		return nil, err
	}
	return &Network{
		Layers:  data.Layers,
		Weights: data.Weights,
		Biases:  data.Biases,
	}, nil
}
