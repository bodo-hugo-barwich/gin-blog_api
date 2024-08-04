package controllers

import (
	"fmt"
)

type (
	APIErrorResponse struct {
		Title            string
		StatusCode       uint
		Page             string
		ErrorMessage     string
		ErrorDescription string
	}

	LoginSuccess struct {
		Title      string
		StatusCode uint
		Page       string
		Message    string
		Token      string
		Expiry     string
	}

	AuthorizationSubject struct {
		ID    uint
		Login string
	}
)

// Validity of a Login Session
var SESSIONEXPIRY uint = 20

func NewAuthorizationSubject(subject map[string]interface{}) AuthorizationSubject {
	var authSubject AuthorizationSubject = AuthorizationSubject{0, ""}

	switch id := subject["ID"].(type) {
	case float64:
		if id >= 0 {
			authSubject.ID = uint(id)
		}
	case int64:
		if id > -1 {
			authSubject.ID = uint(id)
		}
	case int:
		if id > -1 {
			authSubject.ID = uint(id)
		}
	default:
		fmt.Printf("AuthorizationSubject: Subject ID type '%T'\n", subject["ID"])
	}

	if login, ok := subject["Login"].(string); ok {
		authSubject.Login = login
	}

	return authSubject
}
