package main

import (
	"testing"
)

func TestClownFish(t *testing.T) {
	img := readFile("clownfish")
	compareImg := readPngFile("clownfishReal")
	compareHeight := compareImg.Rect.Dy()
	compareWidth := compareImg.Rect.Dx()
	for i := 0; i < 1; i++ {
		res := runIcon(img, 64, false)

		width := res.Rect.Bounds().Dx()
		height := res.Rect.Bounds().Dy()
		if width != compareWidth || height != compareHeight {
			t.Fatalf("WidthxHeight not the same \nold: %dx%d \nnew: %dx%d\n", compareWidth, compareHeight, width, height)
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

func BenchmarkIcon(b *testing.B) {
	img := readFile("clownfish")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runIcon(img, 64, true)
	}
}
