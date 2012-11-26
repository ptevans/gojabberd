package xmpp

import (
	"encoding/xml"
	"strings"
	"gojabberd/xmpp/stanza"
)

func ParseMessage(data string) (message stanza.Message) {
	xml.Unmarshal([]byte(data), &message)
	return
}

func ParseThread(data string) (thread stanza.Thread) {
	xml.Unmarshal([]byte(data), &thread)
	return
}

func ParseStreamOpening(data []byte) (stream stanza.Stream) {
	xml.Unmarshal(data, &stream)
	return
}

func HandleStanza(stanza []byte) (response []byte) {
	return
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
