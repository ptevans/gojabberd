package main

import (
	"gojabberd/server"
	"net"
	"os"
)

func main() {
	println("Starting XMPP Server")

	//domainTable := new(server.DomainTable)
	presenceTable := new(server.PresenceTable)
	presenceTable.Users = make(map[string](*server.ClientConnection))
	listener, err := net.Listen("tcp", "0.0.0.0:5222")
	if err != nil {
		println("error!", err.Error())
		os.Exit(1)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			println("error!", err.Error())
			return
		}
		c := new(server.ClientConnection)
		go c.Go(conn, presenceTable)
	}
}
