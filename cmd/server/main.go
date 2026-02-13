package main

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"lostark-todo-backend/internal/auth"
	"lostark-todo-backend/internal/client"
	"lostark-todo-backend/internal/config"
	"lostark-todo-backend/internal/handler"
	"lostark-todo-backend/internal/middleware"
	"lostark-todo-backend/internal/router"
	"lostark-todo-backend/internal/scheduler"
	"lostark-todo-backend/internal/service"
	"lostark-todo-backend/internal/webhook"
)

func main() {
	// Setup structured logging
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	// 1. Load config
	cfg := config.Load()

	// 2. Setup database
	db, err := config.NewDatabase(cfg)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	slog.Info("database connected")

	// 3. Setup token provider
	tp, err := auth.NewTokenProvider(cfg.JWTKey)
	if err != nil {
		slog.Error("failed to create token provider", "error", err)
		os.Exit(1)
	}

	// 4. Create Lostark API clients
	lostarkClient := client.NewLostarkClient()
	marketClient := client.NewLostarkMarketClient(lostarkClient)

	// 5. Create Discord webhook
	discord := webhook.NewDiscordWebhook(
		cfg.DiscordWebhookURL,
		cfg.DiscordNoticeURL,
		cfg.DiscordMemberURL,
	)

	// 6. Create services
	authService := service.NewAuthService(db, tp)
	memberService := service.NewMemberService(db)
	characterService := service.NewCharacterService(db)
	characterDayService := service.NewCharacterDayService(db)
	characterWeekService := service.NewCharacterWeekService(db)
	friendService := service.NewFriendService(db)
	generalTodoService := service.NewGeneralTodoService(db)
	scheduleService := service.NewScheduleService(db)
	notificationService := service.NewNotificationService(db)
	cubeService := service.NewCubeService(db)
	logsService := service.NewLogsService(db)
	serverTodoService := service.NewServerTodoService(db)
	lifeEnergyService := service.NewLifeEnergyService(db)
	customTodoService := service.NewCustomTodoService(db)
	communityService := service.NewCommunityService(db)
	commentsService := service.NewCommentsService(db)
	contentService := service.NewContentService(db)
	myGameService := service.NewMyGameService(db)
	adminService := service.NewAdminService(db)

	// 7. Create handlers
	authHandler := handler.NewAuthHandler(authService)
	memberHandler := handler.NewMemberHandler(memberService)
	characterHandler := handler.NewCharacterHandler(characterService)
	characterListHandler := handler.NewCharacterListHandler(characterService)
	characterDayHandler := handler.NewCharacterDayHandler(characterDayService)
	characterWeekHandler := handler.NewCharacterWeekHandler(characterWeekService)
	friendHandler := handler.NewFriendHandler(friendService)
	generalTodoHandler := handler.NewGeneralTodoHandler(generalTodoService)
	scheduleHandler := handler.NewScheduleHandler(scheduleService)
	notificationHandler := handler.NewNotificationHandler(notificationService)
	cubeHandler := handler.NewCubeHandler(cubeService)
	logsHandler := handler.NewLogsHandler(logsService)
	serverTodoHandler := handler.NewServerTodoHandler(serverTodoService)
	lifeEnergyHandler := handler.NewLifeEnergyHandler(lifeEnergyService)
	customTodoHandler := handler.NewCustomTodoHandler(customTodoService)
	communityHandler := handler.NewCommunityHandler(communityService)
	commentsHandler := handler.NewCommentsHandler(commentsService)
	contentHandler := handler.NewContentHandler(contentService)
	myGameHandler := handler.NewMyGameHandler(myGameService)
	adminHandler := handler.NewAdminHandler(adminService)

	// 8. Create OAuth2 handler
	oauth2Handler := auth.NewOAuth2Handler(
		cfg.GoogleClientID,
		cfg.GoogleClientSecret,
		cfg.GoogleRedirectURI,
		tp,
		db,
	)

	// 9. Setup middleware
	corsMw := middleware.NewCORS()
	rateLimiter := middleware.NewRateLimiter(50, 100)

	// MemberGetter adapter for auth middleware
	memberGetter := &memberRoleGetter{db: db}
	authMw := auth.JwtMiddleware(tp, memberGetter)

	// 10. Setup router
	r := router.NewRouter(
		authMw,
		corsMw,
		rateLimiter,
		oauth2Handler,
		authHandler,
		memberHandler,
		characterHandler,
		characterListHandler,
		characterDayHandler,
		characterWeekHandler,
		friendHandler,
		generalTodoHandler,
		scheduleHandler,
		notificationHandler,
		cubeHandler,
		logsHandler,
		serverTodoHandler,
		lifeEnergyHandler,
		customTodoHandler,
		communityHandler,
		commentsHandler,
		contentHandler,
		myGameHandler,
		adminHandler,
	)

	// 11. Start scheduler
	shedLock := scheduler.NewShedLock(db)
	sched := scheduler.NewScheduler(
		shedLock,
		db,
		lostarkClient,
		marketClient,
		discord,
		cfg.LostarkAPIKey,
	)
	sched.Start()
	defer sched.Stop()

	// 12. Start server with graceful shutdown
	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("server starting", "port", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
	}

	slog.Info("server stopped")
}

// memberRoleGetter implements auth.MemberGetter for the JWT middleware.
type memberRoleGetter struct {
	db *sql.DB
}

func (m *memberRoleGetter) GetRole(ctx context.Context, username string) (string, error) {
	var role string
	err := m.db.QueryRowContext(ctx,
		`SELECT role FROM member WHERE username = ?`, username,
	).Scan(&role)
	return role, err
}
