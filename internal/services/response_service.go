// internal/service/response_service.go
package service

import (
	"errors"
	"time"
	"survey2earn-backend/internal/models"
	"survey2earn-backend/internal/dto"
	"survey2earn-backend/internal/repository"
)

type ResponseService interface {
	StartSurvey(userID, surveyID uint, req *dto.StartSurveyRequest) (*dto.ResponseStartResponse, error)
	SubmitAnswers(userID uint, responseID uint, answers []dto.SubmitAnswerRequest) error
	CompleteSurvey(userID uint, req *dto.CompleteSurveyRequest) (*dto.CompletionResponse, error)
	GetResponse(userID, responseID uint) (*dto.SurveyResponseResponse, error)
	GetUserResponses(userID uint, req *dto.ListResponsesRequest) (*dto.ResponseListResponse, error)
	GetResponseProgress(userID, responseID uint) (*dto.SurveyProgressResponse, error)
	UpdateAnswer(userID, responseID, questionID uint, req *dto.UpdateAnswerRequest) error
	AbandonSurvey(userID, responseID uint) error
}

type responseService struct {
	responseRepo repository.ResponseRepository
	surveyRepo   repository.SurveyRepository
	rewardRepo   repository.RewardRepository
	userRepo     repository.UserRepository
}

func NewResponseService(
	responseRepo repository.ResponseRepository,
	surveyRepo repository.SurveyRepository,
	rewardRepo repository.RewardRepository,
	userRepo repository.UserRepository,
) ResponseService {
	return &responseService{
		responseRepo: responseRepo,
		surveyRepo:   surveyRepo,
		rewardRepo:   rewardRepo,
		userRepo:     userRepo,
	}
}

func (s *responseService) StartSurvey(userID, surveyID uint, req *dto.StartSurveyRequest) (*dto.ResponseStartResponse, error) {
	// Get survey
	survey, err := s.surveyRepo.GetByID(surveyID)
	if err != nil {
		return nil, err
	}

	// Check if survey is active
	if !survey.IsActive() {
		return nil, errors.New("survey is not active")
	}

	// Check if user can participate
	if survey.RequireLogin && userID == 0 {
		return nil, errors.New("login required to participate")
	}

	// Check if user already responded (if multiple responses not allowed)
	if !survey.AllowMultiple {
		exists, err := s.responseRepo.HasUserResponded(userID, surveyID)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.New("user has already responded to this survey")
		}
	}

	// Check if survey has reached max participants
	if survey.ResponseCount >= survey.MaxResponses {
		return nil, errors.New("survey has reached maximum participants")
	}

	// Create response
	response := &models.Response{
		SurveyID:  surveyID,
		UserID:    userID,
		Status:    models.ResponseStatusStarted,
		StartedAt: time.Now(),
		IPAddress: req.IPAddress,
		UserAgent: req.UserAgent,
		Timezone:  req.Timezone,
		Language:  req.Language,
		IsValid:   true,
	}

	if err := s.responseRepo.Create(response); err != nil {
		return nil, err
	}

	// Calculate time left (if survey has time limit)
	var timeLeft *int
	if survey.EstimatedDuration > 0 {
		timeLeftVal := survey.EstimatedDuration * 60 // convert to seconds
		timeLeft = &timeLeftVal
	}

	return &dto.ResponseStartResponse{
		ResponseID: response.ID,
		SurveyID:   surveyID,
		Status:     string(response.Status),
		StartedAt:  response.StartedAt,
		TimeLeft:   timeLeft,
	}, nil
}

func (s *responseService) SubmitAnswers(userID uint, responseID uint, answers []dto.SubmitAnswerRequest) error {
	// Get response
	response, err := s.responseRepo.GetByID(responseID)
	if err != nil {
		return err
	}

	// Check ownership
	if response.UserID != userID {
		return errors.New("unauthorized")
	}

	// Check if response is still active
	if response.Status != models.ResponseStatusStarted {
		return errors.New("response is not active")
	}

	// Get survey with questions
	survey, err := s.surveyRepo.GetByID(response.SurveyID)
	if err != nil {
		return err
	}

	// Process each answer
	for _, answerReq := range answers {
		// Find the question
		question, err := survey.GetQuestionByID(answerReq.QuestionID)
		if err != nil {
			continue // Skip invalid questions
		}

		// Convert DTO answer to model answer value
		answerValue := models.AnswerValue{
			Type:    answerReq.Answer.Type,
			Content: answerReq.Answer.Content,
			Options: answerReq.Answer.Options,
			Rating:  answerReq.Answer.Rating,
			Scale:   answerReq.Answer.Scale,
			Date:    answerReq.Answer.Date,
		}

		// Create or update answer
		answer := &models.Answer{
			ResponseID:  responseID,
			QuestionID:  answerReq.QuestionID,
			AnswerText:  s.extractAnswerText(answerValue),
			AnswerValue: answerValue,
			TimeSpent:   answerReq.TimeSpent,
			IsSkipped:   answerReq.IsSkipped,
		}

		// Validate answer
		if err := answer.ValidateAnswer(question); err != nil {
			return err
		}

		// Save or update answer
		if err := s.responseRepo.UpsertAnswer(answer); err != nil {
			return err
		}
	}

	return nil
}

func (s *responseService) CompleteSurvey(userID uint, req *dto.CompleteSurveyRequest) (*dto.CompletionResponse, error) {
	// Get response
	response, err := s.responseRepo.GetByID(req.ResponseID)
	if err != nil {
		return nil, err
	}

	// Check ownership
	if response.UserID != userID {
		return nil, errors.New("unauthorized")
	}

	// Check if response is still active
	if response.Status != models.ResponseStatusStarted {
		return nil, errors.New("response is not active")
	}

	// Submit final answers if provided
	if len(req.Answers) > 0 {
		if err := s.SubmitAnswers(userID, req.ResponseID, req.Answers); err != nil {
			return nil, err
		}
	}

	// Get survey
	survey, err := s.surveyRepo.GetByID(response.SurveyID)
	if err != nil {
		return nil, err
	}

	// Mark response as completed
	response.MarkAsCompleted()
	response.Duration = req.Duration

	// Calculate quality score
	response.QualityScore = s.calculateQualityScore(response, survey)

	// Update response
	if err := s.responseRepo.Update(response); err != nil {
		return nil, err
	}

	// Process rewards
	rewardAmount, xpEarned, err := s.processRewards(response, survey)
	if err != nil {
		return nil, err
	}

	// Update survey statistics
	if err := s.surveyRepo.UpdateStatistics(survey.ID); err != nil {
		return nil, err
	}

	// Generate NFT certificate (mock)
	nftCertificate := s.generateNFTCertificate(response, survey)

	return &dto.CompletionResponse{
		ResponseID:      response.ID,
		Status:          string(response.Status),
		CompletedAt:     *response.CompletedAt,
		Duration:        response.Duration,
		RewardEarned:    rewardAmount,
		XpEarned:        xpEarned,
		NFTCertificate:  &nftCertificate,
		TransactionHash: nil, // Will be updated when blockchain transaction is processed
		Message:         "Survey completed successfully! Your rewards will be processed shortly.",
	}, nil
}

func (s *responseService) GetResponse(userID, responseID uint) (*dto.SurveyResponseResponse, error) {
	response, err := s.responseRepo.GetWithAnswers(responseID)
	if err != nil {
		return nil, err
	}

	// Check ownership
	if response.UserID != userID {
		return nil, errors.New("unauthorized")
	}

	return s.responseToDTO(response), nil
}

func (s *responseService) GetUserResponses(userID uint, req *dto.ListResponsesRequest) (*dto.ResponseListResponse, error) {
	responses, total, err := s.responseRepo.GetByUserID(userID, req)
	if err != nil {
		return nil, err
	}

	items := make([]dto.ResponseItemResponse, len(responses))
	for i, response := range responses {
		items[i] = s.responseToItemDTO(&response)
	}

	totalPages := int(total) / req.Limit
	if int(total)%req.Limit > 0 {
		totalPages++
	}

	return &dto.ResponseListResponse{
		Responses:  items,
		Total:      total,
		Page:       req.Page,
		Limit:      req.Limit,
		TotalPages: totalPages,
	}, nil
}

func (s *responseService) GetResponseProgress(userID, responseID uint) (*dto.SurveyProgressResponse, error) {
	response, err := s.responseRepo.GetWithAnswers(responseID)
	if err != nil {
		return nil, err
	}

	// Check ownership
	if response.UserID != userID {
		return nil, errors.New("unauthorized")
	}

	// Get survey
	survey, err := s.surveyRepo.GetByID(response.SurveyID)
	if err != nil {
		return nil, err
	}

	// Calculate progress
	questionsTotal := len(survey.Questions)
	questionsAnswered := len(response.Answers)
	progress := float64(questionsAnswered) / float64(questionsTotal) * 100

	// Calculate time spent
	timeSpent := response.CalculateDuration()

	// Calculate time left
	var timeLeft *int
	if survey.EstimatedDuration > 0 {
		maxTime := survey.EstimatedDuration * 60 // minutes to seconds
		remaining := maxTime - timeSpent
		if remaining > 0 {
			timeLeft = &remaining
		}
	}

	// Get last answered time
	var lastAnsweredAt *time.Time
	if len(response.Answers) > 0 {
		lastAnsweredAt = &response.Answers[len(response.Answers)-1].UpdatedAt
	}

	return &dto.SurveyProgressResponse{
		ResponseID:        response.ID,
		SurveyID:          response.SurveyID,
		Status:            string(response.Status),
		Progress:          progress,
		QuestionsTotal:    questionsTotal,
		QuestionsAnswered: questionsAnswered,
		TimeSpent:         timeSpent,
		TimeLeft:          timeLeft,
		StartedAt:         response.StartedAt,
		LastAnsweredAt:    lastAnsweredAt,
	}, nil
}

func (s *responseService) UpdateAnswer(userID, responseID, questionID uint, req *dto.UpdateAnswerRequest) error {
	// Get response
	response, err := s.responseRepo.GetByID(responseID)
	if err != nil {
		return err
	}

	// Check ownership
	if response.UserID != userID {
		return errors.New("unauthorized")
	}

	// Check if response is still active
	if response.Status != models.ResponseStatusStarted {
		return errors.New("response is not active")
	}

	// Convert DTO answer to model answer value
	answerValue := models.AnswerValue{
		Type:    req.Answer.Type,
		Content: req.Answer.Content,
		Options: req.Answer.Options,
		Rating:  req.Answer.Rating,
		Scale:   req.Answer.Scale,
		Date:    req.Answer.Date,
	}

	// Update answer
	answer := &models.Answer{
		ResponseID:  responseID,
		QuestionID:  questionID,
		AnswerText:  s.extractAnswerText(answerValue),
		AnswerValue: answerValue,
		TimeSpent:   req.TimeSpent,
		IsSkipped:   req.IsSkipped,
	}

	return s.responseRepo.UpsertAnswer(answer)
}

func (s *responseService) AbandonSurvey(userID, responseID uint) error {
	// Get response
	response, err := s.responseRepo.GetByID(responseID)
	if err != nil {
		return err
	}

	// Check ownership
	if response.UserID != userID {
		return errors.New("unauthorized")
	}

	// Check if response can be abandoned
	if response.Status != models.ResponseStatusStarted {
		return errors.New("response cannot be abandoned")
	}

	// Mark as abandoned
	response.MarkAsAbandoned()

	return s.responseRepo.Update(response)
}

// Helper methods

func (s *responseService) extractAnswerText(answerValue models.AnswerValue) string {
	switch answerValue.Type {
	case "text":
		if str, ok := answerValue.Content.(string); ok {
			return str
		}
	case "number":
		if num, ok := answerValue.Content.(float64); ok {
			return fmt.Sprintf("%.2f", num)
		}
	case "boolean":
		if b, ok := answerValue.Content.(bool); ok {
			if b {
				return "true"
			}
			return "false"
		}
	case "array":
		if options := answerValue.Options; len(options) > 0 {
			return strings.Join(options, ", ")
		}
	case "rating":
		if answerValue.Rating != nil {
			return fmt.Sprintf("%d", *answerValue.Rating)
		}
	case "scale":
		if answerValue.Scale != nil {
			return fmt.Sprintf("%d", *answerValue.Scale)
		}
	case "date":
		if answerValue.Date != nil {
			return answerValue.Date.Format("2006-01-02")
		}
	}
	return ""
}

func (s *responseService) calculateQualityScore(response *models.Response, survey *models.Survey) float64 {
	// Simple quality score calculation
	// In a real implementation, this would be more sophisticated
	score := 5.0

	// Check completion rate
	questionsTotal := len(survey.Questions)
	questionsAnswered := len(response.Answers)
	completionRate := float64(questionsAnswered) / float64(questionsTotal)

	score *= completionRate

	// Check time spent (penalize too fast responses)
	avgTimePerQuestion := float64(response.Duration) / float64(questionsAnswered)
	if avgTimePerQuestion < 5 { // Less than 5 seconds per question
		score *= 0.7
	}

	// Check for skipped required questions
	skippedRequired := 0
	for _, answer := range response.Answers {
		if answer.IsSkipped {
			// Find if question was required
			for _, question := range survey.Questions {
				if question.ID == answer.QuestionID && question.Required {
					skippedRequired++
					break
				}
			}
		}
	}

	if skippedRequired > 0 {
		score *= (1.0 - float64(skippedRequired)*0.1)
	}

	if score < 0 {
		score = 0
	}
	if score > 5 {
		score = 5
	}

	return score
}

func (s *responseService) processRewards(response *models.Response, survey *models.Survey) (float64, int, error) {
	// Get reward pool
	pool, err := s.rewardRepo.GetPoolBySurveyID(survey.ID)
	if err != nil {
		return 0, 0, err
	}

	// Check if pool can process reward
	if !pool.CanProcessReward() {
		return 0, 0, errors.New("insufficient reward pool")
	}

	// Calculate rewards based on quality score
	baseReward := survey.RewardPerResponse
	qualityMultiplier := response.QualityScore / 5.0
	finalReward := baseReward * qualityMultiplier

	// Calculate XP (mock calculation)
	xpEarned := int(float64(survey.EstimatedDuration) * 10 * qualityMultiplier)

	// Create reward transaction
	transaction := &models.RewardTransaction{
		UserID:   response.UserID,
		SurveyID: survey.ID,
		ResponseID: &response.ID,
		PoolID:   &pool.ID,
		Type:     models.TransactionTypeReward,
		Amount:   finalReward,
		Status:   models.TransactionStatusPending,
	}

	// Process reward
	if err := s.rewardRepo.ProcessReward(pool, transaction); err != nil {
		return 0, 0, err
	}

	// Update user balance
	if err := s.userRepo.UpdateBalance(response.UserID, finalReward, float64(xpEarned)); err != nil {
		return 0, 0, err
	}

	return finalReward, xpEarned, nil
}

func (s *responseService) generateNFTCertificate(response *models.Response, survey *models.Survey) string {
	// Mock NFT certificate generation
	// In a real implementation, this would interact with NFT smart contract
	return fmt.Sprintf("NFT-CERT-%d-%d-%d", survey.ID, response.UserID, response.ID)
}

func (s *responseService) responseToDTO(response *models.Response) *dto.SurveyResponseResponse {
	answers := make([]dto.AnswerResponse, len(response.Answers))
	for i, answer := range response.Answers {
		answers[i] = dto.AnswerResponse{
			ID:         answer.ID,
			QuestionID: answer.QuestionID,
			Answer: dto.AnswerValue{
				Type:    answer.AnswerValue.Type,
				Content: answer.AnswerValue.Content,
				Options: answer.AnswerValue.Options,
				Rating:  answer.AnswerValue.Rating,
				Scale:   answer.AnswerValue.Scale,
				Date:    answer.AnswerValue.Date,
			},
			TimeSpent: answer.TimeSpent,
			IsSkipped: answer.IsSkipped,
			CreatedAt: answer.CreatedAt,
			UpdatedAt: answer.UpdatedAt,
		}
	}

	// Get reward earned (mock)
	rewardEarned := 0.0
	xpEarned := 0
	if response.Transaction != nil {
		rewardEarned = response.Transaction.Amount
		xpEarned = int(rewardEarned * 10) // Mock XP calculation
	}

	// Generate NFT certificate if completed
	var nftCertificate *string
	if response.IsCompleted() {
		cert := s.generateNFTCertificate(response, &response.Survey)
		nftCertificate = &cert
	}

	return &dto.SurveyResponseResponse{
		ID:             response.ID,
		SurveyID:       response.SurveyID,
		UserID:         response.UserID,
		Status:         string(response.Status),
		StartedAt:      response.StartedAt,
		CompletedAt:    response.CompletedAt,
		Duration:       response.Duration,
		QualityScore:   response.QualityScore,
		IsValid:        response.IsValid,
		Answers:        answers,
		RewardEarned:   rewardEarned,
		XpEarned:       xpEarned,
		NFTCertificate: nftCertificate,
	}
}

func (s *responseService) responseToItemDTO(response *models.Response) dto.ResponseItemResponse {
	// Calculate progress
	progress := 0.0
	if len(response.Survey.Questions) > 0 {
		progress = float64(len(response.Answers)) / float64(len(response.Survey.Questions)) * 100
	}

	// Get reward earned (mock)
	rewardEarned := 0.0
	xpEarned := 0
	if response.Transaction != nil {
		rewardEarned = response.Transaction.Amount
		xpEarned = int(rewardEarned * 10)
	}

	return dto.ResponseItemResponse{
		ID:           response.ID,
		SurveyID:     response.SurveyID,
		SurveyTitle:  response.Survey.Title,
		Status:       string(response.Status),
		StartedAt:    response.StartedAt,
		CompletedAt:  response.CompletedAt,
		Duration:     response.Duration,
		RewardEarned: rewardEarned,
		XpEarned:     xpEarned,
		QualityScore: response.QualityScore,
		Progress:     progress,
	}
}