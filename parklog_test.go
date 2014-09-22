package main

import (
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"testing"
)

const LOG_LINE string = "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Quisque placerat maximus mauris, a suscipit eros feugiat nec."

var confs = []*StreamConfig{
	&StreamConfig{"tcp://localhost:9000", "tcp - ", false},
	&StreamConfig{"unix://socket", "unix - ", false},
}

func launchServer(scheme string, endpoint string, resp chan<- string, ready chan<- bool) {
	l, err := net.Listen(scheme, endpoint)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		l.Close()
		os.Remove("./" + endpoint)
	}()

	ready <- true

	conn, err := l.Accept()
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := conn.Read(buf)
			if err != nil {
				if err != io.EOF {
					log.Fatal(err)
				}
				break
			}
			resp <- string(buf[0:n])

			for i := 0; i < n; i++ {
				buf[i] = 0
			}
		}
	}()

}

func Test_NewStream(t *testing.T) {

	for _, conf := range confs {
		stream, err := NewStream(conf)
		if err != nil {
			t.Error(err)
		}
		//no server here so
		if stream.Status != NOT_CONNECTED {
			t.Error("Stream has status ", stream.Status, " expected NOT_CONNECTED")
		}

		resp := make(chan string)
		ready := make(chan bool)

		//launch a server
		go launchServer(stream.Url.Scheme, stream.Url.Host+stream.Url.Path, resp, ready)
		<-ready

		if err = stream.TryConnect(); err != nil {
			t.Error(err)
		}

		if stream.Status != CONNECTED {
			t.Error("Stream has status ", stream.Status, " expected CONNECTED")
		}

		for i := 0; i < 10; i++ {
			line := LOG_LINE + strconv.Itoa(i)
			err = stream.Write(line)
			if err != nil {
				log.Println(err)
			}
			received := <-resp
			checkWrite(stream, line, received, t)
		}

		stream.Close()
		if stream.Status != NOT_CONNECTED {
			t.Error("Stream has status ", stream.Status, " expected NOT_CONNECTED")
		}
	}

}

func checkWrite(s *Stream, wroteLine, loggedLine string, t *testing.T) {
	if !strings.HasPrefix(loggedLine, s.Conf.Prefix) {
		t.Error("Logged line doesn't have prefix ", s.Conf.Prefix)
	}

	if !strings.HasSuffix(loggedLine, wroteLine) {
		t.Error("Logged line ", loggedLine, " doesn't match ", wroteLine)
	}
}
