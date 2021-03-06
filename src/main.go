package main

import (
	"image"
	"image/jpeg"
	"os"
)

func init() {
	image.RegisterFormat("jpeg", "jpeg", jpeg.Decode, jpeg.DecodeConfig)
}

func main() {
	defer elapsed("imageConverter")()
	fileName := "clownFish"
	if len(os.Args) >= 2 {
		fileName = os.Args[1]
	}

	img := readFile(fileName)
	background := runIcon(img, 64, true)
	writeFile(fileName, background)
}
