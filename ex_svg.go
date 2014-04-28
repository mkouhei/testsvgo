

package main

import (
	"bytes"
	"io"
	"fmt"
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

func generateSVG(canvas *svg.SVG) *svg.SVG {
	width := 500
	height := 500

	canvas.Start(width, height)
	canvas.Circle(width/2, height/2, 200)
	canvas.Text(width/2, height/2, now(),
		"text-anchor: middle; font-size: 16px; fill: white")
	canvas.End()
	return canvas
}

func writeSVG(w http.ResponseWriter, req *http.Request) {
	var b bytes.Buffer
	canvas := svg.New(&b)
	svg_obj := generateSVG(canvas)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, svg_obj.Writer)
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

	var b bytes.Buffer
	canvas := svg.New(&b)
	svg_obj := generateSVG(canvas)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	var v = struct {
		Data io.Writer
	}{
		svg_obj.Writer,
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
			/*
		default:
			b := new(bytes.Buffer)

			canvas := svg.New(b)
			svg_obj := generateSVG(canvas)

			pipeReader, pipeWriter := io.Pipe()
			svg_obj.Writer = pipeWriter
			io.Pipe()
			buf := new(bytes.Buffer)
			buf.ReadFrom(pipeReader)
			s := buf.String()
			pipeWriter.Close()
			
			ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := ws.WriteMessage(websocket.TextMessage, []byte(s)); err != nil {
				return
			}
*/
		}
	}
}


func main() {
 	http.Handle("/test0", http.HandlerFunc(writeSVG))
	http.Handle("/", http.HandlerFunc(serveHome))
	http.Handle("/ws", http.HandlerFunc(serveWs))
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}


const homeHTML = `<!doctype html>
<html>
<head>
<title>test1 SVG</title>
</head>
<body>
<svg svg width="500" height="500"
     xmlns="http://www.w3.org/2000/svg" 
     xmlns:xlink="http://www.w3.org/1999/xlink" id="test0">{{.Data}}</svg>
<script type="text/javascript">
(function() {
var data = document.getElementById("test0");
var conn = new WebSocket("ws://127.0.0.1:8000/ws");
conn.onclose = function(evt) {
data.textContent = "connection closed";
}
conn.onmessage = function(evt) {
console.log(evt.data);
data.textContent = evt.data;
}
})();
</script>
</body>
</html>
`
