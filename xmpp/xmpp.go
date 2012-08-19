package xmpp

import (
	"bytes"
	"encoding/base64"
	"encoding/xml"
	"net"
	"strings"
)

const (
	RECV_BUF_LEN = 16384
)

type Stream struct {
	XMLName xml.Name `xml:"stream"`
	From    string   `xml:"from,attr"`
	Id      string   `xml:"id,attr"`
	To      string   `xml:"to,attr"`
	Version string   `xml:"version,attr"`
	Lang    string   `xml:"lang,attr"`
}

type StreamResponse struct {
	XMLName     xml.Name `xml:"stream:stream"`
	Xmlns       string   `xml:"xmlns,attr"`
	XmlnsStream string   `xml:"xmlns:stream,attr"`
	From        string   `xml:"from,attr"`
	To          string   `xml:"to,attr,omitempty"`
	Id          string   `xml:"id,attr"`
	Version     string   `xml:"version,attr"`
	Lang        string   `xml:"lang,attr"`
	Features    *StreamFeatures
}

type StreamFeatures struct {
	XMLName    xml.Name `xml:"stream:features"`
	Bind       Bind
	Session    Session
	Mechanisms Mechanisms `xml:",omitempty"`
}

type Bind struct {
	XMLName  xml.Name   `xml:"bind"`
	Xmlns    string     `xml:"xmlns,attr"`
	Resource []Resource `xml:"resource,omitempty"`
	Jid      []Jid      `xml:"jid,omitempty"`
}

type Session struct {
	XMLName xml.Name `xml:"session"`
	Xmlns   string   `xml:"xmlns,attr"`
}

type Jid struct {
	XMLName xml.Name `xml:"jid"`
	Data    string   `xml:",chardata"`
}

type Mechanisms struct {
	XMLName   xml.Name `xml:"mechanisms"`
	Xmlns     string   `xml:"xmlns,attr,omitempty"`
	Mechanism []Mechanism
}

type Mechanism struct {
	XMLName xml.Name `xml:"mechanism"`
	Data    string   `xml:",chardata"`
}

type Auth struct {
	XMLName   xml.Name `xml:"auth"`
	Xmlns     string   `xml:"xmlns,attr"`
	Mechanism string   `xml:"mechanism,attr"`
	Data      string   `xml:",chardata"`
}

func (a Auth) GetCredentials() (user string, pass string) {
	buf := bytes.NewBufferString(a.Data)
	decoder := base64.NewDecoder(base64.StdEncoding, buf)
	decodedCreds := make([]byte, 128)
	decoder.Read(decodedCreds)
	creds := bytes.Split(decodedCreds, []byte("\x00"))
	user, pass = string(creds[1]), string(creds[2])
	return
}

type Success struct {
	XMLName xml.Name `xml:"success"`
	Xmlns   string   `xml:"xmlns,attr"`
}

type Message struct {
	XMLName xml.Name `xml:"message"`
	From    string   `xml:"from,attr"`
	Id      string   `xml:"id,attr"`
	To      string   `xml:"to,attr"`
	Type    string   `xml:"type,attr"` // chat, error, groupchat, headline, normal
	Subject Subject
	Body    Body
	Thread  Thread
}

type Body struct {
	XMLName xml.Name `xml:"body"`
	Data    string   `xml:",chardata"`
	Lang    string   `xml:"lang,attr"`
}

type Subject struct {
	XMLName xml.Name `xml:"subject"`
	Data    string   `xml:",chardata"`
	Lang    string   `xml:"lang,attr"`
}

type Thread struct {
	XMLName xml.Name `xml:"thread"`
	Parent  string   `xml:"parent,attr"`
}

type Iq struct {
	XMLName xml.Name `xml:"iq"`
	To      string   `xml:"to,attr,omitempty"`
	From    string   `xml:"from,attr,omitempty"`
	Id      string   `xml:"id,attr"`
	Type    string   `xml:"type,attr"`
	Bind    []Bind   `xml:"bind,omitempty"`
	Query   []Query  `xml:"query,omitempty"`
}

type Query struct {
	XMLName xml.Name `xml:"query"`
	Xmlns   string   `xml:"xmlns,attr"`
}

type Resource struct {
	XMLName xml.Name `xml:"resource"`
	Data    string   `xml:",chardata"`
}

func ParseMessage(data string) (message Message) {
	xml.Unmarshal([]byte(data), &message)
	return
}

func ParseThread(data string) (thread Thread) {
	xml.Unmarshal([]byte(data), &thread)
	return
}

func ParseStreamOpening(data []byte) (stream Stream) {
	xml.Unmarshal(data, &stream)
	return
}

func HandleStanza(stanza []byte) (response []byte) {
	return
}

type PresenceTable struct {
	Users map[string](*ClientConnection)
}

type Domain struct {
	presence    PresenceTable
	domain      string
	mechanisms  []string // a list of SASL mechanims that can be used
	aliases     []string // a list of alternate domain names (e.g. localhost)
	tlsEnabled  bool
	tlsRequired bool
}

type ClientConnection struct {
	authenticated bool
	conn          net.Conn
	Chan          chan *[]byte
	username      string
	domain        string
	resource      string
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
	c.Chan = make(chan *[]byte)
	c.authenticated = false
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
	stream := new(Stream)
	xml.Unmarshal(buf, stream)
	c.domain = stream.To

	response := c.buildStreamOpening()
	c.conn.Write(response)

	// read the client's response
	c.readSocket(buf)
	authStanza := new(Auth)
	xml.Unmarshal(buf, authStanza)
	if authStanza.Mechanism == "PLAIN" {
		c.username, _ = authStanza.GetCredentials()
		c.authenticated = true
	}
	if !c.authenticated {
		println("Authentication failed!")
		// TODO: real auth response
		c.conn.Write([]byte("</stream>"))
		return
	}
	success := Success{Xmlns: authStanza.Xmlns}
	response, _ = xml.Marshal(success)
	c.conn.Write(response)

	// Read the client's stream restart
	c.readSocket(buf)
	response2 := c.buildStreamOpening()
	c.conn.Write(response2)

	// Receive and handle bind request
	c.readSocket(buf)
	iq := new(Iq)
	xml.Unmarshal(buf, iq)
	// TODO: this is terrible, resource won't always be there
	if len(iq.Bind[0].Resource) > 0 {
		c.resource = iq.Bind[0].Resource[0].Data
	} else {
		c.resource = "autogen"
	}
	responseIq := Iq{Type: "result", Id: iq.Id}
	b := Bind{Xmlns: iq.Bind[0].Xmlns}
	b.Jid = make([]Jid, 1)
	b.Jid[0] = Jid{Data: c.username + "@" + c.domain + "/" + c.resource}
	responseIq.Bind = make([]Bind, 1)
	responseIq.Bind[0] = b
	response, _ = xml.Marshal(responseIq)
	c.conn.Write(response)

	// Receive and handle session request
	c.readSocket(buf)
	iq = new(Iq)
	xml.Unmarshal(buf, iq)
	responseIq = Iq{Type: "result", Id: iq.Id, From: c.domain}
	response, _ = xml.Marshal(responseIq)
	c.conn.Write(response)

	// Receive and handle roster request
	c.readSocket(buf)

	iq = new(Iq)
	xml.Unmarshal(buf, iq)
	responseIq = Iq{Type: "result", To: iq.From, Id: iq.Id}
	responseIq.Query = make([]Query, 1)
	responseIq.Query[0].Xmlns = "jabber:iq:roster"
	response, _ = xml.Marshal(responseIq)
	c.conn.Write(response)
	return
}

func (c *ClientConnection) buildStreamOpening() (response []byte) {
	stream := StreamResponse{Id: "xxx", From: "localhost", Version: "1.0", Lang: "en", Xmlns: "jabber:client", XmlnsStream: "http://etherx.jabber.org/streams"}
	streamFeatures := new(StreamFeatures)
	bind := Bind{Xmlns: "urn:ietf:params:xml:ns:xmpp-bind"}
	session := Session{Xmlns: "urn:ietf:params:xml:ns:xmpp-session"}
	if !c.authenticated {
		mechanism := Mechanism{Data: "PLAIN"}
		mechanisms := Mechanisms{Xmlns: "urn:ietf:params:xml:ns:xmpp-sasl"}
		mechanisms.Mechanism = make([]Mechanism, 1)
		mechanisms.Mechanism[0] = mechanism
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
			message := new(Message)
			xml.Unmarshal(buf, message)
			username, _, _ := ParseJid(message.To)
			// TODO: test for this before just overwriting
			message.From = c.username + "@" + c.domain
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

func ParseJid(jid string) (username string, domain string, resource string) {
	components := strings.Split(jid, "/")
	if len(components) == 2 {
		resource = components[1]
		jid = components[0]
	} else {
		resource = "/"
	}
	components = strings.Split(jid, "@")
	if len(components) == 2 {
		domain = components[1]
	}
	username = components[0]
	return
}
