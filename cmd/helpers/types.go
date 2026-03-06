package helpers

type User struct {
	ID           int64
	Name         string
	Email        string
	PasswordHash string
	TokenVersion int64
}
