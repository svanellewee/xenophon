package main_test

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSomething(t *testing.T) {
	t.Run("some test", func(t *testing.T) {
		assert.Equal(t, 1, 0)
	})
}
