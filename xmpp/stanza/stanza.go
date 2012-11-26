
package stanza

import  (
	"bytes"
	"encoding/base64"
	"encoding/xml"
	"gojabberd/auth"
)

type Stream struct {
	XMLName xml.Name `xml:"stream"`
	From    string   `xml:"from,attr"`
	Id      string   `xml:"id,attr"`
	To      string   `xml:"to,attr"`
	Version string   `xml:"version,attr"`
	Lang    string   `xml:"lang,attr,omitempty"`
}

type StreamResponse struct {
	XMLName     xml.Name `xml:"stream:stream"`
	Xmlns       string   `xml:"xmlns,attr"`
	XmlnsStream string   `xml:"xmlns:stream,attr"`
	From        string   `xml:"from,attr"`
	To          string   `xml:"to,attr,omitempty"`
	Id          string   `xml:"id,attr"`
	Version     string   `xml:"version,attr"`
	Lang        string   `xml:"lang,attr,omitempty"`
	Features    *StreamFeatures
}

type StreamFeatures struct {
	XMLName    xml.Name `xml:"stream:features"`
	Bind       Bind
	Session    Session
	Mechanisms Mechanisms
}

type Bind struct {
	XMLName  xml.Name   `xml:"bind"`
	Xmlns    string     `xml:"xmlns,attr"`
	Resource []Resource
	Jid      []Jid
}

type Session struct {
	XMLName xml.Name `xml:"session"`
	Xmlns   string   `xml:"xmlns,attr,omitempty"`
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

type Challenge struct {
	XMLName	xml.Name	`xml:"challenge"`
	Xmlns	string	`xml:"xmlns,attr"`
	Data	string	`xml:",chardata"`
}

func NewChallenge(realm string) (c *Challenge) {
	c = new(Challenge)
	c.Xmlns = "urn:ietf:params:xml:ns:xmpp-sasl"
	c.Data = auth.ChallengeDigestMd5(realm)
	return
}

type Response struct {
	XMLName	xml.Name	`xml:"response"`
	Xmlns	string	`xml:"xmlns,attr"`
	Data	[]byte	`xml:",chardata"`
}

type Success struct {
	XMLName xml.Name `xml:"success"`
	Xmlns   string   `xml:"xmlns,attr"`
}

type Message struct {
	XMLName xml.Name `xml:"message"`
	From    string   `xml:"from,attr,omitempty"`
	Id      string   `xml:"id,attr,omitempty"`
	To      string   `xml:"to,attr,omitempty"`
	Type    string   `xml:"type,attr,omitempty"` // chat, error, groupchat, headline, normal
	Subject *Subject
	Body    *Body
	Thread  *Thread
}

type Body struct {
	XMLName xml.Name `xml:"body"`
	Data    string   `xml:",chardata"`
	Lang    string   `xml:"lang,attr,omitempty""`
}

type Subject struct {
	XMLName xml.Name `xml:"subject"`
	Data    string   `xml:",chardata"`
	Lang    string   `xml:"lang,attr,omitempty"`
}

type Thread struct {
	XMLName xml.Name `xml:"thread"`
	Data    string   `xml:",chardata"`
	Parent  string   `xml:"parent,attr,omitempty"`
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
