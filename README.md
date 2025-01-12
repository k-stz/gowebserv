# Go Web Server
A learning project with the aim to implement a simple webserver. 
Piecing together how HTTP works on top of TCP Sockets.

## Goals
- Respond to GET Requests ✅
- Respond with Pictures like Gif/JPEG ✅
- Resolve complex HTTP requests (loading pictures) ✅
- Use go-templates ✅ (see /template endpoint)
- write testing code for some usecases ✅ (see request_test.go)
- probably goroutine handling and cleaning up each session ✅ (good enough)

# Lessons Learned
## TCP Socket Contains All the Magic

The HTTP protocol defines how raw inputs and outputs in a TCP segment payload are interpreted. Through simple read and write operations, we can perform all the magic of a web server.

Once a client establishes a connection, it can write the HTTP request as raw bytes into the request. The web server can then respond with its HTTP response, and that's it.

Let’s walk through an example. Consider our web server running on `localhost:8080`. A `curl localhost:8080/` command will first attempt to establish a connection with the web server. The server will accept() the connection by creating a **client socket**. After that, both the client (e.g., `curl`) and the web server (the server) can write to their respective ends of the connected pair (each end of the "connected pair" is just a socket).

On the server's side its connected pair socket is referred to as the "client socket,". And, still continuing with the server's point of view, the clients connected pair socket is referred to as the "remote/peer/foreign" socket.

Programmatically, after the socket API's `accept()` system call completes, the three-way TCP handshake will already have established a TCP connection (you can inspect this with Wireshark). The result is a connected pair into which both client and server can write bytes at any time. For example, a `curl localhost:8080/` command writes the following bytes, shown as ASCII, into its socket:

```sh
GET / HTTP/1.1
Host: localhost:8080
User-Agent: curl/8.11.1
Accept: */*
```

Note: Per the standard all lines in HTTP have to be delimited with CR-LF (`\r\n`), (that's the internet standard in most protocols). But clients will usually give you leeway by also accepting a single newline with `\n`.

This is the HTTP request, using the GET method on the URL path /.

The web server can now send an HTTP response by writing the headers (here: Content-Lenght, Content-Type) and body (here: "Hello World!\r\n") directly into the client socket:

```sh
HTTP/1.1 404
Content-Length: 12
Content-Type: text/plain; charset=utf-8

Hello World!
```

Note: The headers and body are separated by an additional blank line `\r\n\r\n`.

The exact protocol definitions can be found in relevant RFCs, such as `RFC 1945` and friends.

## Human-Powered Web Server
To emphasize again the basics of a webserver is just writing the right order of bytes into a TCP socket, even when you do it by hand...! You can emulate a web server manually using tools like netcat. Start a TCP socket listening on port 5000 using a command like:
```sh
netcat -l -p 5000
```

Now visit `localhost:5000` in your browser. You’ll see the HTTP request show up in the netcat interactive session. It will look something like this:

```sh
GET / HTTP/1.1
Host: localhost:5000
...
```

At this point, the session will appear blocked, waiting for a response. You can act as a human-powered web server by typing out the response message manually. For example:

```sh
HTTP/1.1 200 OK
Content-Length: 16

<b>Hi there!</b>
```

Once you press Enter after the HTML-bold-tagged "Hi there!", you’ll see the response rendered in your browser. Yay, human powers!

## Close your HTTP connections
When writing more complex html website, that would cause subsequent HTTP Request from a single visit closing your connection becomes mandatory in order to avoid blocking.
For example when you want to serve a website with an `<img>`-tag that points to an image, the browser will first load the html with the `<img>`-tag and then request the picture inside the tag. Make sure that this request will either close its connection or use in its response the header `Connection: close` in order to signify the client (e.g. browser) that no more data is coming after serving the picture.

Without eiter the `Connection: close` header or closing the connection (in go call `conn.Close()` on the net.Conn interface ) the browser/client might hang till it timeouts and never even fully download the picture and display it.

# Sources:
- regarding the Linux Socket API, highly recommend the book: "The Linux Programming Interface" (2010) by Michael Kerrisk. Especially Chapter 55 "Sockets: Introduction" and Chapter 58 "Sockets: Fundamentals of TCP/IP Network)" 