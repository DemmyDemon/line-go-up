package main

import (
	"embed"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"os"

	"github.com/DemmyDemon/line-go-up/eodhd"
	"github.com/DemmyDemon/line-go-up/labelimage"
	"github.com/golang/freetype/truetype"
)

const (
	refPoint = 237.000
	ISIN     = "SE0001192618"
)

var (
	BadColor  = color.RGBA{128, 60, 60, 255}
	GoodColor = color.RGBA{60, 128, 60, 255}
)

var (
	//go:embed res
	res embed.FS
)

func main() {

	font, err := ReadFont(res, "res/White Rabbit.ttf")
	if err != nil {
		fmt.Printf("Unable to load font: %s\n", err)
	}

	http.HandleFunc("/Handelsbanken_Multi_Asset_50.png",
		func(w http.ResponseWriter, r *http.Request) {
			getImage(w, r, font)
		})
	err = http.ListenAndServe(":2468", nil)
	if err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			fmt.Println("Server closed")
		} else {
			fmt.Printf("Server error: %s\n", err)
			os.Exit(1)
		}
	}
}

func ReadFont(fs embed.FS, fontPath string) (*truetype.Font, error) {

	fontBytes, err := fs.ReadFile(fontPath)
	if err != nil {
		return nil, fmt.Errorf("reading font: %w", err)
	}

	f, err := truetype.Parse(fontBytes)
	if err != nil {
		return nil, fmt.Errorf("parsing font: %w", err)
	}

	return f, nil

}

func getImage(w http.ResponseWriter, r *http.Request, font *truetype.Font) {
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

	lines := []string{
		result.PreviousCloseDate,
		fmt.Sprintf("%s%0.3f", result.Currency, result.PreviousClose),
		offsetPercent,
	}

	img := labelimage.CreateWithFont(image.Rect(0, 0, 192, 64), font, textColor, lines, true, true)
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
