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
	// Though writing the HTTP Code comes sequetnially first in the response,
	// because it is a bytebuffer. It is a better Idea to make it dependent
	// on the body successfully rendeirng. Because it influences the HTTP
	// Status code!
	reponseHeader := fmt.Sprintf("HTTP/1.1 200\nContent-Length: %d\r\nConnection: close\r\n\r\n",
		bodyLen)
	// write actual response
	n, err := conn.Write([]byte(reponseHeader + body + "\r\n"))
	if err != nil {
		slog.Error("Error writing in client socket")
		panic(err)
	}
	slog.Info("TCP Response written.", "total-bytes=", n, "body-bytes", bodyLen)

	conn.Close()
}

func uploadBackend(conn net.Conn, body string, req *http.Request) {
	fmt.Println(" #!# Uploading logic reached #!#")
	if req.Method != "POST" {
		body = "Error: Only supports POST request. Given: " + req.Method
	}
	//
	err := req.ParseMultipartForm(10 << 20) // 10MB 10 ^ (2 * 20)
	if err != nil {
		slog.Error("calling ParseMultipartForm")
		panic(err)
	}

	// retreive file
	multipartFile, header, err := req.FormFile("file")
	if err != nil {
		slog.Error("failed to get file from form")
		panic(err)
	}
	defer multipartFile.Close()

	dst, err := os.Create("uploads/" + header.Filename)
	if err != nil {
		slog.Error("creating file", "path", "uploads/"+header.Filename)
		panic(err)
	}
	defer dst.Close()

	buf := new(bytes.Buffer)
	n64, err := io.Copy(buf, multipartFile)
	if err != nil {
		slog.Error("Buffering POST body")
		panic(err)
	}
	fmt.Println("bytes in Multipart File:", n64)
	// Reading from a bytes.Buffer consumes it!
	// so we need to first park its content in body
	n64, err = io.Copy(dst, buf)
	if err != nil {
		slog.Error("Writing Post Request body to file")
		panic(err)
	}

	body = body + fmt.Sprintf("<br> Upload successful!<br>File uploaded to: %s<br>Bytes written: %d", "uploads/"+header.Filename, n64)
	bodyLen := len(body)

	reponseHeader := fmt.Sprintf("HTTP/1.1 200\nContent-Length: %d\r\nConnection: close\r\n\r\n",
		bodyLen)
	// write actual response
	n, err := conn.Write([]byte(reponseHeader + body + "\r\n"))
	if err != nil {
		slog.Error("Error writing in client socket")
		panic(err)
	}
	slog.Info("TCP Response written.", "total-bytes=", n, "body-bytes", bodyLen)

	conn.Close()
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
	Address string
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

func genTemplate(srv *tcpServer, templateFileName string) string {
	fmt.Println("## PARSING template...")
	buf := new(bytes.Buffer)
	site := htmlSite{
		Title:   "Welcome!",
		DateStr: time.Now().Format("02 Jan 06 15:04:05.000 MST"),
		Date:    time.Now().UTC(),
		Address: srv.address,
	}
	// impute "localhost" if address just port ":8000"
	if len(strings.Split(site.Address, ":")) <= 2 {
		site.Address = "localhost" + site.Address
	}

	tmpl := template.Must(template.New(templateFileName).Funcs(NewFuncMap()).ParseFiles(
		"template/" + templateFileName))

	err := tmpl.Execute(buf, site)
	if err != nil {
		panic(err)
	}
	fmt.Println("rendered templ:", buf)
	return buf.String()
}

// Return string based on method and path
func multiplexRequest(srv *tcpServer, conn net.Conn, method, path string, req *http.Request) {
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
	case "/upload":
		body = "<b>TODO: Time to implement upload!</b>"
		writeHTTPContent(conn, genTemplate(srv, "upload.tpl"))
	case "/favicon.ico":
		writePicture(conn, "image/png", "pics/favicon.png")
	case "/pics/gopher.jpg":
		writePicture(conn, "image/jpeg", "pics/gopher.jpg")
	case "/template":
		body := genTemplate(srv, "example.tpl")
		writeHTTPContent(conn, body)
	case "/upload-backend":
		uploadBackend(conn, "<b>File Upload Result</b>", req)
	default:
		body = fmt.Sprintf("Not implemented!Method=%s Path=%s", method, path)
		writeHTTPContent(conn, body)
	}
}

func writeSimpleResponse(srv *tcpServer, conn net.Conn) {
	method, path, req := getNextRequest(conn)

	fmt.Println("Method=", method, "path=", path)
	fmt.Println("HTTP Request (inside writeSimpleResponse)")

	multiplexRequest(srv, conn, method, path, req)
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

func (srv *tcpServer) startTCPServer(handleConnectionFn func(*tcpServer, net.Conn)) {
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
		go handleConnectionFn(srv, conn)
	}
}

func main() {
	slog.Info("Start main")
	server := NewTcpServer(":8080")
	server.startTCPServer(writeSimpleResponse)
	//serverSocketListener(ln)
}
