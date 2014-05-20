package main

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/url"
	"os"
)

func main() {

	log.SetPrefix(os.Args[0] + " -- ")

	streams := initStreams()

	reader := bufio.NewReader(os.Stdin)

	for {
		line, err := reader.ReadString('\n')

		if err != nil {
			if err != io.EOF {
				log.Fatal(err)
			}
			break
		}
		for _, stream := range streams {
			stream.Log.Print(line)
		}
	}
	for _, stream := range streams {
		stream.Conn.Close()
	}
}

func initStreams() (streams []*Stream) {
	var streamConfigs []StreamConfig
	confs, err := ioutil.ReadFile("parklog.json")
	if err != nil {
		log.Fatal(err)
	}
	if err := json.Unmarshal(confs, &streamConfigs); err != nil {
		log.Fatal(err)
	}

	for _, conf := range streamConfigs {
		s, err := NewStreamer(&conf)
		if err != nil {
			log.Println(err)
			continue
		}
		streams = append(streams, s)
	}
	return
}

type StreamConfig struct {
	Url    string `json:"url"`
	Prefix string `json:"prefix"`
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

func NewStreamer(conf *StreamConfig) (*Stream, error) {
	u, err := url.Parse(conf.Url)
	if err != nil {
		return nil, err
	}

	var conn Streamer

	if u.Scheme == "tls" || u.Scheme == "ssl" {
		config := &tls.Config{InsecureSkipVerify: true}
		conn, err = tls.Dial("tcp", u.Host+u.Path, config)
	} else if u.Scheme == "file" {
		conn, err = os.OpenFile(u.Host+u.Path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	} else {
		conn, err = net.Dial(u.Scheme, u.Host+u.Path)
	}

	if err != nil {
		return nil, err
	}

	l := log.New(conn, conf.Prefix, log.LstdFlags)

	return &Stream{u, conn, l}, nil
}
