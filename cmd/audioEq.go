package main

/*
#include <stdlib.h>
*/
import "C"

import (
	"log"
	"math"
	"unsafe"

	"github.com/C0d3-5t3w/go-audioEq/internal/core"
	"github.com/C0d3-5t3w/go-audioEq/internal/filters"
	"github.com/C0d3-5t3w/go-audioEq/internal/gui"
	"github.com/C0d3-5t3w/go-audioEq/internal/gui/helper"
	"pipelined.dev/audio/vst2"
)

const (
	MasterGain = iota
	NumParams  = 1
)

// Plugin implements the VST2 plugin interface
type Plugin struct {
	host        vst2.HostCallbackFunc
	core        *core.Processor
	gui         *gui.GUI
	paramValues []float32
	paramDefs   map[int]string
}

func Allocator(host vst2.HostCallbackFunc) (vst2.Plugin, vst2.Plugin) {
	p := &Plugin{
		host:        host,
		core:        core.NewProcessor(),
		paramValues: make([]float32, NumParams),
		paramDefs:   map[int]string{MasterGain: "Master dB"},
	}
	p.SetParameter(MasterGain, 0.5)
	p.gui = gui.NewGUI(p.updateFromGUI, func() []filters.Filter { return p.core.GetFilters() })
	return p, p
}

// Info returns plugin information
func (p *Plugin) Info() (name, vendor string, version int32) {
	return "GoAudioEQ", "C0d3-5t3w", 1000
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
	for i := 0; i < chans; i++ {
		for j := int32(0); j < frames; j++ {
			out[i][j] = in[i][j]
		}
	}
	// TODO: Implement actual processing using p.core.Process
}

// ProcessDoubleReplacing processes audio with float64 samples
func (p *Plugin) ProcessDoubleReplacing(in [][]float64, out [][]float64, frames int32) {
	chans := min(len(in), len(out))
	for i := 0; i < chans; i++ {
		for j := int32(0); j < frames; j++ {
			out[i][j] = in[i][j]
		}
	}
	// TODO: Implement actual processing using p.core.ProcessFloat64
}

// Dispatch handles plugin opcodes
func (p *Plugin) Dispatch(opcode vst2.PluginOpcode, index int32, value int64, ptr unsafe.Pointer, opt float32) int64 {
	switch opcode {
	// Add needed opcode handlers
	}
	return 0
}

// helpers for parameters:
func (p *Plugin) getParamName(idx int32) uintptr {
	s := p.paramDefs[int(idx)]
	return uintptr(unsafe.Pointer(C.CString(s)))
}

// similar getParamDisplay, getParamLabel...

func (p *Plugin) updateFromGUI(idx int, v float32) {
	p.SetParameter(int(idx), v)
	// host callback with correct parameter type:
	p.host(vst2.HostOpcode(10), int32(idx), 0, nil, float32(v)) // 10 is equivalent to EffEditSetParameter
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
	gainDB := helper.SliderValueToGain(float64(value), -24.0, 24.0) // Example range
	p.core.SetMasterGain(gainDB)

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
	case "bypass", "receiveVstEvents", "receiveVstMidiEvent":
		return false
	default:
		return false
	}
}

// NumParameters returns the number of parameters
func (p *Plugin) NumParameters() int {
	return NumParams
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func main() {
	log.Println("Starting VST plugin")
	vst2.Main(Allocator)
}

func strToID(s string) int32 {
	b := []byte(s)
	if len(b) < 4 {
		b = append(b, make([]byte, 4-len(b))...)
	}
	return int32(b[0])<<24 | int32(b[1])<<16 | int32(b[2])<<8 | int32(b[3])
}
