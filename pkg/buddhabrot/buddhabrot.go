package buddhabrot

import (
	"math"
	"image"
	"image/color"
	"sync"
	"flag"
)


var (
	maxIterations = flag.Int("max-iterations", 1000, "Number of iterations dones at max")
	width = flag.Int("width", 800, "output image width")
	height = flag.Int("height", 600, "output image height")
	samples = flag.Int("samples", 100, "number of samples")
	mode = flag.String("mode", "seq", "rendering mode [seq|px|workers|row]")
	workers = flag.Int("workers", 1, "number of workers in workers rendering mode")
)

func renderPoint(x int, y int, counters *[][]uint) {
	c := complex(float64(x) * 3.0 / float64(*width) - 2, float64(y) * 3.0 / float64(*height) - 1.5)
	z := complex(0, 0)

	p := math.Sqrt((real(c) - 0.25) * (real(c) - 0.25) + imag(c) * imag(c))
	if (p - 2 * p * p + 0.25 > real(c)) {
		return 
	}

	history := make([]complex128, *maxIterations)
	a := 0;
	for a < *maxIterations {
		z = c + (z * z)
		history[a] = z;
		if (real(z) * real(z) + imag(z) * imag(z) > 4.0) {
			for b := 1; b <= a; b++ {
				var x1 int = int((real(history[b]) + 2) / 3.0 * float64(*width))
				var y1 int = int((imag(history[b]) + 1.5) / 3.0 * float64(*height))
				if (x1 >= 0 && x1 < *width && y1 >= 0 && y1 < *height) {
					(*counters)[x1][y1]++;
				}
			}
			break
		}
		a++;
	}
}

func Generate() *image.RGBA {
	var counters *[][]uint = newCounterTab()
	switch *mode {
	case "seq":
		renderSeq(counters)
	case "px":
		renderPerPixel(counters)
	case "row":
		renderPerRow(counters)
	case "workers":
		renderByWorkers(counters)
	}
	return fillImage(counters) 
}

func renderByWorkers(counters *[][]uint) {
	c := make(chan struct{x, y int})
	for i := 0; i < *workers; i++ {
		go func() {
			for px := range c {
				renderPoint(px.x, px.y, counters)
			}
		}()
	}
	for x := 0; x < *width; x++ {
		for y := 0; y < *height; y++ {
			c <- struct{x, y int}{x, y}
		}
	}
}

func renderPerRow(counters *[][]uint) {
	var wg sync.WaitGroup
	wg.Add(*width)
	for x := 0; x < *width; x++ {
		go func(x int) {
			for y := 0; y < *height; y++ {
				renderPoint(x, y, counters)
			}
			wg.Done()
		}(x)
	}
	wg.Wait()
}

func renderPerPixel(counters *[][]uint) {
	var wg sync.WaitGroup
	wg.Add(*width * *height)
	for x := 0; x < *width; x++ {
		for y := 0; y < *height; y++ {
			go func(x, y int) {
				renderPoint(x, y, counters)
				wg.Done()
			}(x, y)
		}
	}
	wg.Wait()
}

func renderSeq(counters *[][]uint) {
	for x := 0; x < *width; x++ {
		for y := 0; y < *height; y++ {
			renderPoint(x, y, counters)
		}
	}
}

func fillImage(counters *[][]uint) *image.RGBA {
	var total uint = 0
	var max uint = 0
	for x := 0; x < *width; x++ {
		for y := 0; y < *height; y++ {
			total += (*counters)[x][y]
			if ((*counters)[x][y] > max) {
				max = (*counters)[x][y]
			}
		}
	}
	luminosity := float64(total) / float64(*width * *height)
	image := image.NewRGBA(image.Rectangle{image.Point{0,0}, image.Point{*width,*height}})
	for x := 0; x < *width; x++ {
		for y := 0; y < *height; y++ {
			tmp := (int)(float64((*counters)[x][y]) * luminosity)
			image.SetRGBA(x, y, color.RGBA{uint8(tmp), uint8(tmp), uint8(tmp), 255})
		}
	}
	return image
}

func newCounterTab() *[][]uint {
	var counters [][]uint = make([][]uint, *width * *height)
	for x := 0; x < *width; x++ {
		counters[x] = make([]uint, *height)
	}
	return &counters
}
