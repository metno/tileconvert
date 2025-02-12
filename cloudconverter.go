package main

import (
	"fmt"
	"image"
	"io"
	"log"
	"net/http"
	"strings"

	"image/color"
	"image/png"

	_ "golang.org/x/image/webp"
)

func main() {
	converter := CloudConverter{
		upstream: "https://beta.yr-maps.met.no",
	}
	http.Handle("/api/", &converter)
	log.Println("listening on port 8080")
	log.Fatalln(http.ListenAndServe(":8080", nil))
}

// convertClouds converts a cloud tile to a simple black-and-white tile, with pure black or pure white pixels.
func convertClouds(img image.Image, out io.Writer) error {
	bounds := img.Bounds()
	outputImage := image.NewGray(bounds)

	// Loop over each pixel in the input image,
	// and write a calculated value to the corresponding pixel in the output image:
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {

			cloudCoverInPercent := asPercent(img.At(x, y))

			// Simple transformation.
			// If cloud cover in this area (pixel) is larger than 50%,
			// set the output pixel to white (Y=0xff).
			// Otherwise set the pixel value to black (Y=0).
			pixelColor := color.Gray{Y: 0}
			if cloudCoverInPercent > 50 {
				pixelColor.Y = 0xff
			}

			outputImage.Set(x, y, pixelColor)
		}
	}

	return png.Encode(out, outputImage)
}

func asPercent(c color.Color) float64 {
	// Extract byte value from red channel
	r, _, _, _ := c.RGBA()
	// Get the value represented by the byte.
	// In this case each byte increment represents an increase by 0.5 percent points.
	//
	// The reason we right-shift by eigth is that the color is alpha-premultiplied,
	// and we know that alpha is 255.
	// This is specific to the golang standard library.
	// See https://pkg.go.dev/image/color#Color for more info about this.
	return float64(r>>8) / 2
}

// CloudConverter implements a very simple http server,
// proxying another tileserver, and serving modified images.
type CloudConverter struct {
	upstream string
}

// ServeHTTP implements the http serving part, proxying another server, and calling convertClouds when needed.
func (cc *CloudConverter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get(cc.upstream + r.URL.Path)
	if err != nil {
		log.Println(err)
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		http.Error(w, http.StatusText(resp.StatusCode), resp.StatusCode)
		return
	}

	// Handle requests for other paths than images
	if !strings.HasSuffix(r.URL.Path, ".png") && !strings.HasSuffix(r.URL.Path, ".webp") {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		b, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Println(err)
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}
		out := strings.Replace(string(b), cc.upstream, "http://localhost:8080", -1)
		fmt.Fprint(w, out)
		return
	}

	if strings.Contains(r.URL.Path, "cloud") {
		img, _, err := image.Decode(resp.Body)
		if err != nil {
			log.Println(err)
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		if err := convertClouds(img, w); err != nil {
			log.Println(err)
		}
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	io.Copy(w, resp.Body)
}
