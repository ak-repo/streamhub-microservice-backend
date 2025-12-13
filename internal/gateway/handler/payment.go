package handler

import (
	"context"
	"time"

	"github.com/ak-repo/stream-hub/gen/paymentpb"
	"github.com/ak-repo/stream-hub/pkg/errors"
	"github.com/ak-repo/stream-hub/pkg/response"
	"github.com/gofiber/fiber/v2"
)

type PaymentHandler struct {
	client paymentpb.PaymentServiceClient
}

func NewPaymentHandler(cli paymentpb.PaymentServiceClient) *PaymentHandler {
	return &PaymentHandler{client: cli}
}

func (h *PaymentHandler) CreatePaymentSession(c *fiber.Ctx) error {
	req := new(paymentpb.CreatePaymentSessionRequest)
	if err := c.BodyParser(req); err != nil {
		return response.InvalidReqBody(c)
	}

	uid, ok := c.Locals("userID").(string)
	if !ok || uid == "" {
		return response.Error(c, fiber.StatusUnauthorized, fiber.Map{"error": "unauthorized"})
	}
	req.PurchaserUserId = uid // Set the payer's ID

	if req.ChannelId == "" || req.AmountPaidCents <= 0 || req.StorageAddedBytes <= 0 {
		return response.InvalidReqBody(c)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := h.client.CreatePaymentSession(ctx, req)
	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}
	return response.Success(c, "payment session created, proceed to checkout", resp)
}

func (h *PaymentHandler) VerifyPayment(c *fiber.Ctx) error {
	req := new(paymentpb.VerifyPaymentRequest)
	if err := c.BodyParser(req); err != nil {
		return response.InvalidReqBody(c)
	}

	if req.RazorpayOrderId == "" || req.RazorpayPaymentId == "" || req.RazorpaySignature == "" {
		return response.InvalidReqBody(c)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := h.client.VerifyPayment(ctx, req)
	if err != nil {
		code, body := errors.GRPCToFiber(err)
		return response.Error(c, code, body)
	}

	return response.Success(c, "payment successfully processed and storage added", resp)
}
