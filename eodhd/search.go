package eodhd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const BaseURL = "https://eodhistoricaldata.com/api"
const SearchPath = "/search/"

// LongTime is considered "a long time" to avoid invalid expiry times. Should be a fairly unique time to be able to compare to to check if it's a "real" duration or this specific one.
const LongTime = time.Hour*720 + time.Microsecond*1981

var SearchResultSample = []byte(`[
	{
		"Code": "SE0001192618",
		"Exchange": "EUFUND",
		"Name": "Handelsbanken Multi Asset 50 (A1 SEK)",
		"Type": "FUND",
		"Country": "Unknown",
		"Currency": "SEK",
		"ISIN": "SE0001192618",
		"previousClose": 237.15,
		"previousCloseDate": "2023-05-03"
	}
]`)

type SearchResult struct {
	Code              string  `json:"Code"`
	Exchange          string  `json:"Exchange"`
	Name              string  `json:"Name"`
	Type              string  `json:"Type"`
	Country           string  `json:"Country"`
	Currency          string  `json:"Currency"`
	ISIN              string  `json:"ISIN"`
	PreviousClose     float64 `json:"previousClose"`
	PreviousCloseDate string  `json:"previousCloseDate"`
}

type SearchResults []SearchResult

func (srs *SearchResults) IsOlderThan(age time.Duration) (bool, error) {
	maxDate := time.Now().Add(-age)
	for _, result := range *srs {
		closeDate, err := time.Parse("2006-01-02", result.PreviousCloseDate)
		if err != nil {
			return true, fmt.Errorf("previous close date malformed: %w", err)
		}
		if closeDate.Before(maxDate) {
			return true, nil
		}
	}
	return false, nil
}

func getCachedData(code string) (SearchResults, time.Duration, error) {
	file, err := os.Open(code + ".json")
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, LongTime, nil // Because the file not existing isn't an error
		}
		return nil, LongTime, fmt.Errorf("open cache file: %w", err)
	}
	defer file.Close()

	info, err := file.Stat()

	if err != nil {
		return nil, LongTime, fmt.Errorf("get file info: %w", err)
	}

	fileAge := time.Until(info.ModTime()).Abs()

	result := SearchResults{}
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fileAge, fmt.Errorf("reading cached data: %w", err)
	}

	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, fileAge, fmt.Errorf("unmarshalling cached data: %w", err)
	}

	return result, fileAge, nil

}

func writeCache(code string, results SearchResults) error {
	file, err := os.Create(code + ".json")
	if err != nil {
		return fmt.Errorf("could not create cache file: %w", err)
	}
	defer file.Close()

	data, err := json.Marshal(results)
	if err != nil {
		return fmt.Errorf("re-encoding for cache file: %w", err)
	}

	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("writing to cache file: %w", err)
	}

	return nil
}

func Search(code string, api_token string) (SearchResults, error) {
	if api_token == "" {
		sample := SearchResults{}
		err := json.Unmarshal(SearchResultSample, &sample)
		return sample, err
	}

	cached, age, err := getCachedData(code)
	if err != nil {
		return SearchResults{}, fmt.Errorf("loading cached results: %w", err)
	}
	if cached != nil { // It's nil if there is no cache
		old := age.Hours() > 2
		if err != nil {
			return cached, fmt.Errorf("determining cache age: %w", err)
		}
		if !old {
			return cached, nil
		}
	}

	response, err := http.Get(BaseURL + SearchPath + code + "?api_token=" + api_token)
	if err != nil {
		return SearchResults{}, fmt.Errorf("searching for %q: %w", code, err)
	}
	defer response.Body.Close()

	results := SearchResults{}
	dec := json.NewDecoder(response.Body)
	err = dec.Decode(&results)
	if err != nil {
		return results, fmt.Errorf("decoding result: %w", err)
	}

	err = writeCache(code, results)
	if err != nil {
		return results, fmt.Errorf("caching data: %w", err)
	}
	return results, nil

}
