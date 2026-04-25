package domain

type Repository interface {
	FindByID(id UserID) (*User, error)
	FindByUUID(uuid string) (*User, error)
	FindByEmail(email string) (*User, error)
	Create(user *User) error
	Update(user *User) error
	Delete(id UserID) error
}
