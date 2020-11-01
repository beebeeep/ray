package main

import (
	"fmt"
	"image/color"
	"math"
)

type vector struct {
	x, y, z float64
}

type ray struct {
	origin, dir vector
	transformMatrix m44
}

type m44 [4][4]float64

type matrix [][]float64

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

func (v vector) CrossProduct(u vector) vector {
	return vector{v.y*u.z - v.z*u.y,  v.z*u.x - v.x*u.z, v.x * u.y - v.y*u.x}
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

func (v vector) TransformPoint(m m44) vector {
	var x, y, z, w float64
	x = v.x * m[0][0] + v.y * m[1][0] + v.z*m[2][0] + m[3][0]
	y = v.x * m[0][1] + v.y * m[1][1] + v.z*m[2][1] + m[3][1]
	z = v.x * m[0][2] + v.y * m[1][2] + v.z*m[2][2] + m[3][2]
	w = v.x * m[0][3] + v.y * m[1][3] + v.z*m[2][3] + m[3][3]

	return vector{x/w, y/w, z/w}
}

func (v vector) TransformDir(m m44) vector {
	var x, y, z  float64
	x = v.x * m[0][0] + v.y * m[1][0] + v.z*m[2][0]
	y = v.x * m[0][1] + v.y * m[1][1] + v.z*m[2][1]
	z = v.x * m[0][2] + v.y * m[1][2] + v.z*m[2][2]

	return vector{x, y, z}
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

func (a matrix) Multiply(b matrix) matrix {
	ra := len(a)
	ca := len(a[0])
	rb := len(b)
	cb := len(b[0])
	if ca != rb {
		panic(fmt.Sprintf("cannot multiply %dx%d and %dx%d", ra, ca, rb, cb))
	}

	r := make(matrix, ra)
	for i := 0; i < ra; i++ {
		r[i] = make([]float64, cb)
		for j := 0; j < cb; j++ {
			for k := 0; k < ca; k++ {
				r[i][j] += a[i][k]*b[k][j]
			}
		}
	}

	return r
}

func (m m44) Dump() {
	for _, r := range m {
		for _, v := range r {
			fmt.Printf("%v\t", v)
		}
		fmt.Print("\n")
	}
}

// calculateTransformMatrix calculates transformation matrix
// to convert from ray basis to world basis
// in ray basis z-axis goes in -r.dir
func (r *ray) calculateTransformMatrix() {
	tmp := vector{0,1,0}
	forward := r.dir.Multiply(-1).Normalize()
	right := tmp.CrossProduct(forward)
	up := forward.CrossProduct(right)

	r.transformMatrix[0][0] = right.x
	r.transformMatrix[0][1] = right.y
	r.transformMatrix[0][2] = right.z
	r.transformMatrix[0][3] = 0

	r.transformMatrix[1][0] = up.x
	r.transformMatrix[1][1] = up.y
	r.transformMatrix[1][2] = up.z
	r.transformMatrix[1][3] = 0

	r.transformMatrix[2][0] = forward.x
	r.transformMatrix[2][1] = forward.y
	r.transformMatrix[2][2] = forward.z
	r.transformMatrix[2][3] = 0

	r.transformMatrix[3][0] = r.origin.x
	r.transformMatrix[3][1] = r.origin.y
	r.transformMatrix[3][2] = r.origin.z
	r.transformMatrix[3][3] = 1
}
