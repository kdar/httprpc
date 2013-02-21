package soap

import (
  "encoding/xml"
  "errors"
  "fmt"
  "github.com/gorilla/rpc"
  "net/http"
  "reflect"
)

type RequestEnvelope struct {
  XMLName xml.Name    `xml:"http://schemas.xmlsoap.org/soap/envelope/ Envelope"`
  Header  []byte      `xml:"http://schemas.xmlsoap.org/soap/envelope/ Header,omitempty"`
  Body    RequestBody `xml:"http://schemas.xmlsoap.org/soap/envelope/ Body"`
}

type RequestBody struct {
  //Test string `xml:",innerxml"`
  XMLName  xml.Name
  Innerxml string "innerxml"
  Method   Method `xml:",any"`
}

type Method struct {
  XMLName xml.Name
  Params  []byte `xml:",innerxml"`
}

type ResponseEnvelope struct {
  XMLName xml.Name    `xml:"SOAP-ENV:Envelope"`
  ENV     string      `xml:"xmlns:SOAP-ENV,attr"`
  Header  []byte      `xml:"SOAP-ENV:Header,omitempty"`
  Body    interface{} `xml:"SOAP-ENV:Body"`
  //Body ResponseBody `xml:"SOAP-ENV:Body"`
}

type ResponseBody struct {
  XMLName xml.Name    `xml:"SOAP-ENV:Body"`
  Data    interface{} `xml:",innerxml"`
}

type Fault struct {
  XMLName     xml.Name    `xml:"SOAP-ENV:Fault"`
  FaultCode   string      `xml:"faultcode"`
  FaultString string      `xml:"faultstring"`
  Detail      interface{} `xml:"detail"`
}

func NewResponse(data interface{}) *ResponseEnvelope {
  return &ResponseEnvelope{
    ENV:  "http://schemas.xmlsoap.org/soap/envelope/",
    Body: data,
  }
}

func NewFault(s string, detail interface{}) *ResponseEnvelope {
  return NewResponse(&ResponseBody{Data: &Fault{
    FaultCode:   "SOAP-ENV:Client",
    FaultString: s,
    Detail:      detail,
  }})
}

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
  Port     string
  Envelope RequestEnvelope
}
type serverResponse ResponseEnvelope

// newCodecRequest returns a new CodecRequest.
func newCodecRequest(r *http.Request) rpc.CodecRequest {
  req := new(serverRequest)
  err := xml.NewDecoder(r.Body).Decode(&req.Envelope)
  r.Body.Close()

  req.Port = r.Form.Get("_soap_port")

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
    return fmt.Sprintf("%s.%s", c.request.Port, c.request.Envelope.Body.Method.XMLName.Local), nil
  }
  return "", c.err
}

// ReadRequest fills the request object for the RPC method.
func (c *CodecRequest) ReadRequest(args interface{}) error {
  rv := reflect.ValueOf(args)
  if rv.Kind() != reflect.Ptr || rv.IsNil() {
    c.err = errors.New("rpc: args not of proper type")
  } else {
    xmlstr := "<root>" + string(c.request.Envelope.Body.Method.Params) + "</root>"
    c.err = xml.Unmarshal([]byte(xmlstr), args)
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

  var res *ResponseEnvelope

  if methodErr != nil {
    res = NewFault(methodErr.Error(), nil)
  } else {
    res = NewResponse(reply)
  }

  w.Header().Set("Content-Type", "application/soap+xml; charset=utf-8")
  encoder := xml.NewEncoder(w)
  err := encoder.Encode(res)

  return err
}
