package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test__sphere_rayIntersect_inside(t *testing.T) {
	s := sphere{10, vector{0, 0, 0}, material{}}
	d, ok := s.rayIntersect(vector{0, 0, 0}, vector{1, 1, 1})
	assert.True(t, ok)
	assert.Equal(t, float64(10), d)
}
