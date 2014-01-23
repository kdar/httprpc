package httprpc

import (
	"errors"

	gjson "github.com/gorilla/rpc/v2/json"
	gjson2 "github.com/gorilla/rpc/v2/json2"
)

func CallJson(version, address, method string, params, reply interface{}) error {
	switch version {
	case "10", "1.0", "v10", "v1.0":
		return CallRaw(address,
			method, &params, &reply, "application/json",
			gjson.EncodeClientRequest, gjson.DecodeClientResponse)
	case "11", "1.1", "v11", "v1.1":
		// not implemented yet
	case "20", "2.0", "v20", "v2.0":
		return CallRaw(address,
			method, &params, &reply, "application/json",
			gjson2.EncodeClientRequest, gjson2.DecodeClientResponse)
	default:
		return errors.New("JsonRPC Version not recognized")
	}

	return nil
}
