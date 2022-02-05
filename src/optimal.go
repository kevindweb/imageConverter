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
	"sync/atomic"
)

type chunkArea struct {
	dimensions [4]int
	pixels     int
}

// dfs iteratively adds neighbors to the component list to find the entire
// connected icon - can't use recursive, causes stackoverflow with num pixels
func dfsChunk(col int, row int, width int, height int, matrix []componentPixel,
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
			// if col >= width || col < 0 || row >= height || row < 0 {
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

func findIconInChunk(componentInx *uint64,
	startRow int, startCol int, endRow int, endCol int, width int, height int,
	matrix []componentPixel, background [3]uint32) map[int]chunkArea {

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

			componentPixelCount, dimensions := dfsChunk(i, j, width, height, matrix,
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

	return componentDimensionMap
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

	// use perfect square int chunks
	chunks := 36
	// want same number of chunks in the height and width
	chunkRows := int(math.Sqrt(float64(chunks)))
	chunkRowSize := height / chunkRows
	chunkColSize := width / chunkRows

	// will serve as atomic thread-safe counter to avoid component inx collisions
	var componentNum uint64 = 0

	chunkComponentDimensions := make([]map[int]chunkArea, chunks)
	currChunk := 0

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

			// TODO: use a channel instead for goroutine
			chunkComponentDimensions[currChunk] = findIconInChunk(&componentNum, row*chunkRowSize, col*chunkColSize,
				endRow, endCol, width, height, matrix, background)
			currChunk++
		}
	}

	// merge chunks together
	// run union find on the merge chunks
	// run through the intersections of the chunks only (ignore edges of picture as there's no intersections)
	unionFindParents := NewUnionFind(int(componentNum), chunks, chunkComponentDimensions)

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
		unionFindParent := unionFindParents.Root(i)
		if unionFindParent != nil && unionFindParent.totalPixels > maxComponentPixelCount {
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
			// if unionFindParents.root[i].parent == &maxUnionParent {
			// this component belongs to the maximum icon's set
			maxComponentSet[i] = true
		}
	}

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
