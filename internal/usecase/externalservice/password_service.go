package externalservice

type PasswordService interface {
	HashPassword(password string) (string, error)
	ComparePasswords(hashedPassword string, plainPassword []byte) bool
}
