package filters

import (
	"math"
	"math/cmplx"
)

// Added for GetResponse

// Filter defines the interface for all EQ filter types
type Filter interface {
	// Process applies the filter to a single sample for a given channel
	Process(input float64, channel int) float64
	// SetParams updates filter parameters (e.g., frequency, Q, gain) and recalculates coefficients based on sample rate
	SetParams(sampleRate float64, params ...float64) // params could be freq, Q, gain etc.
	// GetParams returns the current parameters of the filter
	GetParams() []float64
	// Type returns the type of the filter (e.g., "Peak", "LowShelf")
	Type() string
	// SetEnabled enables or disables the filter
	SetEnabled(enable bool)
	// ResetState clears the filter's internal state
	ResetState()
	// IsEnabled returns true if the filter is currently enabled
	IsEnabled() bool // Added
	// GetResponse calculates the complex frequency response at a given point z on the unit circle
	GetResponse(z complex128) complex128 // Added
}

// BiquadState holds the state variables for a biquad filter (for one channel)
type BiquadState struct {
	X1, X2 float64 // Input history - Exported
	Y1, Y2 float64 // Output history - Exported
}

// BiquadCoeffs holds the coefficients for a biquad filter
type BiquadCoeffs struct {
	B0, B1, B2 float64 // Feedforward coefficients
	A1, A2     float64 // Feedback coefficients (A0 is implicitly 1)
}

// BaseFilter provides common fields and methods for biquad filters
type BaseFilter struct {
	Coeffs  BiquadCoeffs   // Exported
	State   [2]BiquadState // State for stereo channels (0=L, 1=R) - Exported
	Freq    float64        // Exported
	Q       float64        // Exported
	GainDB  float64        // Exported
	Enabled bool           // Exported
}

// ProcessSample applies the biquad filter algorithm - Exported
func (bf *BaseFilter) ProcessSample(input float64, channel int) float64 {
	if !bf.Enabled || channel < 0 || channel >= len(bf.State) {
		return input // Passthrough if disabled or invalid channel
	}

	s := &bf.State[channel]
	c := &bf.Coeffs

	// Direct Form I calculation
	// Avoid denormals
	output := c.B0*input + c.B1*s.X1 + c.B2*s.X2 - c.A1*s.Y1 - c.A2*s.Y2
	if math.Abs(output) < 1e-10 { // Check against a small threshold
		output = 0.0
	}

	// Update state variables
	s.X2 = s.X1
	s.X1 = input
	s.Y2 = s.Y1
	s.Y1 = output

	return output
}

// SetEnabled enables or disables the filter
func (bf *BaseFilter) SetEnabled(enable bool) {
	bf.Enabled = enable
	if !enable {
		// Reset state when disabling to avoid artifacts when re-enabled
		bf.ResetState()
	}
}

// ResetState clears the filter's internal state
func (bf *BaseFilter) ResetState() {
	bf.State[0] = BiquadState{}
	bf.State[1] = BiquadState{}
}

// IsEnabled returns true if the filter is enabled.
func (bf *BaseFilter) IsEnabled() bool { // Added implementation
	return bf.Enabled
}

// GetResponse calculates the complex frequency response for the biquad filter.
// This should be implemented by specific filter types if they don't use BaseFilter directly,
// but providing a base implementation is useful.
func (bf *BaseFilter) GetResponse(z complex128) complex128 { // Added implementation
	if !bf.Enabled {
		return complex(1.0, 0.0)
	}
	zInv := 1.0 / z
	zInv2 := zInv * zInv
	c := &bf.Coeffs
	num := complex(c.B0, 0) + complex(c.B1, 0)*zInv + complex(c.B2, 0)*zInv2
	// Denominator uses stored A1, A2 which have flipped signs compared to standard formula
	den := complex(1.0, 0.0) + complex(c.A1, 0)*zInv + complex(c.A2, 0)*zInv2
	if cmplx.Abs(den) < 1e-10 {
		// Avoid division by zero, return a large number or handle appropriately
		// Returning 0 might be safer to indicate a problem or lack of response at this point.
		return complex(0.0, 0.0)
	}
	return num / den
}

// GetParams returns the current parameters of the base filter.
// Note: Specific filter types embedding BaseFilter should override this
// if they have different parameter orders or additional parameters.
func (bf *BaseFilter) GetParams() []float64 {
	return []float64{bf.Freq, bf.Q, bf.GainDB}
}

// Common parameters
const (
	ParamFreq = iota
	ParamQ
	ParamGainDB
)

// Compile-time check to ensure BaseFilter satisfies parts of the Filter interface
// (Methods implemented directly on BaseFilter)
var _ interface {
	SetEnabled(bool)
	ResetState()
	GetParams() []float64
	IsEnabled() bool                   // Added check
	GetResponse(complex128) complex128 // Added check
} = (*BaseFilter)(nil)
