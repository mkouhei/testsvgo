package main

import (
	"bytes"
	"time"
	"log"
	"fmt"
	"net/http"
	"github.com/ajstarks/svgo"
)

func main() {
 	http.Handle("/test0", http.HandlerFunc(writeSVG))
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

func now() string {
	t := time.Now()
	return t.Format("2006/01/02 15:05:04")
}

func generateSVG(canvas *svg.SVG) *svg.SVG {
	width := 500
	height := 500
	canvas.Start(width, height)
	canvas.Circle(width/2, height/2, 100)
	canvas.Text(width/2, height/2, now(),
		"text-anchor: middle; font-size: 16px; fill: white")
	canvas.End()
	return canvas
}

func writeSVG(w http.ResponseWriter, req *http.Request) {

	var b bytes.Buffer
	canvas := svg.New(&b)
	svg_obj := generateSVG(canvas)

	w.Header().Set("Content-Type", "image/svg+xml")
	fmt.Fprint(w, svg_obj.Writer)
}
