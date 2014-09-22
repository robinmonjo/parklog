HARDWARE=$(shell uname -m)

test:
	go test

build:
	go build

release:
	mkdir -p release
	GOOS=linux go build -o release/parklog
	cd release && tar -zcf parklog_linux_$(HARDWARE).tgz parklog
	GOOS=darwin go build -o release/parklog
	cd release && tar -zcf parklog_darwin_$(HARDWARE).tgz parklog
	GOOS=linux GOARCH=arm GOARM=5 go build -o release/parklog
	cd release && tar -zcf parklog_linux_pi.tgz parklog
	rm release/parklog

clean:
	rm -rf release
