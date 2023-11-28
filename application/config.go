package application

import (
	"os"
	"strconv"
)

// Config contient la configuration nécessaire pour l'application.
type Config struct {
	RedisAddress string // Adresse du serveur Redis.
	ServerPort   uint16 // Port pour le serveur HTTP.
}

// LoadConfig charge la configuration de l'application.
// Elle lit les variables d'environnement et définit les valeurs par défaut si nécessaire.
func LoadConfig() Config {
	// Configuration par défaut.
	cfg := Config{
		RedisAddress: "localhost:6379", // Valeur par défaut pour Redis.
		ServerPort:   3000,             // Valeur par défaut pour le port du serveur.
	}

	// Recherche et utilisation de la variable d'environnement pour l'adresse Redis, si elle existe.
	if redisAddr, exists := os.LookupEnv("REDIS_ADDR"); exists {
		cfg.RedisAddress = redisAddr
	}

	// Recherche et utilisation de la variable d'environnement pour le port du serveur, si elle existe.
	if serverPort, exists := os.LookupEnv("SERVER_PORT"); exists {
		// Conversion de la valeur de la variable d'environnement en un nombre.
		if port, err := strconv.ParseUint(serverPort, 10, 16); err == nil {
			cfg.ServerPort = uint16(port)
		}
	}

	// Retourne la configuration chargée.
	return cfg
}
