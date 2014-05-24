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

const DEFAULT_ROTATION_TRESHOLD = 10 //bytes

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
			go stream.LogRotateIfNeeded()
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
	Url              string `json:"url"`
	Prefix           string `json:"prefix"`
	AllowSSCert      bool   `json:"allow_self_signed_cert"`
	LogRotate        bool   `json:"log_rotate"`
	RotationTreshold int    `json:"rotation_treshold"`
}

type WriterCloser interface {
	Write(b []byte) (n int, err error)
	Close() error
}

type Stream struct {
	Url              *url.URL
	Conn             WriterCloser
	Log              *log.Logger
	LogRotate        bool
	RotationTreshold int
}

func NewStreamer(conf *StreamConfig) (*Stream, error) {
	u, err := url.Parse(conf.Url)
	if err != nil {
		return nil, err
	}

	var conn WriterCloser

	if u.Scheme == "tls" || u.Scheme == "ssl" {
		config := &tls.Config{InsecureSkipVerify: conf.AllowSSCert}
		conn, err = tls.Dial("tcp", u.Host+u.Path, config)
	} else if u.Scheme == "file" {
		conn, err = os.OpenFile(u.Host+u.Path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	} else {
		conn, err = net.Dial(u.Scheme, u.Host+u.Path)
	}

	if u.Scheme != "file" && conf.LogRotate {
		log.Println("Warning", u.Scheme, "doesn't support log rotation, changing config")
		conf.LogRotate = false
	}

	if conf.LogRotate && conf.RotationTreshold == 0 {
		conf.RotationTreshold = DEFAULT_ROTATION_TRESHOLD
	}

	if err != nil {
		return nil, err
	}

	l := log.New(conn, conf.Prefix, log.LstdFlags)

	return &Stream{u, conn, l, conf.LogRotate, conf.RotationTreshold}, nil
}

func (stream Stream) LogRotateIfNeeded() {
	if !stream.LogRotate {
		return
	}
	log.Println("Will perform log rotation")
}
