package notch

import (
	"math"
	// Ensure cmplx is imported if GetResponse is overridden
	"github.com/C0d3-5t3w/go-audioEq/internal/filters"
)

// NotchFilter implements a notch (band-stop) filter
type NotchFilter struct {
	filters.BaseFilter
}

// NewNotchFilter creates a new notch filter
func NewNotchFilter(freq, q float64) *NotchFilter {
	nf := &NotchFilter{}
	nf.Freq = freq
	nf.Q = q
	nf.GainDB = 0 // Not applicable
	nf.SetEnabled(true)
	return nf
}

// Process applies the filter
func (nf *NotchFilter) Process(input float64, channel int) float64 {
	return nf.ProcessSample(input, channel)
}

// SetParams updates filter parameters and recalculates coefficients
func (nf *NotchFilter) SetParams(sampleRate float64, params ...float64) {
	if len(params) > filters.ParamFreq {
		nf.Freq = params[filters.ParamFreq]
	}
	if len(params) > filters.ParamQ {
		nf.Q = params[filters.ParamQ]
	}

	// Recalculate coefficients (Audio EQ Cookbook - Notch)
	w0 := 2.0 * math.Pi * nf.Freq / sampleRate
	cosW0 := math.Cos(w0)
	sinW0 := math.Sin(w0)
	alpha := sinW0 / (2.0 * nf.Q) // Or use BW: alpha = sin(w0)*sinh( ln(2)/2 * BW * w0/sin(w0) )

	b0 := 1.0
	b1 := -2.0 * cosW0
	b2 := 1.0
	a0 := 1.0 + alpha // Normalization factor
	a1 := -2.0 * cosW0
	a2 := 1.0 - alpha

	nf.Coeffs.B0 = b0 / a0
	nf.Coeffs.B1 = b1 / a0
	nf.Coeffs.B2 = b2 / a0
	nf.Coeffs.A1 = a1 / a0 // Flipped sign
	nf.Coeffs.A2 = a2 / a0 // Flipped sign
}

// GetParams returns the current parameters
func (nf *NotchFilter) GetParams() []float64 {
	return []float64{nf.Freq, nf.Q}
}

// Type returns the filter type
func (nf *NotchFilter) Type() string {
	return "Notch"
}

// GetResponse uses the embedded BaseFilter's implementation

// Compile-time check
var _ filters.Filter = (*NotchFilter)(nil)
