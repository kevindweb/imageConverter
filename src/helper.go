package main

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"io/ioutil"
	"os"
	"time"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

// readFile takes the input string and reads the image from ./images into memory
func readFile(fileToConvert string) image.Image {
	file, err := os.Open("../images/" + fileToConvert + ".jpeg")
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

	err = ioutil.WriteFile("../icons/"+fileName+".png", buf.Bytes(), 0644)
	check(err)
}

// elapsed prints the function duration
// to be used like: defer elapsed("what")()
func elapsed(what string) func() {
	start := time.Now()
	return func() {
		fmt.Printf("%s took %v\n", what, time.Since(start))
	}
}
