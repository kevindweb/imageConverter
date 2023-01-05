package main

import (
	"image"
	"image/color"
	"math"
)

// Legacy code - either not chunked or not threaded

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

func findIconInChunk(componentInx int,
	startRow int, startCol int, endRow int, endCol int, width int, height int,
	matrix []componentPixel, background [3]uint32) (int, map[int]chunkArea) {

	componentDimensionMap := make(map[int]chunkArea)
	reuseComponent := false
	currComponent := componentInx - 1

	for j := startRow; j < endRow; j++ {
		for i := startCol; i < endCol; i++ {
			if !reuseComponent {
				// get a unique inx amongst all chunks
				// only if we actually had a unique component last time
				currComponent += 1
			}

			componentPixelCount, dimensions := dfs(i, j, width, height, matrix,
				background, currComponent, startRow, endRow, startCol, endCol)
			if componentPixelCount > 0 {
				// potential icon
				componentDimensionMap[currComponent] = chunkArea{
					dimensions: dimensions,
					pixels:     componentPixelCount,
				}
				reuseComponent = false
			} else {
				reuseComponent = true
			}
		}
	}

	return currComponent + 1, componentDimensionMap
}

// findIcon takes an image and searches for the connected components
// it then returns the component (and dimensions) with maximum pixel count
func findIconChunk(width int, height int, matrix []componentPixel,
	background [3]uint32, chunks int) ([4]int, map[int]bool) {
	/*
	   * split the image into equal size chunks
	   * in each chunk, find the connected components
	   * find where the chunks intersect, then merge
	   * return the maximum merged icon's dimensions and details
	       * save space by only grabbing the required height/width instead of entire image
	*/

	// want same number of chunks in the height and width
	chunkRows := int(math.Sqrt(float64(chunks)))
	chunkRowSize := height / chunkRows
	chunkColSize := width / chunkRows

	chunkComponentDimensions := make([]map[int]chunkArea, chunks)

	chunk := 0
	componentNum := 0

	for row := 0; row < chunkRows; row++ {
		endRow := (row + 1) * chunkRowSize
		if row == chunkRows-1 && endRow != height {
			// chunks might not have been evenly distributed
			endRow = height
		}
		for col := 0; col < chunkRows; col++ {
			// each chunk updates the matrix in place
			endCol := (col + 1) * chunkColSize
			if col == chunkRows-1 && endCol != width {
				// chunk columns weren't even
				endCol = width
			}

			componentNum, chunkComponentDimensions[chunk] = findIconInChunk(componentNum, row*chunkRowSize, col*chunkColSize,
				endRow, endCol, width, height, matrix, background)
			chunk++
		}
	}

	return handleChunkMerge(componentNum, chunks, chunkRows, chunkRowSize, chunkColSize, width, height, chunkComponentDimensions, matrix, background)
}

// findIcon takes an image and searches for the connected components
// it then returns the component (and dimensions) with maximum pixel count
func findIcon(width int, height int, matrix []componentPixel,
	background [3]uint32) ([4]int, map[int]bool) {
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
				background, components, 0, height, 0, width)
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

	return componentDimensions[maxComponent-1], map[int]bool{maxComponent: true}
}

// oldRunIcon is the main entrypoint into the algorithm
// when given an image, it finds the background, components
// and returns the transparent png result
func oldRunIcon(img image.Image, chunks int, threaded bool) *image.RGBA {
	var pixelMatrix []componentPixel
	var backgroundColor [3]uint32
	var iconDimensions [4]int
	var iconComponentMap map[int]bool

	backgroundWidth := img.Bounds().Dx()
	backgroundHeight := img.Bounds().Dy()

	if chunks > 0 {
		// run by chunking
		if threaded {
			// run chunks in parallel
			backgroundColor, pixelMatrix = findBackgroundColorThreaded(img, backgroundWidth, backgroundHeight, chunks)
			iconDimensions, iconComponentMap = findIconChunkThread(backgroundWidth,
				backgroundHeight, pixelMatrix, backgroundColor, chunks)
		} else {
			backgroundColor, pixelMatrix = findBackgroundColor(img, backgroundWidth, backgroundHeight)
			iconDimensions, iconComponentMap = findIconChunk(backgroundWidth,
				backgroundHeight, pixelMatrix, backgroundColor, chunks)
		}
	} else {
		backgroundColor, pixelMatrix = findBackgroundColor(img, backgroundWidth, backgroundHeight)
		iconDimensions, iconComponentMap = findIcon(backgroundWidth,
			backgroundHeight, pixelMatrix, backgroundColor)
	}

	return buildTransparentImage(pixelMatrix, iconDimensions, iconComponentMap, backgroundWidth)
}
