package mail

import (
	"net/smtp"

	"github.com/emersion/go-sasl"
)

// xoauth2Client implements the SASL "XOAUTH2" mechanism, as documented by Google for
// Gmail's IMAP and SMTP servers: https://developers.google.com/gmail/imap/xoauth2-protocol
// It predates, and is not quite the same wire format as, the standardized OAUTHBEARER
// (RFC 7628), so it needs its own small implementation rather than reusing go-sasl's.
type xoauth2Client struct {
	user  string
	token string
}

// NewXOAuth2Client returns a SASL client for the "XOAUTH2" mechanism used by Gmail.
func NewXOAuth2Client(user, accessToken string) sasl.Client {
	return &xoauth2Client{user: user, token: accessToken}
}

func (a *xoauth2Client) Start() (mech string, ir []byte, err error) {
	ir = []byte("user=" + a.user + "\x01auth=Bearer " + a.token + "\x01\x01")
	return "XOAUTH2", ir, nil
}

// Next is only reached when the server rejects the initial response: it replies with a
// JSON error challenge and expects an empty response to close out the exchange cleanly.
func (a *xoauth2Client) Next(challenge []byte) ([]byte, error) {
	return []byte{}, nil
}

// xoauth2SMTPAuth implements net/smtp.Auth for Gmail's "XOAUTH2" mechanism, the SMTP
// counterpart of xoauth2Client above.
type xoauth2SMTPAuth struct {
	user  string
	token string
}

func (a xoauth2SMTPAuth) Start(_ *smtp.ServerInfo) (proto string, toServer []byte, err error) {
	return "XOAUTH2", []byte("user=" + a.user + "\x01auth=Bearer " + a.token + "\x01\x01"), nil
}

func (a xoauth2SMTPAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		// The server rejected the token and sent a JSON error challenge; reply with an
		// empty response to end the exchange instead of leaving it hanging.
		return []byte{}, nil
	}
	return nil, nil
}
