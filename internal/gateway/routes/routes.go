package routes

import (
	"github.com/ak-repo/stream-hub/config"
	"github.com/ak-repo/stream-hub/internal/gateway/handler"
	"github.com/ak-repo/stream-hub/internal/gateway/middleware"
	"github.com/ak-repo/stream-hub/pkg/grpc/clients"
	"github.com/ak-repo/stream-hub/pkg/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/websocket/v2"
)

// New initializes the Fiber application with global middleware (CORS)
// and groups the administrative and user-facing routes.
func New(app *fiber.App, cfg *config.Config, clients *clients.Clients) {
	// Global Middleware: CORS Configuration
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:5173, http://localhost:3000, http://localhost:3001,http://localhost:3002",
		AllowCredentials: true,
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		ExposeHeaders:    "Authorization",
	}))

	// API Group Prefix
	api := app.Group("/api/v1")

	// Route Definitions
	adminRoutes(api, clients, cfg)
	userRoutes(api, cfg, clients)

}

// ---
// 🔒 Admin Routes
// ---

func adminRoutes(api fiber.Router, clients *clients.Clients, cfg *config.Config) {

	jwtMan := jwt.NewJWTManager(cfg.JWT.Secret, cfg.JWT.Expiry, cfg.JWT.Expiry*7)

	adminR := api.Group("/admin")
	// Requires JWT and Admin Role check in middleware
	adminR.Use(middleware.AuthMiddleware(jwtMan))

	// Admin Handler Initialization
	handler := handler.NewAdminHandler(clients, cfg)

	// Users actions (Moderation and Management)
	adminR.Get("/users", handler.ListUsers)
	adminR.Post("/users/ban", handler.BanUser)
	adminR.Post("/users/unban", handler.UnbanUser)
	adminR.Post("/users/change-role", handler.UpdateRole)
	adminR.Post("/users/uploads-block", handler.BlockUserUpload)
	adminR.Delete("/users/:target_user_id", handler.DeleteUser)

	// Channels actions (Moderation)
	adminR.Get("/channels", handler.ListChannels)
	adminR.Post("/channels/freeze", handler.FreezeChannel)
	adminR.Post("/channels/unfreeze", handler.FreezeChannel) // Note: Same handler used for both
	adminR.Delete("/channels/:channel_id", handler.DeleteChannel)

	// Files actions (Content moderation)
	adminR.Get("/files", handler.ListAllFiles)
	adminR.Delete("/files/:id", handler.DeleteFile)

}

// ---
// 👤 User Routes
// ---

func userRoutes(api fiber.Router, cfg *config.Config, clients *clients.Clients) {
	jwtMan := jwt.NewJWTManager(cfg.JWT.Secret, cfg.JWT.Expiry, cfg.JWT.Expiry*7)

	// --------------------------
	// AUTH HANDLER (Public and Authenticated)
	// --------------------------
	auth := handler.NewAuthHandler(clients.Auth, jwtMan)

	// Public Routes
	public := api.Group("")
	public.Post("/login", auth.Login)       // tested
	public.Post("/register", auth.Register) // tested
	public.Post("/verify-gen", auth.SendMagicLink)
	public.Get("/verify-link", auth.VerifyMagicLink)
	public.Post("/forget-password", auth.ForgetPassword)
	public.Post("/verify-password", auth.VerifyOTPForPasswordReset)

	// Authenticated Auth Routes
	authR := api.Group("/auth")
	authR.Use(middleware.AuthMiddleware(jwtMan))
	authR.Get("/users", auth.SearchUsers)
	authR.Post("/profile-update", auth.UpdateProfile)   // tested
	authR.Post("/change-password", auth.ChangePassword) // tested
	authR.Post("/upload-profile", auth.UploadAvatar)

	// --------------------------
	// FILE HANDLER (Authenticated)
	// --------------------------
	file := handler.NewFileHandler(clients.File)

	fileR := api.Group("/files")
	fileR.Use(middleware.AuthMiddleware(jwtMan))
	fileR.Post("/upload-url", file.CreateUploadUrl)      // tested
	fileR.Post("/confirm", file.CompleteUpload)          // tested
	fileR.Get("/download-url", file.GenerateDownloadURL) // tested
	fileR.Get("/", file.ListFiles)                       // tested
	fileR.Delete("/:file_id", file.DeleteFile)           // tested

	// --------------------------
	// CHANNEL HANDLER (WebSockets and Authenticated)
	// --------------------------
	channel := handler.NewChannelHandler(clients.Channel)

	// Websocket route for real-time communication
	api.Get("/ws", websocket.New(channel.WsHandler))

	ch := api.Group("/channels")
	ch.Use(middleware.AuthMiddleware(jwtMan))

	// Channel CRUD and Listing
	ch.Get("", channel.ListChannels)                      // tested
	ch.Post("/create", channel.CreateChannel)             // tested
	ch.Delete("/leave/:channel_id", channel.LeaveChannel) // tested
	ch.Delete("/:channel_id", channel.DeleteChannel)      // tested
	ch.Get("/search", channel.SearchChannels)
	// Channel Details and History
	ch.Get("/channel/:channelId", channel.GetChannel)   // tested
	ch.Get("/members/:channelId", channel.ListMembers)  // tested
	ch.Get("/:channelId/history", channel.ListMessages) //tested
	ch.Get("/storage/:channel_id", channel.GetChannelStorage)

	// Request Handling (Invites/Joins)
	ch.Post("/sendinvite", channel.SendInvite)     // tested
	ch.Post("/sendjoin", channel.SendJoin)         // tested
	ch.Get("/invites", channel.ListUserInvites)    // tested
	ch.Get("/joins/:id", channel.ListChannelJoins) // tested
	ch.Post("/updatereq", channel.UpdateRequestStatus)

	// --------------------------
	// PAYMENT HANDLER  - Razorpay
	// --------------------------
	payHandler := handler.NewPaymentHandler(clients.Payment)
	pay := api.Group("/payment")
	pay.Use(middleware.AuthMiddleware(jwtMan))

	pay.Post("/session", payHandler.CreatePaymentSession)
	pay.Post("/verify", payHandler.VerifyPayment)

	pay.Get("/subscription/:channel_id", payHandler.SubscriptionPlans)

}
