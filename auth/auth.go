/*
	Package auth implements the various SASL mechanisms required by XMPP
	See RFC 6120 for more details.
*/
package auth

import(
	"bytes"
	"encoding/base64"
	"fmt"
	"crypto/md5"
	"io"
	"time"
)

const (
	SALT = "the rain in spain falls mainly on the plain"
)

// describe the auth to perform, method is required and the rest of the fields
// will depend upon the auth method being invoked
type AuthPacket struct {
	method string
	username string
	password string
	digest string
	domain string
}

// ChallengeDigestMd5 creates a base64 challenge string to be presented to a
// client negotiating SASL authentication.
func ChallengeDigestMd5(realm string) string {
	var challenge bytes.Buffer
	encoder := base64.NewEncoder(base64.StdEncoding, &challenge)

	io.WriteString(encoder, `realm="` + realm + `",`)
	nonce, _ := genNonce(realm)
	io.WriteString(encoder, `nonce="` + nonce + `",`)
	// qop options can also include "auth-int" and "auth-conf"
	// see Robert Norris "crash course in SASL and DIGEST-MD5 for XMPP"
	io.WriteString(encoder, `qop="auth",charset=utf-8,algorithm=md5-sess`)
	encoder.Close()

	return challenge.String()
}

// ResponseDigestMd5 decodes a base64 response string received from the client
// and checks the presented values and credentials. It returns two values, a
// bool that indicates if the user passed authentication and a string that contains
// the next challenge digest that should be sent to the client.
func ResponseDigestMd5(clientResponse []byte) (bool, string) {
	var challenge bytes.Buffer
	buf := bytes.NewBuffer(clientResponse)
	decoder := base64.NewDecoder(base64.StdEncoding, buf)
	decoded := make([]byte, len(clientResponse))
	decoder.Read(decoded)


	encoder := base64.NewEncoder(base64.StdEncoding, &challenge)
	// TODO: make this a legit value a la rfc2831 section 2.1.3
	io.WriteString(encoder, "rspauth=1234")
	encoder.Close()
	return true, challenge.String()
}

// genNonce creates a nonce (number used once). The nonce is the md5 hex digest
// of the realm, SALT, and current timestamp. In theory this could be used to
// reconstruct the nonce and verify that the server produced it.
func genNonce(realm string) (nonce, timestamp string) {
	timestamp = time.Now().String()
	content := realm + SALT + timestamp
	h := md5.New()
	io.WriteString(h, content)
	nonce = fmt.Sprintf("%x", h.Sum(nil))
	return
}
