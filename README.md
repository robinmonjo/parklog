* read a config file in json to know where to send logs
* allow TCP, SYSLOG, SSL, (FILE + log rotation ? by saying I want to keep 5 mb of logs for example)

{
  "streamer": [
    {
      "uri" : "tcp:// ssl:// file:// syslog://",
      "max_size" : "100kb"
    }
  ]
}
