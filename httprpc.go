package httprpc

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

type RPCEncoder func(method string, args interface{}) ([]byte, error)
type RPCDecoder func(r io.Reader, reply interface{}) error

func CallRaw(address, method string, params, reply interface{}, contentType string, encoder RPCEncoder, decoder RPCDecoder) error {
	enc, err := encoder(method, params)
	if err != nil {
		return err
	}

	res, err := http.Post(address, contentType, bytes.NewBuffer(enc))
	if err != nil {
		return err
	}

	resText, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	reader := bytes.NewReader(resText)
	res.Body.Close()
	if res.StatusCode != 200 {
		return errors.New(fmt.Sprintf("%s: %s", res.Status, string(resText)))
	} else {
		reader.UnreadByte()

		err = decoder(reader, &reply)
		if err != nil {
			return err
		}

		return nil
	}

	return nil
}
