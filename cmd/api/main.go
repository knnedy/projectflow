package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/knnedy/projectflow/internal/config"
	"github.com/knnedy/projectflow/internal/handler"
	"github.com/knnedy/projectflow/internal/repository"
	"github.com/knnedy/projectflow/internal/router"
	"github.com/knnedy/projectflow/internal/service"
	"github.com/knnedy/projectflow/internal/token"
)

func main() {
	// load config
	cfg := config.Load()

	// setup structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// connect to database
	db, err := repository.NewDB((cfg.DBUrl))
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Pool.Close()
	slog.Info("connected to the database")

	// init token manager
	tokens := token.NewTokenManager(cfg.JWTSecret)

	// init services
	activitySvc := service.NewActivityService(db)
	authSvc := service.NewAuthService(db.Queries(), tokens)
	userSvc := service.NewUserService(db.Queries())
	orgSvc := service.NewOrgService(db, activitySvc)
	memberSvc := service.NewMemberService(db, activitySvc)
	projectSvc := service.NewProjectService(db, activitySvc)
	issueSvc := service.NewIssueService(db, activitySvc)
	commentSvc := service.NewCommentService(db)

	// init handlers
	authHandler := handler.NewAuthHandler(authSvc)
	userHandler := handler.NewUserHandler(userSvc)
	orgHandler := handler.NewOrgHandler(orgSvc, memberSvc)
	memberHandler := handler.NewMemberHandler(memberSvc)
	projectHandler := handler.NewProjectHandler(projectSvc)
	issueHandler := handler.NewIssueHandler(issueSvc)
	commentHandler := handler.NewCommentHandler(commentSvc)
	activityHandler := handler.NewActivityHandler(activitySvc)

	// init router
	r := router.New(
		db,
		tokens,
		authHandler,
		userHandler,
		orgHandler,
		memberHandler,
		projectHandler,
		issueHandler,
		commentHandler,
		activityHandler,
	)

	// start server
	addr := fmt.Sprintf(":%s", cfg.Port)
	slog.Info("starting server", "addr", addr, "env", cfg.Env)

	if err := http.ListenAndServe(addr, r); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}

}
