package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPathBasenameWithoutExt(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(PathBasenameWithoutExt("./a/test.bin"), "test", "should be equal")
}
