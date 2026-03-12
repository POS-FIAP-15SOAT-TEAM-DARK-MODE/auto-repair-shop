package domain

type (
	User struct {
		ID       string
		Name     string
		Email    string
		Password string
	}

	UserRepository interface {
		Create(user *User) error
	}
)
