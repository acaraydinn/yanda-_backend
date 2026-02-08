package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yandas/backend/internal/services"
)

// FavoriteHandler handles favorite endpoints
type FavoriteHandler struct {
	svcs *services.Services
}

// NewFavoriteHandler creates a new favorite handler
func NewFavoriteHandler(svcs *services.Services) *FavoriteHandler {
	return &FavoriteHandler{svcs: svcs}
}

// Toggle adds or removes a yandaş from favorites
func (h *FavoriteHandler) Toggle(c *gin.Context) {
	yandasID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("invalid yandas ID"))
		return
	}

	added, err := h.svcs.Favorite.Toggle(getUserID(c), yandasID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(gin.H{
		"is_favorited": added,
	}))
}

// List returns user's favorited yandaş profiles
func (h *FavoriteHandler) List(c *gin.Context) {
	page, limit := getPagination(c)
	favs, total, err := h.svcs.Favorite.List(getUserID(c), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponseWithMeta(favs, PaginationMeta(page, limit, total)))
}

// Check checks if a yandaş is favorited
func (h *FavoriteHandler) Check(c *gin.Context) {
	yandasID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("invalid yandas ID"))
		return
	}

	isFavorited := h.svcs.Favorite.IsFavorited(getUserID(c), yandasID)
	c.JSON(http.StatusOK, SuccessResponse(gin.H{
		"is_favorited": isFavorited,
	}))
}

// IDs returns just the yandas IDs that the user has favorited
func (h *FavoriteHandler) IDs(c *gin.Context) {
	ids, err := h.svcs.Favorite.GetFavoriteIDs(getUserID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(ids))
}
