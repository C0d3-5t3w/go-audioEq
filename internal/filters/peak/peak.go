package peak

import (
	"math"
	// Ensure cmplx is imported if GetResponse is overridden
	"github.com/C0d3-5t3w/go-audioEq/internal/filters"
)

// PeakFilter implements a peaking EQ filter
type PeakFilter struct {
	filters.BaseFilter
}

// NewPeakFilter creates a new peak filter
func NewPeakFilter(freq, q, gainDB float64) *PeakFilter {
	pf := &PeakFilter{}
	pf.Freq = freq
	pf.Q = q
	pf.GainDB = gainDB
	pf.SetEnabled(true) // Enabled by default
	return pf
}

// Process applies the filter to a single sample
func (pf *PeakFilter) Process(input float64, channel int) float64 {
	return pf.ProcessSample(input, channel)
}

// SetParams updates filter parameters and recalculates coefficients
func (pf *PeakFilter) SetParams(sampleRate float64, params ...float64) {
	if len(params) > filters.ParamFreq {
		pf.Freq = params[filters.ParamFreq]
	}
	if len(params) > filters.ParamQ {
		pf.Q = params[filters.ParamQ]
	}
	if len(params) > filters.ParamGainDB {
		pf.GainDB = params[filters.ParamGainDB]
	}

	// Recalculate coefficients (using formulas from Audio EQ Cookbook)
	A := math.Pow(10.0, pf.GainDB/40.0) // Gain converted for calculation
	w0 := 2.0 * math.Pi * pf.Freq / sampleRate
	cosW0 := math.Cos(w0)
	sinW0 := math.Sin(w0)
	alpha := sinW0 / (2.0 * pf.Q)

	b0 := 1.0 + alpha*A
	b1 := -2.0 * cosW0
	b2 := 1.0 - alpha*A
	a0 := 1.0 + alpha/A // Normalization factor
	a1 := -2.0 * cosW0
	a2 := 1.0 - alpha/A

	pf.Coeffs.B0 = b0 / a0
	pf.Coeffs.B1 = b1 / a0
	pf.Coeffs.B2 = b2 / a0
	pf.Coeffs.A1 = a1 / a0 // Note: Sign flipped for standard difference equation y[n] = b0x[n] + ... - a1y[n-1] - ...
	pf.Coeffs.A2 = a2 / a0 // Note: Sign flipped

	// Reset state when parameters change significantly to avoid instability
	// pf.ResetState() // Consider if needed
}

// GetParams returns the current parameters
func (pf *PeakFilter) GetParams() []float64 {
	return []float64{pf.Freq, pf.Q, pf.GainDB}
}

// Type returns the filter type
func (pf *PeakFilter) Type() string {
	return "Peak"
}

// GetResponse uses the embedded BaseFilter's implementation

// Compile-time check
var _ filters.Filter = (*PeakFilter)(nil)
