package main

import (
	"fmt"
	"image"
	"image/png"
	"log"
	"math"
	"os"
	"runtime"
	"sync"
	"time"

	"golang.org/x/image/bmp"
)

const (
	RESX = 3840
	RESY = 2160
	_ttl = 4
)

var (
	matIvory       = material{vector{0.24, 0.24, 0.18}, vector{0.03, 0.03, 0.03}, vector{0, 0, 0}, 50, 1}
	matRedMetal    = material{vector{0.37, 0.13, 0.13}, vector{0.06, 0.03, 0.03}, vector{0, 0, 0}, 10, 1}
	matOrangeMetal = material{vector{0.45, 0.25, 0.1}, vector{0.03, 0.02, 0.01}, vector{0, 0, 0}, 10, 1}
	matBlackGlossy = material{vector{0.04, 0.04, 0.04}, vector{0.15, 0.15, 0.15}, vector{0, 0, 0}, 50, 1}
	matMirror      = material{vector{0, 0, 0}, vector{0.9, 0.9, 0.9}, vector{0, 0, 0}, 1000, 1}
	matGlass       = material{vector{0, 0, 0}, vector{0, 0, 0}, vector{0.8, 0.9, 0.8}, 1000, 1.33}

	backgroundColor = vector{0.2, 0.2, 0.2}
)


func startRendering(s scene, img *image.NRGBA64) {
	start := time.Now()
	defer func() {
		log.Printf("rendering took %s", time.Now().Sub(start).Round(time.Millisecond))
	}()

	wg := sync.WaitGroup{}
	n := runtime.NumCPU()
	stripeWidth := RESY / n
	for i := 0; i < n; i++ {
		//r := image.Rect(i*stripeWidth, 0, (i+1)*stripeWidth, RESY)
		r := image.Rect(0, i*stripeWidth, RESX, (i+1)*stripeWidth)
		wg.Add(1)
		go func() {
			t := time.Now()
			s.render(img.SubImage(r).(*image.NRGBA64))
			log.Printf("stripe %v rendered in %s", r, time.Now().Sub(t).Round(time.Millisecond))
			wg.Done()
		}()
	}
	wg.Wait()
}

func renderStatic(s scene, img *image.NRGBA64) {
		startRendering(s, img)
		f, err := os.Create("out.png")
		if err != nil {
			panic(err)
		}
		if err := png.Encode(f, img); err != nil {
			panic(err)
		}
		if err := f.Close(); err != nil {
			panic(err)
		}
	}

func renderOrbit(s scene, img *image.NRGBA64) {
	var orbitR = float64(120)
	var orbitC = vector{7.5, 5, -20}
	fps := 60.0
	frames := 5.0*fps

	i := -1
	for phi := float64(0); phi < 2.0*math.Pi; phi += 2.0*math.Pi/frames {
		r := vector{orbitR*math.Cos(phi), 0, orbitR*math.Sin(phi)}
		i++
		log.Printf("rendering frame %d", i)

		s.camera.origin = orbitC.Add(r)
		s.camera.dir = orbitC.Sub(s.camera.origin).Normalize()
		s.camera.calculateTransformMatrix()
		startRendering(s, img)

		f, err := os.Create(fmt.Sprintf("out-%06d.bmp", i))
		if err != nil {
			panic(err)
		}
		s := time.Now()
		if err := bmp.Encode(f, img); err != nil {
			panic(err)
		}
		log.Printf("image written in %s", time.Now().Sub(s).Round(time.Millisecond))
		if err := f.Close(); err != nil {
			panic(err)
		}
	}
}

func main() {
	start := time.Now()
	defer func() {
		log.Printf("done in %s", time.Now().Sub(start).Round(time.Millisecond))
	}()
	scene := scene{
		objects: []object{
		&sphere{5, vector{2, 2, -20}, matOrangeMetal},
		&sphere{4, vector{-5, 10, -30}, matIvory},
		&sphere{8, vector{13, 5, -20}, matRedMetal},
		&sphere{1, vector{2, 5, -12}, matIvory},
		&sphere{8, vector{28, 15, -12}, matBlackGlossy},
		&sphere{5, vector{7, 20, -30}, matMirror},
		&sphere{3, vector{6, 1, -10}, matGlass},
	},
	 lights: []light{
		{vector{-10, 30, 10}, 1.5},
		{vector{40, 0, 10}, 0.7},
	},
	camera: ray{
		origin: vector{10, 5, 40},
		dir: vector{0, 0, -1},
	},
	fov: math.Pi / 5.0,
}
	scene.camera.calculateTransformMatrix()
	img := image.NewNRGBA64(image.Rect(0, 0, RESX, RESY))
	//renderOrbit(scene, img)
	renderStatic(scene, img)

}
