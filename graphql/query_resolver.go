package main

import (
	"context"
	"log"
	"time"
)

type queryResolver struct {
	server *Server
}

func (q queryResolver) Accounts(ctx context.Context, pagination *PaginationInput, id *string) ([]*Account, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if id != nil {
		a, err := q.server.accountClient.GetAccount(ctx, *id)
		if err != nil {
			return nil, err
		}
		return []*Account{{
			ID:   a.ID,
			Name: a.Name,
		}}, nil
	}
	var skip, take uint64
	if pagination != nil {
		skip = uint64(*pagination.Skip)
		take = uint64(*pagination.Take)
	}

	accounts, err := q.server.accountClient.GetAccounts(ctx, skip, take)
	if err != nil {
		return nil, err
	}

	accs := make([]*Account, 0, len(accounts))
	for _, a := range accounts {
		accs = append(accs, &Account{ID: a.ID, Name: a.Name})
	}
	return accs, nil

}
func (q queryResolver) Products(ctx context.Context, pagination *PaginationInput, query *string, id *string) ([]*Product, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if id != nil {
		p, err := q.server.catalogClient.GetProduct(ctx, *id)
		if err != nil {
			log.Printf("Product query error: %v", err)
			return nil, err
		}
		return []*Product{{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
		}}, nil
	}

	var skip, take uint64
	if pagination != nil && pagination.Skip != nil && pagination.Take != nil {
		skip = uint64(*pagination.Skip)
		take = uint64(*pagination.Take)
	}

	var searchQuery string
	if query != nil {
		searchQuery = *query
	}

	products, err := q.server.catalogClient.GetProducts(ctx, []string{}, searchQuery, skip, take)
	if err != nil {
		log.Printf("Product query error: %v", err)
		return nil, err
	}

	pr := make([]*Product, 0, len(products))
	for _, p := range products {
		pr = append(pr, &Product{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
		})
	}
	return pr, nil

}
