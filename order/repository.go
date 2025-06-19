package order

import (
	"context"
	"database/sql"
	"time"

	"github.com/lib/pq"
)

type Repository interface {
	Close()
	PutOrder(ctx context.Context, o Order) error
	GetOrdersForAccount(ctx context.Context, accountId string) ([]Order, error)
}

type postgresRepository struct {
	db *sql.DB
}

func NewPostgresReposytory(url string) (Repository, error) {
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return &postgresRepository{db}, nil
}

// Close implements Repository.
func (r *postgresRepository) Close() {
	r.db.Close()
}

// PutOrder implements Repository.
func (r *postgresRepository) PutOrder(ctx context.Context, o Order) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		tx.Commit()
	}()

	_, err = tx.ExecContext(ctx, "INSERT INTO orders(id,created_at,account_id,total_price) VALUES($1,$2,$3,$4)",
		o.ID, o.CreatedAt, o.AccountID, o.TotalPrice)
	if err != nil {
		return err
	}

	stmt, _ := tx.PrepareContext(ctx, pq.CopyIn("order_products", "order_id", "product_id", "quantity"))
	for _, p := range o.Products {
		_, err = stmt.ExecContext(ctx, o.ID, p.ID, p.Quantity)
		if err != nil {
			return err
		}
	}
	_, err = stmt.ExecContext(ctx)
	if err != nil {
		return err
	}

	stmt.Close()
	return nil
}

// GetOrdersForAccount implements Repository.
func (r *postgresRepository) GetOrdersForAccount(ctx context.Context, accountId string) ([]Order, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT
         o.id,
         o.account_id,
		 o.created_at,
         o.total_price::money::numeric::float8,
         op.product_id,
         op.quantity
         FROM orders o
         JOIN order_products op
         ON o.id=op.order_id
         WHERE o.account_id=$1
         ORDER BY o.id`,
		accountId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := []Order{}
	var currentOrder *Order
	var lastOrderID string

	for rows.Next() {
		var orderID, accountID string
		var totalPrice float64
		var productID string
		var quantity uint32
		var createdAt time.Time

		if err := rows.Scan(
			&orderID,
			&accountID,
			&createdAt,
			&totalPrice,
			&productID,
			&quantity,
		); err != nil {
			return nil, err
		}

		// Если это новый заказ (отличается от предыдущего)
		if lastOrderID != orderID {
			// Если у нас уже есть текущий заказ, добавляем его в результат
			if currentOrder != nil {
				orders = append(orders, *currentOrder)
			}

			// Создаём новый заказ
			currentOrder = &Order{
				ID:         orderID,
				AccountID:  accountID,
				CreatedAt:  createdAt,
				TotalPrice: totalPrice,
				Products:   []OrderedProduct{},
			}
			lastOrderID = orderID
		}

		// Добавляем продукт в текущий заказ
		currentOrder.Products = append(currentOrder.Products, OrderedProduct{
			ID:       productID,
			Quantity: quantity,
		})
	}

	// Добавляем последний заказ, если он есть
	if currentOrder != nil {
		orders = append(orders, *currentOrder)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return orders, nil
}
