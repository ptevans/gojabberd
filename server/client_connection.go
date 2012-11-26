
package server

import (
	"net"
	"encoding/xml"
	"gojabberd/auth"
	"gojabberd/xmpp"
	"gojabberd/xmpp/stanza"
)

const (
	RECV_BUF_LEN = 16384
)

type ClientConnection struct {
	authenticated bool
	conn          net.Conn
	Chan          chan *[]byte
	username      string
	domainpart    string
	resource      string
	domain        *Domain
	presenceTable *PresenceTable
	kill          chan bool
	died          chan bool
	shutdown      chan bool
}

func (c *ClientConnection) Go(conn net.Conn, presenceTable *PresenceTable) {
	defer func() {
		if err := recover(); err != nil {
			println("Connection terminated!", err)
			c.kill <- true
			c.kill <- true
			delete(c.presenceTable.Users, c.username)
		}
	}()
	c.conn = conn
	c.presenceTable = presenceTable
	c.authenticated = false
	c.Chan = make(chan *[]byte)
	c.died = make(chan bool)
	c.kill = make(chan bool)
	c.shutdown = make(chan bool)
	err := c.negotiateStream()
	if err != nil {
		println("Failed to establish connection!")
		return
	}
	c.presenceTable.Users[c.username] = c
	go c.listenOnSocket()
	go c.listenOnChannel()
	select {
	case <-c.died:
		panic("Error reading from socket! Sutting down connection...")
	case <-c.shutdown:
		c.conn.Write([]byte("</stream>"))
		panic("Closing connection.")
	}
}

func (c *ClientConnection) readSocket(buf []byte) {
	n, err := c.conn.Read(buf)
	if err != nil {
		println("Error reading socket!")
		c.died <- true
	}
	println("received ", n, " bytes of data =", string(buf))
	return
}

func (c *ClientConnection) negotiateStream() (e error) {
	buf := make([]byte, RECV_BUF_LEN)
	c.readSocket(buf)
	stream := new(stanza.Stream)
	xml.Unmarshal(buf, stream)
	c.domainpart = stream.To

	response := c.buildStreamOpening()
	c.conn.Write(response)

	// read the client's response
	c.readSocket(buf)
	authStanza := new(stanza.Auth)
	xml.Unmarshal(buf, authStanza)
	switch {
	case authStanza.Mechanism == "DIGEST-MD5":
		challenge := stanza.NewChallenge(c.domainpart)
		out, _ := xml.Marshal(challenge)
		c.conn.Write(out)
		c.readSocket(buf)
		var response stanza.Response
		xml.Unmarshal(buf, response)
		_, challenge.Data = auth.ResponseDigestMd5(response.Data)
		out, _ = xml.Marshal(challenge)
		c.conn.Write(out)
		c.authenticated = true
		c.readSocket(buf)
	case authStanza.Mechanism == "PLAIN":
		c.username, _ = authStanza.GetCredentials()
		c.authenticated = true
	default:
		// TODO: real auth response
		c.conn.Write([]byte("</stream>"))
	}
	if !c.authenticated {
		println("Authentication failed!")
		// TODO: real auth response
		c.conn.Write([]byte("</stream>"))
		return
	}
	success := stanza.Success{Xmlns: authStanza.Xmlns}
	response, _ = xml.Marshal(success)
	c.conn.Write(response)

	// Read the client's stream restart
	c.readSocket(buf)
	response2 := c.buildStreamOpening()
	c.conn.Write(response2)

	// Receive and handle bind request
	c.readSocket(buf)
	iq := new(stanza.Iq)
	xml.Unmarshal(buf, iq)
	// TODO: this is terrible, resource won't always be there
	if len(iq.Bind[0].Resource) > 0 {
		c.resource = iq.Bind[0].Resource[0].Data
	} else {
		c.resource = "autogen"
	}
	responseIq := stanza.Iq{Type: "result", Id: iq.Id}
	b := stanza.Bind{Xmlns: iq.Bind[0].Xmlns}
	b.Jid = make([]stanza.Jid, 1)
	b.Jid[0] = stanza.Jid{Data: c.username + "@" + c.domainpart + "/" + c.resource}
	responseIq.Bind = make([]stanza.Bind, 1)
	responseIq.Bind[0] = b
	response, _ = xml.Marshal(responseIq)
	c.conn.Write(response)

	// Receive and handle session request
	c.readSocket(buf)
	iq = new(stanza.Iq)
	xml.Unmarshal(buf, iq)
	responseIq = stanza.Iq{Type: "result", Id: iq.Id, From: c.domainpart}
	response, _ = xml.Marshal(responseIq)
	c.conn.Write(response)

	// Receive and handle roster request
	c.readSocket(buf)

	iq = new(stanza.Iq)
	xml.Unmarshal(buf, iq)
	responseIq = stanza.Iq{Type: "result", To: iq.From, Id: iq.Id}
	responseIq.Query = make([]stanza.Query, 1)
	responseIq.Query[0].Xmlns = "jabber:iq:roster"
	response, _ = xml.Marshal(responseIq)
	c.conn.Write(response)
	return
}

func (c *ClientConnection) buildStreamOpening() (response []byte) {
	stream := stanza.StreamResponse{Id: "zzz", From: "localhost", Version: "1.0", Lang: "en", Xmlns: "jabber:client", XmlnsStream: "http://etherx.jabber.org/streams"}
	streamFeatures := new(stanza.StreamFeatures)
	bind := stanza.Bind{Xmlns: "urn:ietf:params:xml:ns:xmpp-bind"}
	session := stanza.Session{Xmlns: "urn:ietf:params:xml:ns:xmpp-session"}
	if !c.authenticated {
		mechanisms := stanza.Mechanisms{Xmlns: "urn:ietf:params:xml:ns:xmpp-sasl"}
		mechanisms.Mechanism = make([]stanza.Mechanism, 4)
		mechanisms.Mechanism[0] = stanza.Mechanism{Data: "DIGEST-MD5"}
		mechanisms.Mechanism[1] = stanza.Mechanism{Data: "PLAIN"}
		streamFeatures.Mechanisms = mechanisms
	}

	streamFeatures.Bind = bind
	streamFeatures.Session = session
	stream.Features = streamFeatures
	response, err := xml.Marshal(stream)
	if err != nil {
		panic("Could not marshal stream opening")
	}
	// Remove the stream close tag as that isn't how XMPP rolls
	last := len(response)
	response = response[0 : last-16]
	return
}

func (c *ClientConnection) listenOnSocket() {
	buf := make([]byte, RECV_BUF_LEN)
	for {
		select {
		case <-c.kill:
			return
		default:
			c.readSocket(buf)
			// TODO: this is a bunch of crap :)
			message := new(stanza.Message)
			xml.Unmarshal(buf, message)
			username, _, _ := xmpp.ParseJid(message.To)
			// TODO: test for this before just overwriting
			message.From = c.username + "@" + c.domainpart
			if recipient, ok := c.presenceTable.Users[username]; ok {
				buf, _ := xml.Marshal(message)
				recipient.Chan <- &buf
			} else {
				println("recipient not found in presence table")
			}
		}
	}
}

func (c *ClientConnection) listenOnChannel() {
	for {
		select {
		case buf := <-c.Chan:
			println("received message on channel: ", string(*buf))
			c.conn.Write(*buf)
		case <-c.kill:
			return
		}
	}
}

