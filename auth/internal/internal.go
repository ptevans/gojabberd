/*
	Package auth/internal implements a user credential table and an
	interface for authenticating users using the various SASL methods
	specified in RFC 6120.

	The purpose of this package is implement a standard interface template
	which can be use for testing and to serve as a template for implementing
	real authentication packages backed by real databases/services.
*/
package internal

type user struct {
	username string
	password string
}

type userTable struct {
	users map[string](user)
}

func (ut *userTable) plain(username, password string) bool {
	u, exists := ut.users[username]
	switch {
	case ! exists:
		return false
	case u.password == password:
		return true
	}
	return false
}

func (ut *userTable) add(username, password string) {
	u := user{username, password}
	ut.users[username] = u
}

