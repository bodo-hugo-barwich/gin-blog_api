package model

import (
	"crypto/sha512"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type (
	User struct {
		gorm.Model
		Name     string `json:"name"`
		Slug     string `json:"slug"`
		Login    string `json:"login"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Articles []Article
	}

	DisplayedUser struct {
		Name     string `json:"name"`
		Slug       string `json:"slug"`
		Email      string `json:"email"`
	}

	Login struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
)

var ENCRYPTIONSALT string = "gin-blog"
var ENCRYPTIONKEY []byte = []byte("gin-blog")

func NewDisplayedUser(user *User) DisplayedUser {
	return DisplayedUser{
		user.Name,
		user.Slug,
		user.Email,
	}
}

func (user *User) Update(update *User) {
	if update.Name != "" {
		user.Name = update.Name
	}

	if update.Slug != "" {
		user.Slug = update.Slug
	}

	if update.Login != "" {
		user.Login = update.Login
	}

	if update.Email != "" {
		user.Email = update.Email
	}

	if update.Password != "" {
		user.Password = update.Password

		if !strings.HasPrefix(user.Password, "*") {
			user.Password = EncryptPassword(user.Password, ENCRYPTIONSALT)
		}
	}
}

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

	prefix := password[:splitIdx]
	suffix := password[splitIdx:]
	salt = "*" + salt + "*"

	encrypted := sha512.Sum512([]byte(prefix + salt + suffix))
	encryptedString := fmt.Sprintf("%x", encrypted)

	// Encrypted Passwords are marked with an leading '*'
	return "*" + encryptedString
}
