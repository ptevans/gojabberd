
package server


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

