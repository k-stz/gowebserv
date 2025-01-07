package main

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"time"
)

var simpleResponse string = "HTTP/1.1 404\nContent-Length: 12\nContent-Type: text/plain; charset=utf-8\n\nHello World!"

var yt string = `<h1>KL's Awesome WebServer!</h1><iframe width="560" height="315" src="https://www.youtube.com/embed/0-L83sDVrpQ?si=XHtKd0MmzcfqEu5S" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" referrerpolicy="strict-origin-when-cross-origin" allowfullscreen></iframe>`

var yt3 string = "<b>hiii</b>"

var hrefGopher string = `<img class="Hero-gopherLadder" src="./favicon.svg" alt="Go Gopher climbing a ladder.">`

var picResponseToContentLen string = `HTTP/1.1 200 OK\nContent-Type: image/gif\nContent-Length: `

// var picGoDevGopher string = `<img src="https://go.dev/blog/gopher/header.jpg" alt="">`
var picGoDevGopherRel string = `<img src="gopher.jpg" alt="Gopher pic alt-text goes here">`

func writePicture(conn net.Conn, filepath string) {
	file, err := os.OpenFile(filepath, os.O_RDONLY, 0600)
	defer func() {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}()

	fmt.Println("writing picture...")
	// TODO how to get size
	reponseHeader := fmt.Sprintf("HTTP/1.1 200 OK\nContent-Type: image/jpeg\nContent-Length: %d\n\n", 35704)
	_, err = conn.Write([]byte(reponseHeader))
	if err != nil {
		panic(err)
	}

	// When writing this directly with
	// curl throws an error
	//* client returned ERROR on write of 16384 bytes
	n, err := io.Copy(conn, file) // WORKs!
	if err != nil {
		fmt.Println("reached?")
		panic(err)
	}
	fmt.Println("Written picture bytes:", n)

}

func writeHTTPContent(conn net.Conn, content string) {
	slog.Debug("writeHTTPContent")
	contentLen := len(content)
	reponseHeader := fmt.Sprintf("HTTP/1.1 200\nContent-Length: %d\n\n",
		contentLen)
	slog.Info("writeHTTPContent", "content-length", contentLen)
	_, err := conn.Write([]byte(reponseHeader))
	if err != nil {
		slog.Error("Error writing  in client socket")
		panic(err)
	}
	// write actual content
	_, err = conn.Write([]byte(content))
	if err != nil {
		slog.Error("Error writing in client socket")
		panic(err)
	}
}

func writeSocket(conn net.Conn, count int) {
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
	fmt.Printf("Connection=%d, bytes-written= %d",
		count, n)
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

	//writePicture(conn, "gopher.jpg")

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

func main() {
	slog.Info("Start main")
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		slog.Error("Error net.Listen", "err", err)
		panic(err)
	}
	serverSocketListener(ln)
}
