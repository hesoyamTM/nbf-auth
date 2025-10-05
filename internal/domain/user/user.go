package user

import "github.com/google/uuid"

type User struct {
	ID      uuid.UUID
	AuthID  string
	Name    string
	Surname string
}

// NewUser return User struct. Id, name and surname are required
func NewUser(id uuid.UUID, authID, name, surname string) (*User, error) {
	// if name == "" {
	// 	return nil, ErrEmptyName
	// }
	// if surname == "" {
	// 	return nil, ErrEmptySurname
	// }

	return &User{
		id,
		authID,
		name,
		surname,
	}, nil
}
