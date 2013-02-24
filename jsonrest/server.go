// A codec for gorilla rpc to handle REST
// Follows the JSONRPC 2.0 spec.
package jsonrest

import (
  "encoding/json"
  "errors"
  "github.com/gorilla/rpc"
  "io/ioutil"
  "net/http"
  "net/url"
  "reflect"
)

var null = json.RawMessage([]byte("null"))

// ----------------------------------------------------------------------------
// Codec
// ----------------------------------------------------------------------------

// NewCodec returns a new RESTRPC Codec.
func NewCodec() *Codec {
  return &Codec{}
}

// Codec creates a CodecRequest to process each request.
type Codec struct {
}

// NewRequest returns a CodecRequest.
func (c *Codec) NewRequest(r *http.Request) rpc.CodecRequest {
  return newCodecRequest(r)
}

// ----------------------------------------------------------------------------
// CodecRequest
// ----------------------------------------------------------------------------

type serverRequest struct {
  Method string     `json:"method"`
  Params url.Values `json:"params"`
  // The request id. This can be of any type. It is used to match the
  // response with the request that it is replying to.
  Id *json.RawMessage `json:"id"`

  // You can post to the REST RPC and the data will be here.
  // This will eventually be unmarshalled into the args before
  // the Params are set.
  Body []byte `json:"-"`
}

type serverResponse struct {
  // The Object that was returned by the invoked method. This must be null
  // in case there was an error invoking the method.
  Result interface{} `json:"result"`
  // An Error object if there was an error invoking the method. It must be
  // null if there was no error.
  Error interface{} `json:"error"`
  // This must be the same id as the request it is responding to.
  Id *json.RawMessage `json:"id"`
}

// newCodecRequest returns a new CodecRequest.
func newCodecRequest(r *http.Request) rpc.CodecRequest {
  var err error
  req := new(serverRequest)
  req.Body, err = ioutil.ReadAll(r.Body)
  r.Body.Close()

  // Has to be populated using whatever web framework you're using.
  req.Method = r.Form.Get("_rest_method")
  req.Params = r.URL.Query()

  // Get the id of the message. If no id is passed, set the id
  // to an empty string.
  id := req.Params.Get("id")
  if len(id) == 0 {
    id = `""`
  }
  reqid := json.RawMessage(id)
  req.Id = &reqid

  return &CodecRequest{request: req, err: err}
}

// CodecRequest decodes and encodes a single request.
type CodecRequest struct {
  request *serverRequest
  err     error
}

// Method returns the RPC method for the current request.
//
// The method uses a dotted notation as in "Service.Method".
func (c *CodecRequest) Method() (string, error) {
  if c.err == nil {
    return c.request.Method, nil
  }
  return "", c.err
}

// ReadRequest fills the request object for the RPC method.
func (c *CodecRequest) ReadRequest(args interface{}) error {
  if len(c.request.Body) > 0 {
    c.err = json.Unmarshal(c.request.Body, &args)
  }

  rv := reflect.ValueOf(args)
  if rv.Kind() != reflect.Ptr || rv.IsNil() {
    c.err = errors.New("rpc: args not of proper type")
  } else {
    elem := rv.Elem()
    for key, value := range c.request.Params {
      v := elem.FieldByName(key)
      if v.IsValid() {
        v.SetString(value[0])
      }
    }
  }

  return c.err
}

// WriteResponse encodes the response and writes it to the ResponseWriter.
//
// The err parameter is the error resulted from calling the RPC method,
// or nil if there was no error.
func (c *CodecRequest) WriteResponse(w http.ResponseWriter, reply interface{}, methodErr error) error {
  if c.err != nil {
    return c.err
  }

  res := &serverResponse{
    Result: reply,
    Error:  &null,
    Id:     c.request.Id,
  }

  if methodErr != nil {
    // Propagate error message as string.
    res.Error = methodErr.Error()
    // Result must be null if there was an error invoking the method.
    res.Result = &null
  }

  w.Header().Set("Content-Type", "application/rest+json; charset=utf-8")
  encoder := json.NewEncoder(w)
  err := encoder.Encode(res)

  return err
}
