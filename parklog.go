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
	"time"
)

const DIAL_TIMEOUT = 5 * time.Second

type StreamStatus int

const (
	CONNECTED StreamStatus = iota
	CONNECTING
	NOT_CONNECTED
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
			stream.Write(line)
		}
	}
	for _, stream := range streams {
		stream.Conn.Close()
	}
}

func initStreams() (streams []*Stream) {
	var streamConfigs []StreamConfig
	file, err := ioutil.ReadFile("parklog.json")
	if err != nil {
		log.Fatal(err)
	}

	confs := os.ExpandEnv(string(file))
	if err := json.Unmarshal([]byte(confs), &streamConfigs); err != nil {
		log.Fatal(err)
	}

	for _, conf := range streamConfigs {
		s, err := NewStream(&conf)
		if err != nil {
			log.Println(err)
			continue
		}
		streams = append(streams, s)
	}
	return
}

type StreamConfig struct {
	Url         string `json:"url"`
	Prefix      string `json:"prefix"`
	AllowSSCert bool   `json:"allow_self_signed_cert"`
}

type Stream struct {
	Url    *url.URL
	Conn   io.WriteCloser
	Conf   *StreamConfig
	Status StreamStatus
}

func NewStream(conf *StreamConfig) (*Stream, error) {
	u, err := url.Parse(conf.Url)
	if err != nil {
		return nil, err
	}

	stream := &Stream{Url: u, Conf: conf, Status: CONNECTING}
	stream.TryConnect()
	return stream, nil
}

func (s *Stream) TryConnect() {
	s.Status = CONNECTING
	if err := s.Connect(); err != nil {
		s.Status = NOT_CONNECTED
		log.Println(err)
	} else {
		s.Status = CONNECTED
	}
}

func (s *Stream) Connect() error {
	var conn io.WriteCloser
	var err error
	path := s.Url.Host + s.Url.Path

	switch {
	case s.Url.Scheme == "tls" || s.Url.Scheme == "ssl":
		config := &tls.Config{InsecureSkipVerify: s.Conf.AllowSSCert}
		conn, err = tls.Dial("tcp", path, config)
	case s.Url.Scheme == "file":
		conn, err = os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	default:
		conn, err = net.DialTimeout(s.Url.Scheme, path, DIAL_TIMEOUT)
	}
	s.Conn = conn
	return err
}

func (s *Stream) Write(line string) {
	if s.Status == CONNECTED {
		if _, err := s.Conn.Write([]byte(line)); err != nil {
			s.Status = NOT_CONNECTED
			log.Println(err)
		}
	} else {
		log.Println(s, "is oflline")
		s.TryConnect()
	}

}
