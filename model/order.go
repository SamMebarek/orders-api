package model

import (
	"time"

	"github.com/google/uuid"
)

// Order représente une commande.
type Order struct {
	OrderID     uint64     `json:"order_id"`     // Identifiant unique de la commande.
	CustomerID  uuid.UUID  `json:"customer_id"`  // Identifiant unique du client.
	LineItems   []LineItem `json:"line_items"`   // Liste des articles de la commande.
	CreatedAt   *time.Time `json:"created_at"`   // Date et heure de création de la commande.
	ShippedAt   *time.Time `json:"shipped_at"`   // Date et heure d'expédition de la commande.
	CompletedAt *time.Time `json:"completed_at"` // Date et heure de finalisation de la commande.
}

// LineItem représente un article d'une commande.
type LineItem struct {
	ItemID   uuid.UUID `json:"item_id"`  // Identifiant unique de l'article.
	Quantity uint      `json:"quantity"` // Quantité commandée de l'article.
	Price    uint      `json:"price"`    // Prix unitaire de l'article.
}
