package repository

import (
	"passkey-golang-backend/entity"
)

type Repository struct {
	users map[string]*entity.User
}

func New() *Repository {
	return &Repository{
		users: make(map[string]*entity.User),
	}
}

func (r *Repository) Find(name string) *entity.User {
	return r.users[name]
}

func (r *Repository) Save(user *entity.User) {
	// id := base64.URLEncoding.EncodeToString(user.WebAuthnID())
	// r.users[id] = user

	r.users[user.Name] = user
}
