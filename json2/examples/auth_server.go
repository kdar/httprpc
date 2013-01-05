package main

import (
	auth "github.com/abbot/go-http-auth"
	"github.com/gorilla/rpc"
	"github.com/gorilla/rpc/json"
	"log"
	"net/http"
)

var (
	rpcServer *rpc.Server
)

type HelloArgs struct {
	Who string
}

type HelloReply struct {
	Message string
}

type HelloService struct{}

func (h *HelloService) Say(r *http.Request, args *HelloArgs, reply *HelloReply) error {
	reply.Message = "Hello, " + args.Who + "!"
	return nil
}

func Secret(user, realm string) string {
	if user == "john" {
		return string(auth.MD5Crypt([]byte("hello"), []byte("435345"), []byte("$1$")))
	}
	return ""
}

func handler(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
	rpcServer.ServeHTTP(w, &r.Request)
}

func main() {
	authenticator := auth.BasicAuthenticator("example.com", Secret)

	rpcServer = rpc.NewServer()
	rpcServer.RegisterCodec(json.NewCodec(), "application/json")
	rpcServer.RegisterService(new(HelloService), "")
	http.HandleFunc("/rpc", authenticator(handler))

	err := http.ListenAndServe(":9000", nil)
	if err != nil {
		log.Fatal(err)
	}
}
