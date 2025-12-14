package port

import (
	"context"

	"github.com/ak-repo/stream-hub/internal/payment_service/domain"
	"github.com/google/uuid"
)

// ApplicationService is the INCOMING port.
// Implemented by app/service.go, called by adapter/grpc/handler.go.
type ApplicationService interface {
	CreatePaymentSession(ctx context.Context, session *domain.PaymentSession) (razorpayOrderID string, err error)
	VerifyPaymentAndAddStorage(ctx context.Context, razorpayOrderID, razorpayPaymentID, razorpaySignature string) error
	GetHistoryByChannel(ctx context.Context, channelID uuid.UUID) ([]*domain.PaymentHistory, error)
	ListSubscriptionPlans(ctx context.Context, requesterID, channelID string) ([]*domain.SubscriptionPlan, error)
	ChannelPlanID(ctx context.Context, channelID string) (string, error)
}

// Repository is an OUTGOING port (for database access).
// Implemented by adapter/db/pgxpool/repository.go.
type Repository interface {
	RecordPaymentHistory(ctx context.Context, history *domain.PaymentHistory) error
	GetHistoryByChannel(ctx context.Context, channelID uuid.UUID) ([]*domain.PaymentHistory, error)
	ListSubscriptionPlans(ctx context.Context) ([]*domain.SubscriptionPlan, error)
	PlanByID(ctx context.Context, planID string) (*domain.SubscriptionPlan, error)
	ChannelPlanID(ctx context.Context, channelID string) (string, error)
}

// PaymentGateway is an OUTGOING port (for Razorpay interaction).
// Implemented by adapter/razorpay/gateway.go.
type PaymentGateway interface {
	CreateOrder(ctx context.Context, amountCents int64) (razorpayOrderID string, err error)
	VerifySignature(
		razorpayOrderID string,
		razorpayPaymentID string,
		razorpaySignature string,
	) error
}

type Redis interface {
	SavePaymentSession(ctx context.Context, session *domain.PaymentSession) error
	GetSessionByOrderID(ctx context.Context, razorpayOrderID string) (*domain.PaymentSession, error)
}
