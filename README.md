## Parklog

Simple tool written in Go to redirect your app stdout/stderr to different local or remote endpoints.

### Usage

`bundle exec rails s | parklog`

This will redirect your output on endpoints specified in the `parklog.json` file. This file must be set in your app root directory (as a `Procfile` for instance) and may look like this:

````json
[
  {
    "url":"file:///dev/stdout",
    "prefix":"Rails app on stdout - "
  },
  {
    "url":"tcp://localhost:9999"
  },
  {
    "url":"tcp://localhost:9998",
    "prefix":"Rails app log on 9998 - "
  },
  {
    "url":"tls://localhost:9997",
    "allow_self_signed_cert": true
  }
]
````
You can redirect to several endpoints including local files, tcp servers, tls/ssl servers etc ...
For a full list of supported endpoints refer to [golang `net` package](http://golang.org/pkg/net/#Dial)

You can also inject environment variable inside the `parklog.json`, wherever you want:

````bash
export PORT_A=9999
export URI="file:///dev/stdout"
export PREFIX="Rails app log on 9998 - "
````

````json
[
  {
    "url":"file://$PWD/log.out"
  },
  {
    "url":"$URI",
    "prefix":"Rails app on stdout - "
  },
  {
    "url":"tcp://localhost:$PORT_A"
  },
  {
    "url":"tcp://localhost:9998",
    "prefix":"$PREFIX"
  },
  {
    "url":"tls://localhost:9997",
    "allow_self_signed_cert": true
  }
]
````

### Installation

## OSX

`curl -sL https://github.com/robinmonjo/parklog/releases/download/v0.1.0/parklog_darwin_x86_64.tgz | tar -C /usr/local/bin -zxf -`

## Linux arm (Raspbery Pi)

`curl -sL https://github.com/robinmonjo/parklog/releases/download/v0.1.0/parklog_linux_pi.tgz | tar -C /usr/local/bin -zxf -`

## Linux

``curl -sL https://github.com/robinmonjo/parklog/releases/download/v0.1.0/parklog_linux_x86_64.tgz | tar -C /usr/local/bin -zxf -``

### License

MIT

### Notes for testing

Generate a private key and certificate and launch a tls server

````bash
openssl req -x509 -nodes -days 365 -newkey rsa:1024 -keyout key.pem -out cert.pem
cat key.pem cert.pem > full_cert.pem
openssl s_server -cert full_cert.pem -port 9997
````
