package render

import (
	"image"
	"image/color"
	"log" // Added for safety check
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"github.com/C0d3-5t3w/go-audioEq/internal/analyzer"
)

// PlotRenderer handles drawing the frequency response plot
type PlotRenderer struct {
	analyzer     *analyzer.Analyzer
	canvasObject fyne.CanvasObject // The canvas object (e.g., Raster) to draw on
	width        float32
	height       float32
	frequencies  []float64 // Cached frequencies from analyzer
	magnitudes   []float64 // Cached magnitudes (dB) from analyzer
	bgColor      color.Color
	gridColor    color.Color
	lineColor    color.Color
	minDB        float64
	maxDB        float64
}

// NewPlotRenderer creates a new renderer for the plot
func NewPlotRenderer(analyzer *analyzer.Analyzer, canvas fyne.CanvasObject, minDB, maxDB float64) *PlotRenderer {
	return &PlotRenderer{
		analyzer:     analyzer,
		canvasObject: canvas,
		bgColor:      color.Gray{Y: 30},
		gridColor:    color.Gray{Y: 60},
		lineColor:    color.RGBA{R: 100, G: 200, B: 255, A: 255},
		minDB:        minDB,
		maxDB:        maxDB,
	}
}

// SetData updates the frequency response data to be plotted
func (r *PlotRenderer) SetData(freqs []float64, mags []float64) {
	r.frequencies = freqs
	r.magnitudes = mags
	canvas.Refresh(r.canvasObject) // Request a redraw
}

// SetSize updates the dimensions of the plot area
func (r *PlotRenderer) SetSize(w, h float32) {
	r.width = w
	r.height = h
}

// Render draws the plot onto the provided image buffer
func (r *PlotRenderer) Render(img *image.RGBA) {
	if r.width == 0 || r.height == 0 || len(r.frequencies) == 0 {
		// Fill with background color if no data or size
		bounds := img.Bounds()
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				img.Set(x, y, r.bgColor)
			}
		}
		return
	}

	// Draw background
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			img.Set(x, y, r.bgColor)
		}
	}

	// Draw Grid (Example: Simple dB lines and octave lines)
	r.drawGrid(img)

	// Draw Frequency Response Line
	r.drawResponseLine(img)
}

func (r *PlotRenderer) drawGrid(img *image.RGBA) {
	bounds := img.Bounds()
	width := float64(bounds.Dx())
	height := float64(bounds.Dy())

	// --- dB Lines (Horizontal) ---
	dbRange := r.maxDB - r.minDB
	if dbRange <= 0 {
		dbRange = 60.0
	} // Avoid division by zero
	// Draw 0dB line thicker or different color?
	zeroDbY := height - ((0.0 - r.minDB) / dbRange * height)
	izeroDbY := bounds.Min.Y + int(math.Round(zeroDbY))
	if izeroDbY >= bounds.Min.Y && izeroDbY < bounds.Max.Y {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			img.Set(x, izeroDbY, color.Gray{Y: 90}) // Brighter gray for 0dB
		}
	}

	// Other dB lines
	dbStep := 6.0 // Draw lines every 6dB
	for db := math.Ceil(r.minDB/dbStep) * dbStep; db <= r.maxDB; db += dbStep {
		if db == 0.0 {
			continue
		} // Skip 0dB line, already drawn
		if db >= r.minDB && db <= r.maxDB {
			y := height - ((db - r.minDB) / dbRange * height)
			iy := bounds.Min.Y + int(math.Round(y))
			if iy >= bounds.Min.Y && iy < bounds.Max.Y {
				for x := bounds.Min.X; x < bounds.Max.X; x++ {
					img.Set(x, iy, r.gridColor)
				}
			}
		}
	}

	// --- Frequency Lines (Vertical - Logarithmic) ---
	minFreq := r.analyzer.MinFreq() // Use getter
	maxFreq := r.analyzer.MaxFreq() // Use getter
	if minFreq <= 0 || maxFreq <= minFreq {
		return // Invalid frequency range
	}
	minLogFreq := math.Log10(minFreq)
	maxLogFreq := math.Log10(maxFreq)
	logRange := maxLogFreq - minLogFreq
	if logRange <= 0 {
		logRange = math.Log10(maxFreq / minFreq) // Recalculate just in case
		if logRange <= 0 {
			return
		} // Still invalid
	}

	// Decades (10, 100, 1k, 10k)
	decadeFreqs := []float64{10, 100, 1000, 10000, 100000}
	for _, freq := range decadeFreqs {
		if freq >= minFreq && freq <= maxFreq {
			logFreq := math.Log10(freq)
			x := (logFreq - minLogFreq) / logRange * width
			ix := bounds.Min.X + int(math.Round(x))
			if ix >= bounds.Min.X && ix < bounds.Max.X {
				for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
					// Make decade lines slightly brighter
					// Assert gridColor to access specific fields like Y
					if gridGray, ok := r.gridColor.(color.Gray); ok {
						img.Set(ix, y, color.Gray{Y: gridGray.Y + 15})
					} else if gridGray16, ok := r.gridColor.(color.Gray16); ok {
						// Handle Gray16 if necessary, maybe average?
						img.Set(ix, y, color.Gray16{Y: gridGray16.Y + (15 * 256)}) // Approximate brightness increase
					} else {
						// Fallback if it's not Gray (e.g., RGBA)
						img.Set(ix, y, r.gridColor) // Or a slightly modified version
						log.Printf("Warning: gridColor is not color.Gray, using original color for decade lines.")
					}
				}
			}
		}
	}
	// Sub-Decade Lines (e.g., 20, 30,... 200, 300,... 2k, 3k...)
	for baseFreq := 1.0; baseFreq < maxFreq*10; baseFreq *= 10 {
		for i := 2; i < 10; i++ {
			freq := baseFreq * float64(i)
			if freq >= minFreq && freq <= maxFreq {
				logFreq := math.Log10(freq)
				x := (logFreq - minLogFreq) / logRange * width
				ix := bounds.Min.X + int(math.Round(x))
				if ix >= bounds.Min.X && ix < bounds.Max.X {
					for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
						img.Set(ix, y, r.gridColor)
					}
				}
			}
		}
	}
}

func (r *PlotRenderer) drawResponseLine(img *image.RGBA) {
	bounds := img.Bounds()
	width := float64(bounds.Dx())
	height := float64(bounds.Dy())
	numPoints := len(r.frequencies)
	if numPoints < 2 {
		return
	}

	minFreq := r.analyzer.MinFreq() // Use getter
	maxFreq := r.analyzer.MaxFreq() // Use getter
	if minFreq <= 0 || maxFreq <= minFreq {
		return // Invalid frequency range
	}
	minLogFreq := math.Log10(minFreq)
	maxLogFreq := math.Log10(maxFreq)
	logRange := maxLogFreq - minLogFreq
	dbRange := r.maxDB - r.minDB

	if logRange <= 0 || dbRange <= 0 {
		return // Invalid ranges
	}

	lastX, lastY := -1, -1

	for i := 0; i < numPoints; i++ {
		// Ensure frequency is within bounds for log calculation
		currentFreq := r.frequencies[i]
		if currentFreq < minFreq {
			currentFreq = minFreq
		}
		if currentFreq > maxFreq {
			currentFreq = maxFreq
		}

		// Map frequency to X (logarithmic)
		logFreq := math.Log10(currentFreq)
		x := (logFreq - minLogFreq) / logRange * width
		ix := bounds.Min.X + int(math.Round(x))

		// Map magnitude to Y (linear)
		mag := math.Max(r.minDB, math.Min(r.maxDB, r.magnitudes[i])) // Clamp magnitude
		y := height - ((mag - r.minDB) / dbRange * height)
		iy := bounds.Min.Y + int(math.Round(y))

		// Clamp coordinates to bounds
		ix = int(math.Max(float64(bounds.Min.X), math.Min(float64(bounds.Max.X-1), float64(ix))))
		iy = int(math.Max(float64(bounds.Min.Y), math.Min(float64(bounds.Max.Y-1), float64(iy))))

		if lastX != -1 {
			// Draw line segment from (lastX, lastY) to (ix, iy)
			drawLineSimple(img, lastX, lastY, ix, iy, r.lineColor)
		}
		lastX, lastY = ix, iy
	}
}

// drawLineSimple draws a basic line (might be jagged)
func drawLineSimple(img *image.RGBA, x0, y0, x1, y1 int, col color.Color) {
	dx := math.Abs(float64(x1 - x0))
	dy := math.Abs(float64(y1 - y0))
	steps := int(math.Max(dx, dy))
	if steps == 0 {
		img.Set(x0, y0, col)
		return
	}

	xInc := float64(x1-x0) / float64(steps)
	yInc := float64(y1-y0) / float64(steps)
	x, y := float64(x0), float64(y0)

	for i := 0; i <= steps; i++ {
		ix, iy := int(math.Round(x)), int(math.Round(y))
		// Basic bounds check
		if ix >= img.Bounds().Min.X && ix < img.Bounds().Max.X && iy >= img.Bounds().Min.Y && iy < img.Bounds().Max.Y {
			img.Set(ix, iy, col)
		}
		x += xInc
		y += yInc
	}
}
