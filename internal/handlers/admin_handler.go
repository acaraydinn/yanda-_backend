package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yandas/backend/internal/models"
	"github.com/yandas/backend/internal/services"
)

type AdminHandler struct {
	svcs *services.Services
}

func NewAdminHandler(svcs *services.Services) *AdminHandler {
	return &AdminHandler{svcs: svcs}
}

func (h *AdminHandler) Dashboard(c *gin.Context) {
	stats, _ := h.svcs.Admin.GetDashboard()
	c.JSON(http.StatusOK, SuccessResponse(stats))
}

func (h *AdminHandler) ListUsers(c *gin.Context) {
	page, limit := getPagination(c)
	users, total, _ := h.svcs.Admin.ListUsers(page, limit, c.Query("role"))
	c.JSON(http.StatusOK, SuccessResponseWithMeta(users, PaginationMeta(page, limit, total)))
}

func (h *AdminHandler) GetUser(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	user, err := h.svcs.Admin.GetUser(id)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, SuccessResponse(user))
}

func (h *AdminHandler) UpdateUser(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	var updates map[string]interface{}
	c.ShouldBindJSON(&updates)
	user, _ := h.svcs.Admin.UpdateUser(id, updates)
	c.JSON(http.StatusOK, SuccessResponse(user))
}

func (h *AdminHandler) DeleteUser(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	h.svcs.Admin.DeleteUser(id)
	c.JSON(http.StatusOK, SuccessResponse(gin.H{"message": "Deleted"}))
}

func (h *AdminHandler) ListApplications(c *gin.Context) {
	page, limit := getPagination(c)
	status := c.Query("status")
	apps, total, _ := h.svcs.Admin.ListApplications(page, limit, status)
	c.JSON(http.StatusOK, SuccessResponseWithMeta(apps, PaginationMeta(page, limit, total)))
}

func (h *AdminHandler) GetApplication(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	app, _ := h.svcs.Admin.GetApplication(id)
	c.JSON(http.StatusOK, SuccessResponse(app))
}

func (h *AdminHandler) ApproveApplication(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	h.svcs.Admin.ApproveApplication(id, getUserID(c))
	c.JSON(http.StatusOK, SuccessResponse(gin.H{"message": "Approved"}))
}

func (h *AdminHandler) RejectApplication(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	var input struct {
		Reason string `json:"reason"`
	}
	c.ShouldBindJSON(&input)
	h.svcs.Admin.RejectApplication(id, getUserID(c), input.Reason)
	c.JSON(http.StatusOK, SuccessResponse(gin.H{"message": "Rejected"}))
}

func (h *AdminHandler) ListOrders(c *gin.Context) {
	page, limit := getPagination(c)
	orders, total, _ := h.svcs.Admin.ListOrders(page, limit, c.Query("status"))
	c.JSON(http.StatusOK, SuccessResponseWithMeta(orders, PaginationMeta(page, limit, total)))
}

func (h *AdminHandler) GetOrder(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	order, _ := h.svcs.Admin.GetOrder(id)
	c.JSON(http.StatusOK, SuccessResponse(order))
}

func (h *AdminHandler) CreateCategory(c *gin.Context) {
	var cat models.Category
	c.ShouldBindJSON(&cat)
	h.svcs.Admin.CreateCategory(&cat)
	c.JSON(http.StatusCreated, SuccessResponse(cat))
}

func (h *AdminHandler) UpdateCategory(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	var cat models.Category
	c.ShouldBindJSON(&cat)
	cat.ID = id
	h.svcs.Admin.UpdateCategory(&cat)
	c.JSON(http.StatusOK, SuccessResponse(cat))
}

func (h *AdminHandler) DeleteCategory(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	h.svcs.Admin.DeleteCategory(id)
	c.JSON(http.StatusOK, SuccessResponse(gin.H{"message": "Deleted"}))
}

func (h *AdminHandler) AnalyticsOverview(c *gin.Context) {
	stats, _ := h.svcs.Admin.GetDashboard()
	c.JSON(http.StatusOK, SuccessResponse(stats))
}

func (h *AdminHandler) AnalyticsRevenue(c *gin.Context) {
	c.JSON(http.StatusOK, SuccessResponse(gin.H{"revenue": 0}))
}

func (h *AdminHandler) AnalyticsUsers(c *gin.Context) {
	c.JSON(http.StatusOK, SuccessResponse(gin.H{"users": 0}))
}

func (h *AdminHandler) AuditLogs(c *gin.Context) {
	page, limit := getPagination(c)
	logs, total, _ := h.svcs.Admin.GetAuditLogs(page, limit, nil, "")
	c.JSON(http.StatusOK, SuccessResponseWithMeta(logs, PaginationMeta(page, limit, total)))
}

// Support Ticket handlers

func (h *AdminHandler) ListSupportTickets(c *gin.Context) {
	page, limit := getPagination(c)
	status := c.Query("status")
	priority := c.Query("priority")
	tickets, total, _ := h.svcs.Admin.ListSupportTickets(page, limit, status, priority)
	c.JSON(http.StatusOK, SuccessResponseWithMeta(tickets, PaginationMeta(page, limit, total)))
}

func (h *AdminHandler) GetSupportTicket(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	ticket, err := h.svcs.Admin.GetSupportTicket(id)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, SuccessResponse(ticket))
}

func (h *AdminHandler) UpdateSupportTicket(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	var updates struct {
		Status     string `json:"status"`
		Priority   string `json:"priority"`
		AssignedTo string `json:"assigned_to"`
	}
	c.ShouldBindJSON(&updates)
	ticket, err := h.svcs.Admin.UpdateSupportTicket(id, updates.Status, updates.Priority, updates.AssignedTo)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, SuccessResponse(ticket))
}

func (h *AdminHandler) ReplySupportTicket(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	var input struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("Content is required"))
		return
	}
	message, err := h.svcs.Admin.ReplySupportTicket(id, getUserID(c), input.Content)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusCreated, SuccessResponse(message))
}

func (h *AdminHandler) GetSupportStats(c *gin.Context) {
	stats, _ := h.svcs.Admin.GetSupportStats()
	c.JSON(http.StatusOK, SuccessResponse(stats))
}
