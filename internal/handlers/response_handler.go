// internal/handler/response_handler.go
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

type ResponseHandler struct {
	responseService service.ResponseService
}

func NewResponseHandler(responseService service.ResponseService) *ResponseHandler {
	return &ResponseHandler{
		responseService: responseService,
	}
}

// StartSurvey godoc
// @Summary Start taking a survey
// @Description Start a new survey response session
// @Tags responses
// @Accept json
// @Produce json
// @Param survey body dto.StartSurveyRequest true "Start survey data"
// @Success 201 {object} dto.ResponseStartResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /responses/start [post]
func (h *ResponseHandler) StartSurvey(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "User authentication required",
		})
		return
	}

	var req dto.StartSurveyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logrus.WithError(err).Error("Invalid start survey request")
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	// Get IP address and User Agent from request
	req.IPAddress = c.ClientIP()
	req.UserAgent = c.GetHeader("User-Agent")

	response, err := h.responseService.StartSurvey(userID, req.SurveyID, &req)
	if err != nil {
		logrus.WithError(err).Error("Failed to start survey")
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "start_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, SuccessResponse{
		Success: true,
		Data:    response,
		Message: "Survey started successfully",
	})
}

// SubmitAnswers godoc
// @Summary Submit answers for a survey
// @Description Submit one or more answers for a survey response
// @Tags responses
// @Accept json
// @Produce json
// @Param id path int true "Response ID"
// @Param answers body []dto.SubmitAnswerRequest true "Answers data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /responses/{id}/answers [post]
func (h *ResponseHandler) SubmitAnswers(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "User authentication required",
		})
		return
	}

	responseID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid response ID",
		})
		return
	}

	var answers []dto.SubmitAnswerRequest
	if err := c.ShouldBindJSON(&answers); err != nil {
		logrus.WithError(err).Error("Invalid submit answers request")
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	err = h.responseService.SubmitAnswers(userID, uint(responseID), answers)
	if err != nil {
		logrus.WithError(err).Error("Failed to submit answers")
		if err.Error() == "unauthorized" {
			c.JSON(http.StatusForbidden, ErrorResponse{
				Error:   "forbidden",
				Message: "You don't have permission to modify this response",
			})
			return
		}
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "submit_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Answers submitted successfully",
	})
}

// CompleteSurvey godoc
// @Summary Complete a survey
// @Description Complete a survey and submit final answers
// @Tags responses
// @Accept json
// @Produce json
// @Param complete body dto.CompleteSurveyRequest true "Complete survey data"
// @Success 200 {object} dto.CompletionResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /responses/complete [post]
func (h *ResponseHandler) CompleteSurvey(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "User authentication required",
		})
		return
	}

	var req dto.CompleteSurveyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logrus.WithError(err).Error("Invalid complete survey request")
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	completion, err := h.responseService.CompleteSurvey(userID, &req)
	if err != nil {
		logrus.WithError(err).Error("Failed to complete survey")
		if err.Error() == "unauthorized" {
			c.JSON(http.StatusForbidden, ErrorResponse{
				Error:   "forbidden",
				Message: "You don't have permission to complete this response",
			})
			return
		}
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "completion_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    completion,
		Message: "Survey completed successfully",
	})
}

// GetResponse godoc
// @Summary Get a survey response
// @Description Get details of a specific survey response
// @Tags responses
// @Accept json
// @Produce json
// @Param id path int true "Response ID"
// @Success 200 {object} dto.SurveyResponseResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /responses/{id} [get]
func (h *ResponseHandler) GetResponse(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "User authentication required",
		})
		return
	}

	responseID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid response ID",
		})
		return
	}

	response, err := h.responseService.GetResponse(userID, uint(responseID))
	if err != nil {
		logrus.WithError(err).Error("Failed to get response")
		if err.Error() == "unauthorized" {
			c.JSON(http.StatusForbidden, ErrorResponse{
				Error:   "forbidden",
				Message: "You don't have permission to view this response",
			})
			return
		}
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "not_found",
			Message: "Response not found",
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    response,
	})
}

// GetUserResponses godoc
// @Summary Get user's survey responses
// @Description Get list of survey responses by the authenticated user
// @Tags responses
// @Accept json
// @Produce json
// @Param status query string false "Response status filter"
// @Param survey_id query int false "Survey ID filter"
// @Param start_date query string false "Start date filter (YYYY-MM-DD)"
// @Param end_date query string false "End date filter (YYYY-MM-DD)"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} dto.ResponseListResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /responses [get]
func (h *ResponseHandler) GetUserResponses(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "User authentication required",
		})
		return
	}

	var req dto.ListResponsesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	// Set defaults
	if req.Page == 0 {
		req.Page = 1
	}
	if req.Limit == 0 {
		req.Limit = 10
	}

	responses, err := h.responseService.GetUserResponses(userID, &req)
	if err != nil {
		logrus.WithError(err).Error("Failed to get user responses")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "fetch_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    responses,
	})
}

// GetResponseProgress godoc
// @Summary Get survey response progress
// @Description Get progress information for an ongoing survey response
// @Tags responses
// @Accept json
// @Produce json
// @Param id path int true "Response ID"
// @Success 200 {object} dto.SurveyProgressResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /responses/{id}/progress [get]
func (h *ResponseHandler) GetResponseProgress(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "User authentication required",
		})
		return
	}

	responseID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid response ID",
		})
		return
	}

	progress, err := h.responseService.GetResponseProgress(userID, uint(responseID))
	if err != nil {
		logrus.WithError(err).Error("Failed to get response progress")
		if err.Error() == "unauthorized" {
			c.JSON(http.StatusForbidden, ErrorResponse{
				Error:   "forbidden",
				Message: "You don't have permission to view this response",
			})
			return
		}
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "not_found",
			Message: "Response not found",
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    progress,
	})
}

// UpdateAnswer godoc
// @Summary Update a specific answer
// @Description Update an answer for a specific question in a survey response
// @Tags responses
// @Accept json
// @Produce json
// @Param response_id path int true "Response ID"
// @Param question_id path int true "Question ID"
// @Param answer body dto.UpdateAnswerRequest true "Updated answer data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /responses/{response_id}/questions/{question_id} [put]
func (h *ResponseHandler) UpdateAnswer(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "User authentication required",
		})
		return
	}

	responseID, err := strconv.ParseUint(c.Param("response_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid response ID",
		})
		return
	}

	questionID, err := strconv.ParseUint(c.Param("question_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid question ID",
		})
		return
	}

	var req dto.UpdateAnswerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logrus.WithError(err).Error("Invalid update answer request")
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	err = h.responseService.UpdateAnswer(userID, uint(responseID), uint(questionID), &req)
	if err != nil {
		logrus.WithError(err).Error("Failed to update answer")
		if err.Error() == "unauthorized" {
			c.JSON(http.StatusForbidden, ErrorResponse{
				Error:   "forbidden",
				Message: "You don't have permission to modify this response",
			})
			return
		}
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "update_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Answer updated successfully",
	})
}

// AbandonSurvey godoc
// @Summary Abandon a survey
// @Description Mark a survey response as abandoned
// @Tags responses
// @Accept json
// @Produce json
// @Param id path int true "Response ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /responses/{id}/abandon [post]
func (h *ResponseHandler) AbandonSurvey(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "User authentication required",
		})
		return
	}

	responseID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid response ID",
		})
		return
	}

	err = h.responseService.AbandonSurvey(userID, uint(responseID))
	if err != nil {
		logrus.WithError(err).Error("Failed to abandon survey")
		if err.Error() == "unauthorized" {
			c.JSON(http.StatusForbidden, ErrorResponse{
				Error:   "forbidden",
				Message: "You don't have permission to modify this response",
			})
			return
		}
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "abandon_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Survey response abandoned",
	})
}