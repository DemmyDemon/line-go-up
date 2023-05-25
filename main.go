package main

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"os"

	"github.com/DemmyDemon/line-go-up/eodhd"
	"github.com/DemmyDemon/line-go-up/labelimage"
)

const (
	refPoint = 237.000
	ISIN     = "SE0001192618"
)

var (
	BadColor  = color.RGBA{128, 60, 60, 255}
	GoodColor = color.RGBA{60, 128, 60, 255}
)

func main() {
	http.HandleFunc("/Handelsbanken_Multi_Asset_50.png", getImage)
	err := http.ListenAndServe(":2468", nil)
	if err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			fmt.Println("Server closed")
		} else {
			fmt.Printf("Server error: %s\n", err)
			os.Exit(1)
		}
	}
}

func getImage(w http.ResponseWriter, r *http.Request) {
	results, err := eodhd.Search(ISIN, os.Getenv("API_TOKEN"))
	if err != nil {
		fmt.Printf("Error fetching oedhd data: %s\n", err)
		w.Header().Add("Content-Type", "text/plain")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	if len(results) == 0 {
		fmt.Println("No data found, somehow")
		w.Header().Add("Content-Type", "text/plain")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Malformed data"))
		return
	}
	result := results[0]

	offsetPercent, good := getOffsetPercent(result.PreviousClose)
	textColor := BadColor
	if good {
		textColor = GoodColor
	}

	// img := labelimage.Create(image.Rect(0, 0, 128, 32), color.RGBA{60, 100, 60, 255}, fmt.Sprintf("%0.4f", result.PreviousClose), true, true)
	img := labelimage.Create(image.Rect(0, 0, 128, 32), textColor, offsetPercent, true, true)
	w.WriteHeader(http.StatusOK)
	err = png.Encode(w, img)
	if err != nil {
		fmt.Printf("Error writing PNG: %s\n", err)
		w.Header().Add("Content-Type", "text/plain")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Add("Content-Type", "image/png")
}

func getOffsetPercent(currentValue float64) (string, bool) {
	diff := currentValue - refPoint
	offset := (diff / refPoint) * 100
	return fmt.Sprintf("%+0.3f%%", offset), offset > 0
}
