package main

import (
	"fmt"

	"github.com/SamMebarek/orders-api/application"
)

func main() {
	app := application.New()
	err := app.Start(context.todo)
	if err != nil {
		fmt.Println("failed to start app:", err)

	}
}
