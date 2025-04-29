package pass

import (
	"math"
	// Ensure cmplx is imported if GetResponse is overridden
	"github.com/C0d3-5t3w/go-audioEq/internal/filters"
)

// AllPassFilter implements an all-pass filter (phase shift)
type AllPassFilter struct {
	filters.BaseFilter
}

// NewAllPassFilter creates a new all-pass filter
func NewAllPassFilter(freq, q float64) *AllPassFilter {
	apf := &AllPassFilter{}
	apf.Freq = freq
	apf.Q = q
	apf.GainDB = 0 // Not applicable
	apf.SetEnabled(true)
	return apf
}

// Process applies the filter
func (apf *AllPassFilter) Process(input float64, channel int) float64 {
	return apf.ProcessSample(input, channel)
}

// SetParams updates filter parameters and recalculates coefficients
func (apf *AllPassFilter) SetParams(sampleRate float64, params ...float64) {
	if len(params) > filters.ParamFreq {
		apf.Freq = params[filters.ParamFreq]
	}
	if len(params) > filters.ParamQ {
		apf.Q = params[filters.ParamQ]
	}

	// Recalculate coefficients (Audio EQ Cookbook - APF)
	w0 := 2.0 * math.Pi * apf.Freq / sampleRate
	cosW0 := math.Cos(w0)
	sinW0 := math.Sin(w0)
	alpha := sinW0 / (2.0 * apf.Q)

	b0 := 1.0 - alpha
	b1 := -2.0 * cosW0
	b2 := 1.0 + alpha
	a0 := 1.0 + alpha // Normalization factor
	a1 := -2.0 * cosW0
	a2 := 1.0 - alpha

	apf.Coeffs.B0 = b0 / a0
	apf.Coeffs.B1 = b1 / a0
	apf.Coeffs.B2 = b2 / a0
	apf.Coeffs.A1 = a1 / a0 // Flipped sign
	apf.Coeffs.A2 = a2 / a0 // Flipped sign
}

// GetParams returns the current parameters
func (apf *AllPassFilter) GetParams() []float64 {
	return []float64{apf.Freq, apf.Q}
}

// Type returns the filter type
func (apf *AllPassFilter) Type() string {
	return "AllPass"
}

// GetResponse uses the embedded BaseFilter's implementation

// Compile-time check
var _ filters.Filter = (*AllPassFilter)(nil)
