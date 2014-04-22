package main

import (
	"log"
	"github.com/ajstarks/svgo"
	"net/http"
)

func main() {
 	http.Handle("/test0", http.HandlerFunc(writeSVG))
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

func generateSVG(canvas *svg.SVG) {
	width := 500
	height := 500
	canvas.Start(width, height)
	canvas.Circle(width/2, height/2, 100)
	canvas.Text(width/2, height/2, "Hello, SVG",
		"text-anchor: middle; font-size: 30px; fill: white")
	canvas.End()
}

func writeSVG(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "image/svg+xml")
	canvas := svg.New(w)
	generateSVG(canvas)
}
