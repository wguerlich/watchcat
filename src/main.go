package main

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"io/ioutil"
	"log"
	"net/http"
	"time"
	//colorful "github.com/lucasb-eyer/go-colorful"
)

var maskImage *image.Image
var minChange float32

var threshold float32 = 0.8

func main() {
	maskFile := flag.String("mask", "mask.jpg", "image mask")
	snapshot := flag.String("snapshot", "http://example.com/snapshot", "snapshot url")
	upload := flag.String("upload", "http://example.com/upload", "upload url")
	mc := flag.Float64("minChange", 0.3, "minimum change")
	flag.Parse()
	mask, err := ioutil.ReadFile(*maskFile)
	minChange = float32(*mc)

	if err != nil {
		log.Fatal(err)
	}
	reader := bytes.NewReader(mask)
	img, err := jpeg.Decode(reader)
	if err != nil {
		log.Fatal(err)
	}
	maskImage = &img

	fmt.Printf("Hello\n")
	go processImages()
	go processUploads(*upload)
	fetchLoop(*snapshot)
}

var imgQueue = make(chan *imageHolder, 5)
var uploadQueue = make(chan *imageHolder, 100)
var backQueue = make(chan *imageHolder, 50)

func fetchLoop(url string) {
	for true {
		fmt.Println("xxx")
		payload := fetchURL(url)
		if payload != nil {
			ih := decodeImage(payload)
			ih.ts = time.Now()
			imgQueue <- ih

			time.Sleep(1 * time.Second)
		} else {
			time.Sleep(10 * time.Second)
		}
	}
}

type imageHolder struct {
	image  *image.Image
	source []byte
	ts     time.Time
	//uploaded bool
}

func decodeImage(source []byte) *imageHolder {
	var ih imageHolder
	ih.source = source
	reader := bytes.NewReader(source)
	img, err := jpeg.Decode(reader)
	if err == nil {
		ih.image = &img
	} else {
		log.Println(err)
		return nil
	}
	return &ih
}

func fetchURL(url string) []byte {
	resp, err := http.Get(url)
	if err == nil {
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			fmt.Println(resp.Status)
			return nil
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			return body
		}
		log.Println(err)
	} else {
		log.Println(err)
	}
	return nil
}

func processUploads(url string) {
	for true {
		ih := <-uploadQueue
		fmt.Printf("upload")

		resp, err := http.Post(url, "image/jpeg", bytes.NewReader(ih.source))

		if err == nil {
			fmt.Println(resp.Status)
		} else {
			log.Println(err)
		}

	}
}

func processImages() {
	var ih0 *imageHolder

	step := 10
	postCount := 0

	for true {
		ih1 := <-imgQueue

		if ih0 != nil {
			img0 := *ih0.image
			img1 := *ih1.image

			width := img0.Bounds().Dx()
			height := img0.Bounds().Dy()

			count := 0
			changed := 0
			for x := 0; x < width; x += step {
				for y := 0; y < height; y += step {

					cm := (*maskImage).At(x, y)

					r, g, b, _ := cm.RGBA()

					if r <= 512 && g <= 512 && b <= 512 {
						c0 := img0.At(x, y)
						c1 := img1.At(x, y)

						if colorChanged(c0, c1) {
							changed++
						}
						count++
					}

				}
			}

			changeRate := float32(changed) / float32(count)

			fmt.Printf("changerate %d\n", changeRate)

			if changeRate > minChange {
				postCount = 10
				for len(backQueue) > 0 {
					uploadQueue <- <-backQueue
				}

				uploadQueue <- ih1
			} else {
				if postCount > 0 {
					postCount--
					uploadQueue <- ih1
				} else {
					backQueue <- ih1
					if len(backQueue) > 10 {
						<-backQueue
					}
				}
			}

			ih0.image = nil
		}

		ih0 = ih1
	}
}

func colorChanged(c0, c1 color.Color) bool {
	r0, g0, b0, _ := c0.RGBA()
	r1, g1, b1, _ := c1.RGBA()

	var rr, rg, rb float32

	if r0 < r1 {
		rr = float32(r0) / float32(r1)
	} else {
		rr = float32(r1) / float32(r0)
	}

	if g0 < g1 {
		rg = float32(g0) / float32(g1)
	} else {
		rg = float32(g1) / float32(g0)
	}

	if b0 < b1 {
		rb = float32(b0) / float32(b1)
	} else {
		rb = float32(b1) / float32(b0)
	}

	if rr < threshold {
		return true
	}
	if rg < threshold {
		return true
	}
	if rb < threshold {
		return true
	}

	return false
}
