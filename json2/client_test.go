package json2

import (
	"bytes"
	"encoding/json"
	//"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

type Params struct {
	A string
	B int
}

func execute(url, method string, req, res interface{}) error {
	buf, _ := EncodeClientRequest(method, req)
	body := bytes.NewBuffer(buf)
	r, _ := http.NewRequest("POST", url, body)
	r.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(r)
	if err != nil {
		return err
	}

	return DecodeClientResponse(resp.Body, res)
}

func newServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}

		var req clientRequest
		err = json.Unmarshal(got, &req)
		if err != nil {
			t.Fatalf("Unmarshal json: %s", err)
		}

		response := &clientResponse{Id: req.Id, Jsonrpc: "2.0"}
		switch req.Method {
		case "Echo":
			result, err := json.Marshal(req.Params)
			if err != nil {
				t.Fatalf("Marshal json: %s", err)
			}
			rtmp := json.RawMessage(result)
			response.Result = &rtmp
		case "WillError":
			w.Write([]byte("dksjflksjdf"))
		}

		responseText, err := json.Marshal(&response)
		//fmt.Println(string(responseText))
		w.Write(responseText)
	}))
}

func TestClient(t *testing.T) {
	ts := newServer(t)
	defer ts.Close()

	var reply1 Params
	params1 := Params{"hey", 5}

	err := execute(ts.URL, "Echo", &params1, &reply1)
	if err != nil {
		t.Fatal(err)
	} else if reply1 != params1 {
		t.Errorf("Expected to get %v, but got %v", params1, reply1)
	}

	err = execute(ts.URL, "WillError", &params1, &reply1)
	if err == nil {
		t.Fatalf("WillError: Expected error")
	}
}
