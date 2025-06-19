package catalog

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

type Repository interface {
	Close() error
	PutProduct(ctx context.Context, p Product) error
	GetProductByID(ctx context.Context, id string) (*Product, error)
	ListProducts(ctx context.Context, skip, take uint64) ([]Product, error)
	ListProductsWithIDs(ctx context.Context, ids []string) ([]Product, error)
	SearchProducts(ctx context.Context, query string, skip, take uint64) ([]Product, error)
}

type ElasticRepository struct {
	client *elasticsearch.Client
}

type searchResponse struct {
	Hits struct {
		Hits []struct {
			Source Product `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

// type ProductDocument struct {
// 	Name        string  `json:"name"`
// 	Description string  `json:"description"`
// 	Price       float64 `json:"price"`
// }

func NewElasticReposytory(url string) (Repository, error) {
	c, err := elasticsearch.NewClient(elasticsearch.Config{Addresses: []string{url}})
	if err != nil {
		return nil, err
	}

	res, err := c.Ping()
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	return &ElasticRepository{
		client: c,
	}, nil
}

func (r *ElasticRepository) Close() error {
	if r.client == nil {
		return fmt.Errorf("client is nil")
	}
	return nil
}

func (r *ElasticRepository) PutProduct(ctx context.Context, p Product) error {
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}

	req := esapi.IndexRequest{
		Index:      "catalog",
		DocumentID: p.ID,
		Body:       bytes.NewReader(data),
		Refresh:    "true",
	}

	res, err := req.Do(ctx, r.client)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return errors.New("elasticsearch error")
	}
	return nil
}

func (r *ElasticRepository) GetProductByID(ctx context.Context, id string) (*Product, error) {
	req := esapi.GetRequest{
		Index:      "catalog",
		DocumentID: id,
	}

	res, err := req.Do(ctx, r.client)
	if err != nil {
		return nil, err
	}
	if res.StatusCode == 404 {
		return nil, fmt.Errorf("product not found")
	}
	if res.IsError() {
		return nil, fmt.Errorf("error getting product: %s", res.String())
	}

	var doc struct {
		Source Product `json:"_source"`
	}

	if err := json.NewDecoder(res.Body).Decode(&doc); err != nil {
		return nil, err
	}
	return &doc.Source, nil
}

// executeSearch выполняет поисковый запрос и парсит ответ
func (r *ElasticRepository) executeSearch(ctx context.Context, index string, body io.Reader) ([]Product, error) {
	res, err := r.client.Search(
		r.client.Search.WithContext(ctx),
		r.client.Search.WithIndex(index),
		r.client.Search.WithBody(body),
	)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		var e struct {
			Error struct {
				Reason string `json:"reason"`
			} `json:"error"`
		}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			return nil, fmt.Errorf("error parsing error response: %w", err)
		}
		return nil, fmt.Errorf("elasticsearch error: %s", e.Error.Reason)
	}

	var result searchResponse
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	products := make([]Product, 0, len(result.Hits.Hits))
	for _, hit := range result.Hits.Hits {
		products = append(products, hit.Source)
	}

	return products, nil
}

// ListProducts возвращает список товаров с пагинацией
func (r *ElasticRepository) ListProducts(ctx context.Context, skip, take uint64) ([]Product, error) {
	query := map[string]interface{}{
		"from": skip,
		"size": take,
		"query": map[string]interface{}{
			"match_all": struct{}{},
		},
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, fmt.Errorf("failed to encode query: %w", err)
	}

	return r.executeSearch(ctx, "catalog", &buf)
}

// ListProductsWithIDs возвращает товары по их IDs
func (r *ElasticRepository) ListProductsWithIDs(ctx context.Context, ids []string) ([]Product, error) {
	query := map[string]interface{}{
		"size": len(ids),
		"query": map[string]interface{}{
			"ids": map[string]interface{}{
				"values": ids,
			},
		},
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, fmt.Errorf("failed to encode query: %w", err)
	}

	return r.executeSearch(ctx, "catalog", &buf)
}

func (r *ElasticRepository) SearchProducts(ctx context.Context, query string, skip, take uint64) ([]Product, error) {
	// 1. Формируем multi-match запрос
	searchQuery := map[string]interface{}{
		"from": skip,
		"size": take,
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":     query,
				"fields":    []string{"name", "description"},
				"type":      "best_fields", // Аналогично NewMultiMatchQuery в olivere
				"fuzziness": "AUTO",        // Опционально: нечёткий поиск
			},
		},
	}

	// 2. Сериализуем запрос
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(searchQuery); err != nil {
		return nil, fmt.Errorf("failed to encode query: %w", err)
	}

	return r.executeSearch(ctx, "catalog", &buf)
}
