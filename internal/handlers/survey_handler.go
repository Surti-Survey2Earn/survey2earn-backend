// internal/handler/survey_handler.go
package handler

import (
	"net/http"
	"strconv"
	"survey2earn-backend/internal/dto"
	"survey2earn-backend/internal/service"
	"survey2earn-backend/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type SurveyHandler struct {
	surveyService service.SurveyService
}

func NewSurveyHandler(surveyService service.SurveyService) *SurveyHandler {
	return &SurveyHandler{
		surveyService: surveyService,
	}
}

// CreateSurvey godoc
// @Summary Create a new survey
// @Description Create a new survey as a draft
// @Tags surveys
// @Accept json
// @Produce json
// @Param survey body dto.CreateSurveyRequest true "Survey data"
// @Success 201 {object} dto.SurveyResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /surveys [post]
func (h *SurveyHandler) CreateSurvey(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "User authentication required",
		})
		return
	}

	var req dto.CreateSurveyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logrus.WithError(err).Error("Invalid survey creation request")
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	survey, err := h.surveyService.CreateSurvey(userID, &req)
	if err != nil {
		logrus.WithError(err).Error("Failed to create survey")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "creation_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, SuccessResponse{
		Success: true,
		Data:    survey,
		Message: "Survey created successfully",
	})
}

// UpdateSurvey godoc
// @Summary Update a survey
// @Description Update a draft survey
// @Tags surveys
// @Accept json
// @Produce json
// @Param id path int true "Survey ID"
// @Param survey body dto.UpdateSurveyRequest true "Survey update data"
// @Success 200 {object} dto.SurveyResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /surveys/{id} [put]
func (h *SurveyHandler) UpdateSurvey(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "User authentication required",
		})
		return
	}

	surveyID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid survey ID",
		})
		return
	}

	var req dto.UpdateSurveyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logrus.WithError(err).Error("Invalid survey update request")
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	survey, err := h.surveyService.UpdateSurvey(userID, uint(surveyID), &req)
	if err != nil {
		logrus.WithError(err).Error("Failed to update survey")
		if err.Error() == "unauthorized" {
			c.JSON(http.StatusForbidden, ErrorResponse{
				Error:   "forbidden",
				Message: "You don't have permission to update this survey",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "update_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    survey,
		Message: "Survey updated successfully",
	})
}

// PublishSurvey godoc
// @Summary Publish a survey
// @Description Publish a draft survey to make it available for responses
// @Tags surveys
// @Accept json
// @Produce json
// @Param id path int true "Survey ID"
// @Param publish body dto.PublishSurveyRequest true "Publish data"
// @Success 200 {object} dto.SurveyResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /surveys/{id}/publish [post]
func (h *SurveyHandler) PublishSurvey(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "User authentication required",
		})
		return
	}

	surveyID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid survey ID",
		})
		return
	}

	var req dto.PublishSurveyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logrus.WithError(err).Error("Invalid publish request")
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	survey, err := h.surveyService.PublishSurvey(userID, uint(surveyID), &req)
	if err != nil {
		logrus.WithError(err).Error("Failed to publish survey")
		if err.Error() == "unauthorized" {
			c.JSON(http.StatusForbidden, ErrorResponse{
				Error:   "forbidden",
				Message: "You don't have permission to publish this survey",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "publish_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    survey,
		Message: "Survey published successfully",
	})
}

// GetSurvey godoc
// @Summary Get a survey
// @Description Get survey details by ID
// @Tags surveys
// @Accept json
// @Produce json
// @Param id path int true "Survey ID"
// @Success 200 {object} dto.SurveyResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /surveys/{id} [get]
func (h *SurveyHandler) GetSurvey(c *gin.Context) {
	surveyID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid survey ID",
		})
		return
	}

	survey, err := h.surveyService.GetSurvey(uint(surveyID))
	if err != nil {
		logrus.WithError(err).Error("Failed to get survey")
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "not_found",
			Message: "Survey not found",
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    survey,
	})
}

// GetUserSurveys godoc
// @Summary Get user's surveys
// @Description Get surveys created by the authenticated user
// @Tags surveys
// @Accept json
// @Produce json
// @Param status query string false "Survey status filter"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} dto.SurveyListResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /surveys/my [get]
func (h *SurveyHandler) GetUserSurveys(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "User authentication required",
		})
		return
	}

	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	surveys, err := h.surveyService.GetUserSurveys(userID, status, page, limit)
	if err != nil {
		logrus.WithError(err).Error("Failed to get user surveys")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "fetch_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    surveys,
	})
}

// GetPublicSurveys godoc
// @Summary Get public surveys
// @Description Get list of public surveys available for participation
// @Tags surveys
// @Accept json
// @Produce json
// @Param category query string false "Category filter"
// @Param status query string false "Status filter"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} dto.SurveyListResponse
// @Failure 500 {object} ErrorResponse
// @Router /surveys [get]
func (h *SurveyHandler) GetPublicSurveys(c *gin.Context) {
	category := c.Query("category")
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	surveys, err := h.surveyService.GetPublicSurveys(page, limit, category, status)
	if err != nil {
		logrus.WithError(err).Error("Failed to get public surveys")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "fetch_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    surveys,
	})
}

// DeleteSurvey godoc
// @Summary Delete a survey
// @Description Delete a draft survey
// @Tags surveys
// @Accept json
// @Produce json
// @Param id path int true "Survey ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /surveys/{id} [delete]
func (h *SurveyHandler) DeleteSurvey(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "User authentication required",
		})
		return
	}

	surveyID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid survey ID",
		})
		return
	}

	err = h.surveyService.DeleteSurvey(userID, uint(surveyID))
	if err != nil {
		logrus.WithError(err).Error("Failed to delete survey")
		if err.Error() == "unauthorized" {
			c.JSON(http.StatusForbidden, ErrorResponse{
				Error:   "forbidden",
				Message: "You don't have permission to delete this survey",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "delete_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Survey deleted successfully",
	})
}

// Common response structures
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}