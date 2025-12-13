package pay

import (
	"context"
	"fmt"

	"github.com/ak-repo/stream-hub/config"
	"github.com/ak-repo/stream-hub/internal/payment_service/port"
	"github.com/razorpay/razorpay-go"
	"github.com/razorpay/razorpay-go/utils"
)

type RazorpayGateway struct {
	client *razorpay.Client
	cfg    *config.Config
}

func NewRazorpayGateway(cfg *config.Config) port.PaymentGateway {
	client := razorpay.NewClient(
		cfg.Razorpay.PublishableKey,
		cfg.Razorpay.SecretKey,
	)

	return &RazorpayGateway{
		client: client,
		cfg:    cfg,
	}
}

// amountPaise example: ₹500 => 50000
func (g *RazorpayGateway) CreateOrder(
	ctx context.Context,
	amountPaise int64,
) (string, error) {

	data := map[string]interface{}{
		"amount":   amountPaise*100,
		"currency": "INR",
		"receipt":  fmt.Sprintf("receipt_%d", amountPaise),
	}

	order, err := g.client.Order.Create(data, nil)
	if err != nil {
		return "", fmt.Errorf("razorpay order creation failed: %w", err)
	}

	orderID, ok := order["id"].(string)
	if !ok {
		return "", fmt.Errorf("razorpay order ID missing: %+v", order)
	}

	return orderID, nil
}

func (g *RazorpayGateway) VerifySignature(
	razorpayOrderID string,
	razorpayPaymentID string,
	razorpaySignature string,
) error {

	params := map[string]interface{}{
		"razorpay_order_id":   razorpayOrderID,
		"razorpay_payment_id": razorpayPaymentID,
	}

	isValid := utils.VerifyPaymentSignature(
		params,
		razorpaySignature,
		g.cfg.Razorpay.SecretKey,
	)

	if !isValid {
		return fmt.Errorf("invalid razorpay payment signature")
	}

	return nil
}
