package main

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var verbose *bool = flag.Bool("v", false, "verbose")
var configPath *string = flag.String("c", "./parklog.json", "config file path")

func _log(v ...interface{}) {
	if *verbose {
		log.Println(v)
	}
}

const DIAL_TIMEOUT = 5 * time.Second

type StreamStatus int

const (
	CONNECTED StreamStatus = iota
	CONNECTING
	NOT_CONNECTED
)

var (
	streams     []*Stream
	streamsLock = new(sync.RWMutex)
)

func main() {
	flag.Parse()

	log.SetPrefix(os.Args[0] + " -- ")

	//listening for SIGUSR2 to provide hot config reload
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGUSR2)
	go func() {
		for {
			<-s
			_log("SIGUSR2 trying to reload config ...")
			err, tmpStreams := initStreams()
			if err != nil {
				_log("Couldn't relaod", err)
				continue
			}
			streamsLock.Lock()
			streams = tmpStreams
			streamsLock.Unlock()
			_log("Config relaoded")
		}
	}()

	var err error
	if err, streams = initStreams(); err != nil {
		//failed to read initial config, aborting
		log.Fatal(err)
	}

	reader := bufio.NewReader(os.Stdin)

	for {
		line, err := reader.ReadString('\n')

		if err != nil {
			if err != io.EOF {
				_log(err)
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

func initStreams() (err error, streams []*Stream) {
	var streamConfigs []StreamConfig
	file, err := ioutil.ReadFile(*configPath)
	if err != nil {
		return
	}

	confs := os.ExpandEnv(string(file))
	if err = json.Unmarshal([]byte(confs), &streamConfigs); err != nil {
		return
	}

	for _, conf := range streamConfigs {
		s, err := NewStream(&conf)
		if err != nil {
			_log(err)
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
		_log(err)
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
		tcpConn, dialErr := net.DialTimeout("tcp", path, DIAL_TIMEOUT)
		if dialErr != nil {
			config := &tls.Config{InsecureSkipVerify: s.Conf.AllowSSCert}
			conn = tls.Client(tcpConn, config)
		} else {
			err = dialErr
		}

	case s.Url.Scheme == "file":
		conn, err = os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)

	default:
		conn, err = net.DialTimeout(s.Url.Scheme, path, DIAL_TIMEOUT)

	}
	s.Conn = conn
	return err
}

func (s *Stream) Write(line string) {
	switch {
	case s.Status == CONNECTED:
		if _, err := s.Conn.Write([]byte(s.Conf.Prefix + line)); err != nil {
			s.Status = NOT_CONNECTED
			_log(err)
		}
	case s.Status == NOT_CONNECTED:
		s.TryConnect()
	}
}
