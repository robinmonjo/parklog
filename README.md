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
    "url":"tls://localhost:9997"
  }
]
````

You can redirect to several endpoints including local files, tcp servers, tls/ssl servers etc ...
For a full list of supported endpoints refer to [golang `net` package](http://golang.org/pkg/net/#Dial)

### Todo

* Allow tls connection to strictly check the server cert
* Web / script hook on start and on stop ?
* Log rotation if file:// ?
* syslog support

### Notes for testing

Generate a private key and certificate and launch a tls server

````bash
openssl req -x509 -nodes -days 365 -newkey rsa:1024 -keyout key.pem -out cert.pem
cat key.pem cert.pem > full_cert.pem
openssl s_server -cert full_cert.pem -port 9997
````
