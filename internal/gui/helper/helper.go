package helper

// This file can contain utility functions for the GUI, if needed.
// For example:
// - Functions to convert between parameter values (0.0-1.0) and display values (Hz, dB, Q).
// - Custom widget implementations.
// - Color scheme definitions.

import (
	"math"
)

// LogScale converts a linear value (0-1) to a logarithmic frequency
func LogScale(value, minFreq, maxFreq float64) float64 {
	if value <= 0 {
		return minFreq
	}
	if value >= 1 {
		return maxFreq
	}
	minLog := math.Log10(minFreq)
	maxLog := math.Log10(maxFreq)
	logFreq := minLog + value*(maxLog-minLog)
	return math.Pow(10, logFreq)
}

// InverseLogScale converts a logarithmic frequency back to a linear value (0-1)
func InverseLogScale(freq, minFreq, maxFreq float64) float64 {
	if freq <= minFreq {
		return 0.0
	}
	if freq >= maxFreq {
		return 1.0
	}
	minLog := math.Log10(minFreq)
	maxLog := math.Log10(maxFreq)
	logFreq := math.Log10(freq)
	return (logFreq - minLog) / (maxLog - minLog)
}

// GainToSliderValue converts dB gain to a slider value (0-1, assuming -24dB to +24dB range)
func GainToSliderValue(gainDB, minDB, maxDB float64) float64 {
	if minDB >= maxDB {
		return 0.5
	} // Avoid division by zero
	val := (gainDB - minDB) / (maxDB - minDB)
	return math.Max(0.0, math.Min(1.0, val)) // Clamp to 0-1
}

// SliderValueToGain converts a slider value (0-1) to dB gain
func SliderValueToGain(value, minDB, maxDB float64) float64 {
	return minDB + value*(maxDB-minDB)
}

// QToSliderValue converts Q factor to a slider value (0-1, example mapping)
func QToSliderValue(q, minQ, maxQ float64) float64 {
	// Q often benefits from a non-linear mapping for better control at low values
	// Example: Logarithmic mapping for Q
	if q <= minQ {
		return 0.0
	}
	if q >= maxQ {
		return 1.0
	}
	minLogQ := math.Log10(minQ)
	maxLogQ := math.Log10(maxQ)
	logQ := math.Log10(q)
	val := (logQ - minLogQ) / (maxLogQ - minLogQ)
	return math.Max(0.0, math.Min(1.0, val))
}

// SliderValueToQ converts slider value back to Q factor
func SliderValueToQ(value, minQ, maxQ float64) float64 {
	minLogQ := math.Log10(minQ)
	maxLogQ := math.Log10(maxQ)
	logQ := minLogQ + value*(maxLogQ-minLogQ)
	return math.Pow(10, logQ)
}
