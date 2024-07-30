package model

import (
	//"fmt"
	"crypto/sha512"

	"gorm.io/gorm"
)

type (
User struct {
	gorm.Model
	Name  string	`json:"name"`
	Login string	`json:"login"`
	Email string	`json:"email"`
	Password string `json:"password"`
	Articles []Article
}

Login struct {
	Login string `json:"login"`
	Password string `json:"password"`
}
)

func (user *User) Auth(login string, password string, salt string) bool {
	if user.Login != login {
		return false
	}

	encrypted := EncryptPassword(password, salt)

	if user.Password != encrypted {
		return false
	}

	return true
}

func (user *User) AuthLogin(logindata *Login, salt string) bool {
	return user.Auth(logindata.Login, logindata.Password, salt)
}

func EncryptPassword(password string, salt string) string {
	length := len(password)

	if length < 2 {
		password = "!" + password + "!"
		length = len(password)
	}

	splitIdx := int(length / 2)

	prefix := password[0:splitIdx]
	suffix := password[splitIdx:length - 1]
	salt = "*" + salt + "*"

	encrypted := sha512.Sum512([]byte(prefix + salt + suffix))
	encryptedString := string(encrypted[:])

	// Encrypted Passwords are marked with an leading '*'
	return "*" + encryptedString
}