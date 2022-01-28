package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io/ioutil"
	"math"
	"os"
	"time"
)

// check raises an error if it exists
func check(e error) {
	if e != nil {
		panic(e)
	}
}

// readFile takes the input string and reads the image from ./images into memory
func readFile(fileToConvert string) image.Image {
	file, err := os.Open("../images/" + fileToConvert + ".jpeg")
	check(err)
	defer file.Close()
	img, _, err := image.Decode(file)
	check(err)

	return img
}

// writeFile ouputs an image.RGBA to png file on disk in ./icons
func writeFile(fileName string, background *image.RGBA) {
	buf := new(bytes.Buffer)
	err := png.Encode(buf, background)
	check(err)

	err = ioutil.WriteFile("../icons/"+fileName+".png", buf.Bytes(), 0644)
	check(err)
}

// elapsed prints the function duration
// to be used like: defer elapsed("what")()
func elapsed(what string) func() {
	start := time.Now()
	return func() {
		fmt.Printf("%s took %v\n", what, time.Since(start))
	}
}

// square returns float64 of the squared number
func square(num uint32) float64 {
	return float64(num * num)
}

// colorDiff compares two RGB colors and returns the result
// the Euclidean distance algorithm is based on this article
// https://en.wikipedia.org/wiki/Color_difference
func colorDiff(c1 color.Color, background [3]uint32) float64 {
	r1, g1, b1, _ := c1.RGBA()
	return math.Sqrt(square(background[0]-r1) +
		square(background[1]-g1) + square(background[2]-b1))
}

const MaxInt = int(^uint(0) >> 1)

type componentPixel struct {
	pixel     color.Color
	component int
}
