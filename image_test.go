// main_test.go
package main

import "testing"

func TestGetVal(t *testing.T) {
	img := readFile("clownfish")
	for i := 0; i < 40; i++ {
		runIcon(img)
	}
}
