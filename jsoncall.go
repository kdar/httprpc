package httprpc

import (
	"errors"
	gjson "github.com/gorilla/rpc/json"
	hjson2 "github.com/kdar/httprpc/json2"
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
			hjson2.EncodeClientRequest, hjson2.DecodeClientResponse)
	default:
		return errors.New("JsonRPC Version not recognized")
	}

	return nil
}
