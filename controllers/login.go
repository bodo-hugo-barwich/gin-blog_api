package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"cxcurrency/model"
	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
)

func RegisterLoginRoutes(engine *gin.Engine) {

	engine.POST("/login", DispatchLogin)
}

func DispatchLogin(c *gin.Context) {
	var userLogin model.Login
	var user *model.User
	var err error

	req := c.Request

	contentType := req.Header.Get("content-type")

	//fmt.Printf("Controller 'Login': Content-Type: %#v\n", contentType)

	if strings.Contains(contentType, "application/json") {
		// Parse into the Login Structure
		c.BindJSON(&userLogin)
	} else {
		// Populate Login from Parameters
		userLogin.Login = c.PostForm("login")
		userLogin.Password = c.PostForm("password")
	}

	//fmt.Printf("Controller 'Login': Login: %#v\n", userLogin)

	if userLogin.Login == "" {
		c.JSON(http.StatusUnprocessableEntity,
			APIErrorResponse{
				"Blog - Error",
				http.StatusUnprocessableEntity,
				"login",
				"Unprocessable Content",
				"User Login: Login Data is incomplete!",
			})

		return
	}

	if user, err = GetUserByLogin(userLogin.Login); user == nil || err != nil {
		c.JSON(http.StatusUnauthorized,
			APIErrorResponse{
				"Blog - Error",
				http.StatusUnprocessableEntity,
				"login",
				"Unauthorized",
				"User Login: Login failed!",
			})

		return
	}

	//fmt.Printf("Controller 'Login': User: %#v; Login: %#v; \n", user, userLogin)

	if user.AuthLogin(&userLogin, model.ENCRYPTIONSALT) {
		// Session Validity
		sessionStart := time.Now()
		sessionMinutes, _ := time.ParseDuration(fmt.Sprintf("%dm", SESSIONEXPIRY))
		sessionExpiry := time.Now().Add(sessionMinutes)

		fmt.Printf("Controller 'Login': Session: iat: '%#v'; exp: '%#v'\n", time.Now().Unix(), time.Now().Add(sessionMinutes).Unix())

		// Create a new JWT
		token := jwt.NewWithClaims(jwt.SigningMethodHS512,
			jwt.MapClaims{
				"iss": "Blog",
				"sub": AuthorizationSubject{user.ID, user.Login},
				"iat": sessionStart.Unix(),
				"exp": sessionExpiry.Unix(),
			})
		tokenString, err := token.SignedString(model.ENCRYPTIONKEY)

		if err == nil {
			c.JSON(http.StatusOK,
				LoginSuccess{
					"Blog - Success",
					http.StatusOK,
					"login",
					"OK",
					tokenString,
					sessionExpiry.Format(time.RFC3339),
				})
		}
	} else {
		c.JSON(http.StatusUnauthorized,
			APIErrorResponse{
				"Blog - Error",
				http.StatusUnprocessableEntity,
				"login",
				"Unauthorized",
				"User Login: Login failed!",
			})
	}
}

func ValidateAuthorizationHeader(c *gin.Context) (*model.User, error) {
	var tokenString string

	authorizationHeader := c.Request.Header["Authorization"]

	if len(authorizationHeader) == 0 {
		return nil, errors.New("Authorization Token: Token is invalid! Message: No Token!")
	}

	bearerString := authorizationHeader[len(authorizationHeader)-1]

	if strings.HasPrefix(bearerString, "Bearer ") {
		bearerFields := strings.Fields(bearerString)

		tokenString = bearerFields[1]
	}

	if tokenString == "" {
		return nil, errors.New("Authorization Token: Token is invalid! Message: No Token!")
	}

	return ValidateToken(tokenString)
}

func ValidateToken(tokenString string) (*model.User, error) {
	var user *model.User
	var err error

	token, err := jwt.Parse(tokenString, GetEncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("Authorization Token: Token is invalid! Message: %v", err)
	}

	fmt.Printf("Controller 'Login': Parsed Token: %#v\n", token)

	if tokenData, ok := token.Claims.(jwt.MapClaims); ok {
		fmt.Printf("Controller 'Login': Token Data: %#v\n", tokenData)

		if subject, ok := tokenData["sub"]; ok {
			fmt.Printf("Controller 'Login': sub: %#v\n", subject)

			authSubject := NewAuthorizationSubjectFromInterface(subject.(map[string]interface{}))

			fmt.Printf("Controller 'Login': auth sub: %#v\n", authSubject)

			if user, err = GetUserByID(authSubject.ID); user == nil || err != nil {
				if err == nil {
					err = errors.New("Authorization Token: User unauthorized!")
				}

				return nil, err
			}

			// User Data Integrity Check
			if user.Login != authSubject.Login {
				return nil, errors.New("Authorization Token: User unauthorized!")
			}
		} else {
			return nil, errors.New("Authorization Token: Payload invalid!")
		}
	} else {
		return nil, errors.New("Authorization Token: Payload invalid!")
	}

	return user, nil
}

func GetEncryptionKey(token *jwt.Token) (interface{}, error) {
	// Don't forget to validate the alg is what you expect:
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("Authorization Token: Token is invalid! Message: Unexpected signing method: %v", token.Header["alg"])
	}

	// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
	return model.ENCRYPTIONKEY, nil
}
