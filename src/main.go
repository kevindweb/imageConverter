package main

import (
	"fmt"
	"net/http"
	"time"

	"image"
	"image/jpeg"

	"github.com/syumai/workers"
)

func setupCORS(w *http.ResponseWriter, req *http.Request) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

func writeErr(code int, err error, w http.ResponseWriter) {
	w.WriteHeader(code)
	w.Write([]byte(err.Error()))
}

func main() {
	http.HandleFunc("/iconify", func(w http.ResponseWriter, req *http.Request) {
		setupCORS(&w, req)

		// Parse our multipart form, 10 << 20 specifies a maximum
		// upload of 10 MB files.
		req.ParseMultipartForm(10 << 20)

		file, _, err := req.FormFile("myFile")
		if err != nil {
			writeErr(http.StatusBadRequest, err, w)
			return
		}
		defer file.Close()

		// _ = time.Now()
		// _ = t.Format("2006-01-02-15:04:05")
		time.Sleep(time.Second)

		fmt.Println("test")
		w.Write([]byte("bad times"))
		r := 5
		for i := 0; i < 1000000000; i++ {
			r += 4
		}

		// _, _, err = image.Decode(file)
		// // img, _, err := image.Decode(file)
		// if err != nil {
		// 	writeErr(http.StatusBadRequest, err, w)
		// 	return
		// }

		// background := runIcon(img, 64, true)
		// buf := new(bytes.Buffer)
		// enc := &png.Encoder{
		// 	CompressionLevel: png.NoCompression,
		// }

		// err = enc.Encode(buf, background)
		// if err != nil {
		// 	writeErr(http.StatusInternalServerError, err, w)
		// 	return
		// }

		// w.Write(buf.Bytes())
		// time.Now().Sub(t)
		// fmt.Println("Time until", thing.Milliseconds())
		w.Write([]byte("here"))
	})
	workers.Serve(nil) // use http.DefaultServeMux
	// http.ListenAndServe(":8090", nil)
}

func init() {
	image.RegisterFormat("jpeg", "jpeg", jpeg.Decode, jpeg.DecodeConfig)
}
