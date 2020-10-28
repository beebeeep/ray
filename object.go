package main

import "math"

type object interface {
	rayIntersect(origin, dir vector) (float64, bool)
	getMaterial() material
	getNormale(p vector) vector
}

type material struct {
	diffuse, specular, ambient vector
	specularExp float64
}

type light struct {
	position vector
	intensity float64
}

type sphere struct {
	radius float64
	center vector
	material material
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
