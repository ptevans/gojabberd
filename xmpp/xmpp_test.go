package xmpp

import "testing"
import "encoding/xml"

func TestParseMessage(t *testing.T) {
	data := `
	<message from="jimbo@localhost" >
		<body xml:lang="en">WTFBBQ</body>
		<thread parent="0000-000-000" />
	</message>`

	message := ParseMessage(data)
	if message.From != "jimbo@localhost" {
		t.Errorf("Message 'from' field mangled: %v", message.To)
	}
	t.Log(message.Body.Data)
}

func TestParseStreamOpening(t *testing.T) {
	data := `<stream:stream xmlns:stream="http://etherx.jabber.org/streams" version="1.0" xmlns="jabber:client" to="localhost" xml:lang="en" xmlns:xml="http://www.w3.org/XML/1998/namespace">`

	stream := ParseStreamOpening([]byte(data))
	t.Log(stream.To, stream.Version, stream.Lang)
}

func TestBuildStreamResponse(t *testing.T) {
	c := ClientConnection{}
	response := c.buildStreamOpening()
	t.Log(string(response))
}

func TestAuthGetCredentials(t *testing.T) {
	auth := Auth{Data: "AHRlc3QxAHRlc3Q="}
	user, pass := auth.GetCredentials()
	println("Got user and pass:", user, pass)
	if user != "test1" {
		t.Errorf("Username decoded incorrectly")
	}
	if pass != "test" {
		t.Errorf("Password decoded incorrectly")
	}
}

func TestIqXmlStructs(t *testing.T) {
	data := `<iq type="set" id="bind_1">
	  <bind xmlns="urn:ietf:params:xml:ns:xmpp-bind">
	    <resource deep="derp">galttop</resource>
	  </bind>
	</iq>`
	iq := new(Iq)
	xml.Unmarshal([]byte(data), iq)
	println(iq.Bind[0].Xmlns)
	println(iq.Bind[0].Resource[0].Data)
	data = `<iq id='wy2xa82b4' type='result'>
	  <bind xmlns='urn:ietf:params:xml:ns:xmpp-bind'>
	    <jid>juliet@im.example.com/balcony</jid>
	  </bind>
	</iq>`
}
