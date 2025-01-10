package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"
	"text/template"
	"time"
)

var simpleResponse string = "HTTP/1.1 404\nContent-Length: 12\nContent-Type: text/plain; charset=utf-8\n\nHello World!"

var yt string = `<h1>KL's Awesome WebServer!</h1><iframe width="560" height="315" src="https://www.youtube.com/embed/0-L83sDVrpQ?si=XHtKd0MmzcfqEu5S" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" referrerpolicy="strict-origin-when-cross-origin" allowfullscreen></iframe>`

var yt3 string = "<b>hiii</b>"

var hrefGopher string = `<img class="Hero-gopherLadder" src="./favicon.svg" alt="Go Gopher climbing a ladder.">`

var picResponseToContentLen string = `HTTP/1.1 200 OK\nContent-Type: image/gif\nContent-Length: `

var picGoDevGopherRel string = `<img src="pics/gopher.jpg" alt="Gopher pic al-text" width="500">`

type tcpServer struct {
	address string // ":8080"
	// there are multiple net.Conn per Server for each client-socket
	// conn     net.Conn
	listener    net.Listener
	connections []net.Conn
}

func NewTcpServer(address string) *tcpServer {
	return &tcpServer{
		address: address,
	}
}

func writePicture(conn net.Conn, contentType, filepath string) {
	slog.Info("writePicture()", "file", filepath)
	file, err := os.OpenFile(filepath, os.O_RDONLY, 0600)
	defer func() {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}()
	fileInfo, err := file.Stat()
	if err != nil {
		panic(err)
	}
	filesize := fileInfo.Size()

	reponseHeader := fmt.Sprintf("HTTP/1.1 200 OK\nContent-Type: %s\nContent-Length: %d\n\n",
		contentType,
		filesize)
	_, err = conn.Write([]byte(reponseHeader))
	if err != nil {
		panic(err)
	}

	n, err := io.Copy(conn, file)
	if err != nil {
		panic(err)
	}
	fmt.Println("Pic size written=", n)
	conn.Close()
}

func writeHTTPContent(conn net.Conn, body string) {
	bodyLen := len(body)
	reponseHeader := fmt.Sprintf("HTTP/1.1 200\nContent-Length: %d\n\n",
		bodyLen)
	// write actual response
	n, err := conn.Write([]byte(reponseHeader + body))
	if err != nil {
		slog.Error("Error writing in client socket")
		panic(err)
	}
	slog.Info("TCP Response written.", "total-bytes=", n, "body-bytes", bodyLen)
	conn.Close() // I need to close it on each call!?
}

func getNextRequest(conn net.Conn) (method, path string, request *http.Request) {
	reader := bufio.NewReader(conn)
	req, err := http.ReadRequest(reader)
	if err != nil {
		panic(err)
	}
	path = req.URL.Path
	method = req.Method
	return method, path, req
}

type htmlSite struct {
	Title   string
	DateStr string
	Date    time.Time
}

func NewFuncMap() template.FuncMap {
	return template.FuncMap{
		"isTrue": func() bool {
			return true
		},
		"upper": func(input string) string {
			return strings.ToUpper(input)
		},
	}
}

func genTemplate(path string) string {
	fmt.Println("## PARSING template...")
	buf := new(bytes.Buffer)
	site := htmlSite{
		Title:   "Welcome!",
		DateStr: time.Now().Format(time.RFC822),
		Date:    time.Now().UTC(),
	}

	tmpl := template.Must(template.New("my.tpl").Funcs(NewFuncMap()).ParseFiles("template/my.tpl"))

	err := tmpl.Execute(buf, site)
	if err != nil {
		panic(err)
	}
	fmt.Println("rendered templ:", buf)
	return "<b>TODO: rendered template here</b><br>" + buf.String()
}

// Return string based on method and path
func multiplexRequest(conn net.Conn, method, path string) {
	var body string
	fmt.Println("  ### SWITCH PATH", path, "###")
	switch path {
	case "/":
		body = fmt.Sprintf("<b>Hello, World!:<br>Method=%s<br> Urlpath=%s",
			method, path)
		// embed pic
		body = body + "<br>" + picGoDevGopherRel
		fmt.Println("writing body:", body)
		writeHTTPContent(conn, body)
	case "/hello":
		body = "<b>Hello Endpoint reached</b>"
		writeHTTPContent(conn, body)
	case "/favicon.ico":
		writePicture(conn, "image/png", "pics/favicon.png")
	case "/pics/gopher.jpg":
		writePicture(conn, "image/jpeg", "pics/gopher.jpg")
	case "/template":
		body := genTemplate("template/first.tpl")
		writeHTTPContent(conn, body)
	default:
		body = fmt.Sprintf("Not implemented! Path=%s", path)
	}
}

func writeSimpleResponse(conn net.Conn) {
	method, path, _ := getNextRequest(conn)

	fmt.Println("Method=", method, "path=", path)
	fmt.Println("HTTP Request (inside writeSimpleResponse)")

	multiplexRequest(conn, method, path)
}

func writeSocket(conn net.Conn) {
	fmt.Println("writing into socket!")
	reponseHeader := fmt.Sprintf("HTTP/1.1 404\nContent-Length: %d\n\n", len(yt))
	fmt.Println("len", len(yt))
	// interesssting only when writing the \n does the
	// client react => must be part of HTTP protocol

	n, err := conn.Write([]byte(reponseHeader))
	time.Sleep(time.Second * 3) // fake longer request
	n, err = conn.Write([]byte(yt))

	if err != nil {
		slog.Error("Error writing  in client socket")
		panic(err)
	}
	fmt.Printf("bytes-written= %d", n)
	fmt.Println("written!")
}

// read from the conn buffer returning
func readRequest(conn net.Conn) (method, path string, e error) {
	buffer := make([]byte, 100)
	n, err := conn.Read(buffer)
	reader := bufio.NewReader(conn)

	for {
		line, err := reader.ReadString('\n')
		if err == nil {
			if err == io.EOF {
				break
			}
			return "", "", err
		}
		fmt.Println("line: ", line)
	}

	if err != nil {
		panic(err)
	}
	fmt.Printf("bytes-read= %d content:\n%s", n, string(buffer))
	return "GET", "/", nil
}

func handleConnection(conn net.Conn, count int) {
	slog.Info("handleConnection",
		"count", count,
		"remoteaddr", conn.RemoteAddr())
	// lets read some
	// lets read all instead
	// _, _, err := readRequest(conn)
	// if err != nil {
	// 	panic(err)
	// }

	// Next lets write into the socket
	//writeSocket(conn, count)
	//writePicture(conn, count)
	// this will attempt to request an img at
	// writeHTTPContent(conn, picGoDevGopherRel)
	writeHTTPContent(conn, simpleResponse)

}

func serverSocketListener(ln net.Listener) {
	for count := range 100 {
		slog.Info("serverSocketListener", "count", count)
		conn, err := ln.Accept()
		if err != nil {
			slog.Error("Error ln.Accept", "err", err)
			panic(err)
		}
		go func() {
			handleConnection(conn, count)
		}()
	}
	slog.Info("serverSocketListener", "msg", "accepting no more connections")

}

func (srv *tcpServer) Shutdown() error {
	return srv.listener.Close()
}

func (srv *tcpServer) startTCPServer(handleConnectionFn func(net.Conn)) {
	ln, err := net.Listen("tcp", srv.address)
	if err != nil {
		panic(err)
	}
	srv.listener = ln

	slog.Info("Listening on TCP Socket.", "address", srv.address)
	for {
		conn, err := srv.listener.Accept()
		if err != nil {
			panic(err)
		}
		srv.connections = append(srv.connections, conn)
		go handleConnectionFn(conn)
	}
}

func main() {
	slog.Info("Start main")
	server := NewTcpServer(":8080")
	server.startTCPServer(writeSimpleResponse)
	//serverSocketListener(ln)
}
