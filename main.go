package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"math"
	"os"
	"time"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func init() {
	image.RegisterFormat("jpeg", "jpeg", jpeg.Decode, jpeg.DecodeConfig)
}

// readFile takes the input string and reads the image from ./images into memory
func readFile(fileToConvert string) image.Image {
	file, err := os.Open("images/" + fileToConvert + ".jpeg")
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

	err = ioutil.WriteFile("icons/"+fileName+".png", buf.Bytes(), 0644)
	check(err)
}

type componentPixel struct {
	pixel     color.Color
	component int
}

// findBackgroundColor scans the image for the most popular colors
// using a hashmap it tracks the highest and returns that as the background
// alongside a 1d representation of the pixels for further computation
func findBackgroundColor(img image.Image, width int, height int) ([3]uint32, []componentPixel) {
	matrix := make([]componentPixel, width*height)
	colorCount := make(map[color.Color]int)

	var popularColor color.Color
	maxPixelCount := 0

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pixel := img.At(x, y)
			pixelCount := 1
			matrix[y*width+x].pixel = pixel

			if count, ok := colorCount[pixel]; ok {
				pixelCount += count
			}
			if pixelCount > maxPixelCount {
				popularColor = pixel
				maxPixelCount = pixelCount
			}
			colorCount[pixel] = pixelCount
		}
	}

	red, green, blue, _ := popularColor.RGBA()
	return [3]uint32{red, green, blue}, matrix
}

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

// dfs iteratively adds neighbors to the component list to find the entire
// connected icon - can't use recursive, causes stackoverflow with num pixels
func dfs(col int, row int, width int, height int, matrix []componentPixel,
	background [3]uint32, component int) (int, [4]int) {
	var stack [][2]int
	stack = append(stack, [2]int{col, row})
	var col_row [2]int

	count := 0
	neighbors := [3]int{-1, 0, 1}
	neighborLength := len(neighbors)

	firstPixel := true
	var pixelSpace [4]int

	for len(stack) > 0 {
		col_row, stack = stack[len(stack)-1], stack[:len(stack)-1]
		col = col_row[0]
		row = col_row[1]

		if col >= width || col < 0 || row >= height || row < 0 {
			continue
		}

		inx := row*width + col
		if matrix[inx].component != 0 {
			// we've visited this pixel
			continue
		}

		if colorDiff(matrix[inx].pixel, background) < 15000 {
			// this is the background, don't want this component
			matrix[inx].component = -1
			continue
		}

		if firstPixel {
			// initialize the dimensions only when necessary
			firstPixel = false
			pixelSpace = [4]int{MaxInt, 0, MaxInt, 0}
		}

		// update dimensions of this icon [top, bottom, left, right]
		if row < pixelSpace[0] {
			pixelSpace[0] = row
		} else if row > pixelSpace[1] {
			pixelSpace[1] = row
		}
		if col < pixelSpace[2] {
			pixelSpace[2] = col
		} else if col > pixelSpace[3] {
			pixelSpace[3] = col
		}

		matrix[inx].component = component

		for i := 0; i < neighborLength; i++ {
			neighborColumn := col + neighbors[i]
			for j := 0; j < neighborLength; j++ {
				stack = append(stack, [2]int{neighborColumn, row + neighbors[j]})
			}
		}

		count += 1
	}

	return count, pixelSpace
}

// findIcon takes an image and searches for the connected components
// it then returns the component (and dimensions) with maximum pixel count
func findIcon(width int, height int, matrix []componentPixel,
	background [3]uint32) ([]componentPixel, int, [4]int) {
	/*
		 * find connected components
		 	* connected components are surrounded by "background" color
			* which separates them from other components
		 * find connected component with highest numPixels
		 * trim the image into only the icon's dimensions to save space
	*/

	components := 1
	maxComponent := 1
	maxComponentPixelCount := 0
	var componentDimensions [][4]int

	for j := 0; j < height; j++ {
		for i := 0; i < width; i++ {
			componentPixelCount, dimensions := dfs(i, j, width, height, matrix,
				background, components)
			if componentPixelCount > 0 {
				if componentPixelCount > maxComponentPixelCount {
					maxComponentPixelCount = componentPixelCount
					maxComponent = components
				}

				componentDimensions = append(componentDimensions, dimensions)
				components += 1
			}
		}
	}

	return matrix, maxComponent, componentDimensions[maxComponent-1]
}

// runIcon is the main entrypoint into the algorithm
// when given an image, it finds the background, components
// and returns the transparent png result
func runIcon(img image.Image) *image.RGBA {
	backgroundWidth := img.Bounds().Dx()
	backgroundHeight := img.Bounds().Dy()
	transparentColor := image.Transparent

	backgroundColor, pixelMatrix := findBackgroundColor(img, backgroundWidth, backgroundHeight)
	matrix, iconComponent, iconDimensions := findIcon(backgroundWidth,
		backgroundHeight, pixelMatrix, backgroundColor)

	topPixel := iconDimensions[0]
	bottomPixel := iconDimensions[1]
	leftPixel := iconDimensions[2]
	rightPixel := iconDimensions[3]

	iconWidth := rightPixel - leftPixel
	iconHeight := bottomPixel - topPixel
	background := image.NewRGBA(image.Rect(0, 0, iconWidth, iconHeight))
	matrixInx := topPixel*backgroundWidth + leftPixel
	var row int

	for j := 0; j < iconHeight; j++ {
		row = j*backgroundWidth + matrixInx
		for i := 0; i < iconWidth; i++ {
			// original matrix index is (j+topPixel)*backgroundWidth+(i+leftPixel)
			// accessing 2d matrix as 1d array https://stackoverflow.com/a/2151141
			pixel := matrix[row+i]
			if pixel.component == iconComponent {
				background.Set(i, j, pixel.pixel)
			} else {
				background.Set(i, j, transparentColor)
			}
		}
	}

	return background
}

func elapsed(what string) func() {
	start := time.Now()
	return func() {
		fmt.Printf("%s took %v\n", what, time.Since(start))
	}
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
