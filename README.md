# Go Web Server
A learning project with the aim to implement a simple webserver. 
Piecing together how HTTP works on top of TCP Sockets.

## Goals
- Respond to GET Requests
- Respond with Pictures like Gif/JPEG
- Resolve complex HTTP requests (loading pictures)
- Use go-templates
- write testing code for some usecases
- probably goroutine handling and cleaning up each session 

# Lessons Learned
## TCP Socket Contains All the Magic

The HTTP protocol defines how raw inputs and outputs in a TCP segment payload are interpreted. Through simple read and write operations, we can perform all the magic of a web server.

Once a client establishes a connection, it can write the HTTP request as raw bytes into the request. The web server can then respond with its HTTP response, and that's it.

Let’s walk through an example. Consider our web server running on `localhost:8080`. A `curl localhost:8080/` command will first attempt to establish a connection with the web server. The server will accept() the connection by creating a **client socket**. After that, both the client (e.g., `curl`) and the web server (the server) can write to their respective ends of the connected pair (each end of the "connected pair" is just a socket).

On the server's side its connected pair socket is referred to as the "client socket,". And, still continuing with the server's point of view, the clients connected pair socket is referred to as the "remote/peer/foreign" socket.

Programmatically, after the socket API's `accept()` system call completes, the TCP handshake will already have occurred (you can inspect this with Wireshark). The result is a connected pair into which both client and server can write bytes at any time. For example, a `curl localhost:8080/` command writes the following bytes, shown as ASCII, into its socket:

```sh
GET / HTTP/1.1
Host: localhost:8080
User-Agent: curl/8.11.1
Accept: */*
```

This is the HTTP request, using the GET method on the URL path /.

The web server can now send an HTTP response by writing the following directly into the client socket:

```sh
HTTP/1.1 404
Content-Length: 12
Content-Type: text/plain; charset=utf-8

Hello World!
```

The exact protocol definitions can be found in relevant RFCs, such as `RFC 1945` and friends.
