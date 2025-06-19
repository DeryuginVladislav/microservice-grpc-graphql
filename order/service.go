package order

import (
	"context"
	"time"

	"github.com/segmentio/ksuid"
)

type Service interface {
	PostOrder(ctx context.Context, accountID string, products []OrderedProduct) (*Order, error)
	GetOrdersForAccount(ctx context.Context, accountID string) ([]Order, error)
}

type Order struct {
	ID         string           `json:"id"`
	AccountID  string           `json:"account_id"`
	CreatedAt  time.Time        `json:"created_at"`
	TotalPrice float64          `json:"total_price"`
	Products   []OrderedProduct `json:"products"`
}

type OrderedProduct struct {
	ID          string
	Name        string
	Description string
	Price       float64
	Quantity    uint32
}

type orderService struct {
	repository Repository
}

func NewService(r Repository) Service {
	return &orderService{
		repository: r,
	}
}

func (s orderService) PostOrder(ctx context.Context, accountID string, products []OrderedProduct) (*Order, error) {
	totalPrice := 0.0
	for _, v := range products {
		totalPrice += v.Price * float64(v.Quantity)
	}

	order := Order{
		ID:         ksuid.New().String(),
		AccountID:  accountID,
		CreatedAt:  time.Now(),
		Products:   products,
		TotalPrice: totalPrice,
	}

	err := s.repository.PutOrder(ctx, order)
	if err != nil {
		return nil, err
	}

	return &order, nil
}
func (s orderService) GetOrdersForAccount(ctx context.Context, accountID string) ([]Order, error) {
	return s.repository.GetOrdersForAccount(ctx, accountID)
}
