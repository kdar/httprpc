package soap

import (
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	rpc "github.com/gorilla/rpc/v2"
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
	XMLName xml.Name    `xml:"Envelope"`
	ENV     string      `xml:"xmlns,attr"`
	Header  []byte      `xml:"Header,omitempty"`
	Body    interface{} `xml:"Body"`
	//Body ResponseBody `xml:"SOAP-ENV:Body"`
}

type ResponseBody struct {
	XMLName xml.Name    `xml:"Body"`
	Data    interface{} `xml:",innerxml"`
}

type Fault struct {
	XMLName     xml.Name    `xml:"Fault"`
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

func NewFault(c, s string, detail interface{}) *ResponseEnvelope {
	if c == "" {
		c = "Client"
	}
	return NewResponse(&ResponseBody{Data: &Fault{
		FaultCode:   c,
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

	path := r.URL.Path
	index := strings.LastIndex(path, "/")
	if index < 0 {
		return &CodecRequest{request: req, err: fmt.Errorf("soap: no port: %s", path)}
	}
	req.Port = path[index+1:]

	//req.Port = r.Form.Get("_soap_port")

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
	if c.err == nil {
		rv := reflect.ValueOf(args)
		if rv.Kind() != reflect.Ptr || rv.IsNil() {
			c.err = errors.New("rpc: args not of proper type")
		} else {
			xmlstr := "<root>" + string(c.request.Envelope.Body.Method.Params) + "</root>"
			c.err = xml.Unmarshal([]byte(xmlstr), args)
		}
	}

	return c.err
}

// WriteResponse encodes the response and writes it to the ResponseWriter.
func (c *CodecRequest) WriteResponse(w http.ResponseWriter, reply interface{}) {
	res := NewResponse(reply)
	w.Header().Set("Content-Type", "application/soap+xml; charset=utf-8")
	encoder := xml.NewEncoder(w)
	encoder.Encode(res)
}

func (c *CodecRequest) WriteError(w http.ResponseWriter, status int, err error) {
	res := NewFault("", err.Error(), nil)
	w.Header().Set("Content-Type", "application/soap+xml; charset=utf-8")
	w.WriteHeader(status)
	encoder := xml.NewEncoder(w)
	encoder.Encode(res)
}
