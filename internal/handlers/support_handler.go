package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yandas/backend/internal/services"
)

// SupportHandler handles user-facing support endpoints
type SupportHandler struct {
	svcs *services.Services
}

// NewSupportHandler creates a new support handler
func NewSupportHandler(svcs *services.Services) *SupportHandler {
	return &SupportHandler{svcs: svcs}
}

// CreateTicket creates a new support ticket
func (h *SupportHandler) CreateTicket(c *gin.Context) {
	var input services.CreateTicketInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}

	ticket, err := h.svcs.Support.CreateTicket(getUserID(c), &input)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusCreated, SuccessResponse(ticket))
}

// ListTickets returns user's support tickets
func (h *SupportHandler) ListTickets(c *gin.Context) {
	page, limit := getPagination(c)
	tickets, total, err := h.svcs.Support.ListUserTickets(getUserID(c), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponseWithMeta(tickets, PaginationMeta(page, limit, total)))
}

// GetTicket returns a specific support ticket
func (h *SupportHandler) GetTicket(c *gin.Context) {
	ticketID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("invalid ticket ID"))
		return
	}

	ticket, err := h.svcs.Support.GetUserTicket(getUserID(c), ticketID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(ticket))
}

// ReplyTicket adds a message to a support ticket
func (h *SupportHandler) ReplyTicket(c *gin.Context) {
	ticketID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("invalid ticket ID"))
		return
	}

	var input struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}

	msg, err := h.svcs.Support.ReplyTicket(getUserID(c), ticketID, input.Content)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusCreated, SuccessResponse(msg))
}

// SearchHandler handles search endpoints
type SearchHandler struct {
	svcs *services.Services
}

// NewSearchHandler creates a new search handler
func NewSearchHandler(svcs *services.Services) *SearchHandler {
	return &SearchHandler{svcs: svcs}
}

// SearchYandas searches yandaş profiles
func (h *SearchHandler) SearchYandas(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse("query parameter 'q' is required"))
		return
	}

	page, limit := getPagination(c)
	profiles, total, err := h.svcs.Yandas.Search(query, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponseWithMeta(profiles, PaginationMeta(page, limit, total)))
}
