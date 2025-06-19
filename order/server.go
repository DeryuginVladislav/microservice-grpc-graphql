package order

import (
	"context"
	"fmt"

	"go-microservice/account"
	"go-microservice/catalog"
	"go-microservice/order/pb"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type grpcServer struct {
	service       Service
	accountClient *account.Client
	catalogClient *catalog.Client
	pb.UnimplementedOrderServiceServer
}

func ListenGRPC(s Service, accountURL, catalogURL string, port int) error {
	accountClient, err := account.NewClient(accountURL)
	if err != nil {
		return err
	}

	catalogClient, err := catalog.NewClient(catalogURL)
	if err != nil {
		return err
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		catalogClient.Close()
		accountClient.Close()
		return err
	}
	serv := grpc.NewServer()
	pb.RegisterOrderServiceServer(serv, &grpcServer{
		service:       s,
		accountClient: accountClient,
		catalogClient: catalogClient,
	})
	reflection.Register(serv)
	return serv.Serve(lis)
}

func (s *grpcServer) GetOrdersForAccount(ctx context.Context, r *pb.GetOrdersForAccountRequest) (*pb.GetOrdersForAccountResponse, error) {
	// Валидация
	if r.AccountId == "" {
		return nil, status.Error(codes.InvalidArgument, "accountId is required")
	}

	// Получение заказов
	orders, err := s.service.GetOrdersForAccount(ctx, r.AccountId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get orders: %v", err)
	}

	// Сбор всех ID продуктов для batch-запроса
	productIDs := make([]string, 0)
	for _, o := range orders {
		for _, p := range o.Products {
			productIDs = append(productIDs, p.ID)
		}
	}

	// Получение информации о продуктах одним запросом
	productsMap := make(map[string]*catalog.Product)
	if len(productIDs) > 0 {
		products, err := s.catalogClient.GetProducts(ctx, productIDs, "", 0, 0)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to get products: %v", err)
		}
		for _, p := range products {
			productsMap[p.ID] = &p
		}
	}

	// Конвертация в protobuf
	pbOrders := make([]*pb.Order, 0, len(orders))
	for _, o := range orders {
		pbOrder := &pb.Order{
			Id:         o.ID,
			AccountId:  o.AccountID,
			TotalPrice: o.TotalPrice,
		}

		createdAtBytes, err := o.CreatedAt.MarshalBinary()
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to marshal createdAt: %v", err)
		}
		pbOrder.CreatedAt = createdAtBytes

		// Обработка продуктов
		pbProducts := make([]*pb.Order_OrderProduct, 0, len(o.Products))
		for _, p := range o.Products {
			product, exists := productsMap[p.ID]
			if !exists {
				continue // или возвращаем ошибку, в зависимости от требований
			}

			pbProducts = append(pbProducts, &pb.Order_OrderProduct{
				Id:          p.ID,
				Quantity:    p.Quantity,
				Name:        product.Name,
				Description: product.Description,
				Price:       product.Price,
			})
		}
		pbOrder.Products = pbProducts
		pbOrders = append(pbOrders, pbOrder)
	}

	return &pb.GetOrdersForAccountResponse{Orders: pbOrders}, nil
}

func (s *grpcServer) PostOrder(ctx context.Context, r *pb.PostOrderRequest) (*pb.PostOrderResponse, error) {
	// Валидация
	if r.AccountId == "" {
		return nil, status.Error(codes.InvalidArgument, "accountId is required")
	}
	if len(r.Products) == 0 {
		return nil, status.Error(codes.InvalidArgument, "at least one product is required")
	}

	// Проверка аккаунта
	if _, err := s.accountClient.GetAccount(ctx, r.AccountId); err != nil {
		return nil, status.Errorf(codes.NotFound, "account not found: %v", err)
	}

	// Сбор ID продуктов для запроса
	productIDs := make([]string, 0, len(r.Products))
	productMap := make(map[string]uint32, len(r.Products))
	for _, p := range r.Products {
		productIDs = append(productIDs, p.ProductId)
		productMap[p.ProductId] = p.Quantity
	}

	// Получение данных о продуктах
	catalogProducts, err := s.catalogClient.GetProducts(ctx, productIDs, "", 0, 0)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get products: %v", err)
	}

	// Проверка что все продукты найдены
	if len(catalogProducts) != len(productIDs) {
		return nil, status.Error(codes.InvalidArgument, "some products not found in catalog")
	}

	// Создание списка продуктов для заказа
	products := make([]OrderedProduct, 0, len(catalogProducts))
	for _, p := range catalogProducts {
		quantity, exists := productMap[p.ID]
		if !exists || quantity == 0 {
			continue // этого не должно происходить после предыдущих проверок
		}
		products = append(products, OrderedProduct{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			Quantity:    quantity,
		})
	}

	// Создание заказа
	order, err := s.service.PostOrder(ctx, r.AccountId, products)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create order: %v", err)
	}

	// Конвертация в protobuf
	createdAtBytes, err := order.CreatedAt.MarshalBinary()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to marshal createdAt: %v", err)
	}

	orderPb := &pb.Order{
		Id:         order.ID,
		AccountId:  order.AccountID,
		TotalPrice: order.TotalPrice,
		CreatedAt:  createdAtBytes,
		Products:   make([]*pb.Order_OrderProduct, 0, len(order.Products)),
	}

	for _, p := range order.Products {
		orderPb.Products = append(orderPb.Products, &pb.Order_OrderProduct{
			Id:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			Quantity:    p.Quantity,
		})
	}

	return &pb.PostOrderResponse{Order: orderPb}, nil
}
