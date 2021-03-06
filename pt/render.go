package pt

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"math/rand"
	"runtime"
	"time"
)

func showProgress(start time.Time, i, h int) {
	pct := int(100 * float64(i) / float64(h))
	elapsed := time.Since(start)
	fmt.Printf("\r%4d / %d (%3d%%) [", i, h, pct)
	for p := 0; p < 100; p += 2 {
		if pct > p {
			fmt.Print("=")
		} else {
			fmt.Print(" ")
		}
	}
	fmt.Printf("] %s ", DurationString(elapsed))
}

func Render(scene *Scene, camera *Camera, w, h, cameraSamples, hitSamples, depth int) image.Image {
	ncpu := runtime.NumCPU()
	runtime.GOMAXPROCS(ncpu)
	scene.Compile()
	image := image.NewNRGBA(image.Rect(0, 0, w, h))
	ch := make(chan int, h)
	fmt.Printf("%d x %d pixels, %d x %d = %d samples, %d bounces, %d cores\n",
		w, h, cameraSamples, hitSamples, cameraSamples*hitSamples, depth, ncpu)
	start := time.Now()
	for i := 0; i < ncpu; i++ {
		go func(i int) {
			n := int(math.Sqrt(float64(cameraSamples)))
			rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
			for y := i; y < h; y += ncpu {
				for x := 0; x < w; x++ {
					c := Color{}
					for u := 0; u < n; u++ {
						for v := 0; v < n; v++ {
							fu := (float64(u) + 0.5) / float64(n)
							fv := (float64(v) + 0.5) / float64(n)
							ray := camera.CastRay(x, y, w, h, fu, fv)
							c = c.Add(scene.Sample(ray, hitSamples, depth, rnd))
						}
					}
					c = c.Div(float64(n * n))
					r := uint8(math.Min(255, c.R*255))
					g := uint8(math.Min(255, c.G*255))
					b := uint8(math.Min(255, c.B*255))
					image.SetNRGBA(x, y, color.NRGBA{r, g, b, 255})
				}
				ch <- 1
			}
		}(i)
	}
	showProgress(start, 0, h)
	for i := 0; i < h; i++ {
		<-ch
		showProgress(start, i+1, h)
	}
	fmt.Println()
	return image
}
