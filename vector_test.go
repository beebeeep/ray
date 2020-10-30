package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test__refract_no_refraction(t *testing.T) {
	v := vector{0.26338599340389107, 0.29598121859703125, -0.9181628051257913}
	n := vector{-0.529572584399768, 0.6833382022188227, 0.5025950449820392}
	r := v.Refract(n, 1)
	assert.InEpsilon(t, v.x, r.x, 1e-6)
	assert.InEpsilon(t, v.y, r.y, 1e-6)
	assert.InEpsilon(t, v.z, r.z, 1e-6)
}
