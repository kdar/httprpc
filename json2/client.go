// HTTP JSON Gorilla RPC for golang
package json2

import (
  "encoding/json"
  "fmt"
  "io"
  "math/rand"
)

// clientRequest represents a JSON-RPC v2.0 request sent by a client.
type clientRequest struct {
  // A String containing the name of the method to be invoked.
  Method string `json:"method"`
  // Object to pass as request parameter to the method.
  Params interface{} `json:"params"`
  // The request id. This can be of any type. It is used to match the
  // response with the request that it is replying to.
  Id uint64 `json:"id"`
  // Always set to 2.0
  Jsonrpc string `json:"jsonrpc"`
}

// clientResponse represents a JSON-RPC response returned to a client.
type clientResponse struct {
  Result  *json.RawMessage `json:"result"`
  Error   interface{}      `json:"error"` // {"code": -32700, "message": "Parse error"}
  Id      uint64           `json:"id"`
  Jsonrpc string           `json:"jsonrpc"`
}

// EncodeClientRequest encodes parameters for a JSON-RPC v2.0 client request.
func EncodeClientRequest(method string, args interface{}) ([]byte, error) {
  c := &clientRequest{
    Method:  method,
    Params:  args,
    Id:      uint64(rand.Int63()),
    Jsonrpc: "2.0",
  }
  return json.Marshal(c)
}

// DecodeClientResponse decodes the response body of a client request into
// the interface reply.
func DecodeClientResponse(r io.Reader, reply interface{}) error {
  var c clientResponse
  if err := json.NewDecoder(r).Decode(&c); err != nil {
    return err
  }
  if c.Error != nil {
    return fmt.Errorf("%v", c.Error)
  }
  return json.Unmarshal(*c.Result, reply)
}
