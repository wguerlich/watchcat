package main

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func main() {
	fmt.Printf("Hello\n")
	url := "http://192.168.2.232/cgi-bin/anv/images_cgi?channel=0&user=cam123&pwd=cam123"
	fetchLoop(url)
}

func fetchLoop(url string) {
	for true {
		fmt.Println("xxx")
		payload := fetchURL(url)
		if payload != nil {
			ih := decodeImage(payload)
			img := *ih.image

			fmt.Printf("%d %d", img.Bounds().Dx(), img.Bounds().Dy())
			time.Sleep(1 * time.Second)
		} else {
			time.Sleep(10 * time.Second)
		}
	}
}

type imageHolder struct {
	image  *image.Image
	source []byte
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
