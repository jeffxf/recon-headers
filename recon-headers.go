package main

import (
	"bytes"
	"crypto/rand"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

var port, name, logfile string
var beginningOfTime = time.Unix(0, 0).Format(time.RFC1123)

// Returns a random int between 0 and 255
func randInt255() int {
	b := make([]byte, 1)
	rand.Read(b)
	return int(b[0])
}

// Generates a unique png with a random size & writes it back to the provided buffer
func generatePng(buffer io.Writer) {
	width := (randInt255() / 5) + 1
	height := (randInt255() / 5) + 1
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	img.Set(width/3, height/2, color.RGBA{255, 255, 255, 255})
	img.Set(height/4, width/4, color.RGBA{255, 255, 255, 255})
	png.Encode(buffer, img)
	return
}

// Parse a RemoteAddr to get the IP
func remoteAddrIP(r *http.Request) string {
	splitAddr := strings.Split(r.RemoteAddr, ":")
	return strings.Join(splitAddr[0:len(splitAddr)-1], ":")
}

// Parse a RemoteAddr to get the Port
func remoteAddrPort(r *http.Request) string {
	splitAddr := strings.Split(r.RemoteAddr, ":")
	return splitAddr[len(splitAddr)-1]
}

// Concat headers into a single string
func headerString(r *http.Request) string {
	var headers bytes.Buffer

	// Sort the Headers
	var sortedHeaders []string
	for k := range r.Header {
		sortedHeaders = append(sortedHeaders, k)
	}
	sort.Strings(sortedHeaders)

	// Add each header to the buffer
	for _, header := range sortedHeaders {
		headers.WriteString(strings.ToLower(header) + ":")
		values := strings.Join(r.Header[header], ", ")
		values = strings.Replace(values, `"`, ``, -1)
		values = strconv.Quote(values)
		headers.WriteString(values)
		headers.WriteString(" ")
	}

	return headers.String()
}

// returnImage replies to a web request with a unique png image and attempts to prevent caching
func returnImage(w http.ResponseWriter, r *http.Request) {

	// Delete headers we receive that could allow caching or setting the allowed origin
	deleteHeaders := []string{"ETag", "If-Modified-Since", "If-Match", "If-None-Match", "If-Range", "If-Unmodified-Since", "Origin"}
	for _, h := range deleteHeaders {
		w.Header().Del(h)
	}

	// Set headers to prevent caching
	setHeaders := map[string]string{
		"Expires":         beginningOfTime,
		"Cache-Control":   "no-cache, private, max-age=0",
		"Pragma":          "no-cache",
		"X-Accel-Expires": "0",
	}
	for h, v := range setHeaders {
		w.Header().Set(h, v)
	}

	// Generate unique png image and send it off
	buffer := new(bytes.Buffer)
	generatePng(buffer)
	w.Write(buffer.Bytes())

	// Log the request data
	log.Printf("%v %v %v %v", remoteAddrIP(r), remoteAddrPort(r), r.URL.String(), headerString(r))
}

// handleFlags stores the command line argument options
func handleFlags() {
	flag.StringVar(&port, "port", "8080", "The port number the web server should listen on")
	flag.StringVar(&name, "name", "/", `The specific file path and name to listen for
Examples:
	"/" 		= 	Respond to any path or file name (wildcard path and file name)
	"/jeffxf" 	= 	Only respond to exact match of "/jeffxf"
	"/jeffxf.png"	= 	Only respond to exact match of "/jeffxf.png"
	"/jeffxf/" 	= 	Respond to "/jeffxf/" and anything after (wildcard file name)
`)
	flag.StringVar(&logfile, "logfile", "recon-headers.log", "The name of the log file")
	flag.Parse()
}

// setupLogger configures the loggers
func setupLogger(logfile string) *os.File {
	// Create log file
	file, err := os.OpenFile(logfile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("[*] Error creating log file: ", logfile)
	}
	log.SetOutput(file)
	return file
}

func main() {
	// Parse command line arguments
	handleFlags()

	// Configure settings for logging
	l := setupLogger(logfile)
	defer l.Close()

	// If a file name is provided, ensure it starts with a "/"
	if !strings.HasPrefix(name, "/") {
		name = "/" + name
	}

	http.HandleFunc(name, returnImage)
	fmt.Println("[*] Starting web server on port:", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		panic(err)
	}
}
