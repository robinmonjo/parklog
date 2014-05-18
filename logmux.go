package main

import (
	"bufio"
	"io"
	"log"
	"net"
	"net/url"
	"os"
)

func main() {

	tcpStream := NewStream("file:///dev/stdout")
	reader := bufio.NewReader(os.Stdin)

	for {
		line, err := reader.ReadString('\n')

		if err != nil {
			if err != io.EOF {
				log.Fatal(err)
			}
			break
		}
		tcpStream.Log.Print(line)
	}
	tcpStream.Conn.Close()
}

type Streamer interface {
	Write(b []byte) (n int, err error)
	Close() error
}

type Stream struct {
	Url  *url.URL
	Conn Streamer
	Log  *log.Logger
}

func NewStream(uri string) *Stream {
	u, err := url.Parse(uri)
	if err != nil {
		log.Fatal(err)
	}

	var conn Streamer

	if u.Scheme == "tls" {
		//todo handle tls
	} else if u.Scheme == "file" {
		conn, err = os.OpenFile(u.Host+u.Path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	} else {
		conn, err = net.Dial(u.Scheme, u.Host)
	}

	if err != nil {
		log.Fatal(err)
	}

	l := log.New(conn, "", log.LstdFlags)

	return &Stream{u, conn, l}
}
