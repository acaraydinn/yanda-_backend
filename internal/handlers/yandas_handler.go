package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yandas/backend/internal/services"
)

type YandasHandler struct {
	svcs *services.Services
}

func NewYandasHandler(svcs *services.Services) *YandasHandler {
	return &YandasHandler{svcs: svcs}
}

func (h *YandasHandler) ListPublic(c *gin.Context) {
	page, limit := getPagination(c)
	category := c.Query("category")
	city := c.Query("city")

	yandas, total, err := h.svcs.Yandas.ListPublic(page, limit, category, city)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, SuccessResponseWithMeta(yandas, PaginationMeta(page, limit, total)))
}

func (h *YandasHandler) GetPublic(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	yandas, err := h.svcs.Yandas.GetPublic(id)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, SuccessResponse(yandas))
}

func (h *YandasHandler) GetServices(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	services, err := h.svcs.Yandas.GetServices(id)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, SuccessResponse(services))
}

func (h *YandasHandler) GetReviews(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	page, limit := getPagination(c)
	reviews, total, _ := h.svcs.Yandas.GetReviews(id, page, limit)
	c.JSON(http.StatusOK, SuccessResponseWithMeta(reviews, PaginationMeta(page, limit, total)))
}

// saveUploadedFile saves an uploaded file and returns the URL path
func saveUploadedFile(c *gin.Context, fieldName string, userID uuid.UUID) (string, error) {
	file, err := c.FormFile(fieldName)
	if err != nil {
		return "", err
	}

	// Create uploads/documents directory if not exists
	uploadDir := filepath.Join(".", "uploads", "documents")
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", err
	}

	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%s_%s_%d%s", userID.String(), fieldName, time.Now().UnixNano(), ext)
	filePath := filepath.Join(uploadDir, filename)

	// Save the file
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		return "", err
	}

	// Return URL path
	return fmt.Sprintf("/uploads/documents/%s", filename), nil
}

func (h *YandasHandler) Apply(c *gin.Context) {
	userID := getUserID(c)

	// Handle both JSON and multipart form
	contentType := c.ContentType()

	var input services.ApplicationInput

	if contentType == "application/json" {
		// JSON request (backward compatible)
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
			return
		}
	} else {
		// Multipart form request
		input.Bio = c.PostForm("bio")
		input.InstagramHandle = c.PostForm("instagram_handle")

		// Parse service cities from form
		serviceCities := c.PostFormArray("service_cities[]")
		if len(serviceCities) == 0 {
			serviceCities = c.PostFormArray("service_cities")
		}
		input.ServiceCities = serviceCities

		// Parse category IDs
		categoryIDs := c.PostFormArray("category_ids[]")
		if len(categoryIDs) == 0 {
			categoryIDs = c.PostFormArray("category_ids")
		}
		input.CategoryIDs = categoryIDs

		// Save uploaded files - Front and Back for ID and License, PDF for Criminal Record
		if kimlikOnURL, err := saveUploadedFile(c, "kimlik_on", userID); err == nil {
			input.KimlikOnURL = kimlikOnURL
		}
		if kimlikArkaURL, err := saveUploadedFile(c, "kimlik_arka", userID); err == nil {
			input.KimlikArkaURL = kimlikArkaURL
		}
		if ehliyetOnURL, err := saveUploadedFile(c, "ehliyet_on", userID); err == nil {
			input.EhliyetOnURL = ehliyetOnURL
		}
		if ehliyetArkaURL, err := saveUploadedFile(c, "ehliyet_arka", userID); err == nil {
			input.EhliyetArkaURL = ehliyetArkaURL
		}
		if adliSicilPDFURL, err := saveUploadedFile(c, "adli_sicil_pdf", userID); err == nil {
			input.AdliSicilPDFURL = adliSicilPDFURL
		}
	}

	profile, err := h.svcs.Yandas.Apply(userID, &input)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusCreated, SuccessResponse(profile))
}

func (h *YandasHandler) ApplicationStatus(c *gin.Context) {
	profile, err := h.svcs.Yandas.GetApplicationStatus(getUserID(c))
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse("No application found"))
		return
	}
	c.JSON(http.StatusOK, SuccessResponse(profile))
}

func (h *YandasHandler) UpdateProfile(c *gin.Context) {
	var input services.UpdateYandasProfileInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}
	profile, err := h.svcs.Yandas.UpdateProfile(getUserID(c), &input)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, SuccessResponse(profile))
}

func (h *YandasHandler) UpdateAvailability(c *gin.Context) {
	var input struct {
		Available bool `json:"available"`
	}
	c.ShouldBindJSON(&input)
	h.svcs.Yandas.UpdateAvailability(getUserID(c), input.Available)
	c.JSON(http.StatusOK, SuccessResponse(gin.H{"available": input.Available}))
}

func (h *YandasHandler) UpdateLocation(c *gin.Context) {
	var input struct {
		Lat float64 `json:"latitude"`
		Lng float64 `json:"longitude"`
	}
	c.ShouldBindJSON(&input)
	h.svcs.Yandas.UpdateLocation(getUserID(c), input.Lat, input.Lng)
	c.JSON(http.StatusOK, SuccessResponse(gin.H{"message": "Location updated"}))
}

func (h *YandasHandler) CreateService(c *gin.Context) {
	var input services.ServiceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}
	svc, err := h.svcs.Yandas.CreateService(getUserID(c), &input)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusCreated, SuccessResponse(svc))
}

func (h *YandasHandler) UpdateService(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	var input services.ServiceInput
	c.ShouldBindJSON(&input)
	svc, err := h.svcs.Yandas.UpdateService(getUserID(c), id, &input)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, SuccessResponse(svc))
}

func (h *YandasHandler) DeleteService(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	h.svcs.Yandas.DeleteService(getUserID(c), id)
	c.JSON(http.StatusOK, SuccessResponse(gin.H{"message": "Deleted"}))
}

func (h *YandasHandler) GetMyServices(c *gin.Context) {
	userID := getUserID(c)
	profile, err := h.svcs.Yandas.GetApplicationStatus(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse("Yandaş profile not found"))
		return
	}
	services, err := h.svcs.Yandas.GetServices(profile.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, SuccessResponse(services))
}

func (h *YandasHandler) GetOrders(c *gin.Context) {
	page, limit := getPagination(c)
	orders, total, _ := h.svcs.Yandas.GetOrders(getUserID(c), page, limit, c.Query("status"))
	c.JSON(http.StatusOK, SuccessResponseWithMeta(orders, PaginationMeta(page, limit, total)))
}

func (h *YandasHandler) AcceptOrder(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	if err := h.svcs.Yandas.AcceptOrder(getUserID(c), id); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, SuccessResponse(gin.H{"message": "Accepted"}))
}

func (h *YandasHandler) RejectOrder(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	var input struct {
		Reason string `json:"reason"`
	}
	c.ShouldBindJSON(&input)
	h.svcs.Yandas.RejectOrder(getUserID(c), id, input.Reason)
	c.JSON(http.StatusOK, SuccessResponse(gin.H{"message": "Rejected"}))
}

func (h *YandasHandler) StartOrder(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	h.svcs.Yandas.StartOrder(getUserID(c), id)
	c.JSON(http.StatusOK, SuccessResponse(gin.H{"message": "Started"}))
}

func (h *YandasHandler) CompleteOrder(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	var input struct {
		Notes string `json:"notes"`
	}
	c.ShouldBindJSON(&input)
	h.svcs.Yandas.CompleteOrder(getUserID(c), id, input.Notes)
	c.JSON(http.StatusOK, SuccessResponse(gin.H{"message": "Completed"}))
}

func (h *YandasHandler) GetStats(c *gin.Context) {
	stats, _ := h.svcs.Yandas.GetStats(getUserID(c))
	c.JSON(http.StatusOK, SuccessResponse(stats))
}
