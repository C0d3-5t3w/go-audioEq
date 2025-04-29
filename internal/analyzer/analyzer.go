package analyzer

import (
	"log" // Added for logging
	"math"
	"math/cmplx"

	"github.com/C0d3-5t3w/go-audioEq/internal/filters"
	// Consider adding an FFT library dependency if needed, e.g., "github.com/mjibson/go-dsp/fft"
	// Or implement a basic DFT/FFT here.
)

// Analyzer calculates the frequency response of the filter chain
type Analyzer struct {
	sampleRate float64
	fftSize    int
	minFreq    float64
	maxFreq    float64
}

// NewAnalyzer creates a frequency response analyzer
func NewAnalyzer(sampleRate float64, fftSize int, minFreq, maxFreq float64) *Analyzer {
	// Ensure minFreq is positive
	if minFreq <= 0 {
		minFreq = 1.0 // Set a small positive minimum frequency
		log.Printf("Warning: Analyzer minFrequency was <= 0, adjusted to %f Hz", minFreq)
	}
	// Ensure maxFreq is greater than minFreq and less than Nyquist
	nyquist := sampleRate / 2.0
	if maxFreq <= minFreq {
		maxFreq = nyquist
		log.Printf("Warning: Analyzer maxFrequency was <= minFrequency, adjusted to Nyquist (%f Hz)", maxFreq)
	} else if maxFreq > nyquist {
		maxFreq = nyquist
		log.Printf("Warning: Analyzer maxFrequency was > Nyquist, adjusted to Nyquist (%f Hz)", maxFreq)
	}

	return &Analyzer{
		sampleRate: sampleRate,
		fftSize:    fftSize,
		minFreq:    minFreq,
		maxFreq:    maxFreq,
	}
}

// CalculateResponse computes the magnitude response of a series of filters
// Returns frequencies and corresponding magnitudes (in dB)
func (a *Analyzer) CalculateResponse(filterChain []filters.Filter, numPoints int) ([]float64, []float64) {
	if numPoints <= 1 {
		return []float64{}, []float64{} // Need at least 2 points
	}
	if a.minFreq <= 0 || a.maxFreq <= a.minFreq {
		log.Printf("Error: Invalid frequency range for analyzer [%f, %f]", a.minFreq, a.maxFreq)
		return []float64{}, []float64{}
	}

	frequencies := make([]float64, numPoints)
	magnitudesDB := make([]float64, numPoints)

	// Generate logarithmically spaced frequencies for plotting
	minLogFreq := math.Log10(a.minFreq)
	maxLogFreq := math.Log10(a.maxFreq)

	for i := 0; i < numPoints; i++ {
		// Avoid issues at the exact endpoints with division by zero if numPoints=1
		var logFreq float64
		if numPoints == 1 {
			logFreq = minLogFreq // Or average, doesn't matter much for 1 point
		} else {
			logFreq = minLogFreq + (maxLogFreq-minLogFreq)*float64(i)/float64(numPoints-1)
		}
		freq := math.Pow(10, logFreq)
		frequencies[i] = freq

		// Calculate the complex response H(z) for this frequency
		// z = e^(j * 2 * pi * freq / sampleRate)
		omega := 2.0 * math.Pi * freq / a.sampleRate
		z := cmplx.Exp(complex(0, omega)) // z = e^(j*omega)

		totalResponse := complex(1.0, 0.0) // Start with unity gain

		for _, filter := range filterChain {
			// Use interface methods
			if !filter.IsEnabled() {
				continue // Skip disabled filters
			}

			// Calculate response for this filter using its GetResponse method
			response := filter.GetResponse(z)
			totalResponse *= response
		}

		// Convert magnitude to dB
		magnitude := cmplx.Abs(totalResponse)
		if magnitude < 1e-10 { // Avoid log(0)
			magnitudesDB[i] = -200.0 // Represents very low dB
		} else {
			magnitudesDB[i] = 20.0 * math.Log10(magnitude)
		}
	}

	return frequencies, magnitudesDB
}

// MinFreq returns the minimum frequency for analysis.
func (a *Analyzer) MinFreq() float64 {
	return a.minFreq
}

// MaxFreq returns the maximum frequency for analysis.
func (a *Analyzer) MaxFreq() float64 {
	return a.maxFreq
}

// TODO: Add GetResponse method to BaseFilter or Filter interface
// func (bf *BaseFilter) GetResponse(z complex128) complex128 { ... } // Done in filters.go
