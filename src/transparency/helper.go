package transparency

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

// ReadFile takes the input string and reads the image from ./images into memory
func ReadFile(fileToConvert string) image.Image {
	file, err := os.Open("../images/" + fileToConvert + ".jpeg")
	check(err)
	defer file.Close()
	img, _, err := image.Decode(file)
	check(err)

	return img
}

func ReadPngFile(fileToRead string) *image.NRGBA {
	file, err := os.Open("../icons/" + fileToRead + ".png")
	check(err)
	defer file.Close()

	img, _, err := image.Decode(file)
	check(err)

	return img.(*image.NRGBA)
}

// WriteFile ouputs an image.RGBA to png file on disk in ./icons
func WriteFile(fileName string, background *image.RGBA) {
	buf := new(bytes.Buffer)
	err := png.Encode(buf, background)
	check(err)

	err = ioutil.WriteFile("../icons/"+fileName+".png", buf.Bytes(), 0644)
	check(err)
}

// Elapsed prints the function duration
// usage: defer Elapsed("what")()
func Elapsed(what string) func() {
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

func ColorCompare(c1, c2 color.Color) bool {
	r1, g1, b1, _ := c1.RGBA()
	r2, g2, b2, _ := c2.RGBA()
	return r1 == r2 && g1 == g2 && b1 == b2
}

const MaxInt = int(^uint(0) >> 1)

type componentPixel struct {
	pixel     color.Color
	component int
}
