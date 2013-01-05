package main

import (
  "fmt"
  "github.com/kdar/httprpc"
  "log"
)

type Args struct {
  Who string
}

type Reply struct {
  Message string
}

func main() {
  var reply Reply
  err := httprpc.CallJson("1.0", "http://localhost:9000/rpc", "HelloService.Say", &Args{"kevin"}, &reply)
  if err != nil {
    log.Fatal(err)
  }
  fmt.Println(reply.Message)
}
