package catalog

import (
	"context"
	"fmt"
	"go-microservice/catalog/pb"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type grpcServer struct {
	service Service
	pb.UnimplementedCatalogServiceServer
}

func ListenGRPC(s Service, port int) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	serv := grpc.NewServer()
	pb.RegisterCatalogServiceServer(serv, &grpcServer{service: s})
	reflection.Register(serv)
	return serv.Serve(lis)
}

func toProtoProduct(p *Product) *pb.Product {
	return &pb.Product{
		Id:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
	}
}

func makeProductsResponse(products []Product) *pb.ProductsResponse {
	pbProducts := make([]*pb.Product, len(products))
	for i, p := range products {
		pbProducts[i] = toProtoProduct(&p)
	}
	return &pb.ProductsResponse{Products: pbProducts}
}

func (s *grpcServer) GetProduct(ctx context.Context, r *pb.GetProductRequest) (*pb.ProductResponse, error) {
	product, err := s.service.GetProduct(ctx, r.Id)
	if err != nil {
		return nil, err
	}
	return &pb.ProductResponse{Product: toProtoProduct(product)}, nil
}

func (s *grpcServer) GetProducts(ctx context.Context, r *pb.GetProductsRequest) (*pb.ProductsResponse, error) {
	var products []Product
	var err error
	if r.Query != "" {
		products, err = s.service.SearchProducts(ctx, r.Query, r.Skip, r.Take)
	} else if len(r.Ids) != 0 {
		products, err = s.service.GetProductsByIDs(ctx, r.Ids)
	} else {
		products, err = s.service.GetProducts(ctx, r.Skip, r.Take)
	}

	if err != nil {
		return nil, err
	}

	return makeProductsResponse(products), nil
}

func (s *grpcServer) PostProduct(ctx context.Context, r *pb.PostProductRequest) (*pb.ProductResponse, error) {
	product, err := s.service.PostProduct(ctx, r.Name, r.Description, r.Price)
	if err != nil {
		return nil, err
	}
	return &pb.ProductResponse{Product: toProtoProduct(product)}, nil
}
