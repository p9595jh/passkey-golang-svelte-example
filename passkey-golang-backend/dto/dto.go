package dto

import "github.com/go-webauthn/webauthn/protocol"

type UserRegistrationDTO struct {
	Name      string `json:"name"`
	Email     string `json:"email"`
	BirthYear int    `json:"birthYear"`
}

type UserLoginDTO struct {
	Name string `json:"name"`
}

type BeginRegistrationDTO struct {
	Options *protocol.CredentialCreation `json:"options"`
	SID     string                       `json:"sid"`
}

type BeginLoginDTO struct {
	Options *protocol.CredentialAssertion `json:"options"`
	SID     string                        `json:"sid"`
}

type FinishRegistrationDTO struct {
	Message string `json:"message"`
}

type FinishLoginDTO struct {
	SID string `json:"sid"`
}
