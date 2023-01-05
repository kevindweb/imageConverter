package main

import (
	"image"
	"image/color"
	"math"
	"sync/atomic"
)

type chunkArea struct {
	dimensions [4]int
	pixels     int
}

func findBackgroundColorChunk(img image.Image, matrix []componentPixel, colorMapChannel chan map[color.Color]int,
	startRow, startCol, endRow, endCol, width, height int) {
	colorCount := make(map[color.Color]int)

	for j := startRow; j < endRow; j++ {
		for i := startCol; i < endCol; i++ {
			pixel := img.At(i, j)
			pixelCount := 1
			matrix[j*width+i].pixel = pixel

			if count, ok := colorCount[pixel]; ok {
				pixelCount += count
			}

			colorCount[pixel] = pixelCount
		}
	}

	colorMapChannel <- colorCount
}

func findBackgroundColorThreaded(img image.Image, width, height, chunks int) ([3]uint32, []componentPixel) {
	matrix := make([]componentPixel, width*height)
	colorCount := make(map[color.Color]int)
	c1 := make(chan map[color.Color]int)

	var popularColor color.Color
	maxPixelCount := 0

	// want same number of chunks in the height and width
	chunkRows := int(math.Sqrt(float64(chunks)))
	chunkRowSize := height / chunkRows
	chunkColSize := width / chunkRows

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

			go findBackgroundColorChunk(img, matrix, c1, row*chunkRowSize, col*chunkColSize,
				endRow, endCol, width, height)
		}
	}

	for chunk := 0; chunk < chunks; chunk++ {
		colorMap := <-c1

		for pixel, count := range colorMap {
			if globalCount, ok := colorCount[pixel]; ok {
				count += globalCount
			}
			if count > maxPixelCount {
				popularColor = pixel
				maxPixelCount = count
			}
			colorCount[pixel] = count
		}
	}

	red, green, blue, _ := popularColor.RGBA()
	return [3]uint32{red, green, blue}, matrix
}

// dfs iteratively adds neighbors to the component list to find the entire
// connected icon - can't use recursive, causes stackoverflow with num pixels
func dfs(col int, row int, width int, height int, matrix []componentPixel,
	background [3]uint32, component int, top, bottom, left, right int) (int, [4]int) {
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

		if row < top || row >= bottom || col < left || col >= right {
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

func buildTransparentImage(matrix []componentPixel, iconDimensions [4]int,
	iconComponents map[int]bool, backgroundWidth int) *image.RGBA {

	topPixel := iconDimensions[0]
	bottomPixel := iconDimensions[1]
	leftPixel := iconDimensions[2]
	rightPixel := iconDimensions[3]

	iconWidth := rightPixel - leftPixel
	iconHeight := bottomPixel - topPixel
	background := image.NewRGBA(image.Rect(0, 0, iconWidth, iconHeight))

	for j := 0; j < iconHeight; j++ {
		for i := 0; i < iconWidth; i++ {
			// accessing 2d matrix as 1d array https://stackoverflow.com/a/2151141
			pixel := matrix[(j+topPixel)*backgroundWidth+(i+leftPixel)]
			if _, ok := iconComponents[pixel.component]; ok {
				// if this pixel is in any of the "icon" components, set the pixel
				background.Set(i, j, pixel.pixel)
			}
		}
	}

	return background
}

func findIconInChunkThreaded(componentInx *uint64,
	startRow int, startCol int, endRow int, endCol int, width int, height int,
	matrix []componentPixel, background [3]uint32, channel chan map[int]chunkArea) {

	componentDimensionMap := make(map[int]chunkArea)
	reuseComponent := false
	var currComponent int

	for j := startRow; j < endRow; j++ {
		for i := startCol; i < endCol; i++ {
			if !reuseComponent {
				// get a unique inx amongst all chunks
				// only if we actually had a unique component last time
				currComponent = int(atomic.AddUint64(componentInx, 1))
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

	channel <- componentDimensionMap
}

func mergeOnEitherSideByRow(matrix []componentPixel, background [3]uint32, unionFindArray *UnionFind, row, width int) {
	for col := 0; col < width; col++ {
		lowerPixelComponent := matrix[row*width+col].component
		upperPixelComponent := matrix[(row-1)*width+col].component
		if lowerPixelComponent > 0 {
			if upperPixelComponent > 0 {
				// these components should merge
				unionFindArray.Union(upperPixelComponent, lowerPixelComponent)
			} else if col != width-1 {
				// check upper diagonal
				upperDiagonalPixelComponent := matrix[(row-1)*width+(col+1)].component
				if upperDiagonalPixelComponent > 0 {
					unionFindArray.Union(upperDiagonalPixelComponent, lowerPixelComponent)
				}
			}
		}
	}
}

func mergeOnEitherSideByCol(matrix []componentPixel, background [3]uint32, unionFindArray *UnionFind, col, height, width int) {
	for row := 0; row < height; row++ {
		rightPixelComponent := matrix[row*width+col].component
		leftPixelComponent := matrix[row*width+(col-1)].component
		if rightPixelComponent > 0 {
			if leftPixelComponent > 0 {
				// these components should merge
				unionFindArray.Union(rightPixelComponent, leftPixelComponent)
			} else if row != height-1 {
				// check upper diagonal
				leftDiagonalPixelComponent := matrix[(row+1)*width+(col-1)].component
				if leftDiagonalPixelComponent > 0 {
					unionFindArray.Union(leftDiagonalPixelComponent, rightPixelComponent)
				}
			}
		}
	}
}

func handleChunkMerge(componentNum, chunks, chunkRows, chunkRowSize, chunkColSize, width, height int, chunkComponentDimensions []map[int]chunkArea, matrix []componentPixel, background [3]uint32) ([4]int, map[int]bool) {
	// merge chunks together
	// run union find on the merge chunks
	// run through the intersections of the chunks only (ignore edges of picture as there's no intersections)
	unionFindParents := NewUnionFind(componentNum, chunks, chunkComponentDimensions)

	// merge chunks by the intersections
	// ignore the outsides of the image because they won't have any merging
	for rowIntersection := 1; rowIntersection < chunkRows; rowIntersection++ {
		mergeOnEitherSideByRow(matrix, background, unionFindParents, rowIntersection*chunkRowSize, width)
	}

	for colIntersection := 1; colIntersection < chunkRows; colIntersection++ {
		mergeOnEitherSideByCol(matrix, background, unionFindParents, colIntersection*chunkColSize, height, width)
	}

	maxComponentInx := 0
	maxComponentPixelCount := 0

	// when two icon components merge, the entire icon dimensions need to merge
	var maxComponentDimensions [4]int

	// start at 1, component 0 is reserved and unused
	for i := 1; i < len(unionFindParents.root); i++ {
		unionFindParent := unionFindParents.area[unionFindParents.Root(i)]
		if unionFindParent.totalPixels > maxComponentPixelCount {
			maxComponentInx = i
			maxComponentPixelCount = unionFindParent.totalPixels
			maxComponentDimensions = unionFindParent.totalDimensions
		}
	}

	// group chunk components together after merge for final component check
	// eg chunk1 had component65 that merged with chunk2's component800
	// the result icon's component hashset is {65, 800}
	maxComponentSet := make(map[int]bool)
	maxComponentSet[maxComponentInx] = true

	for i := 1; i < len(unionFindParents.root); i++ {
		if unionFindParents.Connected(i, maxComponentInx) {
			// this component belongs to the maximum icon's set
			maxComponentSet[i] = true
		}
	}

	return maxComponentDimensions, maxComponentSet
}

// it then returns the component (and dimensions) with maximum pixel count
func findIconChunkThread(width int, height int, matrix []componentPixel,
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

	// will serve as atomic thread-safe counter to avoid component inx collisions
	var componentNum uint64 = 0

	chunkComponentDimensions := make([]map[int]chunkArea, chunks)

	c1 := make(chan map[int]chunkArea)

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

			go findIconInChunkThreaded(&componentNum, row*chunkRowSize, col*chunkColSize,
				endRow, endCol, width, height, matrix, background, c1)
		}
	}

	for chunk := 0; chunk < chunks; chunk++ {
		data := <-c1
		chunkComponentDimensions[chunk] = data
	}

	return handleChunkMerge(int(componentNum), chunks, chunkRows, chunkRowSize, chunkColSize, width, height, chunkComponentDimensions, matrix, background)
}

// runIcon is the main entrypoint into the algorithm
// when given an image, it finds the background, components
// and returns the transparent png result
func runIcon(img image.Image, chunks int, threaded bool) *image.RGBA {
	var pixelMatrix []componentPixel
	var backgroundColor [3]uint32
	var iconDimensions [4]int
	var iconComponentMap map[int]bool

	backgroundWidth := img.Bounds().Dx()
	backgroundHeight := img.Bounds().Dy()

	backgroundColor, pixelMatrix = findBackgroundColorThreaded(img, backgroundWidth, backgroundHeight, chunks)
	iconDimensions, iconComponentMap = findIconChunkThread(backgroundWidth,
		backgroundHeight, pixelMatrix, backgroundColor, chunks)

	return buildTransparentImage(pixelMatrix, iconDimensions, iconComponentMap, backgroundWidth)
}
