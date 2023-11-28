package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/SamMebarek/orders-api/application"
)

func main() {
	// Initialisation de l'application avec la configuration chargée depuis LoadConfig().
	app := application.New(application.LoadConfig())

	// Préparation à gérer l'interruption du programme (comme un CTRL+C) de façon gracieuse.
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel() // S'assure que les ressources du contexte sont libérées à la fin.

	// Démarrage de l'application. Si une erreur survient, elle sera affichée.
	err := app.Start(ctx)
	if err != nil {
		fmt.Println("failed to start app:", err)
	}
}
