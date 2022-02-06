package main

// Code referenced from https://github.com/theodesp/unionfind/blob/master/unionfind.go

type UnionFindArea struct {
	totalPixels     int
	totalDimensions [4]int
}

type UnionFind struct {
	root []int
	size []int
	area []UnionFindArea
}

// New returns an initialized list of size
func NewUnionFind(components, chunks int, chunkComponentDimensions []map[int]chunkArea) *UnionFind {
	return new(UnionFind).init(components, chunks, chunkComponentDimensions)
}

func (uf *UnionFind) init(componentNum int, chunks int, chunkComponentDimensions []map[int]chunkArea) *UnionFind {
	uf.root = make([]int, componentNum)
	uf.size = make([]int, componentNum)
	uf.area = make([]UnionFindArea, componentNum)

	for chunk := 0; chunk < chunks; chunk++ {
		for k, v := range chunkComponentDimensions[chunk] {
			item := UnionFindArea{
				totalPixels:     v.pixels,
				totalDimensions: v.dimensions,
			}

			uf.root[k] = k
			uf.size[k] = 1
			uf.area[k] = item
		}
	}

	return uf
}

func min(num1, num2 int) int {
	if num1 < num2 {
		return num1
	}
	return num2
}

func max(num1, num2 int) int {
	if num1 < num2 {
		return num2
	}
	return num1
}

func mergeDimensions(dimension1, dimension2 [4]int) [4]int {
	return [4]int{
		min(dimension1[0], dimension2[0]),
		max(dimension1[1], dimension2[1]),
		min(dimension1[2], dimension2[2]),
		max(dimension1[3], dimension2[3]),
	}
}

// Union/merge smaller area into larger area
func (uf *UnionFind) mergeInto(larger int, smaller int) {
	rootLarge := uf.root[larger]
	rootSmall := uf.root[smaller]

	largeArea := uf.area[rootLarge]
	smallArea := uf.area[rootSmall]

	largeArea.totalDimensions = mergeDimensions(largeArea.totalDimensions, smallArea.totalDimensions)
	largeArea.totalPixels += smallArea.totalPixels
	uf.area[rootLarge] = largeArea
}

// Union connects p and q by finding their roots and comparing their respective
// size arrays to keep the tree flat
func (uf *UnionFind) Union(p int, q int) {
	qRoot := uf.Root(q)
	pRoot := uf.Root(p)

	if qRoot == pRoot {
		return
	}

	if uf.size[qRoot] < uf.size[pRoot] {
		uf.mergeInto(pRoot, qRoot)
		uf.root[qRoot] = uf.root[pRoot]
		uf.size[pRoot] += uf.size[qRoot]
	} else {
		uf.mergeInto(qRoot, pRoot)
		uf.root[pRoot] = uf.root[qRoot]
		uf.size[qRoot] += uf.size[pRoot]
	}
}

// Root or Find traverses each parent element while compressing the
// levels to find the root element of p
// If we attempt to access an element outside the array it returns -1
func (uf *UnionFind) Root(p int) int {
	if p > len(uf.root)-1 {
		return -1
	}

	for uf.root[p] != p {
		uf.root[p] = uf.root[uf.root[p]]
		p = uf.root[p]
	}

	return p
}

// Root or Find
func (uf *UnionFind) Find(p int) int {
	return uf.Root(p)
}

// Get the details of the component's root
func (uf *UnionFind) FindArea(p int) UnionFindArea {
	return uf.area[uf.Root(p)]
}

// Check if items p,q are connected
func (uf *UnionFind) Connected(p int, q int) bool {
	return uf.Root(p) == uf.Root(q)
}
