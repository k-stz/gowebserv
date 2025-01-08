package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestServer(t *testing.T) {
	// Start the server.
	srv := NewTcpServer(":5000")
	go srv.startTCPServer(func(conn net.Conn) {
		writeSimpleResponse(conn)
	})
	// Wait a moment for the server to start
	time.Sleep(500 * time.Millisecond)

	// client := &http.Client{
	//CheckRedirect:
	//	redirectPolicyFunc,
	// }
	//req, err := http.NewRequest("GET", "http://localhost:5000/hello", nil)
	// resp, err := client.Do(req)
	// if err != nil {
	// 	panic(err)
	// }

	// Make an HTTP request to the server
	urlpath := "/hello"
	resp, err := http.Get("http://localhost:5000" + urlpath)
	if err != nil {
		t.Fatalf("Failed to make GET request: %s", err)
	}
	fmt.Println("Client() HTTP Response")
	buf := new(strings.Builder)
	_, err = io.Copy(buf, resp.Body)
	fmt.Println("resp:", buf)
	//defer resp.Body.Close()
	//fmt.Println("resp:", resp)
	// // Check if the status code is 200.
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status code 200, got %d", resp.StatusCode)
	}

	//server.listener.Close()
	// body, err := io.ReadAll(resp.Body)
	// fmt.Println("response body:\n", string(body))
}
