package main

import (
	"fmt"
	"log"
	"math"
	"unsafe"

	"github.com/C0d3-5t3w/go-audioEq/internal/core"    // Needed for FilterProvider
	"github.com/C0d3-5t3w/go-audioEq/internal/filters" // Import filters for GetFilters return type
	"github.com/C0d3-5t3w/go-audioEq/internal/gui"
	"pipelined.dev/audio/vst2"
)

// Define parameter indices
const (
	// Start with one example parameter (e.g., Master Gain)
	MasterGain = iota
	// Add more parameters here as needed (e.g., for bands)
	NumParams = 1 // Update this count as parameters are added
)

// Plugin is the main VST plugin structure
type Plugin struct {
	vst2.Plugin // Embed vst2.Plugin to get default implementations
	host        vst2.Host
	core        *core.Processor
	gui         *gui.GUI
	// params      []*vst2.ParameterProperties // ParameterProperties is empty, store details differently
	paramValues []float32               // Current parameter values (0.0-1.0)
	paramDefs   map[int]paramDefinition // Store parameter definitions locally
}

// Local struct to hold parameter details since vst2.ParameterProperties is empty
type paramDefinition struct {
	Name      string
	UnitLabel string
	// Add Min/Max/Steps if needed for display/conversion logic
}

// strToInt32 converts a 4-character string to a VST2 unique ID (int32).
func strToInt32(str string) int32 {
	if len(str) != 4 {
		panic("VST2 Unique ID must be 4 characters long")
	}
	b := []byte(str)
	return int32(b[0])<<24 | int32(b[1])<<16 | int32(b[2])<<8 | int32(b[3])
}

// Allocator returns a new plugin instance - conforms to vst2.AllocatorFunc
// The second return value should also be vst2.Plugin (which Plugin implements via embedding and Dispatch)
func Allocator(host vst2.Host) (vst2.Plugin, vst2.Plugin) {
	p := &Plugin{
		host:        host,
		core:        core.NewProcessor(),
		paramValues: make([]float32, NumParams),
		paramDefs:   make(map[int]paramDefinition),
	}

	// Initialize parameter definitions
	p.paramDefs[MasterGain] = paramDefinition{
		Name:      "Master",
		UnitLabel: "dB",
	}
	// Add other definitions...

	// Set initial default value (0dB -> 0.5 normalized)
	initialMasterGainValue := helper.GainToSliderValue(0.0, MinGainDB, MaxGainDB)
	// Call SetParameter directly to initialize core and value
	p.SetParameter(MasterGain, float32(initialMasterGainValue))

	// Initialize GUI (passing parameter update callback and filter provider)
	// Need to adjust the callback signature if necessary, or wrap it.
	p.gui = gui.NewGUI(p.updateParameterFromGUI, func() []filters.Filter { return p.core.GetFilters() }) // Pass core.GetFilters via closure

	// Return p which implements vst2.Plugin interface (via embedding) and the dispatcher
	return p, p // Return p for both plugin and dispatcher roles
}

// --- vst2.Plugin Implementation (Overrides embedded defaults) ---

// Info returns the plugin's static properties.
func (p *Plugin) Info() vst2.Info { // Changed return type to vst2.Info
	return vst2.Info{
		Name:           "GoAudioEQ",
		Vendor:         "C0d3-5t3w",
		Category:       vst2.PluginCategoryEffect,
		UniqueID:       strToInt32("GEQ1"), // Use helper
		Version:        1000,
		InputChannels:  2,
		OutputChannels: 2,
		Parameters:     NumParams, // Set NumParams here
	}
}

// Process audio samples (float32).
func (p *Plugin) Process(io *vst2.IO) { // Signature matches core.Process
	p.core.Process(io) // Pass the IO struct directly
}

// ProcessDouble audio samples (float64).
func (p *Plugin) ProcessDouble(io *vst2.IO) { // Signature matches core.ProcessFloat64
	p.core.ProcessFloat64(io) // Pass the IO struct directly
}

// --- Dispatcher Methods (Handle VST opcodes) ---

// Dispatch handles incoming VST opcodes
func (p *Plugin) Dispatch(op vst2.Opcode, index int32, value int64, ptr unsafe.Pointer, opt float32) uintptr {
	switch op {
	case vst2.OpSetSampleRate:
		p.SetSampleRate(float64(opt))
	case vst2.OpSetBufferSize:
		p.SetBufferSize(int(value))
	case vst2.OpProcessEvents:
		// Handle MIDI events if needed: p.ProcessEvents(ptr)
	case vst2.OpGetParamName:
		return p.getParameterName(int(index))
	case vst2.OpGetParamDisplay:
		return p.getParameterDisplay(int(index))
	case vst2.OpGetParamLabel:
		return p.getParameterLabel(int(index))
	case vst2.OpSetParam:
		p.SetParameter(int(index), opt)
	case vst2.OpGetParam:
		return uintptr(math.Float32bits(p.ParameterValue(int(index))))
	case vst2.OpCanDo:
		return p.canDo(vst2.CanDoString(ptr))
	case vst2.OpGetPlugCategory:
		return uintptr(vst2.PluginCategoryEffect)
	case vst2.OpEditOpen:
		p.OpenEditor(ptr) // Pass the HWND pointer
	case vst2.OpEditClose:
		p.CloseEditor()
	case vst2.OpEditIdle:
		// Optional: Perform idle tasks for the editor
	case vst2.OpShutdown:
		p.Shutdown()
	// Add other opcodes as needed (Resume, Suspend, GetChunk, SetChunk, etc.)
	default:
		log.Printf("Plugin Dispatch: Received unknown opcode %v\n", op)
	}
	return 0
}

// --- Parameter Handling Methods ---

// getParameterName returns the name for a parameter index.
func (p *Plugin) getParameterName(index int) uintptr {
	if def, ok := p.paramDefs[index]; ok {
		return vst2.StringPtr(def.Name)
	}
	return vst2.StringPtr("")
}

// getParameterDisplay returns the formatted string representation of the current parameter value.
func (p *Plugin) getParameterDisplay(index int) uintptr {
	if index < 0 || index >= len(p.paramValues) {
		return vst2.StringPtr("")
	}
	value := p.paramValues[index]

	var display string
	switch index {
	case MasterGain:
		gainDB := helper.SliderValueToGain(float64(value), MinGainDB, MaxGainDB)
		display = fmt.Sprintf("%.1f", gainDB)
	// Add cases for other EQ parameters
	default:
		display = fmt.Sprintf("%.2f", value)
	}
	return vst2.StringPtr(display)
}

// getParameterLabel returns the unit label for a parameter index.
func (p *Plugin) getParameterLabel(index int) uintptr {
	if def, ok := p.paramDefs[index]; ok {
		return vst2.StringPtr(def.UnitLabel)
	}
	return vst2.StringPtr("")
}

// SetParameter updates a parameter value (called by host via dispatcher).
func (p *Plugin) SetParameter(index int, value float32) {
	if index < 0 || index >= len(p.paramValues) {
		return
	}
	// Clamp value just in case
	value = float32(math.Max(0.0, math.Min(1.0, float64(value))))
	if p.paramValues[index] == value {
		return // No change
	}
	p.paramValues[index] = value

	// Update core processor based on parameter change
	switch index {
	case MasterGain:
		gainDB := helper.SliderValueToGain(float64(value), MinGainDB, MaxGainDB)
		p.core.SetMasterGain(gainDB)
		// Add cases for other EQ parameters (Freq, Q, Gain for each band)
	}

	// Update GUI if it's open
	if p.gui != nil {
		// Consider running GUI updates on a separate thread or using a channel
		// to avoid blocking the audio thread if GUI operations are slow.
		p.gui.UpdateParameter(index, value)
	}
}

// ParameterValue returns the current value of a parameter.
func (p *Plugin) ParameterValue(index int) float32 {
	if index < 0 || index >= len(p.paramValues) {
		return 0.0
	}
	return p.paramValues[index]
}

// --- Other Plugin Methods ---

// canDo checks plugin capabilities.
func (p *Plugin) canDo(canDo vst2.CanDoString) uintptr {
	switch canDo {
	case vst2.CanDoReceiveEvents, vst2.CanDoReceiveMidiEvent, vst2.CanDoProcessReplacing:
		return 1 // Example: supports MIDI events and replacing process
	case vst2.CanDoPlugAsChannelInsert, vst2.CanDoPlugAsSend:
		return 1 // Example: can be used as insert or send effect
	case vst2.CanDoProcessFloat64Replacing:
		return 1 // Supports double precision processing
	default:
		log.Printf("Plugin CanDo: Received unknown string %v\n", canDo)
		return 0
	}
}

// SetBufferSize updates the buffer size.
func (p *Plugin) SetBufferSize(size int) {
	p.core.SetBufferSize(size)
}

// SetSampleRate updates the sample rate.
func (p *Plugin) SetSampleRate(sampleRate float64) {
	p.core.SetSampleRate(sampleRate)
}

// OpenEditor opens the plugin GUI.
func (p *Plugin) OpenEditor(hwnd unsafe.Pointer) {
	if p.gui != nil {
		// Run GUI opening in a separate goroutine to avoid blocking host
		go func() {
			p.gui.Open(hwnd)
			// After opening, maybe update all parameters in the GUI
			for i, v := range p.paramValues {
				p.gui.UpdateParameter(i, v)
			}
		}()
	}
}

// CloseEditor closes the plugin GUI.
func (p *Plugin) CloseEditor() {
	if p.gui != nil {
		p.gui.Close()
	}
}

// Shutdown performs cleanup when the plugin is unloaded.
func (p *Plugin) Shutdown() {
	log.Println("Shutting down GoAudioEQ plugin...")
	p.CloseEditor() // Ensure GUI is closed
	// Add any other cleanup needed
}

// --- End Plugin Implementation ---

// updateParameterFromGUI is called by the GUI when a control changes
func (p *Plugin) updateParameterFromGUI(index int, value float32) {
	// Set the parameter value internally (this also updates the core)
	p.SetParameter(index, value)

	// Notify the host about the parameter change
	// Call the host callback function directly with the Automate opcode.
	if p.host != nil { // Check if the host callback function is assigned
		p.host(vst2.HostAutomate, int32(index), 0, nil, value) // Correctly call the host function
	}
}

// main is the entry point for the VST plugin binary.
func main() {
	// Use vst2.PluginMain instead of vst2.Serve
	vst2.PluginMain(Allocator)
}

// --- Helper types/constants from GUI needed for parameter mapping ---
// These should ideally be defined in a shared place or passed during init.
const (
	MaxGainDB = 24.0
	MinGainDB = -24.0
	// Add Min/Max Freq/Q if needed for parameter display/conversion here
)

// Re-declare helper package locally or import it properly
// Import the actual helper package instead of redeclaring
// import "github.com/C0d3-5t3w/go-audioEq/internal/gui/helper"
// For now, using the local struct alias:
var helper = guiHelper{} // Use a local alias or import

type guiHelper struct{}

func (guiHelper) GainToSliderValue(gainDB, minDB, maxDB float64) float64 {
	if minDB >= maxDB {
		return 0.5
	}
	val := (gainDB - minDB) / (maxDB - minDB)
	return math.Max(0.0, math.Min(1.0, val))
}

func (guiHelper) SliderValueToGain(value, minDB, maxDB float64) float64 {
	return minDB + value*(maxDB-minDB)
}

// Add other necessary helpers (LogScale, QToSliderValue etc.) if needed here
// or ensure proper import and usage from the actual helper package.
