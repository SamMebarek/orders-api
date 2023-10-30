package main

import (
	"github.com/SamMebarek/orders-api/application"
)

func main() {
	app := application.New()
	err := app.Start()

}
