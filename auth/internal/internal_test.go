package internal

import "testing"

func Test_plain(t *testing.T) {
	ut := new(userTable)
	ut.users = make(map[string]user)
	u := user{username: "testuser", password: "testpass"}
	ut.users["testuser"] = u

	authenticated := ut.plain("testuser", "testpass")
	if ! authenticated {
		t.Errorf("Expected successful authentication, failed.")
	}
	authenticated = ut.plain("testuser", "testwrongpass")
	if authenticated {
		t.Errorf("Expected failed authentication, got success.")
	}
	authenticated = ut.plain("someuser", "somepass")
	if authenticated {
		t.Errorf("Expected failed authentication, got success.")
	}
}

func Test_add(t *testing.T) {
	ut := new(userTable)
	ut.users = make(map[string]user)

	authenticated := ut.plain("testuser", "testpass")
	if authenticated {
		t.Errorf("Expected failed authentication, got success.")
	}
	ut.add("testuser", "testpass")
	authenticated = ut.plain("testuser", "testpass")
	if ! authenticated {
		t.Errorf("Expected successful authentication, failed.")
	}
}

