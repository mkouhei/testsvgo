package main

import (
	"log"
	"flag"
	"net/http"
	"time"
	"text/template"

	"github.com/ajstarks/svgo"
	"github.com/gorilla/websocket"
)

const (
	writeWait = 10 * time.Second
	pongWait = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
)

var (
	addr = flag.String("addr", ":8080", "http service address")
	homeTempl = template.Must(template.New("").Parse(homeHTML))
	upgrader = websocket.Upgrader{
		ReadBufferSize: 1024,
		WriteBufferSize: 1024,
	}
)


func now() string {
	t := time.Now()
	return t.Format("2006/01/02 15:04:05")
}

func generateSVG(canvas *svg.SVG, msg string) *svg.SVG {
	width := 500
	height := 500
	now := now()

	canvas.Start(width, height)
	canvas.Circle(width/2, height/2, 200)
	canvas.Text(width/2, height/2, now+"\n"+msg,
		"text-anchor: middle; font-size: 30px; fill: white")
	canvas.End()
	return canvas
}

func writeSVG(w http.ResponseWriter, req *http.Request) {
	//w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	canvas := svg.New(w)
	generateSVG(canvas, "hoge")
}

func serveHome(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		http.Error(w, "Not found", 404)
		return
	}
	if req.Method != "GET" {
		http.Error(w, "Method not allowd", 405)
		return
	}

/*
	w.Header().Set("Content-Type", "image/svg+xml")
	canvas := svg.New(w)
	generateSVG(canvas, "moge")
*/
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	now := now()
	hoge := "<circle cx=\"250\" cy=\"250\" r=\"200\"/><text x=\"250\" y=\"250\" style=\"text-anchor: middle; font-size: 16px; fill: white\">" + now + "</text>"

	var v = struct {
		Data string
	}{
		hoge,
	}
	homeTempl.Execute(w, &v)
}

func serveWs(w http.ResponseWriter, req *http.Request) {
	ws, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			log.Println(err)
		}
		return
	}

	go writer(ws)
	reader(ws)
}

func reader(ws *websocket.Conn) {
	defer ws.Close()
	ws.SetReadLimit(512)
	ws.SetReadDeadline(time.Now().Add(pongWait))
	ws.SetPongHandler(func(string) error { ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			break
		}
	}
}

func writer(ws *websocket.Conn) {
	pingTicker := time.NewTicker(pingPeriod)
	defer func() {
		pingTicker.Stop()
		ws.Close()
	}()
	for {
		select {
		case <-pingTicker.C:
			ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := ws.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}


func main() {
 	http.Handle("/test0", http.HandlerFunc(writeSVG))
	http.Handle("/", http.HandlerFunc(serveHome))
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}


const homeHTML = `<!doctype html>
<html>
<head>
<title>test SVG</title>
</head>
<body>
<svg svg width="500" height="500"
     xmlns="http://www.w3.org/2000/svg" 
     xmlns:xlink="http://www.w3.org/1999/xlink" id="test0">{{.Data}}</svg>
</body>
</html>
`
