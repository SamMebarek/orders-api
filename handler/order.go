package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/SamMebarek/orders-api/model"
	"github.com/SamMebarek/orders-api/repository/order"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Order struct {
	Repo *order.RedisRepo // Référence à un dépôt Redis pour les opérations sur les commandes.
}

// Create est une méthode HTTP pour créer une nouvelle commande.
func (h *Order) Create(w http.ResponseWriter, r *http.Request) {
	// Définition d'un struct pour décoder le corps de la requête JSON.
	var body struct {
		CustomerID uuid.UUID        `json:"customer_id"` // ID du client pour la commande.
		LineItems  []model.LineItem `json:"line_items"`  // Articles de la commande.
	}

	// Décodage du corps de la requête JSON. Si cela échoue, renvoie une erreur 400 (Bad Request).
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Obtention de la date et heure actuelle en UTC.
	now := time.Now().UTC()

	// Création d'une nouvelle commande avec les données fournies.
	order := model.Order{
		OrderID:    rand.Uint64(),   // ID de commande généré aléatoirement.
		CustomerID: body.CustomerID, // ID du client issu du corps de la requête.
		LineItems:  body.LineItems,  // Articles de la commande issus du corps de la requête.
		CreatedAt:  &now,            // Date de création fixée à l'heure actuelle.
	}

	// Insertion de la commande dans Redis.
	err := h.Repo.Insert(r.Context(), order)
	if err != nil {
		fmt.Println("failed to insert:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Sérialisation de la commande en JSON pour la réponse.
	res, err := json.Marshal(order)
	if err != nil {
		fmt.Println("failed to marshal:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Envoi de la réponse avec le statut 201 (Created) et les données de la commande.
	w.WriteHeader(http.StatusCreated)
	w.Write(res)
}

// List est une méthode HTTP pour lister les commandes.
func (h *Order) List(w http.ResponseWriter, r *http.Request) {
	// Récupération du paramètre 'cursor' de l'URL, utilisé pour la pagination.
	cursorStr := r.URL.Query().Get("cursor")
	if cursorStr == "" {
		cursorStr = "0"
	}

	// Conversion du 'cursor' en type uint64. Si échec, renvoie une erreur 400 (Bad Request).
	const decimal = 10
	const bitSize = 64
	cursor, err := strconv.ParseUint(cursorStr, decimal, bitSize)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Définition de la taille de la page pour la liste des commandes.
	const size = 50
	res, err := h.Repo.FindAll(r.Context(), order.FindAllPage{
		Offset: cursor, // Point de départ pour la pagination.
		Size:   size,   // Nombre de commandes à retourner.
	})
	if err != nil {
		fmt.Println("failed to find all:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Préparation de la réponse contenant les commandes et le prochain 'cursor'.
	var response struct {
		Items []model.Order `json:"items"`          // Liste des commandes.
		Next  uint64        `json:"next,omitempty"` // Cursor pour la pagination.
	}
	response.Items = res.Orders
	response.Next = res.Cursor

	// Sérialisation et envoi de la réponse.
	data, err := json.Marshal(response)
	if err != nil {
		fmt.Println("failed to marshal:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(data)
}

// GetByID est une méthode HTTP pour obtenir une commande par son ID.
func (h *Order) GetByID(w http.ResponseWriter, r *http.Request) {
	// Extraction de l'ID de commande de l'URL.
	idParam := chi.URLParam(r, "id")

	// Conversion de l'ID en type uint64. Si échec, renvoie une erreur 400 (Bad Request).
	const base = 10
	const bitSize = 64
	orderID, err := strconv.ParseUint(idParam, base, bitSize)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Recherche de la commande par son ID dans Redis.
	o, err := h.Repo.FindByID(r.Context(), orderID)
	if errors.Is(err, order.ErrNotExist) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Println("failed to find by id:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Envoi de la commande en réponse si trouvée.
	if err := json.NewEncoder(w).Encode(o); err != nil {
		fmt.Println("failed to marshal:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// UpdateByID met à jour le statut d'une commande spécifiée par son ID.
func (h *Order) UpdateByID(w http.ResponseWriter, r *http.Request) {
	// Structure pour décoder le corps de la requête JSON.
	var body struct {
		Status string `json:"status"` // Nouveau statut de la commande.
	}

	// Décodage du corps de la requête. Si échec, renvoie une erreur 400 (Bad Request).
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Extraction de l'ID de la commande à partir de l'URL.
	idParam := chi.URLParam(r, "id")

	// Conversion de l'ID en type uint64.
	const base = 10
	const bitSize = 64
	orderID, err := strconv.ParseUint(idParam, base, bitSize)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Recherche de la commande par son ID.
	theOrder, err := h.Repo.FindByID(r.Context(), orderID)
	if errors.Is(err, order.ErrNotExist) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Println("failed to find by id:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Mise à jour du statut de la commande en fonction du corps de la requête.
	const completedStatus = "completed"
	const shippedStatus = "shipped"
	now := time.Now().UTC()

	switch body.Status {
	case shippedStatus:
		// Si l'état demandé est "expédié", vérifie si la commande a déjà été expédiée.
		if theOrder.ShippedAt != nil {
			// Si la commande a déjà une date d'expédition, renvoie une erreur 400 (Bad Request).
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// Sinon, met à jour la date d'expédition de la commande avec l'heure actuelle.
		theOrder.ShippedAt = &now

	case completedStatus:
		// Si l'état demandé est "complété", vérifie si la commande est déjà complétée ou pas encore expédiée.
		if theOrder.CompletedAt != nil || theOrder.ShippedAt == nil {
			// Si la commande est déjà complétée ou pas encore expédiée, renvoie une erreur 400 (Bad Request).
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// Sinon, met à jour la date de complétion de la commande avec l'heure actuelle.
		theOrder.CompletedAt = &now

	default:
		// Si l'état demandé n'est ni "expédié" ni "complété", renvoie une erreur 400 (Bad Request).
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Mise à jour de la commande dans Redis.
	err = h.Repo.Update(r.Context(), theOrder)
	if err != nil {
		fmt.Println("failed to insert:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Envoi de la commande mise à jour en réponse.
	if err := json.NewEncoder(w).Encode(theOrder); err != nil {
		fmt.Println("failed to marshal:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// DeleteByID supprime une commande spécifiée par son ID.
func (h *Order) DeleteByID(w http.ResponseWriter, r *http.Request) {
	// Extraction de l'ID de la commande de l'URL.
	idParam := chi.URLParam(r, "id")

	// Conversion de l'ID en type uint64.
	const base = 10
	const bitSize = 64
	orderID, err := strconv.ParseUint(idParam, base, bitSize)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Suppression de la commande par son ID dans Redis.
	err = h.Repo.DeleteByID(r.Context(), orderID)
	if errors.Is(err, order.ErrNotExist) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Println("failed to find by id:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
