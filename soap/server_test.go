package soap

import (
	"bytes"
	"encoding/xml"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/rpc/v2"
)

var ErrResponseError = errors.New("response error")

type clientResponseEnvelope struct {
	XMLName xml.Name   `xml:"Envelope"`
	ENV     string     `xml:"xmlns,attr"`
	Header  clientData `xml:"Header,omitempty"`
	Body    clientData `xml:"Body"`
}

type clientData struct {
	Data []byte `xml:",innerxml"`
}

type Service1Request struct {
	A int
	B int
}

func (s *Service1Request) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	e.EncodeElement(s.A, xml.StartElement{Name: xml.Name{Local: "A"}})
	e.EncodeElement(s.B, xml.StartElement{Name: xml.Name{Local: "B"}})
	return nil
}

type Service1BadRequest struct {
}

type Service1Response struct {
	Result       int
	ErrorMessage string `xml:"Fault>faultstring,omitempty"`
}

type Service1 struct {
}

func (t *Service1) Multiply(r *http.Request, req *Service1Request, res *Service1Response) error {
	res.Result = req.A * req.B
	return nil
}

func (t *Service1) ResponseError(r *http.Request, req *Service1Request, res *Service1Response) error {
	return ErrResponseError
}

func execute(t *testing.T, s *rpc.Server, port, method string, req, res interface{}) (int, error) {
	if !s.HasMethod(port + "." + method) {
		t.Fatal("Expected to be registered:", port+"."+method)
	}

	params, err := xml.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}

	buf := &bytes.Buffer{}
	buf.WriteString(`
		<Envelope xmlns="http://schemas.xmlsoap.org/soap/envelope/">
		  <Body>
		    <`)
	buf.WriteString(method)
	buf.WriteString(">")
	buf.Write(params)
	buf.WriteString("</")
	buf.WriteString(method)
	buf.WriteString(`>
		  </Body>
		</Envelope>`)

	r, _ := http.NewRequest("POST", "http://localhost:8080/"+port, buf)
	r.Header.Set("Content-Type", "application/soap+xml")

	w := httptest.NewRecorder()
	s.ServeHTTP(w, r)

	var response clientResponseEnvelope
	err = xml.NewDecoder(w.Body).Decode(&response)
	if err != nil {
		return w.Code, err
	}

	err = xml.Unmarshal([]byte("<root>"+string(response.Body.Data)+"</root>"), res)

	return w.Code, err
}

func TestService(t *testing.T) {
	s := rpc.NewServer()
	s.RegisterCodec(NewCodec(), "application/soap+xml")
	s.RegisterService(new(Service1), "")

	var res Service1Response
	if _, err := execute(t, s, "Service1", "Multiply", &Service1Request{4, 2}, &res); err != nil {
		t.Error("Expected err to be nil, but got:", err)
	}
	if res.Result != 8 {
		t.Error("Expected res.Result to be 8, but got:", res.Result)
	}
	if res.ErrorMessage != "" {
		t.Error("Expected error_message to be empty, but got:", res.ErrorMessage)
	}
	if code, err := execute(t, s, "Service1", "ResponseError", &Service1Request{4, 2}, &res); err != nil || code != 400 {
		t.Errorf("Expected code to be 400 and error to be nil, but got %v (%v)", code, err)
	}
	if res.ErrorMessage == "" {
		t.Errorf("Expected error_message to be %q, but got %q", ErrResponseError, res.ErrorMessage)
	}
}
