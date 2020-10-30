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
	if l == 0 {
		return vector{0, 0, 0}
	}
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

func (v vector) EntrywiseProduct(a vector) vector {
	return vector{v.x * a.x, v.y * a.y, v.z * a.z}
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

func (v vector) IsZero() bool {
	return v.x == 0 && v.y == 0 && v.z == 0
}

func (v vector) Reflect(n vector) vector {
	// assuming v is normalized
	return v.Sub(n.Multiply(2.0 * v.DotProduct(n)))
}

func (v vector) Refract(normale vector, refractiveIndex float64) vector {
	if refractiveIndex == 1 {
		// shortcut for non-refracting materials
		return v
	}

	// assuming v is normalized and refraction happens on material/vacuum or vacuum/material boundary
	r := 1 / refractiveIndex
	cosTheta := -math.Max(-1.0, math.Min(1.0, v.DotProduct(normale)))
	if cosTheta < 0 {
		// refraction from material to vacuum, invert normal and refractiveIndex
		normale = normale.Multiply(-1)
		r = refractiveIndex
	}
	k := 1.0 - r*r*(1-cosTheta*cosTheta)
	if k < 0 {
		// total internal reflection
		//return v.Reflect(normale)
		return vector{0, 0, 0}
	}
	return v.Multiply(r).Add(normale.Multiply(r*cosTheta - math.Sqrt(k)))
}

func (v vector) Offset(dir, normale vector, dist float64) vector {
	if dir.DotProduct(normale) < 0 {
		return v.Sub(normale.Multiply(1e-3))
	}
	return v.Add(normale.Multiply(1e-3))
}

func (v vector) toNRGBA64() color.NRGBA64 {
	var s float64 = 1<<16 - 1
	R := v.x * s
	G := v.y * s
	B := v.z * s
	if R > s {
		R = s
	}
	if G > s {
		G = s
	}
	if B > s {
		B = s
	}
	return color.NRGBA64{uint16(R), uint16(G), uint16(B), uint16(s)}
}
