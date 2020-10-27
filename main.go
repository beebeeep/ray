package main

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
)

const (
	RESX = 1280
	RESY = 1024
)

var (
	matRed = material{color.NRGBA{0xd0, 0x30, 0x30, 0xff}}
	matGreen = material{color.NRGBA{0x30, 0xd0, 0x30, 0xff}}
	matOrange = material{color.NRGBA{0xd0, 0x90, 0x20, 0xff}}
)

type vector struct {
	x, y, z float64
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

type material struct {
	diffuseColor color.NRGBA
}

type sphere struct {
	radius float64
	center vector
	material material
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

func colorMultiply(c color.NRGBA, i float64) color.NRGBA {
	if i < 0 {
		return color.NRGBA{}
	}
	R := float64(c.R) * i
	G := float64(c.G) * i
	B := float64(c.B) * i
	if R > 255 {
		R = 255
	}
	if G > 255 {
		G = 255
	}
	if B > 255 {
		B = 255
	}

	return color.NRGBA{uint8(R), uint8(G), uint8(B), c.A}
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
	t0 := plm - di
	t1 := plm + di
	if t0 < 0 {
		t0 = t1
	}
	if t0 < 0 {
		return 0, false
	}
	return t0, true
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

func castRay(origin, dir vector, scene []object, lights []light) color.NRGBA {
	intersection, normale, obj := sceneIntersec(origin, dir, scene)
	if obj == nil {
		return color.NRGBA{0x40, 0x40, 0x40, 0xff}
	}
	var diffuseLightIntensity float64
	for i := range lights {
		lightDir := lights[i].position.Sub(intersection).Normalize()
		diffuseLightIntensity += lights[i].intensity*math.Max(0, lightDir.DotProduct(normale))
	}
	return colorMultiply(obj.getMaterial().diffuseColor, diffuseLightIntensity)
}

func render(img *image.RGBA) {
	fov := math.Pi / 4.0
	ft := math.Tan(fov/2.0)
	camera := vector{0, 0, 50}
	scene := []object{
		&sphere{5, vector{2, 2, -20}, matRed},
		&sphere{2, vector{-5, 10, -30}, matGreen},
		&sphere{8, vector{13, 5, -20}, matOrange},
	}
	lights := []light{
		{vector{-10, 30, 10}, 1},
		{vector{110, 0, 10}, 0.2},
	}

	for y := 0; y < RESY; y++ {
		for x := 0; x < RESX; x++ {
			dx := (2*(float64(x)+0.5)/float64(RESX) - 1.0) * ft * float64(RESX) / float64(RESY)
			dy := -(2*(float64(y)+0.5)/float64(RESY) - 1.0) * ft
			dir := NewNormalized(dx, dy, -1)
			img.Set(x, y, castRay(camera, dir, scene, lights))
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
