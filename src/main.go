package main

import (
	"image"
	"image/jpeg"
	"imageconverter/src/transparency"
	"os"
)

func init() {
	image.RegisterFormat("jpeg", "jpeg", jpeg.Decode, jpeg.DecodeConfig)
}

func main() {
	defer transparency.Elapsed("imageConverter")()
	fileName := "clownFish"
	if len(os.Args) >= 2 {
		fileName = os.Args[1]
	}

	img := transparency.ReadFile(fileName)
	background := transparency.RunIcon(img, 64, true)
	transparency.WriteFile(fileName, background)
}
