package order

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/SamMebarek/orders-api/model"

	"github.com/redis/go-redis/v9"
)

// RedisRepo est un struct pour interagir avec Redis. Il contient un client Redis.
type RedisRepo struct {
	Client *redis.Client
}

// orderIDKey génère une clé Redis pour une commande en utilisant son ID.
func orderIDKey(id uint64) string {
	return fmt.Sprintf("order:%d", id)
}

// Insert ajoute une nouvelle commande dans Redis.
func (r *RedisRepo) Insert(ctx context.Context, order model.Order) error {
	// Convertit la commande en JSON.
	data, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to marshal order: %w", err)
	}

	// Crée une transaction Redis pour assurer que toutes les opérations soient effectuées atomiquement.
	txn := r.Client.TxPipeline()

	// Ajoute la commande avec une clé unique.
	res := txn.SetNX(ctx, orderIDKey(order.OrderID), string(data), 0)
	// Gère les erreurs de l'opération SetNX.
	if res.Err() != nil {
		txn.Discard()
		return fmt.Errorf("failed to insert set: %w", res.Err())
	}

	// Ajoute la clé de la commande à un ensemble pour faciliter les recherches.
	if err := txn.SAdd(ctx, "orders", orderIDKey(order.OrderID)).Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("failed to add order to set: %w", err)
	}

	// Exécute la transaction.
	if _, err := txn.Exec(ctx); err != nil {
		return fmt.Errorf("failed to exec: %w", err)
	}

	return nil
}

// ErrNotExist est une erreur retournée lorsqu'une commande n'est pas trouvée dans Redis.
var ErrNotExist = errors.New("order does not exist")

// FindByID trouve une commande par son ID.
func (r *RedisRepo) FindByID(ctx context.Context, id uint64) (model.Order, error) {
	// Obtient la commande de Redis en utilisant sa clé.
	value, err := r.Client.Get(ctx, orderIDKey(id)).Result()
	// Gère les cas où la commande n'existe pas ou d'autres erreurs Redis.
	if errors.Is(err, redis.Nil) {
		return model.Order{}, ErrNotExist
	} else if err != nil {
		return model.Order{}, fmt.Errorf("failed to get order: %w", err)
	}

	// Convertit la commande JSON en struct Order.
	var order model.Order
	err = json.Unmarshal([]byte(value), &order)
	if err != nil {
		return model.Order{}, fmt.Errorf("failed to unmarshal order: %w", err)
	}

	return order, nil
}

// DeleteByID supprime une commande de Redis en utilisant son ID.
func (r *RedisRepo) DeleteByID(ctx context.Context, id uint64) error {
	// Crée une transaction Redis.
	txn := r.Client.TxPipeline()

	// Supprime la commande de Redis.
	err := txn.Del(ctx, orderIDKey(id)).Err()
	if errors.Is(err, redis.Nil) {
		txn.Discard()
		return ErrNotExist
	} else if err != nil {
		txn.Discard()
		return fmt.Errorf("get order: %w", err)
	}

	// Supprime la clé de la commande de l'ensemble.
	if err := txn.SRem(ctx, "orders", orderIDKey(id)).Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("failed to remove from orders set: %w", err)
	}

	// Exécute la transaction.
	if _, err := txn.Exec(ctx); err != nil {
		return fmt.Errorf("failed to exec: %w", err)
	}

	return nil
}

// Update met à jour une commande existante dans Redis.
func (r *RedisRepo) Update(ctx context.Context, order model.Order) error {
	// Convertit la commande en JSON pour la mise à jour.
	data, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to marshal order: %w", err)
	}

	// Met à jour la commande dans Redis.
	err = r.Client.SetXX(ctx, orderIDKey(order.OrderID), string(data), 0).Err()
	// Gère les erreurs potentielles, y compris le cas où la commande n'existe pas.
	if errors.Is(err, redis.Nil) {
		return ErrNotExist
	} else if err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	return nil
}

// FindAllPage est un struct pour paginer les résultats lors de la recherche de commandes.
type FindAllPage struct {
	Size   uint64 // Nombre de commandes à retourner par page.
	Offset uint64 // Offset pour la pagination.
}

// FindResult est un struct pour retourner les résultats d'une recherche de commandes.
type FindResult struct {
	Orders []model.Order // Liste des commandes trouvées.
	Cursor uint64        // Cursor pour la pagination.
}

// FindAll trouve toutes les commandes avec une pagination.
func (r *RedisRepo) FindAll(ctx context.Context, page FindAllPage) (FindResult, error) {
	// Utilise SScan pour récupérer les clés des commandes de l'ensemble Redis.
	res := r.Client.SScan(ctx, "orders", page.Offset, "*", int64(page.Size))

	// Obtient les résultats du SScan.
	keys, cursor, err := res.Result()
	if err != nil {
		return FindResult{}, fmt.Errorf("failed to get order ids: %w", err)
	}

	// Si aucune clé n'est trouvée, retourne un résultat vide.
	if len(keys) == 0 {
		return FindResult{
			Orders: []model.Order{},
		}, nil
	}

	// Obtient les commandes de Redis en utilisant les clés trouvées.
	xs, err := r.Client.MGet(ctx, keys...).Result()
	if err != nil {
		return FindResult{}, fmt.Errorf("failed to get orders: %w", err)
	}

	// Convertit les commandes de JSON en struct Order.
	orders := make([]model.Order, len(xs))
	for i, x := range xs {
		x := x.(string)
		var order model.Order

		err := json.Unmarshal([]byte(x), &order)
		if err != nil {
			return FindResult{}, fmt.Errorf("failed to unmarshal order: %w", err)
		}

		orders[i] = order
	}

	// Retourne les commandes trouvées avec le cursor pour la pagination.
	return FindResult{
		Orders: orders,
		Cursor: cursor,
	}, nil
}
