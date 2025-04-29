package shelf

import (
	"math"
	// Ensure cmplx is imported if GetResponse is overridden
	"github.com/C0d3-5t3w/go-audioEq/internal/filters"
)

// HighShelfFilter implements a high-shelf EQ filter
type HighShelfFilter struct {
	filters.BaseFilter
}

// NewHighShelfFilter creates a new high-shelf filter
func NewHighShelfFilter(freq, q, gainDB float64) *HighShelfFilter {
	hsf := &HighShelfFilter{}
	hsf.Freq = freq
	hsf.Q = q // Or interpret as slope parameter S
	hsf.GainDB = gainDB
	hsf.SetEnabled(true)
	return hsf
}

// Process applies the filter
func (hsf *HighShelfFilter) Process(input float64, channel int) float64 {
	return hsf.ProcessSample(input, channel)
}

// SetParams updates filter parameters and recalculates coefficients
func (hsf *HighShelfFilter) SetParams(sampleRate float64, params ...float64) {
	if len(params) > filters.ParamFreq {
		hsf.Freq = params[filters.ParamFreq]
	}
	if len(params) > filters.ParamQ {
		hsf.Q = params[filters.ParamQ]
	}
	if len(params) > filters.ParamGainDB {
		hsf.GainDB = params[filters.ParamGainDB]
	}

	// Recalculate coefficients (Audio EQ Cookbook)
	A := math.Pow(10.0, hsf.GainDB/40.0)
	w0 := 2.0 * math.Pi * hsf.Freq / sampleRate
	cosW0 := math.Cos(w0)
	sinW0 := math.Sin(w0)
	// alpha := sinW0 / (2.0 * hsf.q) // Using Q
	// Or using Shelf Slope S
	S := 1.0 // Example slope parameter
	alpha := sinW0 / 2.0 * math.Sqrt((A+1.0/A)*(1.0/S-1.0)+2.0)

	sqrtA2Alpha := 2.0 * math.Sqrt(A) * alpha

	b0 := A * ((A + 1.0) + (A-1.0)*cosW0 + sqrtA2Alpha)
	b1 := -2.0 * A * ((A - 1.0) + (A+1.0)*cosW0)
	b2 := A * ((A + 1.0) + (A-1.0)*cosW0 - sqrtA2Alpha)
	a0 := (A + 1.0) - (A-1.0)*cosW0 + sqrtA2Alpha // Normalization factor
	a1 := 2.0 * ((A - 1.0) - (A+1.0)*cosW0)
	a2 := (A + 1.0) - (A-1.0)*cosW0 - sqrtA2Alpha

	hsf.Coeffs.B0 = b0 / a0
	hsf.Coeffs.B1 = b1 / a0
	hsf.Coeffs.B2 = b2 / a0
	hsf.Coeffs.A1 = a1 / a0 // Flipped sign
	hsf.Coeffs.A2 = a2 / a0 // Flipped sign
}

// GetParams returns the current parameters
func (hsf *HighShelfFilter) GetParams() []float64 {
	return []float64{hsf.Freq, hsf.Q, hsf.GainDB}
}

// Type returns the filter type
func (hsf *HighShelfFilter) Type() string {
	return "HighShelf"
}

// GetResponse uses the embedded BaseFilter's implementation

// Compile-time check
var _ filters.Filter = (*HighShelfFilter)(nil)
