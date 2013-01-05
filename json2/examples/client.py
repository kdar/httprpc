from lovely.jsonrpc import proxy

s = proxy.ServerProxy('http://localhost:9000/rpc', send_id=True)
reply = getattr(s, "HelloService.Say")({'Who':'kevin'})
print(reply)