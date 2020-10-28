package main

import (
	"image/color"
	"math"
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

func (v vector) Length() float64 {
	return math.Sqrt(v.x*v.x + v.y*v.y + v.z*v.z)
}

func (v vector) Reflect(n vector) vector {
	// assuming v is normalized
	return v.Sub(n.Multiply(2.0*v.DotProduct(n)))
}

func (v vector) toRGBA() color.NRGBA {
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
	return color.NRGBA{uint8(R), uint8(G), uint8(B), 255}
}
