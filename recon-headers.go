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

var ip, port, uri, logfile string
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
		headers.WriteString(strconv.Quote(strings.ToLower(header)) + ":")
		values := strings.Join(r.Header[header], ",")
		values = strings.Replace(values, `"`, ``, -1)
		values = strconv.Quote(values)
		headers.WriteString(values)
		headers.WriteString(" ")
	}

	return headers.String()
}

// Replies to a web request with a unique png image, attempts to prevent caching, and logs request data
func handler(w http.ResponseWriter, r *http.Request) {
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

	// Log the request data regardless of expecting the requested path
	log.Printf("%v %v %v %v %v", "(Src IP Redacted)", remoteAddrPort(r), 200, strconv.Quote(r.URL.String()), headerString(r))
	return
}

// Replies to an unexpected web request with a 404 and logs request data
func handlerCatchAll(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)
	fmt.Fprint(w, "404 Not Found")
	// Log the request data
	log.Printf("%v %v %v %v %v", "(Src IP Redacted)", remoteAddrPort(r), 404, strconv.Quote(r.URL.String()), headerString(r))
	return
}

// Demo - display logs
func handlerDemo(w http.ResponseWriter, r *http.Request) {
	file, err := os.OpenFile(logfile, os.O_RDONLY, 0644)
	if err != nil {
		w.WriteHeader(400)
		fmt.Fprint(w, "400 Error occurred accessing logs")
		return
	}
	defer file.Close()

	var buffer []byte
	stat, err := os.Stat(logfile)
	var start int64
	if stat.Size() > 20480 {
		buffer = make([]byte, 20480) // 20KB
		start = stat.Size() - 20480
	} else {
		buffer = make([]byte, stat.Size())
		start = 0
	}
	_, err = file.ReadAt(buffer, start)
	if err != nil {
		w.WriteHeader(400)
		fmt.Fprint(w, "400 Error occurred reading logs")
		return
	}

	lines := bytes.Split(buffer, []byte("\n"))
	linesConcat := bytes.Join(lines[1:len(lines)], []byte("\n\n"))
	trimmed := bytes.Trim(linesConcat, "\x00")

	w.Write(trimmed)
}

// Handles command line arguments
func handleFlags() {
	flag.StringVar(&ip, "ip", "All interfaces", "The local IP address the web server should listen on")
	flag.StringVar(&port, "port", "8080", "The port number the web server should listen on")
	flag.StringVar(&uri, "uri", "/", `The URI that returns an image
Examples:
	"/" 		= 	Respond to any path or file name (wildcard path and file name)
	"/recon" 	= 	Only respond to exact match of "/recon"
	"/recon.png"	= 	Only respond to exact match of "/recon.png"
	"/recon/" 	= 	Respond to "/recon/" and anything after (wildcard file name)
`)
	flag.StringVar(&logfile, "logfile", "recon-headers.log", "The name of the log file")
	flag.Parse()
}

// Configures the logger
func setupLogger(logfile string) *os.File {
	// Create log file
	file, err := os.OpenFile(logfile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("[*] Error creating log file: ", logfile)
	}

	file.Truncate(0)       // For Demo
	file.WriteString("\n") // For Demo

	log.SetOutput(file)
	return file
}

func main() {
	// Parse command line arguments
	handleFlags()

	// Configure settings for logging
	l := setupLogger(logfile)
	defer l.Close()

	//Ensure the URI starts with a "/"
	if !strings.HasPrefix(uri, "/") {
		uri = "/" + uri
	}

	http.HandleFunc("/logs", handlerDemo) // For Demo

	// Handle requests to the provided URI
	http.HandleFunc(uri, handler)
	// If a specific URI is provided, return a 404 to unexpected URI requests
	if uri != "/" {
		http.HandleFunc("/", handlerCatchAll)
	}

	// If an IP isn't provided, listen on all interfaces
	if ip == "All interfaces" {
		ip = ""
	}

	// Ignore DOS-like requests
	server := http.Server{
		Addr:           ip + ":" + os.Getenv("PORT"), // For Demo (getenv port)
		ReadTimeout:    time.Duration(1 * time.Second),
		WriteTimeout:   time.Duration(1 * time.Second),
		MaxHeaderBytes: 4096,
	}

	fmt.Printf("[*] Starting web server (%v:%v)", ip, port)
	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}
