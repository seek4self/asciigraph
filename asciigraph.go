package asciigraph

import (
	"bytes"
	"fmt"
	"math"
	"strings"
)

type Graph struct {
	series   []float64
	conf     *config
	min, max float64
	interval float64
}

func NewGraph(series []float64) *Graph {
	return &Graph{
		series: series,
	}
}

func (g *Graph) parseData(options ...Option) {
	g.conf = configure(config{
		Offset:    3,
		Precision: 2,
	}, options)
	if g.conf.Width > 0 {
		g.series = interpolateArray(g.series, g.conf.Width)
	}
	g.min, g.max = minMaxFloat64Slice(g.series)
	g.interval = math.Abs(g.max - g.min)
	if g.conf.Height <= 0 {
		g.conf.Height = int(g.interval)
		if int(g.interval) <= 0 {
			g.conf.Height = int(g.interval * math.Pow10(int(math.Ceil(-math.Log10(g.interval)))))
		}
	}

}

func (g *Graph) build() {

}

// Plot returns ascii graph for a series.
func Plot(series []float64, options ...Option) string {
	var logMaximum float64
	config := configure(config{
		Offset:    3,
		Precision: 2,
	}, options)

	if config.Width > 0 {
		series = interpolateArray(series, config.Width)
	}

	minimum, maximum := minMaxFloat64Slice(series)
	interval := math.Abs(maximum - minimum)

	if config.Height <= 0 {
		config.Height = int(interval)
		if int(interval) <= 0 {
			config.Height = int(interval * math.Pow10(int(math.Ceil(-math.Log10(interval)))))
		}
	}

	if config.Offset <= 0 {
		config.Offset = 3
	}

	ratio := 1.0
	if interval != 0 {
		ratio = float64(config.Height) / interval
	}
	min2 := round(minimum * ratio)
	max2 := round(maximum * ratio)

	intmin2 := int(min2)
	intmax2 := int(max2)

	rows := int(math.Abs(float64(intmax2 - intmin2)))
	width := len(series) + config.Offset

	plot := make([][]string, rows+1)

	// initialise empty 2D grid
	for i := 0; i < rows+1; i++ {
		line := make([]string, width)
		for j := 0; j < width; j++ {
			line[j] = " "
		}
		plot[i] = line
	}

	precision := config.Precision
	logMaximum = math.Log10(math.Max(math.Abs(maximum), math.Abs(minimum))) //to find number of zeros after decimal
	if minimum == float64(0) && maximum == float64(0) {
		logMaximum = float64(-1)
	}

	if logMaximum < 0 {
		// negative log
		if math.Mod(logMaximum, 1) != 0 {
			// non-zero digits after decimal
			precision += uint(math.Abs(logMaximum))
		} else {
			precision += uint(math.Abs(logMaximum) - 1.0)
		}
	} else if logMaximum > 2 {
		precision = 0
	}

	maxNumLength := len(fmt.Sprintf("%0.*f", precision, maximum))
	minNumLength := len(fmt.Sprintf("%0.*f", precision, minimum))
	maxWidth := int(math.Max(float64(maxNumLength), float64(minNumLength)))

	// axis and labels
	for y := intmin2; y < intmax2+1; y++ {
		magnitude := float64(y)
		if rows > 0 {
			magnitude = maximum - (float64(y-intmin2) * interval / float64(rows))
		}

		label := fmt.Sprintf("%*.*f", maxWidth+1, precision, magnitude)
		w := y - intmin2
		h := int(math.Max(float64(config.Offset)-float64(len(label)), 0))

		plot[w][h] = label
		plot[w][config.Offset-1] = "┤"
	}

	var y0, y1 int

	if !math.IsNaN(series[0]) {
		y0 = int(round(series[0]*ratio) - min2)
		plot[rows-y0][config.Offset-1] = "┼" // first value
	}

	for x := 0; x < len(series)-1; x++ { // plot the line

		d0 := series[x]
		d1 := series[x+1]

		if math.IsNaN(d0) && math.IsNaN(d1) {
			continue
		}

		if math.IsNaN(d1) && !math.IsNaN(d0) {
			y0 = int(round(d0*ratio) - float64(intmin2))
			plot[rows-y0][x+config.Offset] = "╴"
			continue
		}

		if math.IsNaN(d0) && !math.IsNaN(d1) {
			y1 = int(round(d1*ratio) - float64(intmin2))
			plot[rows-y1][x+config.Offset] = "╶"
			continue
		}

		y0 = int(round(d0*ratio) - float64(intmin2))
		y1 = int(round(d1*ratio) - float64(intmin2))

		if y0 == y1 {
			plot[rows-y0][x+config.Offset] = "─"
		} else {
			if y0 > y1 {
				plot[rows-y1][x+config.Offset] = "╰"
				plot[rows-y0][x+config.Offset] = "╮"
			} else {
				plot[rows-y1][x+config.Offset] = "╭"
				plot[rows-y0][x+config.Offset] = "╯"
			}

			start := int(math.Min(float64(y0), float64(y1))) + 1
			end := int(math.Max(float64(y0), float64(y1)))
			for y := start; y < end; y++ {
				plot[rows-y][x+config.Offset] = "│"
			}
		}
	}

	// join columns
	var lines bytes.Buffer
	for h, horizontal := range plot {
		if h != 0 {
			lines.WriteRune('\n')
		}

		// remove trailing spaces
		lastCharIndex := 0
		for i := width - 1; i >= 0; i-- {
			if horizontal[i] != " " {
				lastCharIndex = i
				break
			}
		}

		for _, v := range horizontal[:lastCharIndex+1] {
			lines.WriteString(v)
		}
	}

	// add caption if not empty
	if config.Caption != "" {
		lines.WriteRune('\n')
		lines.WriteString(strings.Repeat(" ", config.Offset+maxWidth))
		if len(config.Caption) < len(series) {
			lines.WriteString(strings.Repeat(" ", (len(series)-len(config.Caption))/2))
		}
		lines.WriteString(config.Caption)
	}

	return lines.String()
}
