package main

/*
//  Do not change this CGO preamble - this specific format is required for VST entry point
#include <stdlib.h>
#include <stdint.h>

// Define hostCallback type for CGO
typedef intptr_t (*hostCallback)(intptr_t, intptr_t, intptr_t, void*, float);

// Helper function to call the hostCallback function pointer
static inline intptr_t bridge_host_callback(hostCallback cb,
                                        intptr_t opcode,
                                        intptr_t index,
                                        intptr_t value,
                                        void* ptr,
                                        float opt) {
    return cb(opcode, index, value, ptr, opt);
}

void* VSTPluginMain(hostCallback);
*/
import "C"

import (
	"log"
	"math"
	"os"
	"path/filepath"
	"unsafe"

	"github.com/C0d3-5t3w/go-audioEq/internal/config"
	"github.com/C0d3-5t3w/go-audioEq/internal/core"
	"github.com/C0d3-5t3w/go-audioEq/internal/filters"
	"github.com/C0d3-5t3w/go-audioEq/internal/filters/peak"
	"github.com/C0d3-5t3w/go-audioEq/internal/filters/shelf"
	"github.com/C0d3-5t3w/go-audioEq/internal/gui"
	"github.com/C0d3-5t3w/go-audioEq/internal/gui/helper"
	"github.com/C0d3-5t3w/go-audioEq/internal/presets"
	"pipelined.dev/audio/vst2"
)

const (
	MasterGain = iota
	NumParams  = 1 // Just master gain for now, will expand with band controls
)

// Plugin implements the VST2 plugin interface
type Plugin struct {
	host         vst2.HostCallbackFunc
	core         *core.Processor
	gui          *gui.GUI
	paramValues  []float32
	paramDefs    map[int]string
	sampleRate   float64
	presetMgr    *presets.Manager
	configLoaded bool
}

var plugin *Plugin
var logFile *os.File

// Initialize logger for debugging
func init() {
	// Create log directory if it doesn't exist
	logDir := filepath.Join(os.TempDir(), "GoAudioEQ")
	os.MkdirAll(logDir, 0755)

	// Open log file
	var err error
	logFile, err = os.OpenFile(filepath.Join(logDir, "vst_debug.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err == nil {
		log.SetOutput(logFile)
		log.Println("GoAudioEQ plugin initializing")
	}

	// Initialize config
	_, configErr := config.LoadConfig("pkg/config.yaml")
	if configErr != nil {
		log.Printf("Failed to load config: %v", configErr)
	}
}

//export VSTPluginMain
func VSTPluginMain(cb C.hostCallback) unsafe.Pointer {
	log.Println("VSTPluginMain called")
	// Convert C callback to Go callback
	hostCallback := func(opcode vst2.HostOpcode, index int32, value int64, ptr unsafe.Pointer, opt float32) int64 {
		return int64(C.bridge_host_callback(cb,
			C.intptr_t(opcode),
			C.intptr_t(index),
			C.intptr_t(value),
			ptr,
			C.float(opt),
		))
	}

	// Create plugin instance if it doesn't exist
	if plugin == nil {
		processor := core.NewProcessor()
		plugin = &Plugin{
			host:         hostCallback,
			core:         processor,
			paramValues:  make([]float32, NumParams),
			paramDefs:    map[int]string{MasterGain: "Master Gain"},
			sampleRate:   44100.0, // Default, will be updated by the host
			configLoaded: false,
		}

		// Initialize with default parameter values
		plugin.SetParameter(MasterGain, 0.5) // 0.5 is 0dB (unity gain)

		// Add some default filters
		plugin.core.AddFilter(shelf.NewLowShelfFilter(100, 0.707, 0.0))
		plugin.core.AddFilter(peak.NewPeakFilter(1000, 1.0, 0.0))
		plugin.core.AddFilter(shelf.NewHighShelfFilter(5000, 0.707, 0.0))

		// Initialize GUI
		plugin.gui = gui.NewGUI(plugin.updateFromGUI, func() []filters.Filter {
			return plugin.core.GetFilters()
		})

		// Load presets if available
		plugin.presetMgr, _ = presets.NewManager("pkg/presets.json")
	}

	// Return the memory address as an unsafe pointer to be used by the VST host
	return unsafe.Pointer(plugin)
}

// Info returns plugin information
func (p *Plugin) Info() (name, vendor string, version int32) {
	return "GoAudioEQ", "C0d3-5t3w", 1000 // Version 1.0.0.0
}

// Category returns the plugin category
func (p *Plugin) Category() vst2.PluginCategory {
	return vst2.PluginCategoryEffect
}

// UniqueID returns the plugin unique ID
func (p *Plugin) UniqueID() int32 {
	return strToID("GEQ1")
}

// InputChannels returns the number of input channels
func (p *Plugin) InputChannels() int {
	return 2
}

// OutputChannels returns the number of output channels
func (p *Plugin) OutputChannels() int {
	return 2
}

// ProcessReplacing processes audio with float32 samples
func (p *Plugin) ProcessReplacing(in [][]float32, out [][]float32, frames int32) {
	chans := min(len(in), len(out))
	// Integrate with our core processor
	p.core.ProcessFloat32Samples(in, out, int(frames), chans)
}

// ProcessDoubleReplacing processes audio with float64 samples
func (p *Plugin) ProcessDoubleReplacing(in [][]float64, out [][]float64, frames int32) {
	chans := min(len(in), len(out))
	// Integrate with our core processor
	p.core.ProcessFloat64Samples(in, out, int(frames), chans)
}

// Dispatch handles plugin opcodes
func (p *Plugin) Dispatch(opcode vst2.PluginOpcode, index int32, value int64, ptr unsafe.Pointer, opt float32) int64 {
	switch opcode {
	case vst2.PlugGetPlugCategory:
		return int64(p.Category())

	case vst2.PlugGetVstVersion:
		return 2400 // VST 2.4

	case vst2.PlugOpen:
		log.Println("Plugin opened")
		return 0

	case vst2.PlugClose:
		log.Println("Plugin closed")
		return 0

	case vst2.PlugSetSampleRate:
		p.sampleRate = float64(opt)
		p.core.SetSampleRate(p.sampleRate)
		log.Printf("Sample rate set to %.1f Hz", p.sampleRate)
		return 0

	case vst2.PlugSetBlockSize:
		p.core.SetBufferSize(int(value))
		log.Printf("Block size set to %d samples", value)
		return 0

	case vst2.PlugMainsChanged:
		if value == 0 {
			// Plugin is being suspended
			log.Println("Plugin suspended")
		} else {
			// Plugin is being resumed
			log.Println("Plugin resumed")
		}
		return 0

	case vst2.PlugEditOpen:
		log.Println("Editor opening")
		p.gui.Open(ptr)
		return 0

	case vst2.PlugEditClose:
		log.Println("Editor closing")
		p.gui.Close()
		return 0

	case vst2.PlugEditIdle:
		// Called regularly when the editor is open
		return 0

	case vst2.PlugEditTop:
		// Bring editor window to the front
		return 0
	}

	return 0
}

// Parameter handling functions
func (p *Plugin) updateFromGUI(idx int, v float32) {
	p.SetParameter(int(idx), v)
	// Host callback with parameter change notification
	p.host(vst2.AudioMasterAutomate, int32(idx), 0, nil, v)
}

// SetParameter implements parameter value changes
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
		gainDB := helper.SliderValueToGain(float64(value), -24.0, 24.0)
		p.core.SetMasterGain(gainDB)
		log.Printf("Master gain set to %.1f dB", gainDB)
	}

	// Update GUI if it's open
	if p.gui != nil {
		p.gui.UpdateParameter(index, value)
	}
}

// ParameterValue gets a parameter's current value
func (p *Plugin) ParameterValue(index int) float32 {
	if index < 0 || index >= len(p.paramValues) {
		return 0.0
	}
	return p.paramValues[index]
}

// CanDo reports plugin capabilities
func (p *Plugin) CanDo(feature string) bool {
	switch feature {
	case "bypass", "sendVstEvents", "sendVstMidiEvent":
		return true
	default:
		return false
	}
}

// NumParameters returns the number of parameters
func (p *Plugin) NumParameters() int {
	return NumParams
}

// Helper functions
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func main() {
	log.Println("Starting GoAudioEQ VST plugin")
	// This is just for the build process - actual VST init happens through the exported function
}

func strToID(s string) int32 {
	b := []byte(s)
	if len(b) < 4 {
		b = append(b, make([]byte, 4-len(b))...)
	}
	return int32(b[0])<<24 | int32(b[1])<<16 | int32(b[2])<<8 | int32(b[3])
}
