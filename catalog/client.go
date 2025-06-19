package catalog

import (
	"context"
	"go-microservice/catalog/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn   *grpc.ClientConn
	client pb.CatalogServiceClient
}

func NewClient(url string) (*Client, error) {
	conn, err := grpc.NewClient(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	client := pb.NewCatalogServiceClient(conn)
	return &Client{conn: conn, client: client}, nil
}

func (c *Client) Close() {
	c.conn.Close()
}

func (c *Client) GetProduct(ctx context.Context, id string) (*Product, error) {
	p, err := c.client.GetProduct(ctx, &pb.GetProductRequest{Id: id})
	if err != nil {
		return nil, err
	}
	return &Product{
		ID:          p.Product.Id,
		Name:        p.Product.Name,
		Description: p.Product.Description,
		Price:       p.Product.Price,
	}, nil
}

func (c *Client) PostProduct(ctx context.Context, price float64, name, description string) (*Product, error) {
	p, err := c.client.PostProduct(ctx, &pb.PostProductRequest{
		Name:        name,
		Description: description,
		Price:       price,
	})

	if err != nil {
		return nil, err
	}

	return &Product{
		ID:          p.Product.Id,
		Name:        p.Product.Name,
		Description: p.Product.Description,
		Price:       p.Product.Price,
	}, nil
}
func (c *Client) GetProducts(ctx context.Context, ids []string, query string, skip, take uint64) ([]Product, error) {
	p, err := c.client.GetProducts(ctx, &pb.GetProductsRequest{
		Ids:   ids,
		Query: query,
		Skip:  skip,
		Take:  take,
	})
	if err != nil {
		return nil, err
	}
	products := make([]Product, len(p.Products))
	for i, v := range p.Products {
		products[i] = Product{
			ID:          v.Id,
			Name:        v.Name,
			Description: v.Description,
			Price:       v.Price,
		}
	}
	return products, nil
}
