package application

import (
	"net/http"

	"github.com/SamMebarek/orders-api/handler"
	"github.com/SamMebarek/orders-api/repository/order"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// loadRoutes configure les routes principales pour l'application.
func (a *App) loadRoutes() {
	// Initialisation d'un nouveau routeur avec chi.
	router := chi.NewRouter()

	// Utilisation d'un middleware pour logger automatiquement les requêtes.
	router.Use(middleware.Logger)

	// Définition d'une route racine simple qui répond avec un statut HTTP 200 OK.
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Configuration des routes pour la gestion des commandes.
	// 'loadOrderRoutes' est appelée pour définir les routes spécifiques aux commandes.
	router.Route("/orders", a.loadOrderRoutes)

	// Enregistrement du routeur configuré dans l'application.
	a.router = router
}

// loadOrderRoutes définit les routes spécifiques pour les opérations sur les commandes.
// Cette méthode est utilisée pour associer les chemins d'accès aux méthodes du gestionnaire de commandes.
func (a *App) loadOrderRoutes(router chi.Router) {
	// Création d'un gestionnaire pour les commandes.
	// Ce gestionnaire utilise Redis pour stocker et récupérer les données des commandes.
	orderHandler := &handler.Order{
		Repo: &order.RedisRepo{
			Client: a.rdb, // Le client Redis est fourni par l'application.
		},
	}

	// Association des routes avec les méthodes spécifiques du gestionnaire de commandes.
	router.Post("/", orderHandler.Create)           // Route pour créer une nouvelle commande.
	router.Get("/", orderHandler.List)              // Route pour lister toutes les commandes.
	router.Get("/{id}", orderHandler.GetByID)       // Route pour obtenir une commande par son ID.
	router.Put("/{id}", orderHandler.UpdateByID)    // Route pour mettre à jour une commande par ID.
	router.Delete("/{id}", orderHandler.DeleteByID) // Route pour supprimer une commande par ID.
}
