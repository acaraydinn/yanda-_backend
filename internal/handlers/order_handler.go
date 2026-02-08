package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yandas/backend/internal/services"
)

type OrderHandler struct {
	svcs *services.Services
}

func NewOrderHandler(svcs *services.Services) *OrderHandler {
	return &OrderHandler{svcs: svcs}
}

func (h *OrderHandler) Create(c *gin.Context) {
	var input services.CreateOrderInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}
	order, err := h.svcs.Order.Create(getUserID(c), &input)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusCreated, SuccessResponse(order))
}

func (h *OrderHandler) List(c *gin.Context) {
	page, limit := getPagination(c)
	orders, total, _ := h.svcs.Order.List(getUserID(c), page, limit, c.Query("status"))
	c.JSON(http.StatusOK, SuccessResponseWithMeta(orders, PaginationMeta(page, limit, total)))
}

func (h *OrderHandler) Get(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	order, err := h.svcs.Order.Get(id, getUserID(c))
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, SuccessResponse(order))
}

func (h *OrderHandler) Cancel(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	var input struct{ Reason string `json:"reason"` }
	c.ShouldBindJSON(&input)
	if err := h.svcs.Order.Cancel(id, getUserID(c), input.Reason); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, SuccessResponse(gin.H{"message": "Cancelled"}))
}

func (h *OrderHandler) Review(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	var input services.ReviewInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}
	review, err := h.svcs.Order.Review(id, getUserID(c), &input)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusCreated, SuccessResponse(review))
}

type CategoryHandler struct {
	svcs *services.Services
}

func NewCategoryHandler(svcs *services.Services) *CategoryHandler {
	return &CategoryHandler{svcs: svcs}
}

func (h *CategoryHandler) List(c *gin.Context) {
	categories, _ := h.svcs.Category.List()
	c.JSON(http.StatusOK, SuccessResponse(categories))
}
