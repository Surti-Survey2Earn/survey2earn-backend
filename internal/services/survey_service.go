// internal/service/survey_service.go
package service

import (
	"errors"
	"time"
	"survey2earn-backend/internal/models"
	"survey2earn-backend/internal/dto"
	"survey2earn-backend/internal/repository"
	"gorm.io/gorm"
)

type SurveyService interface {
	CreateSurvey(userID uint, req *dto.CreateSurveyRequest) (*dto.SurveyResponse, error)
	UpdateSurvey(userID, surveyID uint, req *dto.UpdateSurveyRequest) (*dto.SurveyResponse, error)
	PublishSurvey(userID, surveyID uint, req *dto.PublishSurveyRequest) (*dto.SurveyResponse, error)
	GetSurvey(surveyID uint) (*dto.SurveyResponse, error)
	GetUserSurveys(userID uint, status string, page, limit int) (*dto.SurveyListResponse, error)
	GetPublicSurveys(page, limit int, category, status string) (*dto.SurveyListResponse, error)
	DeleteSurvey(userID, surveyID uint) error
	GetSurveyAnalytics(userID, surveyID uint) (*dto.SurveyAnalyticsResponse, error)
}

type surveyService struct {
	surveyRepo   repository.SurveyRepository
	userRepo     repository.UserRepository
	rewardRepo   repository.RewardRepository
}

func NewSurveyService(
	surveyRepo repository.SurveyRepository,
	userRepo repository.UserRepository,
	rewardRepo repository.RewardRepository,
) SurveyService {
	return &surveyService{
		surveyRepo: surveyRepo,
		userRepo:   userRepo,
		rewardRepo: rewardRepo,
	}
}

func (s *surveyService) CreateSurvey(userID uint, req *dto.CreateSurveyRequest) (*dto.SurveyResponse, error) {
	// Validate user exists
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Parse estimated time to minutes
	estimatedMinutes := s.parseEstimatedTime(req.EstimatedTime)

	// Calculate total reward pool
	totalRewardPool := req.RewardAmount * float64(req.MaxParticipants)

	// Create survey model
	survey := &models.Survey{
		Creator: dto.UserResponse{
			ID:              survey.Creator.ID,
			WalletAddress:   survey.Creator.WalletAddress,
			Username:        survey.Creator.Username,
			ReputationScore: survey.Creator.ReputationScore,
		},
		Progress: progress,
	}
}ID:         userID,
		Title:             req.Title,
		Description:       req.Description,
		Category:          req.Category,
		Status:            models.SurveyStatusDraft,
		MaxResponses:      req.MaxParticipants,
		RewardPerResponse: req.RewardAmount,
		TotalRewardPool:   totalRewardPool,
		EstimatedDuration: estimatedMinutes,
		IsAnonymous:       req.IsAnonymous,
		IsPublic:          req.IsPublic,
		RequireLogin:      req.RequireLogin,
		AllowMultiple:     req.AllowMultiple,
		StartDate:         req.StartDate,
		EndDate:           req.EndDate,
	}

	// Create questions
	questions := make([]models.Question, len(req.Questions))
	for i, q := range req.Questions {
		options := make(models.QuestionOptions, len(q.Options))
		for j, opt := range q.Options {
			options[j] = models.QuestionOption{
				ID:    opt.ID,
				Label: opt.Label,
				Value: opt.Value,
				Order: opt.Order,
			}
		}

		questions[i] = models.Question{
			Type:        models.QuestionType(q.Type),
			Text:        q.Title,
			Description: q.Description,
			Options:     options,
			Required:    q.Required,
			Order:       q.Order,
			MinLength:   q.MinLength,
			MaxLength:   q.MaxLength,
			MinValue:    q.MinValue,
			MaxValue:    q.MaxValue,
		}
	}

	survey.Questions = questions

	// Save survey
	if err := s.surveyRepo.Create(survey); err != nil {
		return nil, err
	}

	// Convert to response DTO
	return s.surveyToDTO(survey), nil
}

func (s *surveyService) UpdateSurvey(userID, surveyID uint, req *dto.UpdateSurveyRequest) (*dto.SurveyResponse, error) {
	// Get survey
	survey, err := s.surveyRepo.GetByID(surveyID)
	if err != nil {
		return nil, err
	}

	// Check ownership
	if survey.CreatorID != userID {
		return nil, errors.New("unauthorized")
	}

	// Check if survey can be edited
	if !survey.CanBeEdited() {
		return nil, errors.New("survey cannot be edited after publishing")
	}

	// Update fields
	if req.Title != nil {
		survey.Title = *req.Title
	}
	if req.Description != nil {
		survey.Description = *req.Description
	}
	if req.Category != nil {
		survey.Category = *req.Category
	}
	if req.EstimatedTime != nil {
		survey.EstimatedDuration = s.parseEstimatedTime(*req.EstimatedTime)
	}
	if req.RewardAmount != nil {
		survey.RewardPerResponse = *req.RewardAmount
		survey.TotalRewardPool = *req.RewardAmount * float64(survey.MaxResponses)
	}
	if req.MaxParticipants != nil {
		survey.MaxResponses = *req.MaxParticipants
		survey.TotalRewardPool = survey.RewardPerResponse * float64(*req.MaxParticipants)
	}
	if req.IsAnonymous != nil {
		survey.IsAnonymous = *req.IsAnonymous
	}
	if req.IsPublic != nil {
		survey.IsPublic = *req.IsPublic
	}
	if req.RequireLogin != nil {
		survey.RequireLogin = *req.RequireLogin
	}
	if req.AllowMultiple != nil {
		survey.AllowMultiple = *req.AllowMultiple
	}

	// Update questions if provided
	if req.Questions != nil {
		// Delete existing questions
		if err := s.surveyRepo.DeleteQuestions(surveyID); err != nil {
			return nil, err
		}

		// Create new questions
		questions := make([]models.Question, len(req.Questions))
		for i, q := range req.Questions {
			options := make(models.QuestionOptions, len(q.Options))
			for j, opt := range q.Options {
				options[j] = models.QuestionOption{
					ID:    opt.ID,
					Label: opt.Label,
					Value: opt.Value,
					Order: opt.Order,
				}
			}

			questions[i] = models.Question{
				SurveyID:    surveyID,
				Type:        models.QuestionType(q.Type),
				Text:        q.Title,
				Description: q.Description,
				Options:     options,
				Required:    q.Required,
				Order:       q.Order,
				MinLength:   q.MinLength,
				MaxLength:   q.MaxLength,
				MinValue:    q.MinValue,
				MaxValue:    q.MaxValue,
			}
		}

		survey.Questions = questions
	}

	// Save survey
	if err := s.surveyRepo.Update(survey); err != nil {
		return nil, err
	}

	return s.surveyToDTO(survey), nil
}

func (s *surveyService) PublishSurvey(userID, surveyID uint, req *dto.PublishSurveyRequest) (*dto.SurveyResponse, error) {
	// Get survey
	survey, err := s.surveyRepo.GetByID(surveyID)
	if err != nil {
		return nil, err
	}

	// Check ownership
	if survey.CreatorID != userID {
		return nil, errors.New("unauthorized")
	}

	// Check if survey can be published
	if survey.Status != models.SurveyStatusDraft {
		return nil, errors.New("only draft surveys can be published")
	}

	// Validate survey has questions
	if len(survey.Questions) == 0 {
		return nil, errors.New("survey must have at least one question")
	}

	// Update survey status and dates
	survey.Status = models.SurveyStatusPublished
	if req.StartDate != nil {
		survey.StartDate = req.StartDate
	}
	if req.EndDate != nil {
		survey.EndDate = req.EndDate
	}

	// Create reward pool
	rewardPool := &models.RewardPool{
		SurveyID:          surveyID,
		TotalAmount:       survey.TotalRewardPool,
		RewardPerResponse: survey.RewardPerResponse,
		MaxResponses:      survey.MaxResponses,
		RemainingAmount:   survey.TotalRewardPool,
		IsActive:          true,
	}

	// Save in transaction
	err = s.surveyRepo.PublishWithRewardPool(survey, rewardPool)
	if err != nil {
		return nil, err
	}

	return s.surveyToDTO(survey), nil
}

func (s *surveyService) GetSurvey(surveyID uint) (*dto.SurveyResponse, error) {
	survey, err := s.surveyRepo.GetByID(surveyID)
	if err != nil {
		return nil, err
	}

	return s.surveyToDTO(survey), nil
}

func (s *surveyService) GetUserSurveys(userID uint, status string, page, limit int) (*dto.SurveyListResponse, error) {
	surveys, total, err := s.surveyRepo.GetByUserID(userID, status, page, limit)
	if err != nil {
		return nil, err
	}

	items := make([]dto.SurveyItemResponse, len(surveys))
	for i, survey := range surveys {
		items[i] = s.surveyToItemDTO(&survey)
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return &dto.SurveyListResponse{
		Surveys:    items,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

func (s *surveyService) GetPublicSurveys(page, limit int, category, status string) (*dto.SurveyListResponse, error) {
	surveys, total, err := s.surveyRepo.GetPublicSurveys(page, limit, category, status)
	if err != nil {
		return nil, err
	}

	items := make([]dto.SurveyItemResponse, len(surveys))
	for i, survey := range surveys {
		items[i] = s.surveyToItemDTO(&survey)
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return &dto.SurveyListResponse{
		Surveys:    items,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

func (s *surveyService) DeleteSurvey(userID, surveyID uint) error {
	// Get survey
	survey, err := s.surveyRepo.GetByID(surveyID)
	if err != nil {
		return err
	}

	// Check ownership
	if survey.CreatorID != userID {
		return errors.New("unauthorized")
	}

	// Check if survey can be deleted
	if survey.Status != models.SurveyStatusDraft {
		return errors.New("only draft surveys can be deleted")
	}

	return s.surveyRepo.Delete(surveyID)
}

func (s *surveyService) GetSurveyAnalytics(userID, surveyID uint) (*dto.SurveyAnalyticsResponse, error) {
	// Implementation for analytics
	// This would include response statistics, demographics, etc.
	return nil, errors.New("not implemented")
}

// Helper methods

func (s *surveyService) parseEstimatedTime(timeStr string) int {
	// Parse time strings like "5-10 min", "15+ min" to minutes
	switch timeStr {
	case "1-3 min":
		return 3
	case "3-5 min":
		return 5
	case "5-10 min":
		return 10
	case "10-15 min":
		return 15
	case "15+ min":
		return 20
	default:
		return 10
	}
}

func (s *surveyService) surveyToDTO(survey *models.Survey) *dto.SurveyResponse {
	questions := make([]dto.QuestionResponse, len(survey.Questions))
	for i, q := range survey.Questions {
		options := make([]dto.QuestionOptionResponse, len(q.Options))
		for j, opt := range q.Options {
			options[j] = dto.QuestionOptionResponse{
				ID:    opt.ID,
				Label: opt.Label,
				Value: opt.Value,
				Order: opt.Order,
			}
		}

		questions[i] = dto.QuestionResponse{
			ID:          q.ID,
			Type:        string(q.Type),
			Text:        q.Text,
			Description: q.Description,
			Required:    q.Required,
			Order:       q.Order,
			Options:     options,
			MinLength:   q.MinLength,
			MaxLength:   q.MaxLength,
			MinValue:    q.MinValue,
			MaxValue:    q.MaxValue,
		}
	}

	return &dto.SurveyResponse{
		ID:                survey.ID,
		CreatorID:         survey.CreatorID,
		Title:             survey.Title,
		Description:       survey.Description,
		Category:          survey.Category,
		Status:            string(survey.Status),
		MaxResponses:      survey.MaxResponses,
		RewardPerResponse: survey.RewardPerResponse,
		TotalRewardPool:   survey.TotalRewardPool,
		EstimatedDuration: survey.EstimatedDuration,
		ResponseCount:     survey.ResponseCount,
		CompletionRate:    survey.CompletionRate,
		AverageRating:     survey.AverageRating,
		IsAnonymous:       survey.IsAnonymous,
		IsPublic:          survey.IsPublic,
		RequireLogin:      survey.RequireLogin,
		AllowMultiple:     survey.AllowMultiple,
		StartDate:         survey.StartDate,
		EndDate:           survey.EndDate,
		CreatedAt:         survey.CreatedAt,
		UpdatedAt:         survey.UpdatedAt,
		Questions:         questions,
		Creator: dto.UserResponse{
			ID:              survey.Creator.ID,
			WalletAddress:   survey.Creator.WalletAddress,
			Username:        survey.Creator.Username,
			ReputationScore: survey.Creator.ReputationScore,
		},
	}
}

func (s *surveyService) surveyToItemDTO(survey *models.Survey) dto.SurveyItemResponse {
	progress := float64(survey.ResponseCount) / float64(survey.MaxResponses) * 100
	if progress > 100 {
		progress = 100
	}

	return dto.SurveyItemResponse{
		ID:                survey.ID,
		Title:             survey.Title,
		Description:       survey.Description,
		Category:          survey.Category,
		Status:            string(survey.Status),
		RewardPerResponse: survey.RewardPerResponse,
		XpReward:          survey.EstimatedDuration * 10, // Mock XP calculation
		EstimatedDuration: survey.EstimatedDuration,
		ResponseCount:     survey.ResponseCount,
		MaxResponses:      survey.MaxResponses,
		CompletionRate:    survey.CompletionRate,
		AverageRating:     survey.AverageRating,
		CreatedAt:         survey.CreatedAt,
	}
}