package main

import (
	"context"
	"log"

	"log-llm/backend/internal/config"
	"log-llm/backend/internal/database"
	"log-llm/backend/internal/detector"
	"log-llm/backend/internal/handler"
	"log-llm/backend/internal/llm"
	"log-llm/backend/internal/repository"
	"log-llm/backend/internal/service"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	if err := database.RunMigrations(ctx, db, cfg.MigrationsDir); err != nil {
		log.Fatalf("run migrations: %v", err)
	}

	repo := repository.New(db)
	llmRouter := llm.NewRouter(cfg)
	ruleDetector := detector.NewRuleDetector()
	judgeDetector := detector.NewLLMJudgeDetector(llmRouter, repo, cfg.DefaultJudgeProvider, cfg.OllamaJudgeModel)
	detectionService := detector.NewService(ruleDetector, judgeDetector)

	chatService := service.NewChatService(cfg, repo, llmRouter, detectionService)
	dashboardService := service.NewDashboardService(repo)
	httpHandler := handler.New(cfg, chatService, dashboardService)

	log.Printf("backend listening on :%s", cfg.BackendPort)
	if err := httpHandler.Router().Run(":" + cfg.BackendPort); err != nil {
		log.Fatalf("run server: %v", err)
	}
}
