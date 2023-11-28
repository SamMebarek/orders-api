package application

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

// App représente l'application avec le routeur, le client Redis, et la configuration.
type App struct {
	router http.Handler  // Gestionnaire HTTP pour router les requêtes.
	rdb    *redis.Client // Client pour interagir avec la base de données Redis.
	config Config        // Configuration de l'application.
}

// New crée et initialise une nouvelle instance de l'application.
func New(config Config) *App {
	// Initialisation de l'application avec un client Redis et la configuration.
	app := &App{
		rdb: redis.NewClient(&redis.Options{
			Addr: config.RedisAddress, // Adresse du serveur Redis depuis la configuration.
		}),
		config: config,
	}

	// Chargement des routes pour le serveur HTTP.
	app.loadRoutes()

	// Retourne l'instance de l'application initialisée.
	return app
}

// Start lance le serveur HTTP de l'application et gère les connexions entrantes.
func (a *App) Start(ctx context.Context) error {
	// Configuration du serveur HTTP avec l'adresse et le gestionnaire de route.
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", a.config.ServerPort),
		Handler: a.router,
	}

	// Vérification de la connexion à Redis.
	err := a.rdb.Ping(ctx).Err()
	if err != nil {
		return fmt.Errorf("failed to connect to redis: %w", err)
	}

	// Fermeture de la connexion Redis lors de l'arrêt de l'application.
	defer func() {
		if err := a.rdb.Close(); err != nil {
			fmt.Println("failed to close redis", err)
		}
	}()

	fmt.Println("Starting server")

	// Canal pour gérer les erreurs potentielles du serveur.
	ch := make(chan error, 1)

	// Démarrage du serveur dans une goroutine.
	go func() {
		err = server.ListenAndServe()
		if err != nil {
			ch <- fmt.Errorf("failed to start server: %w", err)
		}
		close(ch)
	}()

	// Attente d'une erreur du serveur ou d'une interruption du contexte.
	select {
	case err = <-ch:
		return err
	case <-ctx.Done():
		// Création d'un contexte avec un délai pour la fermeture gracieuse du serveur.
		timeout, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		return server.Shutdown(timeout)
	}

	return nil
}
