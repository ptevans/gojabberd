
package server

import (
	"errors"
)

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

func NewDomain(domain string) (d *Domain) {
	d = &Domain{domain: domain}
	//d.presence = new(PresenceTable)
	d.presence.Users = make(map[string](*ClientConnection))
	return
}

type DomainTable struct {
	domains	map[string](*Domain)
}

// Register a domain with the global domain table. If the domain is already
// registered, an error will be returned.
func (dt *DomainTable) RegisterDomain(d *Domain) (e error) {
	exists := dt.CheckDomainRegistration(d)
	if exists {
		e = errors.New("Domain already registered!")
		return
	}
	dt.domains[d.domain] = d
	for _, alias := range d.aliases {
		dt.domains[alias] = d
	}
	return
}

// Test to see if a Domain is registered in the domain table, this checks the
// primary domain name as well as any aliases.
func (dt *DomainTable) CheckDomainRegistration(d *Domain) (bool) {
	_, exists := dt.domains[d.domain]
	if exists {
		return true
	}
	for _, alias := range d.aliases {
		_, exists = dt.domains[alias]
		if exists {
			return true
		}
	}
	return false
}

// Lookup a domain and return a pointer to the Domain. Returns an error if the
// domain cannot be found.
func (dt *DomainTable) GetDomain(hostname string) (*Domain, error) {
	d, ok := dt.domains[hostname]
	if ok {
		return d, nil
	}
	return nil, errors.New("Domain could not be found!")
}
