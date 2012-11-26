package server

import (
	"net"
	"os"
)

func Start() {
	println("Starting XMPP Server")

	domainTable := new(DomainTable)
	domainTable.domains = make(map[string](*Domain))
	domainTable.RegisterDomain(NewDomain("localhost"))
	//presenceTable := new(PresenceTable)
	//presenceTable.Users = make(map[string](*ClientConnection))
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
		c := new(ClientConnection)
		go c.Go(conn, domainTable)
	}
}

