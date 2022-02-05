package main

type UnionFindItem struct {
	parent          *UnionFindItem
	size            int
	totalPixels     int
	totalDimensions [4]int
}

type UnionFind struct {
	root []*UnionFindItem
}

// New returns an initialized list of size
func NewUnionFind(components, chunks int, chunkComponentDimensions []map[int]chunkArea) *UnionFind {
	return new(UnionFind).init(components, chunks, chunkComponentDimensions)
}

// Constructor initializes root and size arrays
func (uf *UnionFind) init(componentNum int, chunks int, chunkComponentDimensions []map[int]chunkArea) *UnionFind {
	uf = new(UnionFind)
	uf.root = make([]*UnionFindItem, componentNum)

	for chunk := 0; chunk < chunks; chunk++ {
		for k, v := range chunkComponentDimensions[chunk] {
			item := UnionFindItem{
				totalPixels:     v.pixels,
				totalDimensions: v.dimensions,
				size:            1,
			}
			item.parent = &item

			uf.root[k] = &item
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

// Union connects p and q by finding their roots and comparing their respective
// size arrays to keep the tree flat
func (uf *UnionFind) Union(p int, q int) {
	qRoot := uf.Root(q)
	pRoot := uf.Root(p)

	if qRoot.parent == pRoot.parent {
		return
	}

	if qRoot.size < pRoot.size {
		qRoot.parent = pRoot.parent
		pRoot.size += qRoot.size
		pRoot.parent.totalPixels += qRoot.totalPixels
		pRoot.parent.totalDimensions = mergeDimensions(pRoot.totalDimensions, qRoot.totalDimensions)
		pRoot.totalDimensions = pRoot.parent.totalDimensions
		pRoot.totalPixels = pRoot.parent.totalPixels
		pRoot.parent.size = pRoot.size
	} else {
		pRoot.parent = qRoot.parent
		qRoot.size += pRoot.size
		qRoot.parent.totalPixels += pRoot.totalPixels
		qRoot.parent.totalDimensions = mergeDimensions(pRoot.totalDimensions, qRoot.totalDimensions)
		qRoot.totalDimensions = qRoot.parent.totalDimensions
		qRoot.totalPixels = qRoot.parent.totalPixels
		qRoot.parent.size = qRoot.size
	}
}

// Root or Find traverses each parent element while compressing the
// levels to find the root element of p
// If we attempt to access an element outside the array it returns -1
func (uf *UnionFind) Root(p int) *UnionFindItem {
	item := uf.root[p]
	if item == nil {
		return nil
	}

	for *item != *item.parent {
		item.parent = item.parent.parent
		*item = *item.parent
	}

	return item
}

// Root or Find
func (uf *UnionFind) Find(p int) *UnionFindItem {
	return uf.Root(p)
}

// Check if items p,q are connected
func (uf *UnionFind) Connected(p, q int) bool {
	pRoot := uf.Root(p)
	if pRoot == nil {
		return false
	}
	qRoot := uf.Root(q)
	if qRoot == nil {
		return false
	}
	return pRoot.parent == qRoot.parent
}
