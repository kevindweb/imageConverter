// main_test.go
package main

import (
	"testing"
)

func TestGetVal(t *testing.T) {
	img := readFile("clownfish")
	compareImg := readPngFile("clownfishReal")
	compareHeight := compareImg.Rect.Dy()
	compareWidth := compareImg.Rect.Dx()
	for i := 0; i < 1; i++ {
		res := runIconOptimal(img)

		width := res.Rect.Bounds().Dx()
		height := res.Rect.Bounds().Dy()
		if width != compareWidth || height != compareHeight {
			t.Fatalf("Width/Height not the same \nold: %dx%d \nnew: %dx%d\n", compareWidth, compareHeight, width, height)
		}

		for j := 0; j < height; j++ {
			for i := 0; i < width; i++ {
				if !colorCompare(res.At(i, j), compareImg.At(i, j)) {
					t.Fatalf("Colors are off at pixel %d,%d", i, j)
				}
			}
		}
	}
}

// func BechmarkIcon(b *testing.B) {
// 	img := readFile("clownfish")
// 	for i := 0; i < b.N; i++ {
// 		runIcon(img)
// 	}
// }
