package core

import (
	"log"
	"math"

	"github.com/C0d3-5t3w/go-audioEq/internal/filters"
	"pipelined.dev/audio/vst2"
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

// Process applies the EQ filters to the input buffer (float32)
func (p *Processor) Process(io *vst2.IO) { // Changed signature to accept *vst2.IO
	numSamples := io.NumSamples()
	numChannels := io.NumChannels()
	// vst2.IO handles input/output buffers internally, no need for separate output arg check here

	// Get channel slices
	inChans := io.InputFloat32()   // Get input buffers
	outChans := io.OutputFloat32() // Get output buffers

	if len(outChans) != numChannels || len(inChans) != numChannels {
		log.Printf("Error: Channel count mismatch in IO buffer (In %d, Out %d, Expected %d)",
			len(inChans), len(outChans), numChannels)
		io.Clear() // Clear output if mismatch
		return
	}

	for ch := 0; ch < numChannels; ch++ {
		in := inChans[ch]
		out := outChans[ch]
		if len(in) != numSamples || len(out) != numSamples {
			log.Printf("Error: Sample count mismatch in IO buffer channel %d (In %d, Out %d, Expected %d)",
				ch, len(in), len(out), numSamples)
			// Optionally clear just this channel or the whole buffer
			io.Clear()
			return // Stop processing on error
		}
		for i := 0; i < numSamples; i++ {
			sample := float64(in[i])
			for _, f := range p.filters {
				sample = f.Process(sample, ch) // Process channel 'ch'
			}
			out[i] = float32(sample * p.masterGain)
		}
	}
}

// ProcessFloat64 applies the EQ filters to the input buffer (float64)
func (p *Processor) ProcessFloat64(io *vst2.IO) { // Changed signature to accept *vst2.IO
	numSamples := io.NumSamples()
	numChannels := io.NumChannels()

	// Get channel slices
	inChans := io.InputFloat64()   // Get input buffers
	outChans := io.OutputFloat64() // Get output buffers

	if len(outChans) != numChannels || len(inChans) != numChannels {
		log.Printf("Error: Channel count mismatch in IO buffer (In %d, Out %d, Expected %d)",
			len(inChans), len(outChans), numChannels)
		io.Clear()
		return
	}

	for ch := 0; ch < numChannels; ch++ {
		in := inChans[ch]
		out := outChans[ch]
		if len(in) != numSamples || len(out) != numSamples {
			log.Printf("Error: Sample count mismatch in IO buffer channel %d (In %d, Out %d, Expected %d)",
				ch, len(in), len(out), numSamples)
			io.Clear()
			return
		}
		for i := 0; i < numSamples; i++ {
			sample := in[i] // Already float64
			for _, f := range p.filters {
				sample = f.Process(sample, ch) // Process channel 'ch'
			}
			out[i] = sample * p.masterGain
		}
	}
}
