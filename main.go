package main

import (
	"image"
	_color "image/color"
	"image/png"
	"math"
	"os"
)

const (
	RESX = 3840
	RESY = 2160
)

var (
	matIvory = material{vector{0.24, 0.24, 0.18}, vector{0.3, 0.3, 0.3}, vector{0, 0, 0}, 50}
	matRedMetal = material{vector{0.37, 0.13, 0.13}, vector{0.06, 0.03, 0.03}, vector{0, 0, 0}, 10}
	matOrangeMetal = material{vector{0.45, 0.25, 0.1}, vector{0.03, 0.02, 0.01}, vector{0, 0, 0}, 10}

	backgroundColor = vector{0.2, 0.2, 0.2}
)

type vector struct {
	x, y, z float64
}

type material struct {
	diffuse, specular, ambient vector
	specularExp float64
}

type light struct {
	position vector
	intensity float64
}

type object interface {
	rayIntersect(origin, dir vector) (float64, bool)
	getMaterial() material
	getNormale(p vector) vector
}

type sphere struct {
	radius float64
	center vector
	material material
}



func NewNormalized(x, y, z float64) vector {
	l := math.Sqrt(x*x + y*y + z*z)
	return vector{x / l, y / l, z / l}
}

func (v vector) Add(u vector) vector {
	return vector{v.x + u.x, v.y + u.y, v.z + u.z}
}

func (v vector) Sub(u vector) vector {
	return vector{v.x - u.x, v.y - u.y, v.z - u.z}
}

func (v vector) Multiply(a float64) vector {
	return vector{v.x * a, v.y * a, v.z * a}
}

func (v vector) DotProduct(u vector) float64 {
	return v.x*u.x + v.y*u.y + v.z*u.z
}

func (v vector) Normalize() vector {
	return NewNormalized(v.x, v.y, v.z)
}

func (v vector) Length() float64 {
	return math.Sqrt(v.x*v.x + v.y*v.y + v.z*v.z)
}

func (v vector) Reflect(n vector) vector {
	// assuming v is normalized
	return v.Sub(n.Multiply(2.0*v.DotProduct(n)))
}

func (v vector) toRGBA() _color.NRGBA {
	R := v.x * 255
	G := v.y * 255
	B := v.z * 255
	if R > 255 {
		R = 255
	}
	if G > 255 {
		G = 255
	}
	if B > 255 {
		B = 255
	}
	return _color.NRGBA{uint8(R), uint8(G), uint8(B), 255}
}

func (s sphere) getMaterial() material {
	return s.material
}

func (s sphere) getNormale(p vector) vector {
	return p.Sub(s.center).Normalize()
}

// rayIntersect returns distance at which ray dir from origin intersects the sphere
func (s sphere) rayIntersect(origin, dir vector) (float64, bool) {
	// note: dir must be normalized

	l := s.center.Sub(origin)       // vector from origin to center
	plm := l.DotProduct(dir)        // length of projection of l on the ray
	d2 := l.DotProduct(l) - plm*plm // squared distance between center and the ray

	if d2 > s.radius*s.radius {
		return 0, false
	}

	di := math.Sqrt(float64(s.radius*s.radius - d2)) // distance between 1st intersection and projection of center to the ray
	if r := plm - di; r >= 0 {
		return r, true
	}
	if r := plm + di; r >= 0 {
		return r, true
	}
	return 0, false
}

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

func isShadowed(l light, scene []object, dir, intersection, normale vector) bool {
		var shadowOrig vector
		lightDist := l.position.Sub(intersection).Length()
		if dir.DotProduct(normale) < 0 {
			shadowOrig = intersection.Sub(normale.Multiply(1e-3))
		} else {
			shadowOrig = intersection.Add(normale.Multiply(1e-3))
		}
		if shadowPoint, _, o := sceneIntersec(shadowOrig, dir, scene); o != nil {
			return shadowPoint.Sub(shadowOrig).Length() < lightDist
		}
		return false
}

func castRay(origin, dir vector, scene []object, lights []light) vector {
	intersection, normale, obj := sceneIntersec(origin, dir, scene)
	if obj == nil {
		return backgroundColor
	}

	// calculate each factor for current intersection
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

		m := obj.getMaterial()
		diffuseIntensity = diffuseIntensity.Add(m.diffuse.Multiply(diffuseFactor))
		specularIntensity = specularIntensity.Add(m.diffuse.Multiply(specularFactor))
	}

	return specularIntensity.Add(diffuseIntensity)
}

func render(img *image.RGBA) {
	fov := math.Pi / 4.0
	ft := math.Tan(fov/2.0)
	camera := vector{0, 0, 50}
	scene := []object{
		&sphere{5, vector{2, 2, -20}, matOrangeMetal},
		&sphere{2, vector{-5, 10, -30}, matIvory},
		&sphere{8, vector{13, 5, -20}, matRedMetal},
		&sphere{1, vector{2, 5, -12}, matIvory},
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
			img.Set(x, y, castRay(camera, dir, scene, lights).toRGBA())
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
