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
* Install golang and execute `go build && ./main`
    * The default image is "clownfish" 
    * You can run another in the images folder with `go run converter.go <image name>`
        * Example: `go run converter.go cloudformation`
        * The output will go into "icons"

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