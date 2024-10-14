package entity

import (
	"github.com/go-webauthn/webauthn/webauthn"
)

var _ webauthn.User = (*User)(nil)

type User struct {
	ID          []byte                `json:"id"`
	Credentials []webauthn.Credential `json:"credentials"`
	Name        string                `json:"name"`
	Email       string                `json:"email"`
	BirthYear   int                   `json:"birthYear"`
}

func (u *User) WebAuthnID() []byte {
	return u.ID
}

func (u *User) WebAuthnName() string {
	return u.Name
}

func (u *User) WebAuthnDisplayName() string {
	return u.Name
}

func (u *User) WebAuthnCredentials() []webauthn.Credential {
	return u.Credentials
}

func (u *User) AddCredential(cred *webauthn.Credential) {
	u.Credentials = append(u.Credentials, *cred)
}
