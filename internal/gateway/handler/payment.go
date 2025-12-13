package handler

// import (
// 	"encoding/json"
// 	"fmt"
// 	"os"

// 	"github.com/gofiber/fiber/v2"
// 	"github.com/stripe/stripe-go/v78"
// 	"github.com/stripe/stripe-go/v78/checkout/session"
// 	"github.com/stripe/stripe-go/v78/webhook"
// )

// type Handler struct {
// 	ChannelRepo *payment.Repository
// }

// func NewPaymentHandler(repo *payment.Repository) *Handler {
// 	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
// 	return &Handler{ChannelRepo: repo}
// }

// // CreateCheckoutSession: Admin clicks "Buy Storage for Channel"
// func (h *Handler) CreateCheckoutSession(c *fiber.Ctx) error {
// 	type Request struct {
// 		PlanID    string `json:"plan_id"`
// 		ChannelID string `json:"channel_id"` // <--- NEW: The channel getting the storage
// 		UserID    string `json:"user_id"`    // The admin buying it
// 	}
// 	var req Request
// 	if err := c.BodyParser(&req); err != nil {
// 		return c.Status(400).SendString("Invalid Input")
// 	}

// 	// TODO: You should add a check here to ensure req.UserID is actually an Admin of req.ChannelID

// 	// Hardcoded plans (Example)
// 	var priceID string
// 	var bytesToAdd int64
// 	switch req.PlanID {
// 	case "channel_boost_100gb":
// 		priceID = "price_1P..." // Replace with real Stripe Price ID
// 		bytesToAdd = 100 * 1024 * 1024 * 1024
// 	default:
// 		return c.Status(400).SendString("Invalid Plan ID")
// 	}

// 	params := &stripe.CheckoutSessionParams{
// 		Mode: stripe.String(string(stripe.CheckoutSessionModePayment)),
// 		LineItems: []*stripe.CheckoutSessionLineItemParams{
// 			{
// 				Price:    stripe.String(priceID),
// 				Quantity: stripe.Int64(1),
// 			},
// 		},
// 		SuccessURL: stripe.String(os.Getenv("SERVER_URL") + "/payment/success"),
// 		CancelURL:  stripe.String(os.Getenv("SERVER_URL") + "/payment/cancel"),

// 		// CRITICAL CHANGE: We now attach ChannelID in metadata
// 		Metadata: map[string]string{
// 			"channel_id":   req.ChannelID,
// 			"user_id":      req.UserID, // We still track who paid
// 			"bytes_to_add": fmt.Sprintf("%d", bytesToAdd),
// 			"plan_id":      req.PlanID,
// 		},
// 	}

// 	sess, err := session.New(params)
// 	if err != nil {
// 		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
// 	}

// 	return c.JSON(fiber.Map{"url": sess.URL})
// }

// // HandleWebhook: Updates the Channel table
// func (h *Handler) HandleWebhook(c *fiber.Ctx) error {
// 	webhookSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
// 	payload := c.Body()
// 	sigHeader := c.Get("Stripe-Signature")

// 	event, err := webhook.ConstructEvent(payload, sigHeader, webhookSecret)
// 	if err != nil {
// 		return c.Status(400).SendString("Webhook Error")
// 	}

// 	if event.Type == "checkout.session.completed" {
// 		var sess stripe.CheckoutSession
// 		if err := json.Unmarshal(event.Data.Raw, &sess); err != nil {
// 			return c.Status(400).SendString("Error parsing session")
// 		}

// 		// Retrieve Metadata
// 		channelID := sess.Metadata["channel_id"] // <--- Get Channel ID
// 		userID := sess.Metadata["user_id"]       // <--- Get Purchaser ID

// 		// (In real code, parse bytes_to_add from string to int64)
// 		var bytesToAdd int64 = 100 * 1024 * 1024 * 1024

// 		fmt.Printf("💰 Admin %s bought storage for Channel %s\n", userID, channelID)

// 		// UPDATE CHANNEL IN DB
// 		err = h.ChannelRepo.AddChannelStorage(c.Context(), channelID, userID, bytesToAdd, sess.ID, int(sess.AmountTotal))
// 		if err != nil {
// 			fmt.Printf("❌ Failed to update channel storage: %v\n", err)
// 			return c.Status(500).SendString("Database Error")
// 		}

// 		fmt.Println("✅ Channel storage updated successfully!")
// 	}

// 	return c.SendStatus(200)
// }
