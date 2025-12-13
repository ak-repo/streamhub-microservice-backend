package app

import (
	"context"

	"time"

	"github.com/ak-repo/stream-hub/pkg/errors"
	"github.com/ak-repo/stream-hub/pkg/logger"
	"go.uber.org/zap"

	"github.com/ak-repo/stream-hub/internal/payment_service/domain"
	"github.com/ak-repo/stream-hub/internal/payment_service/port"
	"github.com/google/uuid"
)

type paymentService struct {
	repo  port.Repository
	redis port.Redis
	pg    port.PaymentGateway
}

func NewPaymentService(repo port.Repository, pg port.PaymentGateway, redis port.Redis) port.ApplicationService {
	return &paymentService{repo: repo, pg: pg, redis: redis}
}

func (s *paymentService) CreatePaymentSession(ctx context.Context, session *domain.PaymentSession) (string, error) {

	plan, err := s.repo.PlanByID(ctx, session.PlanID)
	if err != nil {
		return "", errors.New(errors.CodeNotFound, "failed to find plan", err)
	}

	orderID, err := s.pg.CreateOrder(ctx, int64(plan.PriceINR))
	if err != nil {
		return "", errors.New(errors.CodeInternal, "failed to create order id", err)
	}
	session.RazorpayOrderID = orderID

	if err := s.redis.SavePaymentSession(ctx, session); err != nil {
		return "", errors.New(errors.CodeInternal, "failed to save order session", err)
	}

	return orderID, nil
}

func (s *paymentService) VerifyPaymentAndAddStorage(ctx context.Context, razorpayOrderID, razorpayPaymentID, razorpaySignature string) error {
	session, err := s.redis.GetSessionByOrderID(ctx, razorpayOrderID)
	if err != nil {
		return errors.New(errors.CodeNotFound, "payment session not found or database error", err)
	}

	if err := s.pg.VerifySignature(razorpayOrderID, razorpayPaymentID, razorpaySignature); err != nil {
		return errors.New(errors.CodeForbidden, "payment verification failed: invalid signature", err)
	}

	// 3. Payment is confirmed and verified. Record the transaction in history.
	history := &domain.PaymentHistory{
		ID:                uuid.New().String(),
		ChannelID:         session.ChannelID,
		PurchaserUserID:   session.PurchaserUserID,
		RazorpayPaymentID: razorpayPaymentID,
		RazorpayOrderID:   razorpayOrderID,
		PlanID:            session.PlanID,
		AmountPaidINR:     session.AmountINR,
		CreatedAt:         time.Now().UTC(),
	}

	if err := s.repo.RecordPaymentHistory(ctx, history); err != nil {
		// TODO Log the details
		logger.Log.Error("payment verification sucessfull, but failed to save in DB", zap.Error(err))

		return errors.New(errors.CodeInternal, "payment verification sucessfull, but failed to save in DB", err)
	}

	// TODO update channels plans in channels table

	return nil
}

func (s *paymentService) GetHistoryByChannel(ctx context.Context, channelID uuid.UUID) ([]*domain.PaymentHistory, error) {
	return s.repo.GetHistoryByChannel(ctx, channelID)
}

func (s *paymentService) GetSubscriptionPlans(ctx context.Context, requesterID, channelID string) ([]*domain.SubscriptionPlan, error) {
	// TODO verify the channelID and user are valid
	return s.repo.GetSubscriptionPlans(ctx)
}
