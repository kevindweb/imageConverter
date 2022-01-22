package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"math"
	"os"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func init() {
	image.RegisterFormat("jpeg", "jpeg", jpeg.Decode, jpeg.DecodeConfig)
}

func maxUpdate(highestColor *color.Color, contenderColor color.Color, colorMap map[color.Color]int) {
	if colorMap[contenderColor] > colorMap[*highestColor] {
		*highestColor = contenderColor
	}
}

func findBackgroundColor(img image.Image, width int, height int) color.Color {
	colorCount := make(map[color.Color]int)
	popularColor := img.At(0, 0)
	colorCount[popularColor] = 1

	for x := 1; x < width; x++ {
		for y := 0; y < height; y++ {
			pixel := img.At(x, y)
			if count, ok := colorCount[pixel]; ok {
				colorCount[pixel] = count + 1
			} else {
				colorCount[pixel] = 1
			}

			maxUpdate(&popularColor, pixel, colorCount)
		}
	}

	return popularColor
}

func square(num uint32) float64 {
	return float64(num * num)
}

func colorDiff(c1 color.Color, c2 color.Color) float64 {
	r2, g2, b2, _ := c2.RGBA()
	r1, g1, b1, _ := c1.RGBA()
	diff := math.Sqrt(square(r2-r1) + square(g2-g1) + square(b2-b1))
	return diff
}

func transparentImage(width int, height int) *image.RGBA {
	transparentColor := image.Transparent
	background := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(background, background.Bounds(), transparentColor, image.Point{0, 0}, draw.Src)
	return background
}

func readFile() (image.Image, string) {
	fileToConvert := "cloudformation"
	if len(os.Args) >= 2 {
		fileToConvert = os.Args[1]
	}

	file, err := os.Open(fileToConvert + ".jpeg")
	check(err)
	defer file.Close()
	img, _, err := image.Decode(file)
	check(err)

	return img, fileToConvert
}

type componentPixel struct {
	pixel     color.Color
	component int
	visited   bool
}

func (c *componentPixel) visit(pixel color.Color, component int) {
	c.pixel = pixel
	c.component = component
	c.visited = true
}

func dfs(col int, row int, width int, height int, matrix []componentPixel, img image.Image, background color.Color, component int) int {
	var stack [][2]int
	stack = append(stack, [2]int{col, row})
	var col_row [2]int

	count := 0
	neighbors := [3]int{-1, 0, 1}
	for len(stack) > 0 {
		col_row, stack = stack[len(stack)-1], stack[:len(stack)-1]
		col = col_row[0]
		row = col_row[1]

		if col >= width || col < 0 || row >= height || row < 0 {
			continue
		}

		if matrix[row*width+col].visited {
			continue
		}
		pixel := img.At(col, row)
		if colorDiff(pixel, background) < 10000 {
			// this is the background, don't want this component
			continue
		}

		matrix[row*width+col].visit(pixel, component)

		for i := 0; i < len(neighbors); i++ {
			for j := 0; j < len(neighbors); j++ {
				stack = append(stack, [2]int{col + neighbors[i], row + neighbors[j]})
			}
		}

		count += 1
	}

	return count
}

func findIcon(width int, height int, background color.Color, img image.Image) ([]componentPixel, int) {
	/*
	 * find connected components
	 * find connected component with highest numPixels
	 * only add that component to the transparent list
	 * may be able to create a more efficient "adjcency list" instead of 2d matrix
	 	* loop over and draw only the actual component instead of "searching" for a non-transparent pixel
	*/
	matrix := make([]componentPixel, width*height)

	components := 0
	maxComponent := 0
	maxComponentPixelCount := 0

	for i := 0; i < width; i++ {
		for j := 0; j < height; j++ {
			componentPixelCount := dfs(i, j, width, height, matrix, img, background, components)
			if componentPixelCount > 0 {
				if componentPixelCount > maxComponentPixelCount {
					maxComponentPixelCount = componentPixelCount
					maxComponent = components
				}
				components += 1
			}
		}
	}

	fmt.Println("Iconcomponent Pixelcount", maxComponentPixelCount)
	return matrix, maxComponent
}

/*
features
* make width and height only the actual icon instead of original photo
* cli argument for the 10000 number, find a name for this
*/

func main() {
	img, fileName := readFile()

	backgroundWidth := img.Bounds().Dx()
	backgroundHeight := img.Bounds().Dy()
	backgroundColor := findBackgroundColor(img, backgroundWidth, backgroundHeight)
	matrix, iconComponent := findIcon(backgroundWidth, backgroundHeight, backgroundColor, img)

	background := transparentImage(backgroundWidth, backgroundHeight)

	for i := 0; i < backgroundWidth; i++ {
		for j := 0; j < backgroundHeight; j++ {
			pixel := matrix[j*backgroundWidth+i]
			if pixel.pixel != nil && pixel.component == iconComponent {
				background.Set(i, j, pixel.pixel)
			}
		}
	}

	buf := new(bytes.Buffer)
	err := png.Encode(buf, background)
	check(err)

	err = ioutil.WriteFile(fileName+".png", buf.Bytes(), 0644)
	check(err)
}
