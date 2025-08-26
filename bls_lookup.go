package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	CACHE_DB                = "localdb.json"
	API_THROTTLE_DELAY_MSEC = 250
)

type BLSCache struct {
	cache map[string]float64
	m     sync.RWMutex
}

func (c *BLSCache) Load() {
	// Ensure the file exists
	f, err := os.OpenFile(CACHE_DB, os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal("Error creating cache file", err)
	}
	f.Close()

	data, err := os.ReadFile(CACHE_DB)
	if err != nil {
		log.Fatal("Couldn't read cache file for searching", err)
	}

	if err := json.Unmarshal(data, &c.cache); err != nil {
		fmt.Println("cache file was empty or didn't parse", err)
		c.cache = map[string]float64{}
	}
}

func (c *BLSCache) Find(url string) (float64, error) {
	c.m.RLock()
	defer c.m.RUnlock()
	val, ok := c.cache[url]
	if !ok {
		return 0, fmt.Errorf("didn't find url (%v) in cache", val)
	}
	return val, nil
}

func (c *BLSCache) Add(url string, value float64) {
	c.m.Lock()
	defer c.m.Unlock()
	c.cache[url] = value
	data, err := json.Marshal(c.cache)
	if err != nil {
		fmt.Println("marshalling to json failed", err)
	}

	if err := os.WriteFile(CACHE_DB, data, os.ModeAppend); err != nil {
		fmt.Println("Error writing cache file contents", err)
	}
}

func NewCache() *BLSCache {
	cache := BLSCache{}
	cache.Load()
	return &cache
}

// Is a BLS URL call about to ask for a transation for a 0 value?
var costMatchRegex = regexp.MustCompile(`\?cost1=0&`)

// Pick out the result from the response body
var blsResponseParse = regexp.MustCompile(`<p><span id="answer">\$(.*)<\/span><\/p>`)

func GetBlsValue(url string, cache *BLSCache) (float64, error) {
	// look up the value in the cache, and return it if we find it
	val, err := cache.Find(url)
	if err == nil {
		return val, nil
	}

	// If the url is requesting a 0.0 value, we know the result is 0
	if isMatch := costMatchRegex.MatchString(url); isMatch {
		return 0.0, nil
	}

	// Call the BLS API and scrape out the result
	// We don't want heat from the Commerce department, so rate limit the call to their API
	defer time.Sleep(API_THROTTLE_DELAY_MSEC * time.Millisecond)
	log.Println("Calling BLS for data:", url)
	resp, err := http.Get(url)
	if err != nil {
		return 0, errors.New("call to BLS URL errored")
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, errors.New("error retrieving body from response")
	}
	resp.Body.Close()
	result := blsResponseParse.FindSubmatch(body)
	new_val, err := strconv.ParseFloat(strings.ReplaceAll(string(result[1]), ",", ""), 64)
	if err != nil {
		return 0, fmt.Errorf("extracted value %v did not parse as a float", string(result[1]))
	}

	cache.Add(url, new_val)
	return new_val, nil
}
