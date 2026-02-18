package productv1

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ProductServiceServer interface {
	CreateProduct(context.Context, *CreateProductRequest) (*CreateProductReply, error)
	UpdateProduct(context.Context, *UpdateProductRequest) (*UpdateProductReply, error)
	ActivateProduct(context.Context, *ActivateProductRequest) (*ActivateProductReply, error)
	DeactivateProduct(context.Context, *DeactivateProductRequest) (*DeactivateProductReply, error)
	ApplyDiscount(context.Context, *ApplyDiscountRequest) (*ApplyDiscountReply, error)
	RemoveDiscount(context.Context, *RemoveDiscountRequest) (*RemoveDiscountReply, error)
	GetProduct(context.Context, *GetProductRequest) (*GetProductReply, error)
	ListProducts(context.Context, *ListProductsRequest) (*ListProductsReply, error)
}

type UnimplementedProductServiceServer struct{}

func (UnimplementedProductServiceServer) CreateProduct(context.Context, *CreateProductRequest) (*CreateProductReply, error) {
	return nil, status.Error(codes.Unimplemented, "method CreateProduct not implemented")
}
func (UnimplementedProductServiceServer) UpdateProduct(context.Context, *UpdateProductRequest) (*UpdateProductReply, error) {
	return nil, status.Error(codes.Unimplemented, "method UpdateProduct not implemented")
}
func (UnimplementedProductServiceServer) ActivateProduct(context.Context, *ActivateProductRequest) (*ActivateProductReply, error) {
	return nil, status.Error(codes.Unimplemented, "method ActivateProduct not implemented")
}
func (UnimplementedProductServiceServer) DeactivateProduct(context.Context, *DeactivateProductRequest) (*DeactivateProductReply, error) {
	return nil, status.Error(codes.Unimplemented, "method DeactivateProduct not implemented")
}
func (UnimplementedProductServiceServer) ApplyDiscount(context.Context, *ApplyDiscountRequest) (*ApplyDiscountReply, error) {
	return nil, status.Error(codes.Unimplemented, "method ApplyDiscount not implemented")
}
func (UnimplementedProductServiceServer) RemoveDiscount(context.Context, *RemoveDiscountRequest) (*RemoveDiscountReply, error) {
	return nil, status.Error(codes.Unimplemented, "method RemoveDiscount not implemented")
}
func (UnimplementedProductServiceServer) GetProduct(context.Context, *GetProductRequest) (*GetProductReply, error) {
	return nil, status.Error(codes.Unimplemented, "method GetProduct not implemented")
}
func (UnimplementedProductServiceServer) ListProducts(context.Context, *ListProductsRequest) (*ListProductsReply, error) {
	return nil, status.Error(codes.Unimplemented, "method ListProducts not implemented")
}

func RegisterProductServiceServer(s grpc.ServiceRegistrar, srv ProductServiceServer) {
	s.RegisterService(&grpc.ServiceDesc{ServiceName: "product.v1.ProductService", HandlerType: (*ProductServiceServer)(nil)}, srv)
}
