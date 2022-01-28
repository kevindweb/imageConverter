package main

/*
split image into M equally sized chunks

go routine to execute findIcon on each chunk
- start without multi-threading but convert once tests are valid

merge each chunk by scanning the edges for boundary icons
- if there's a hit, add these two components together in an array
    - merge the height and width
- find the biggest combined component

run build image on merged components
- check if pixel component is in the list of grouped icon chunks

examples:
https://github.com/bonej-org/BoneJ2/blob/da5aa63cdc15516605e8dcb77458eb34b0f00b85/Legacy/bonej/src/main/java/org/bonej/plugins/ConnectedComponents.java#L540
https://github.com/opencv/opencv/issues/7270

It's called multithreaded connected component labeling

*/

import (
	"image"
	"math"
)

// dfs iteratively adds neighbors to the component list to find the entire
// connected icon - can't use recursive, causes stackoverflow with num pixels
func dfsOptimal(col int, row int, width int, height int, matrix []componentPixel,
	background [3]uint32, component int) (int, [4]int) {
	stack := [][2]int{{col, row}}
	var col_row [2]int

	count := 0
	stackPointer := 0
	neighbors := [3]int{-1, 0, 1}
	neighborLength := len(neighbors)

	firstPixel := true
	var pixelSpace [4]int

	for stackPointer > -1 {
		col_row = stack[stackPointer]
		stackPointer--
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
				if i == 1 && j == 1 {
					// don't redo our own (col, row) again
					continue
				}

				stackPointer++
				rc := [2]int{neighborColumn, row + neighbors[j]}
				// try to add to original stack w/o append
				if stackPointer < len(stack) {
					stack[stackPointer] = rc
				} else {
					stack = append(stack, rc)
				}
			}
		}

		count += 1
	}

	return count, pixelSpace
}

type chunkArea struct {
	dimensions [4]int
	pixels     int
}

func findIconInChunk(componentInx int,
	startRow int, startCol int, endRow int, endCol int, width int, height int,
	matrix []componentPixel, background [3]uint32) map[int]chunkArea {

	componentDimensionMap := make(map[int]chunkArea)

	for j := startRow; j < endRow; j++ {
		for i := startCol; i < endCol; i++ {
			componentPixelCount, dimensions := dfsOptimal(i, j, width, height, matrix,
				background, componentInx)
			if componentPixelCount > 0 {
				componentDimensionMap[componentInx] = chunkArea{
					dimensions: dimensions,
					pixels:     componentPixelCount,
				}
				componentInx += 1
			}
		}
	}

	return componentDimensionMap
}

// findIcon takes an image and searches for the connected components
// it then returns the component (and dimensions) with maximum pixel count
func findIconOptimal(width int, height int, matrix []componentPixel,
	background [3]uint32) ([4]int, map[int]bool) {
	/*
	   * split the image into equal size chunks
	   * in each chunk, find the connected components
	   * find where the chunks intersect, then merge
	   * return the maximum merged icon's dimensions and details
	       * save space by only grabbing the required height/width instead of entire image
	*/

	chunks := 4
	// want same number of chunks in the height and width
	chunkRows := int(math.Sqrt(float64(chunks)))
	chunkRowSize := height / chunkRows
	chunkColSize := width / chunkRows

	// TODO: figure out out to make unique component ids without coordination between chunks/threads
	// some thread-safe hash generator or something?

	for row := 0; row < chunkRows; row++ {
		for col := 0; col < chunkRows; col++ {
			// each chunk updates the matrix in place
			chunkComponents := findIconInChunk(1, row*chunkRowSize, col*chunkColSize,
				(row+1)*chunkRowSize, (col+1)*chunkColSize, width, height, matrix, background)
		}
	}

	// when two icon components merge, the entire icon dimensions need to merge
	var maxComponentDimensions [4]int
	// this will "group" icons from different chunks
	// when they merge together
	// so we don't have to rewrite each pixels' component number
	maxComponentSet := make(map[int]bool)

	return maxComponentDimensions, maxComponentSet
}

// runIcon is the main entrypoint into the algorithm
// when given an image, it finds the background, components
// and returns the transparent png result
func runIconOptimal(img image.Image) *image.RGBA {
	backgroundWidth := img.Bounds().Dx()
	backgroundHeight := img.Bounds().Dy()

	backgroundColor, pixelMatrix := findBackgroundColor(img, backgroundWidth, backgroundHeight)
	iconDimensions, iconComponentMap := findIconOptimal(backgroundWidth,
		backgroundHeight, pixelMatrix, backgroundColor)

	return buildTransparentImage(pixelMatrix, iconDimensions, iconComponentMap, backgroundWidth)
}
