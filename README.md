# ImageConverter

## Description
* The goal is to take a jpeg image and convert it to png
* It strips the "background" away to make it a transparent icon
* There is no user input, the program determines the background automatically

## Example
We want to input this jpeg image  
![Regular AWS lambda jpeg](/images/lambda.jpeg)  


And turn it into an icon with a transparent background like this  
![Exciting icon without background](/icons/lambda.png)  


## Running the program
* Install golang and execute `cd src && go build && ./src`
    * The default image is "clownfish" 
    * You can run another in the images folder with `./src <image-name>`
        * Example: `./src cloudformation` (ignore the file type)
        * The output will go into the "icons" directory

## Key algorithms
* Find background color - O(n) image lookup with hash table for the color:pixelCount
    * Find the most "popular" color (set as the background color)
* Find connected components - O(n) dfs to find separate connected components/potential icons
    * We loop through pixels and group them by neighbors, making sure to `visit` them only once
    * This process also "fits" the image by finding the component's dimensions
* Build transparent image - O(n) only add color of pixels if they are in the correct "component"
    * n is smaller here typically because we only process the icon dimensions (not the whole original)

(where n is the total number of pixels - width*height)

## Key points and features
* The algorithm performs especially well in jpegs that "should" be icons with semi-uniform background colors
* The output png has its' dimensions fixed to the icon to reduce the photo size

## Large JPEG Optimizations
Some images can be quite large, eg >20mb - the initial algorithm does no optimizations. There is another optimized version (run with a `chunk` parameter to `runIcon` if a speedup is required)

**Multithreaded connected component labeling**

Inspiration for this came from the following examples:
https://github.com/bonej-org/BoneJ2/blob/da5aa63cdc15516605e8dcb77458eb34b0f00b85/Legacy/bonej/src/main/java/org/bonej/plugins/ConnectedComponents.java#L540

https://github.com/opencv/opencv/issues/7270

The divide, conquer, and merge process goes as follows
1. Split the image into M equally sized chunks (M should be a perfect square)
2. Send the chunks to goroutines for parallel processing
3. Traverse along the edges of the chunks (by row/col) as seen in `mergeOnEitherSideBy<Col/Row>`
    * This means we don't need to rescan the entire image to "merge" the chunks
    * Try to find where one chunk's component "intersects" with an adjacent
    * Avoid checking the image edges (row0, col0) because there's no overlap
4. Find the biggest (in pixel count) icon after the merge
5. Add all adjacent components to a hashset `iconComponentMap`
6. Check if pixel's component is in the hashset, if it is, add to output png, else make that pixel transparent

**Data Structure**

The main part of this algorithm is using the Union-Find (disjoint set) to group the chunks

Why is this necessary
* Each chunk has to use unique component indices in order to effectively convey two components (and their pixels) are different
* In the "merge" process, we need an efficient way to say component1 and component2 are intersecting and should both be included in the result icon
* If we didn't use this data structure, we would need to go back through every pixel and set its' component index to the "max icon component" for the buildTransparentImage process

How it works
* [Union Find](https://en.wikipedia.org/wiki/Disjoint-set_data_structure) is used in a lot of graph processing (designed for Kruskal's minimum spanning tree)
* It efficiently groups components (indices) together using the Union function
* When ready, we use the Connected function to see if two components ever merged
* The algorithm uses path compression to speed up subsequent component accesses
    * Path compression means if we Union(2, 3), Union(3, 4), Union(4, 5) - Connected(2, 5) would need to access 3, 4, then 5 to see if they're connected. For effiency, after one Connected call, we make a direct reference from 2-5 to avoid checking 3 and 4 next time

### Multithreaded approach downfalls
* Setting up the chunks, and threading is quite expensive (in terms of time). Therefore, small images will most likely see a downgrade. Tuning is required to determine the lower bound on file size to see when the algorithm should perform best
* Every execution of the build shows the execution time, use this as a quick metric to determine which algorithm is best for specific images
