package soap

import (
	"encoding/xml"
	"log"
	"testing"
)

type Response struct {
	Nested struct {
		Success bool
	}
}

func TestXMLResponse(t *testing.T) {
	reply := &Response{}
	reply.Nested.Success = true
	res := NewResponse(reply)
	b, err := xml.Marshal(res)
	if err != nil {
		t.Fatal(err)
	}

	log.Println(string(b))
}

func TestXMLFault(t *testing.T) {
	reply := &Response{}
	reply.Nested.Success = true
	res := NewFault("test", reply)
	b, err := xml.Marshal(res)
	if err != nil {
		t.Fatal(err)
	}

	log.Println(string(b))
}
