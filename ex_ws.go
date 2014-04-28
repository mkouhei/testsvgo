package main

import (
	"flag"
	"time"
	"log"
	"net/http"
	"text/template"
	"github.com/gorilla/websocket"
)

const (
	writeWait = 10 * time.Second
	pongWait = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
	pollPeriod = 1 * time.Second
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
	return t.Format(time.RFC3339)
}

func reader(ws *websocket.Conn) {
	defer ws.Close()
	ws.SetReadLimit(512)
	ws.SetReadDeadline(time.Now().Add(pongWait))
	ws.SetPongHandler(func(string) error {
		ws.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			break
		}
	}
}

func writer(ws *websocket.Conn) {
	pingTicker := time.NewTicker(pingPeriod)
	pollTicker := time.NewTicker(pollPeriod)
	defer func() {
		pingTicker.Stop()
		pollTicker.Stop()
		ws.Close()
	}()
	for {
		select {
		case <- pollTicker.C:
			ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := ws.WriteMessage(websocket.TextMessage, []byte(now())); err != nil {
				return
			}
		case <- pingTicker.C:
			ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := ws.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

func serveWs(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			log.Println(err)
		}
		return
	}
	go writer(ws)
	reader(ws)
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", 404)
		return
	}
	if r. Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	var v = struct {
		Host string
		Now string
	}{
		r.Host,
		now(),
	}
	homeTempl.Execute(w, &v)
}

func main() {
 	http.Handle("/", http.HandlerFunc(serveHome))
 	http.Handle("/ws", http.HandlerFunc(serveWs))
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

const homeHTML = `<!doctype html>
<html>
<head>
<title>test websocket</title>
</head>
<body>
<div id="now">{{.Now}}</div>
<script type="text/javascript">
(function() {
var data = document.getElementById("now");
var conn = new WebSocket("ws://{{.Host}}/ws");
conn.onclose = function(evt) {
data.textContent = 'Connection closed';
}
conn.onmessage = function(evt) {
console.log(evt.data);
data.textContent= evt.data;
}
})();
</script>
</body>
</html>
`
