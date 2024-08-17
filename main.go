package main

import (
	"fmt"
	"os"

	"gin-blog/app"
)

func main() {
	var code int = 0

	err := app.Start()

	if err != nil {
		fmt.Printf("Application failed! %v", err)

		code = 1
	}

	os.Exit(code)
}
