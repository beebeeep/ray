package main

import (
	"image"
	"image/png"
	"math"
	"os"
)

const (
	RESX = 3840
	RESY = 2160
)

var (
	matIvory = material{vector{0.24, 0.24, 0.18}, vector{0.03, 0.03, 0.03}, 50}
	matRedMetal = material{vector{0.37, 0.13, 0.13}, vector{0.06, 0.03, 0.03},  10}
	matOrangeMetal = material{vector{0.45, 0.25, 0.1}, vector{0.03, 0.02, 0.01}, 10}
	matBlackGlossy = material{vector{0.04, 0.04, 0.04}, vector{0.15, 0.15, 0.15}, 50}
	matMirror = material{vector{0, 0, 0}, vector{0.9, 0.9, 0.9}, 1000}

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
func castRay(origin, dir vector, scene []object, lights []light, depth int) vector {
	if depth <= 0 {
		return backgroundColor
	}

	intersection, normale, obj := sceneIntersec(origin, dir, scene)
	if obj == nil {
		return backgroundColor
	}

	reflectDir := dir.Reflect(normale)
	reflectOrig := intersection.Offset(reflectDir, normale, 1e-3)
	reflectColor := castRay(reflectOrig, reflectDir, scene, lights, depth-1)
	reflectIntensity := obj.getMaterial().specular.EntrywiseProduct(reflectColor)

	var diffuseIntensity, specularIntensity vector
	var diffuseFactor, specularFactor float64
	for i := range lights {
		lightDir := lights[i].position.Sub(intersection).Normalize()

		if isShadowed(lights[i], scene, lightDir, intersection, normale) {
			continue
		}

		diffuseFactor = lights[i].intensity*math.Max(0.0, lightDir.DotProduct(normale))
		specularFactor = lights[i].intensity*math.Pow(
			math.Max(0.0, lightDir.Reflect(normale).DotProduct(dir)),
			obj.getMaterial().specularExp,
		)

		// calculate lighting from i-th light for each component and color channel
		m := obj.getMaterial()
		diffuseIntensity = diffuseIntensity.Add(m.diffuse.Multiply(diffuseFactor))
		specularIntensity = specularIntensity.Add(m.diffuse.Multiply(specularFactor))
	}

	return specularIntensity.Add(diffuseIntensity).Add(reflectIntensity)
}

func render(img *image.RGBA) {
	fov := math.Pi / 5.0
	ft := math.Tan(fov/2.0)
	camera := vector{10, 5, 40}
	scene := []object{
		&sphere{5, vector{2, 2, -20}, matOrangeMetal},
		&sphere{4, vector{-5, 10, -30}, matIvory},
		&sphere{8, vector{13, 5, -20}, matRedMetal},
		&sphere{1, vector{2, 5, -12}, matIvory},
		&sphere{8, vector{28, 15, -12}, matBlackGlossy},
		&sphere{3, vector{7, -7, -18}, matMirror},
	}
	lights := []light{
		{vector{-10, 30, 10}, 1.5},
		{vector{40, 0, 10}, 0.7},
	}

	for y := 0; y < RESY; y++ {
		for x := 0; x < RESX; x++ {
			dx := (2*(float64(x)+0.5)/float64(RESX) - 1.0) * ft * float64(RESX) / float64(RESY)
			dy := -(2*(float64(y)+0.5)/float64(RESY) - 1.0) * ft
			dir := NewNormalized(dx, dy, -1)
			img.Set(x, y, castRay(camera, dir, scene, lights, 4).toRGBA())
		}
	}
}

func main() {
	img := image.NewRGBA(image.Rect(0, 0, RESX, RESY))
	render(img)
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
