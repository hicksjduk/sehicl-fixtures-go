package main

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestParseBreakpointsNotValid(t *testing.T) {
	checker := func(s string) {
		bp := parseBreakpoints([]byte(s))
		assert.Nil(t, bp)
	}
	checker("rubbish")
	checker("1asgasdgas2")
	checker("84, 3")
}

func TestParseBreakpointsValid(t *testing.T) {
	bp := parseBreakpoints([]byte("   1241   32  1 3   1  "))
	assert.Equal(t, []int{1241, 32, 1, 3, 1}, bp)
}
