module cxcurrency

go 1.15

replace cxcurrency => ./

replace cxcurrency/app => ./app

replace cxcurrency/controllers => ./controllers

replace cxcurrency/model => ./model

require (
	github.com/client9/misspell v0.3.4 // indirect
	github.com/gin-gonic/gin v1.10.0
	github.com/golang-jwt/jwt/v5 v5.2.1
	github.com/sergi/go-diff v1.1.0 // indirect
	golang.org/x/exp v0.0.0-20220722155223-a9213eeb770e // indirect
	golang.org/x/tools/gopls v0.15.3 // indirect
	gopkg.in/errgo.v2 v2.1.0 // indirect
	gorm.io/driver/postgres v1.5.7
	gorm.io/gorm v1.25.10
)
