package asciigraph

import (
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"runtime"
)

func minMaxFloat64Slice(v []float64) (min, max float64) {
	min = math.Inf(1)
	max = math.Inf(-1)

	if len(v) == 0 {
		panic("Empty slice")
	}

	for _, e := range v {
		if e < min {
			min = e
		}
		if e > max {
			max = e
		}
	}
	return
}

func round(input float64) float64 {
	if math.IsNaN(input) {
		return math.NaN()
	}
	sign := 1.0
	if input < 0 {
		sign = -1
		input *= -1
	}
	_, decimal := math.Modf(input)
	var rounded float64
	if decimal >= 0.5 {
		rounded = math.Ceil(input)
	} else {
		rounded = math.Floor(input)
	}
	return rounded * sign
}

// (y-y0)/(x-x0) = (y1-y0)/(x1-x0)
func linearInterpolate(data []float64, x float64) float64 {
	x0, x1 := math.Floor(x), math.Ceil(x)
	y0, y1 := data[int(x0)], data[int(x1)]
	return y0 + (y1-y0)*(x-x0)
}

func interpolateArray(data []float64, fitCount int) []float64 {
	interpolatedData := make([]float64, 0, fitCount)

	springFactor := float64(len(data)-1) / float64(fitCount-1)
	interpolatedData = append(interpolatedData, data[0])

	for i := 1; i < fitCount-1; i++ {
		interpolatedData = append(interpolatedData, linearInterpolate(data, float64(i)*springFactor))
	}
	interpolatedData = append(interpolatedData, data[len(data)-1])
	return interpolatedData
}

// clear terminal screen
var Clear func()

func init() {
	platform := runtime.GOOS

	if platform == "windows" {
		Clear = func() {
			cmd := exec.Command("cmd", "/c", "cls")
			cmd.Stdout = os.Stdout
			if err := cmd.Run(); err != nil {
				log.Fatal(err)
			}
		}
	} else {
		Clear = func() {
			fmt.Print("\033[2J\033[H")
		}
	}
}
