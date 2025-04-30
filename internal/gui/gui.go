package gui

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"sync"
	"time"
	"unsafe"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/C0d3-5t3w/go-audioEq/internal/analyzer"
	"github.com/C0d3-5t3w/go-audioEq/internal/config"
	"github.com/C0d3-5t3w/go-audioEq/internal/core/render"
	"github.com/C0d3-5t3w/go-audioEq/internal/filters" // Assuming Filter interface is needed
	"github.com/C0d3-5t3w/go-audioEq/internal/gui/helper"
)

// ParameterCallback defines the function signature for notifying the plugin about GUI changes.
type ParameterCallback func(index int, value float32)

// Placeholder parameter indices - these should be defined properly in cmd/audioEq.go
const (
	MasterGain = 0
	// Example band parameter offsets (assuming 8 bands, 5 params per band)
	Band1Type    = 1
	Band1Freq    = 2
	Band1Q       = 3
	Band1Gain    = 4
	Band1Enabled = 5
	// ... Band2Type = 6 etc.
	NumBands       = 8 // Example: 8 bands
	ParamsPerBand  = 5 // Type, Freq, Q, Gain, Enabled
	MaxFreq        = 20000.0
	MinFreq        = 20.0
	MaxQ           = 18.0
	MinQ           = 0.1
	MaxGainDB      = 24.0
	MinGainDB      = -24.0
	PlotUpdateRate = 30 // Hz
)

// FilterProvider defines a function type to get the current filter chain
type FilterProvider func() []filters.Filter

// GUI manages the plugin's graphical user interface
type GUI struct {
	app            fyne.App
	window         fyne.Window
	paramCb        ParameterCallback
	filterProvider FilterProvider // Function to get filters from core

	masterGain      *widget.Slider
	masterGainLabel *widget.Label // Label to display master gain value

	// Band Controls (Example for 8 bands)
	bandControls []*BandControlWidgets

	// Plotting
	plotArea     *canvas.Raster
	plotRenderer *render.PlotRenderer
	analyzer     *analyzer.Analyzer
	plotTicker   *time.Ticker
	plotMutex    sync.RWMutex // To protect plot data access

	isVisible bool // Add a manual visibility state

	// Add other controls: band selectors, freq/q/gain sliders/knobs, plot area
}

// BandControlWidgets holds the widgets for a single EQ band
type BandControlWidgets struct {
	container *fyne.Container
	// typeSelect *widget.Select // TODO: Implement type selection
	freqSlider  *widget.Slider
	qSlider     *widget.Slider
	gainSlider  *widget.Slider
	enableCheck *widget.Check
	freqLabel   *widget.Label
	qLabel      *widget.Label
	gainLabel   *widget.Label
}

// NewGUI creates a new GUI instance
func NewGUI(paramCb ParameterCallback, filterProvider FilterProvider) *GUI {
	// Load config for GUI/Analyzer settings
	cfg := config.GetConfig() // Assuming config is loaded elsewhere

	// Create Analyzer
	// TODO: Get sample rate dynamically if possible, or use a default/config value
	analyzerInstance := analyzer.NewAnalyzer(44100.0, cfg.Analyzer.FFTSize, cfg.Analyzer.MinFrequency, cfg.Analyzer.MaxFrequency)

	g := &GUI{
		paramCb:        paramCb,
		filterProvider: filterProvider,
		analyzer:       analyzerInstance,
		bandControls:   make([]*BandControlWidgets, NumBands),
	}
	return g
}

// Open creates and shows the GUI window.
func (g *GUI) Open(hwnd unsafe.Pointer) {
	// Check if window already exists
	if g.window != nil {
		g.window.Show()
		return
	}

	// Use existing Fyne app instance or create one
	// Note: Managing the Fyne app lifecycle within a VST can be complex.
	// This basic implementation assumes a single app instance.
	// Consider using fyne.CurrentApp() if available or managing the app instance carefully.
	if g.app == nil {
		// Set environment variable to prevent Fyne from exiting when window closes
		// os.Setenv("FYNE_NON_NATIVE_CLOSE", "1") // May be needed depending on Fyne version/behavior
		g.app = app.New()
	}

	// Create the main window
	g.window = g.app.NewWindow("Go Audio EQ")

	// --- Create GUI Elements ---

	// Master Gain
	g.masterGain = widget.NewSlider(0.0, 1.0)
	g.masterGain.Step = 0.01
	g.masterGain.SetValue(0.5)                    // Default value (0dB)
	g.masterGainLabel = widget.NewLabel("0.0 dB") // Initial label value
	g.masterGain.OnChanged = func(val float64) {
		if g.paramCb != nil {
			// Update display immediately
			gainDB := helper.SliderValueToGain(val, MinGainDB, MaxGainDB) // Assuming same range for master
			g.masterGainLabel.SetText(fmt.Sprintf("%.1f dB", gainDB))     // Update label
			g.paramCb(MasterGain, float32(val))
		}
	}
	gainLabel := widget.NewLabel("Master:")

	// Frequency Response Plot Area
	g.plotArea = canvas.NewRaster(g.drawPlot)
	g.plotArea.SetMinSize(fyne.NewSize(400, 150)) // Adjust size as needed

	// Create Plot Renderer
	// TODO: Get min/max dB range from config or constants
	g.plotRenderer = render.NewPlotRenderer(g.analyzer, g.plotArea, MinGainDB-6, MaxGainDB+6) // Slightly larger range for plot

	// Band Controls Container
	bandContainer := container.New(layout.NewGridLayout(NumBands)) // Grid layout for bands

	for i := 0; i < NumBands; i++ {
		bandWidgets := &BandControlWidgets{}
		bandIndex := i // Capture loop variable for closures

		// Parameter indices for this band
		// TODO: Define these robustly in cmd/audioEq.go
		pIdxFreq := Band1Freq + bandIndex*ParamsPerBand
		pIdxQ := Band1Q + bandIndex*ParamsPerBand
		pIdxGain := Band1Gain + bandIndex*ParamsPerBand
		// pIdxEnabled := Band1Enabled + bandIndex*ParamsPerBand
		// pIdxType := Band1Type + bandIndex*ParamsPerBand

		// --- Frequency Slider ---
		bandWidgets.freqLabel = widget.NewLabel("F: 1kHz") // Default display
		bandWidgets.freqSlider = widget.NewSlider(0.0, 1.0)
		bandWidgets.freqSlider.Step = 0.001                                             // Finer steps for frequency
		bandWidgets.freqSlider.SetValue(helper.InverseLogScale(1000, MinFreq, MaxFreq)) // Default 1kHz
		bandWidgets.freqSlider.OnChanged = func(val float64) {
			freq := helper.LogScale(val, MinFreq, MaxFreq)
			bandWidgets.freqLabel.SetText(fmt.Sprintf("F: %.0fHz", freq))
			if g.paramCb != nil {
				g.paramCb(pIdxFreq, float32(val))
			}
		}

		// --- Q Slider ---
		bandWidgets.qLabel = widget.NewLabel("Q: 0.71") // Default display
		bandWidgets.qSlider = widget.NewSlider(0.0, 1.0)
		bandWidgets.qSlider.Step = 0.01
		bandWidgets.qSlider.SetValue(helper.QToSliderValue(0.71, MinQ, MaxQ)) // Default Q
		bandWidgets.qSlider.OnChanged = func(val float64) {
			q := helper.SliderValueToQ(val, MinQ, MaxQ)
			bandWidgets.qLabel.SetText(fmt.Sprintf("Q: %.2f", q))
			if g.paramCb != nil {
				g.paramCb(pIdxQ, float32(val))
			}
		}

		// --- Gain Slider ---
		bandWidgets.gainLabel = widget.NewLabel("G: 0.0dB") // Default display
		bandWidgets.gainSlider = widget.NewSlider(0.0, 1.0)
		bandWidgets.gainSlider.Step = 0.01
		bandWidgets.gainSlider.SetValue(helper.GainToSliderValue(0.0, MinGainDB, MaxGainDB)) // Default 0dB
		bandWidgets.gainSlider.OnChanged = func(val float64) {
			gain := helper.SliderValueToGain(val, MinGainDB, MaxGainDB)
			bandWidgets.gainLabel.SetText(fmt.Sprintf("G: %.1fdB", gain))
			if g.paramCb != nil {
				g.paramCb(pIdxGain, float32(val))
			}
		}

		// --- Enable Checkbox (Example) ---
		// bandWidgets.enableCheck = widget.NewCheck("On", func(checked bool) {
		//     // TODO: Map bool to float32 (0.0 or 1.0)
		//     // if g.paramCb != nil { g.paramCb(pIdxEnabled, ...) }
		// })
		// bandWidgets.enableCheck.SetChecked(true) // Default enabled

		// --- Layout for single band ---
		bandWidgets.container = container.NewVBox(
			widget.NewLabel(fmt.Sprintf("Band %d", bandIndex+1)),
			// bandWidgets.enableCheck, // Add enable check
			bandWidgets.freqLabel,
			bandWidgets.freqSlider,
			bandWidgets.qLabel,
			bandWidgets.qSlider,
			bandWidgets.gainLabel,
			bandWidgets.gainSlider,
			// TODO: Add Type selector
		)
		g.bandControls[i] = bandWidgets
		bandContainer.Add(bandWidgets.container)
	}

	// --- Main Layout ---
	controlsArea := container.NewVBox(
		// Use a border layout to put the label next to the slider
		container.NewBorder(nil, nil, gainLabel, g.masterGainLabel, g.masterGain),
		widget.NewSeparator(),
		bandContainer,
	)

	content := container.NewBorder(nil, controlsArea, nil, nil, g.plotArea) // Plot on top, controls below

	g.window.SetContent(content)
	g.window.Resize(fyne.NewSize(600, 500)) // Increased size

	// Handle window closing - just hide it, don't exit the app
	g.window.SetCloseIntercept(func() {
		g.stopPlotUpdates()
		g.isVisible = false // Update visibility state
		g.window.Hide()
	})

	// Embed the window if hwnd is provided (Platform specific - might need adjustments)
	// This part is highly dependent on the VST host and OS.
	// Fyne's direct embedding capabilities might be limited or require specific drivers.
	// For now, we'll just show the window independently.
	// if drv, ok := g.app.Driver().(desktop.Driver); ok && hwnd != nil {
	// 	// Attempt to embed - This API might change or not exist
	// 	drv.SetParent(g.window, hwnd) // Hypothetical embedding function
	// 	log.Println("Attempting to embed GUI...") // Placeholder log
	// } else {
	// 	log.Println("Cannot embed GUI, showing as separate window.")
	// }

	g.window.Show()
	g.startPlotUpdates()

	// Run the Fyne app event loop in a separate goroutine
	// to avoid blocking the VST audio thread.
	go func() {
		log.Println("Starting Fyne event loop...")
		g.app.Run() // This will block until the app exits (which it shouldn't in VST context)
		log.Println("Fyne event loop finished.")
	}()
}

// drawPlot is called by Fyne to render the plot raster
func (g *GUI) drawPlot(w, h int) image.Image {
	g.plotMutex.RLock()
	defer g.plotMutex.RUnlock()

	img := image.NewRGBA(image.Rect(0, 0, w, h))
	if g.plotRenderer == nil {
		// Draw background if renderer not ready
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				img.Set(x, y, color.Gray{Y: 30}) // Default bg
			}
		}
		return img
	}

	// Update renderer size and render
	g.plotRenderer.SetSize(float32(w), float32(h))
	g.plotRenderer.Render(img) // Render draws background, grid, and line
	return img
}

// startPlotUpdates starts a ticker to periodically update the plot
func (g *GUI) startPlotUpdates() {
	if g.plotTicker != nil {
		g.plotTicker.Stop()
	}
	cfg := config.GetConfig()
	updateInterval := time.Second / time.Duration(cfg.GUI.PlotUpdateRateHz)
	g.plotTicker = time.NewTicker(updateInterval)

	go func() {
		for range g.plotTicker.C {
			// Check if window exists and is visible before updating plot
			if g.window != nil && g.isVisible { // Use Visible() method
				g.updatePlot()
			}
		}
	}()
}

// stopPlotUpdates stops the plot update ticker
func (g *GUI) stopPlotUpdates() {
	if g.plotTicker != nil {
		g.plotTicker.Stop()
		g.plotTicker = nil
	}
}

// Close hides the GUI window and stops updates
func (g *GUI) Close() {
	if g.window != nil {
		g.stopPlotUpdates()
		g.window.Hide()
	}
}

// UpdateParameter updates a GUI control based on changes from the host/plugin logic
func (g *GUI) UpdateParameter(index int, value float32) {
	if !g.isVisible { // Use manual visibility state
		return
	}

	// Update Master Gain
	if index == MasterGain {
		if g.masterGain != nil && float32(g.masterGain.Value) != value {
			g.masterGain.SetValue(float64(value))
			gainDB := helper.SliderValueToGain(float64(value), MinGainDB, MaxGainDB)
			g.masterGainLabel.SetText(fmt.Sprintf("%.1f dB", gainDB)) // Update label
			// No need to refresh label explicitly, SetText handles it.
			// g.masterGain.Refresh() // Slider refresh might still be needed visually
		}
		g.triggerPlotUpdate() // Master gain doesn't change plot shape, but good practice
		return
	}

	// Update Band Parameters
	if index > MasterGain {
		bandIndex := (index - Band1Type) / ParamsPerBand
		paramType := (index - Band1Type) % ParamsPerBand

		if bandIndex >= 0 && bandIndex < NumBands && g.bandControls[bandIndex] != nil {
			bc := g.bandControls[bandIndex]
			val64 := float64(value)

			switch paramType {
			// case 0: // Type - TODO
			case 1: // Freq
				if bc.freqSlider != nil && float32(bc.freqSlider.Value) != value {
					bc.freqSlider.SetValue(val64)
					freq := helper.LogScale(val64, MinFreq, MaxFreq)
					bc.freqLabel.SetText(fmt.Sprintf("F: %.0fHz", freq))
					bc.freqLabel.Refresh()
					bc.freqSlider.Refresh()
				}
			case 2: // Q
				if bc.qSlider != nil && float32(bc.qSlider.Value) != value {
					bc.qSlider.SetValue(val64)
					q := helper.SliderValueToQ(val64, MinQ, MaxQ)
					bc.qLabel.SetText(fmt.Sprintf("Q: %.2f", q))
					bc.qLabel.Refresh()
					bc.qSlider.Refresh()
				}
			case 3: // Gain
				if bc.gainSlider != nil && float32(bc.gainSlider.Value) != value {
					bc.gainSlider.SetValue(val64)
					gain := helper.SliderValueToGain(val64, MinGainDB, MaxGainDB)
					bc.gainLabel.SetText(fmt.Sprintf("G: %.1fdB", gain))
					bc.gainLabel.Refresh()
					bc.gainSlider.Refresh()
				}
				// case 4: // Enabled - TODO
				// if bc.enableCheck != nil {
				// 	checked := value >= 0.5
				// 	if bc.enableCheck.Checked != checked {
				// 		bc.enableCheck.SetChecked(checked)
				// 	}
				// }
			}
			g.triggerPlotUpdate() // Update plot when band params change
		}
	}
}

// triggerPlotUpdate requests a plot update (can be called frequently)
func (g *GUI) triggerPlotUpdate() {
	// This function is less critical now with the ticker,
	// but can be used for immediate updates after parameter changes if desired.
	// For now, rely on the ticker.
}

// updatePlot recalculates and redraws the frequency response plot
func (g *GUI) updatePlot() {
	if g.filterProvider == nil || g.analyzer == nil || g.plotRenderer == nil || g.plotArea == nil {
		log.Println("GUI: Plot update skipped, components not ready.")
		return
	}

	filters := g.filterProvider()
	if filters == nil {
		log.Println("GUI: Plot update skipped, no filters provided.")
		return
	}

	// Calculate response (using more points for smoother plot)
	freqs, mags := g.analyzer.CalculateResponse(filters, 512) // Use more points for plot

	g.plotMutex.Lock()
	g.plotRenderer.SetData(freqs, mags)
	g.plotMutex.Unlock()

	// Request canvas refresh
	g.plotArea.Refresh()
	// log.Println("GUI: Plot updated.") // Optional: for debugging
}
