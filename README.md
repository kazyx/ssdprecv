# ssdprecv
SSDP receiver module written in Golang

```go
// Pass information of HTTP server to SsdpReceiver
http := new(ssdprecv.HttpServer)
http.Port = 9000 // port of HTTP server
http.Path = "/description.xml" // path to device description
http.Uuid = "uuid:00000000-0000-0000-0000-000000000000" // UDN of your server
http.ST = []string{"your-search-target-to-respond"} // multiple STs are supported
http.NIF = nif // network interface to listen

receiver := ssdprecv.New(http)
ch := make(chan int)

receiver.Boot(ch)
```

### Note
To complete SSDP server function, HTTP server is required aside from this.
