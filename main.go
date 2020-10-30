package main

import (
	"image"
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

// sceneIntersec returns object, point of intersection and its normale if ray intersects something
// if there is no intersection, then obj == nil
func sceneIntersec(origin, dir vector, scene []object) (intersection, normale vector, obj object) {
	minDist := math.MaxFloat64
	for i := range scene {
		if dist, intersects := scene[i].rayIntersect(origin, dir); intersects && dist < minDist {
			minDist = dist
			obj = scene[i]
		}
	}
	if obj != nil {
		intersection = origin.Add(dir.Multiply(minDist))
		normale = obj.getNormale(intersection)
	}
	return intersection, normale, obj
}

// isShadowed returns true if light l is visible in direction dir from intersection
func isShadowed(l light, scene []object, dir, intersection, normale vector) bool {
	lightDist := l.position.Sub(intersection).Length()

	// rise intersection point above the surface a bit
	// to make sure we won't intersect with itself
	shadowOrig := intersection.Offset(dir, normale, 1e-3)

	if shadowPoint, _, o := sceneIntersec(shadowOrig, dir, scene); o != nil {
		return shadowPoint.Sub(shadowOrig).Length() < lightDist
	}
	return false
}

// castRay casts ray from origin to dir and returns resulting color
func castRay(origin, dir vector, scene []object, lights []light, ttl int) vector {
	if ttl <= 0 {
		return backgroundColor
	}

	intersection, normale, obj := sceneIntersec(origin, dir, scene)
	if obj == nil {
		return backgroundColor
	}
	m := obj.getMaterial()

	reflectDir := dir.Reflect(normale)
	reflectOrig := intersection.Offset(reflectDir, normale, 1e-3)
	reflectColor := castRay(reflectOrig, reflectDir, scene, lights, ttl-1)
	reflectIntensity := m.specularRef.EntrywiseProduct(reflectColor)

	var refractDir, refractOrig, refractColor, refractIntensity vector
	refractDir = dir.Refract(normale, m.refractiveIndex).Normalize()
	if !refractDir.IsZero() {
		refractOrig = intersection.Offset(refractDir, normale, 1e-3)
		refractColor = castRay(refractOrig, refractDir, scene, lights, ttl-1)
		refractIntensity = m.transparency.EntrywiseProduct(refractColor)
	}

	var diffuseIntensity, specularIntensity vector
	var diffuseFactor, specularFactor float64
	for i := range lights {
		lightDir := lights[i].position.Sub(intersection).Normalize()

		if isShadowed(lights[i], scene, lightDir, intersection, normale) {
			continue
		}

		diffuseFactor = lights[i].intensity * math.Max(0.0, lightDir.DotProduct(normale))
		specularFactor = lights[i].intensity * math.Pow(
			math.Max(0.0, lightDir.Reflect(normale).DotProduct(dir)),
			m.specularExp,
		)

		// calculate lighting from i-th light for each component and color channel
		diffuseIntensity = diffuseIntensity.Add(m.diffuseRef.Multiply(diffuseFactor))
		specularIntensity = specularIntensity.Add(m.diffuseRef.Multiply(specularFactor))
	}

	return specularIntensity.Add(diffuseIntensity).Add(reflectIntensity).Add(refractIntensity)
}

func render(scene []object, lights []light, img *image.NRGBA64) {
	fov := math.Pi / 5.0
	ft := math.Tan(fov / 2.0)
	camera := vector{10, 5, 40}

	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			dx := (2*(float64(x)+0.5)/float64(RESX) - 1.0) * ft * float64(RESX) / float64(RESY)
			dy := -(2*(float64(y)+0.5)/float64(RESY) - 1.0) * ft
			dir := NewNormalized(dx, dy, -1)
			img.Set(x, y, castRay(camera, dir, scene, lights, 4).toNRGBA64())
		}
	}
}

func startRendering(scene []object, lights []light, img *image.NRGBA64) {
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
			render(scene, lights, img.SubImage(r).(*image.NRGBA64))
			log.Printf("stripe %v rendered in %s", r, time.Now().Sub(t).Round(time.Millisecond))
			wg.Done()
		}()
	}
	wg.Wait()
}

func main() {
	start := time.Now()
	defer func() {
		log.Printf("done in %s", time.Now().Sub(start).Round(time.Millisecond))
	}()
	scene := []object{
		&sphere{5, vector{2, 2, -20}, matOrangeMetal},
		&sphere{4, vector{-5, 10, -30}, matIvory},
		&sphere{8, vector{13, 5, -20}, matRedMetal},
		&sphere{1, vector{2, 5, -12}, matIvory},
		&sphere{8, vector{28, 15, -12}, matBlackGlossy},
		&sphere{5, vector{7, 20, -30}, matMirror},
		&sphere{3, vector{6, 1, -10}, matGlass},
	}
	lights := []light{
		{vector{-10, 30, 10}, 1.5},
		{vector{40, 0, 10}, 0.7},
	}

	img := image.NewNRGBA64(image.Rect(0, 0, RESX, RESY))
	startRendering(scene, lights, img)

	f, err := os.Create("out.bmp")
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
