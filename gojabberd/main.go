package main

import (
	"gojabberd/xmpp"
	"net"
	"os"
)

func main() {
	println("Starting XMPP Server")

	presenceTable := new(xmpp.PresenceTable)
	presenceTable.Users = make(map[string](*xmpp.ClientConnection))
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
		c := new(xmpp.ClientConnection)
		go c.Go(conn, presenceTable)
	}
}
