package core

import (
	"log"
	"math"

	"github.com/C0d3-5t3w/go-audioEq/internal/filters"
)

// Processor handles the core audio processing logic
type Processor struct {
	sampleRate float64
	bufferSize int
	filters    []filters.Filter // Chain of EQ filters
	masterGain float64          // Linear gain
	// Add mutex if filters are accessed/modified concurrently (e.g., by GUI)
}

// NewProcessor creates a new audio processor instance
func NewProcessor() *Processor {
	return &Processor{
		sampleRate: 44100.0,                   // Default sample rate
		bufferSize: 512,                       // Default buffer size
		filters:    make([]filters.Filter, 0), // Initialize empty filter chain
		masterGain: 1.0,                       // Default gain (0dB)
	}
}

// SetSampleRate updates the sample rate and recalculates filter coefficients
func (p *Processor) SetSampleRate(sr float64) {
	p.sampleRate = sr
	// Recalculate coefficients for all filters
	for _, f := range p.filters {
		f.SetParams(p.sampleRate) // Assuming SetParams takes sample rate
	}
}

// SetBufferSize updates the buffer size
func (p *Processor) SetBufferSize(size int) {
	p.bufferSize = size
	// Potentially resize internal buffers if needed
}

// SetMasterGain sets the overall output gain in dB
func (p *Processor) SetMasterGain(gainDB float64) {
	p.masterGain = math.Pow(10.0, gainDB/20.0) // Convert dB to linear gain
}

// AddFilter adds a new filter to the processing chain
func (p *Processor) AddFilter(f filters.Filter) {
	p.filters = append(p.filters, f)
	f.SetParams(p.sampleRate) // Initialize filter with current sample rate
}

// RemoveFilter removes a filter (implementation depends on how filters are identified)
func (p *Processor) RemoveFilter(index int) {
	if index >= 0 && index < len(p.filters) {
		p.filters = append(p.filters[:index], p.filters[index+1:]...)
	}
}

// GetFilters returns a slice of the currently active filters.
// Consider adding a mutex lock/unlock if filters can be modified concurrently.
func (p *Processor) GetFilters() []filters.Filter {
	// Return a copy or ensure thread safety if needed
	return p.filters
}

// ProcessFloat32Samples processes float32 audio samples directly
func (p *Processor) ProcessFloat32Samples(in [][]float32, out [][]float32, numSamples, numChannels int) {
	if len(out) < numChannels || len(in) < numChannels {
		log.Printf("Error: Channel count mismatch in buffer arrays (In %d, Out %d, Expected %d)",
			len(in), len(out), numChannels)
		return
	}

	for ch := 0; ch < numChannels; ch++ {
		inCh := in[ch]
		outCh := out[ch]
		if len(inCh) < numSamples || len(outCh) < numSamples {
			log.Printf("Error: Sample count mismatch in buffer channel %d (In %d, Out %d, Expected %d)",
				ch, len(inCh), len(outCh), numSamples)
			return
		}

		for i := 0; i < numSamples; i++ {
			sample := float64(inCh[i])
			for _, f := range p.filters {
				sample = f.Process(sample, ch) // Process channel 'ch'
			}
			outCh[i] = float32(sample * p.masterGain)
		}
	}
}

// ProcessFloat64Samples processes float64 audio samples directly
func (p *Processor) ProcessFloat64Samples(in [][]float64, out [][]float64, numSamples, numChannels int) {
	if len(out) < numChannels || len(in) < numChannels {
		log.Printf("Error: Channel count mismatch in buffer arrays (In %d, Out %d, Expected %d)",
			len(in), len(out), numChannels)
		return
	}

	for ch := 0; ch < numChannels; ch++ {
		inCh := in[ch]
		outCh := out[ch]
		if len(inCh) < numSamples || len(outCh) < numSamples {
			log.Printf("Error: Sample count mismatch in buffer channel %d (In %d, Out %d, Expected %d)",
				ch, len(inCh), len(outCh), numSamples)
			return
		}

		for i := 0; i < numSamples; i++ {
			sample := inCh[i]
			for _, f := range p.filters {
				sample = f.Process(sample, ch) // Process channel 'ch'
			}
			outCh[i] = sample * p.masterGain
		}
	}
}
