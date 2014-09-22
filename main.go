package main

import (
	"bufio"
	"flag"
	"io"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var (
	verbose    *bool   = flag.Bool("v", false, "verbose")
	configPath *string = flag.String("c", "./parklog.json", "config file path")

	streams     Streams
	streamsLock = new(sync.RWMutex)
)

func _log(v ...interface{}) {
	if *verbose {
		log.Println(v)
	}
}

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
			err, tmpStreams := InitStreams(*configPath)
			if err != nil {
				_log("Couldn't reload", err)
				continue
			}
			streamsLock.Lock()
			streams.CloseAll()
			streams = tmpStreams
			tmpStreams = nil
			streamsLock.Unlock()
			_log("Config reloaded")
		}
	}()

	var err error
	if err, streams = InitStreams(*configPath); err != nil {
		//failed with initial config, aborting
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
		if errs := streams.WriteAll(line); err != nil {
			_log(errs)
		}
	}
	streams.CloseAll()
}
