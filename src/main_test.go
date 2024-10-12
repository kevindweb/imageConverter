package main

import (
	"imageconverter/src/transparency"
	"testing"
)

func TestClownFish(t *testing.T) {
	img := transparency.ReadFile("clownfish")
	compareImg := transparency.ReadPngFile("clownfishReal")
	compareHeight := compareImg.Rect.Dy()
	compareWidth := compareImg.Rect.Dx()
	for i := 0; i < 1; i++ {
		res := transparency.RunIcon(img, 64, true)

		width := res.Rect.Bounds().Dx()
		height := res.Rect.Bounds().Dy()
		if width != compareWidth || height != compareHeight {
			t.Fatalf(
				"WidthxHeight not the same \nold: %dx%d \nnew: %dx%d\n",
				compareWidth,
				compareHeight,
				width,
				height,
			)
		}

		for j := 0; j < height; j++ {
			for i := 0; i < width; i++ {
				if !transparency.ColorCompare(res.At(i, j), compareImg.At(i, j)) {
					t.Fatalf("Colors are off at pixel %d,%d", i, j)
				}
			}
		}
	}
}

func BechmarkIcon(b *testing.B) {
	img := transparency.ReadFile("clownfish")
	for i := 0; i < b.N; i++ {
		transparency.RunIcon(img, 0, false)
	}
}
