package pass

import (
	"math"
	// Ensure cmplx is imported if GetResponse is overridden
	"github.com/C0d3-5t3w/go-audioEq/internal/filters"
)

// LowPassFilter implements a low-pass filter (LPF)
type LowPassFilter struct {
	filters.BaseFilter
}

// NewLowPassFilter creates a new low-pass filter
func NewLowPassFilter(freq, q float64) *LowPassFilter {
	lpf := &LowPassFilter{}
	lpf.Freq = freq
	lpf.Q = q
	lpf.GainDB = 0 // Not applicable for standard LPF
	lpf.SetEnabled(true)
	return lpf
}

// Process applies the filter
func (lpf *LowPassFilter) Process(input float64, channel int) float64 {
	return lpf.ProcessSample(input, channel)
}

// SetParams updates filter parameters and recalculates coefficients
func (lpf *LowPassFilter) SetParams(sampleRate float64, params ...float64) {
	if len(params) > filters.ParamFreq {
		lpf.Freq = params[filters.ParamFreq]
	}
	if len(params) > filters.ParamQ {
		lpf.Q = params[filters.ParamQ]
	}

	// Recalculate coefficients (Audio EQ Cookbook - LPF)
	w0 := 2.0 * math.Pi * lpf.Freq / sampleRate
	cosW0 := math.Cos(w0)
	sinW0 := math.Sin(w0)
	alpha := sinW0 / (2.0 * lpf.Q)

	b0 := (1.0 - cosW0) / 2.0
	b1 := 1.0 - cosW0
	b2 := (1.0 - cosW0) / 2.0
	a0 := 1.0 + alpha // Normalization factor
	a1 := -2.0 * cosW0
	a2 := 1.0 - alpha

	lpf.Coeffs.B0 = b0 / a0
	lpf.Coeffs.B1 = b1 / a0
	lpf.Coeffs.B2 = b2 / a0
	lpf.Coeffs.A1 = a1 / a0 // Flipped sign
	lpf.Coeffs.A2 = a2 / a0 // Flipped sign
}

// GetParams returns the current parameters
func (lpf *LowPassFilter) GetParams() []float64 {
	return []float64{lpf.Freq, lpf.Q}
}

// Type returns the filter type
func (lpf *LowPassFilter) Type() string {
	return "LowPass"
}

// GetResponse uses the embedded BaseFilter's implementation

// Compile-time check
var _ filters.Filter = (*LowPassFilter)(nil)
