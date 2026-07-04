package handler

import (
	"net/http"
	"strconv"

	"log-llm/backend/internal/config"
	"log-llm/backend/internal/repository"
	"log-llm/backend/internal/service"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	cfg       config.Config
	chat      *service.ChatService
	dashboard *service.DashboardService
}

func New(cfg config.Config, chat *service.ChatService, dashboard *service.DashboardService) *Handler {
	return &Handler{
		cfg:       cfg,
		chat:      chat,
		dashboard: dashboard,
	}
}

func (h *Handler) Router() *gin.Engine {
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     h.cfg.CORSAllowedOrigins,
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodOptions},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
	}))

	router.GET("/health", h.health)

	api := router.Group("/api")
	api.POST("/chat", h.postChat)
	api.GET("/conversations", h.getConversations)
	api.GET("/conversations/:id/messages", h.getMessages)

	dashboard := api.Group("/dashboard")
	dashboard.GET("/risk-events", h.getRiskEvents)
	dashboard.GET("/risk-summary", h.getRiskSummary)
	dashboard.GET("/conversations/:id", h.getConversationDetail)
	dashboard.GET("/detector-runs", h.getDetectorRuns)
	dashboard.GET("/detector-runs/:requestId", h.getDetectorRuns)

	return router
}

func (h *Handler) health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) postChat(c *gin.Context) {
	var request service.ChatRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		writeError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	response, err := h.chat.Chat(c.Request.Context(), request)
	if err != nil {
		if repository.IsNotFound(err) {
			writeError(c, http.StatusNotFound, "conversation not found")
			return
		}
		writeError(c, http.StatusBadGateway, "chat request could not be completed")
		return
	}
	c.JSON(http.StatusOK, response)
}

func (h *Handler) getConversations(c *gin.Context) {
	conversations, err := h.dashboard.ListConversations(c.Request.Context())
	if err != nil {
		writeError(c, http.StatusInternalServerError, "could not load conversations")
		return
	}
	c.JSON(http.StatusOK, gin.H{"conversations": conversations})
}

func (h *Handler) getMessages(c *gin.Context) {
	messages, err := h.dashboard.ListMessages(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeError(c, http.StatusInternalServerError, "could not load messages")
		return
	}
	c.JSON(http.StatusOK, gin.H{"messages": messages})
}

func (h *Handler) getRiskEvents(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	events, err := h.dashboard.RiskEvents(c.Request.Context(), c.Query("conversationId"), limit)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "could not load risk events")
		return
	}
	c.JSON(http.StatusOK, gin.H{"riskEvents": events})
}

func (h *Handler) getRiskSummary(c *gin.Context) {
	summary, err := h.dashboard.RiskSummary(c.Request.Context())
	if err != nil {
		writeError(c, http.StatusInternalServerError, "could not load risk summary")
		return
	}
	c.JSON(http.StatusOK, summary)
}

func (h *Handler) getConversationDetail(c *gin.Context) {
	detail, err := h.dashboard.ConversationDetail(c.Request.Context(), c.Param("id"))
	if err != nil {
		if repository.IsNotFound(err) {
			writeError(c, http.StatusNotFound, "conversation not found")
			return
		}
		writeError(c, http.StatusInternalServerError, "could not load conversation")
		return
	}
	c.JSON(http.StatusOK, detail)
}

func (h *Handler) getDetectorRuns(c *gin.Context) {
	requestID := c.Param("requestId")
	if requestID == "" {
		requestID = c.Query("requestId")
	}
	if requestID == "" {
		writeError(c, http.StatusBadRequest, "requestId is required")
		return
	}
	runs, err := h.dashboard.DetectorRunsByRequest(c.Request.Context(), requestID)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "could not load detector runs")
		return
	}
	c.JSON(http.StatusOK, gin.H{"detectorRuns": runs})
}

func writeError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"error": message})
}
