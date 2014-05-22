### Parklog

Tool to route your app stdout/stderr.

### Usage

`bundle exec rails s | parklog`

This will redirect your output on endpoints specified in the `parklog.json` file. This file may look like this:

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

````json
export PORT_A=9999
export URI="file:///dev/stdout"
export PREFIX="Rails app log on 9998 - "

[
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

### Notes for testing

Generate a private key and certificate and launch a tls server

````bash
openssl req -x509 -nodes -days 365 -newkey rsa:1024 -keyout key.pem -out cert.pem
cat key.pem cert.pem > full_cert.pem
openssl s_server -cert full_cert.pem -port 9997
````
