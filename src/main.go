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
	fileName := "clownfish"
	if len(os.Args) >= 2 {
		fileName = os.Args[1]
	}

	img := readFile(fileName)
	background := runIcon(img)
	writeFile(fileName, background)
}
