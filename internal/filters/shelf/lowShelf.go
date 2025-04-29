package shelf

import (
	"math"
	// Ensure cmplx is imported if GetResponse is overridden
	"github.com/C0d3-5t3w/go-audioEq/internal/filters"
)

// LowShelfFilter implements a low-shelf EQ filter
type LowShelfFilter struct {
	filters.BaseFilter
}

// NewLowShelfFilter creates a new low-shelf filter
func NewLowShelfFilter(freq, q, gainDB float64) *LowShelfFilter {
	lsf := &LowShelfFilter{}
	lsf.Freq = freq     // Use exported Freq
	lsf.Q = q           // Use exported Q // Or interpret as slope parameter S if needed
	lsf.GainDB = gainDB // Use exported GainDB
	lsf.SetEnabled(true)
	return lsf
}

// Process applies the filter
func (lsf *LowShelfFilter) Process(input float64, channel int) float64 {
	return lsf.ProcessSample(input, channel) // Use exported ProcessSample
}

// SetParams updates filter parameters and recalculates coefficients
func (lsf *LowShelfFilter) SetParams(sampleRate float64, params ...float64) {
	if len(params) > filters.ParamFreq {
		lsf.Freq = params[filters.ParamFreq] // Use exported Freq
	}
	if len(params) > filters.ParamQ {
		lsf.Q = params[filters.ParamQ] // Use exported Q // Use Q or calculate S from it if desired
	}
	if len(params) > filters.ParamGainDB {
		lsf.GainDB = params[filters.ParamGainDB] // Use exported GainDB
	}

	// Recalculate coefficients (Audio EQ Cookbook)
	A := math.Pow(10.0, lsf.GainDB/40.0)        // Use exported GainDB
	w0 := 2.0 * math.Pi * lsf.Freq / sampleRate // Use exported Freq
	cosW0 := math.Cos(w0)
	sinW0 := math.Sin(w0)
	// alpha := sinW0 / (2.0 * lsf.q) // Using Q
	// Or using Shelf Slope S (steeper slope for lower S, typically 1)
	S := 1.0 // Example slope parameter, could be derived from Q or set directly
	alpha := sinW0 / 2.0 * math.Sqrt((A+1.0/A)*(1.0/S-1.0)+2.0)

	sqrtA2Alpha := 2.0 * math.Sqrt(A) * alpha

	b0 := A * ((A + 1.0) - (A-1.0)*cosW0 + sqrtA2Alpha)
	b1 := 2.0 * A * ((A - 1.0) - (A+1.0)*cosW0)
	b2 := A * ((A + 1.0) - (A-1.0)*cosW0 - sqrtA2Alpha)
	a0 := (A + 1.0) + (A-1.0)*cosW0 + sqrtA2Alpha // Normalization factor
	a1 := -2.0 * ((A - 1.0) + (A+1.0)*cosW0)
	a2 := (A + 1.0) + (A-1.0)*cosW0 - sqrtA2Alpha

	lsf.Coeffs.B0 = b0 / a0 // Use exported Coeffs
	lsf.Coeffs.B1 = b1 / a0 // Use exported Coeffs
	lsf.Coeffs.B2 = b2 / a0 // Use exported Coeffs
	lsf.Coeffs.A1 = a1 / a0 // Use exported Coeffs // Flipped sign
	lsf.Coeffs.A2 = a2 / a0 // Use exported Coeffs // Flipped sign
}

// GetParams returns the current parameters
func (lsf *LowShelfFilter) GetParams() []float64 {
	return []float64{lsf.Freq, lsf.Q, lsf.GainDB} // Use exported Freq, Q, GainDB
}

// Type returns the filter type
func (lsf *LowShelfFilter) Type() string {
	return "LowShelf"
}

// GetResponse uses the embedded BaseFilter's implementation
// func (lsf *LowShelfFilter) GetResponse(z complex128) complex128 {
// 	 return lsf.BaseFilter.GetResponse(z)
// }

// Compile-time check
var _ filters.Filter = (*LowShelfFilter)(nil)
