package paymentgrpc

import (
	"context"
	"time"

	"github.com/ak-repo/stream-hub/gen/paymentpb"
	"github.com/ak-repo/stream-hub/internal/payment_service/domain"
	"github.com/ak-repo/stream-hub/internal/payment_service/port"
)

// Server is the gRPC handler that implements the PaymentServiceServer interface
type Server struct {
	paymentpb.UnimplementedPaymentServiceServer
	appService port.ApplicationService
}

func NewGrpcServer(app port.ApplicationService) *Server {
	return &Server{appService: app}
}

func (s *Server) CreatePaymentSession(ctx context.Context, req *paymentpb.CreatePaymentSessionRequest) (*paymentpb.CreatePaymentSessionResponse, error) {

	session := &domain.PaymentSession{
		ChannelID:       req.ChannelId,
		PurchaserUserID: req.PurchaserUserId,
		AmountCents:     req.GetAmountPaidCents(),
		StorageBytes:    req.GetStorageAddedBytes(),
		CreatedAt:       time.Now().UTC(),
		Status:          "pending",
	}

	orderID, err := s.appService.CreatePaymentSession(ctx, session)
	if err != nil {
		return nil, err
	}

	return &paymentpb.CreatePaymentSessionResponse{
		RazorpayOrderId: orderID,
	}, nil
}

func (s *Server) VerifyPayment(ctx context.Context, req *paymentpb.VerifyPaymentRequest) (*paymentpb.VerifyPaymentResponse, error) {
	err := s.appService.VerifyPaymentAndAddStorage(
		ctx,
		req.GetRazorpayOrderId(),
		req.GetRazorpayPaymentId(),
		req.GetRazorpaySignature(),
	)
	if err != nil {
		return nil, err
	}
	return &paymentpb.VerifyPaymentResponse{}, nil
}
