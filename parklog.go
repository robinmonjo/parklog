package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"time"
)

//Constants
const DIAL_TIMEOUT = 5 * time.Second

type StreamStatus int

const (
	CONNECTED StreamStatus = iota
	CONNECTING
	NOT_CONNECTED
)

//Types
type Streams []*Stream

func (streams *Streams) CloseAll() {
	for _, stream := range *streams {
		stream.Conn.Close()
	}
}

func (streams *Streams) WriteAll(line string) (errors []error) {
	for _, stream := range *streams {
		if err := stream.Write(line); err != nil {
			errors = append(errors, err)
		}
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

func (s *Stream) TryConnect() error {
	s.Status = CONNECTING
	if err := s.Connect(); err != nil {
		s.Status = NOT_CONNECTED
		return err
	}
	s.Status = CONNECTED
	return nil
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

func (s *Stream) Write(line string) error {
	switch {
	case s.Status == CONNECTED:
		toWrite := []byte(s.Conf.Prefix + line)
		n, err := s.Conn.Write(toWrite)
		if err != nil {
			s.Status = NOT_CONNECTED
			return err
		}
		if n != len(toWrite) {
			return errors.New("Failed to write some bytes on " /* + s.Url*/)
		}
	case s.Status == NOT_CONNECTED:
		s.TryConnect()
	}
	return nil
}

func InitStreams(configPath string) (error, Streams) {
	var streamConfigs []StreamConfig
	file, err := ioutil.ReadFile(configPath)
	if err != nil {
		return err, nil
	}

	confs := os.ExpandEnv(string(file))
	if err = json.Unmarshal([]byte(confs), &streamConfigs); err != nil {
		return err, nil
	}

	var streams Streams
	for _, conf := range streamConfigs {
		s, err := NewStream(&conf)
		if err != nil {
			return err, nil
		}
		streams = append(streams, s)
	}
	return nil, streams
}
