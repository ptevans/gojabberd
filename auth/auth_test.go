
package auth

import "testing"

func Test_plain(t *testing.T) {
}

func Test_genNonce(t *testing.T) {
	nonce, timestamp := genNonce("localhost")
	t.Logf(nonce + " " + timestamp)
}

func TestChallengeDigestMd5(t *testing.T) {
	challenge := ChallengeDigestMd5("localhost")
	t.Logf(string(challenge))
}

func TestResponseDigestMd5(t *testing.T) {
	sample := []byte("dXNlcm5hbWU9InRlc3QxIixyZWFsbT0ibG9jYWxob3N0Iixub25jZT0iNjAwOTZjZjlmN2ExMGEwMGIyMGNlMzE0MjhiMmNlYTIiLGNub25jZT0iUjhMMFZPTFJScmFwL01YSk5GLzJqWEpUSHhKRzJQenR2aXNTMThRVGxRdz0iLG5jPTAwMDAwMDAxLGRpZ2VzdC11cmk9InhtcHAvbG9jYWxob3N0Iixxb3A9YXV0aCxyZXNwb25zZT1jZGM3YzE5MmI3MzY5NGMzNGYyZmRlZDY4MzZhYWIwMCxjaGFyc2V0PXV0Zi04")
	success, _ := ResponseDigestMd5(sample)
	if success {
		t.Logf("true")
	}
}
