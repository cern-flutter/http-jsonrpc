http-jsonrc
===========
RPC client codec implementing jsonrpc over http.

## Usage
```go
func foo() {
    codec, err := http_jsonrpc.NewClientCodec("http://localhost:8080/rpc")
    if err != nil {
        log.Fatal(err)
    }

    client := rpc.NewClientWithCodec(codec)

    var request string
    var reply string
    request = "Hello there"
    if err := client.Call("Test.Echo", &request, &reply); err != nil {
        log.Fatal(err)
    }
}
```