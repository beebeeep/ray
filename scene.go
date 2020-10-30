package main

import (
	"image"
	"math"
)

type scene struct {
	objects []object
	lights []light
	fov float64
	camera vector
}

// intersec returns object, point of intersection and its normale if ray intersects something
// if there is no intersection, then obj == nil
func (s *scene) intersec(origin, dir vector) (intersection, normale vector, obj object) {
	minDist := math.MaxFloat64
	for i := range s.objects {
		if dist, intersects := s.objects[i].rayIntersect(origin, dir); intersects && dist < minDist {
			minDist = dist
			obj = s.objects[i]
		}
	}
	if obj != nil {
		intersection = origin.Add(dir.Multiply(minDist))
		normale = obj.getNormale(intersection)
	}
	return intersection, normale, obj
}

// isShadowed returns true if light l is visible in direction dir from intersection
func (s *scene) isShadowed(l light, dir, intersection, normale vector) bool {
	lightDist := l.position.Sub(intersection).Length()

	// rise intersection point above the surface a bit
	// to make sure we won't intersect with itself
	shadowOrig := intersection.Offset(dir, normale, 1e-3)

	if shadowPoint, _, o := s.intersec(shadowOrig, dir); o != nil {
		return shadowPoint.Sub(shadowOrig).Length() < lightDist
	}
	return false
}

// castRay casts ray from origin to dir and returns resulting color
func (s *scene) castRay(origin, dir vector, ttl int) vector {
	if ttl <= 0 {
		return backgroundColor
	}

	intersection, normale, obj := s.intersec(origin, dir)
	if obj == nil {
		return backgroundColor
	}
	m := obj.getMaterial()

	reflectDir := dir.Reflect(normale)
	reflectOrig := intersection.Offset(reflectDir, normale, 1e-3)
	reflectColor := s.castRay(reflectOrig, reflectDir, ttl-1)
	reflectIntensity := m.specularRef.EntrywiseProduct(reflectColor)

	var refractDir, refractOrig, refractColor, refractIntensity vector
	refractDir = dir.Refract(normale, m.refractiveIndex).Normalize()
	if !refractDir.IsZero() {
		refractOrig = intersection.Offset(refractDir, normale, 1e-3)
		refractColor = s.castRay(refractOrig, refractDir, ttl-1)
		refractIntensity = m.transparency.EntrywiseProduct(refractColor)
	}

	var diffuseIntensity, specularIntensity vector
	var diffuseFactor, specularFactor float64
	for i := range s.lights {
		lightDir := s.lights[i].position.Sub(intersection).Normalize()

		if s.isShadowed(s.lights[i], lightDir, intersection, normale) {
			continue
		}

		diffuseFactor = s.lights[i].intensity * math.Max(0.0, lightDir.DotProduct(normale))
		specularFactor = s.lights[i].intensity * math.Pow(
			math.Max(0.0, lightDir.Reflect(normale).DotProduct(dir)),
			m.specularExp,
		)

		// calculate lighting from i-th light for each component and color channel
		diffuseIntensity = diffuseIntensity.Add(m.diffuseRef.Multiply(diffuseFactor))
		specularIntensity = specularIntensity.Add(m.diffuseRef.Multiply(specularFactor))
	}

	return specularIntensity.Add(diffuseIntensity).Add(reflectIntensity).Add(refractIntensity)
}

func (s *scene) render(img *image.NRGBA64) {

	ft := math.Tan(s.fov / 2.0)

	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			dx := (2*(float64(x)+0.5)/float64(RESX) - 1.0) * ft * float64(RESX) / float64(RESY)
			dy := -(2*(float64(y)+0.5)/float64(RESY) - 1.0) * ft
			dir := NewNormalized(dx, dy, -1)
			img.Set(x, y, s.castRay(s.camera, dir, _ttl).toNRGBA64())
		}
	}
}
