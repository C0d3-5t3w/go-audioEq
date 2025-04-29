package pass

import (
	"math"
	// Ensure cmplx is imported if GetResponse is overridden
	"github.com/C0d3-5t3w/go-audioEq/internal/filters"
)

// HighPassFilter implements a high-pass filter (HPF)
type HighPassFilter struct {
	filters.BaseFilter
}

// NewHighPassFilter creates a new high-pass filter
func NewHighPassFilter(freq, q float64) *HighPassFilter {
	hpf := &HighPassFilter{}
	hpf.Freq = freq
	hpf.Q = q
	hpf.GainDB = 0 // Not applicable
	hpf.SetEnabled(true)
	return hpf
}

// Process applies the filter
func (hpf *HighPassFilter) Process(input float64, channel int) float64 {
	return hpf.ProcessSample(input, channel)
}

// SetParams updates filter parameters and recalculates coefficients
func (hpf *HighPassFilter) SetParams(sampleRate float64, params ...float64) {
	if len(params) > filters.ParamFreq {
		hpf.Freq = params[filters.ParamFreq]
	}
	if len(params) > filters.ParamQ {
		hpf.Q = params[filters.ParamQ]
	}

	// Recalculate coefficients (Audio EQ Cookbook - HPF)
	w0 := 2.0 * math.Pi * hpf.Freq / sampleRate
	cosW0 := math.Cos(w0)
	sinW0 := math.Sin(w0)
	alpha := sinW0 / (2.0 * hpf.Q)

	b0 := (1.0 + cosW0) / 2.0
	b1 := -(1.0 + cosW0)
	b2 := (1.0 + cosW0) / 2.0
	a0 := 1.0 + alpha // Normalization factor
	a1 := -2.0 * cosW0
	a2 := 1.0 - alpha

	hpf.Coeffs.B0 = b0 / a0
	hpf.Coeffs.B1 = b1 / a0
	hpf.Coeffs.B2 = b2 / a0
	hpf.Coeffs.A1 = a1 / a0 // Flipped sign
	hpf.Coeffs.A2 = a2 / a0 // Flipped sign
}

// GetParams returns the current parameters
func (hpf *HighPassFilter) GetParams() []float64 {
	return []float64{hpf.Freq, hpf.Q}
}

// Type returns the filter type
func (hpf *HighPassFilter) Type() string {
	return "HighPass"
}

// GetResponse uses the embedded BaseFilter's implementation

// Compile-time check
var _ filters.Filter = (*HighPassFilter)(nil)
