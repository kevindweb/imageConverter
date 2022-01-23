// image_test.go
package main

import (
	"testing"
)

func TestGetVal(t *testing.T) {
	img := readFile("clownfish")
	for i := 0; i < 40; i++ {
		runIcon(img)
	}
}

func BechmarkIcon(b *testing.B) {
	img := readFile("clownfish")
	for i := 0; i < b.N; i++ {
		runIcon(img)
	}
}
