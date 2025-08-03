# Survey2Earn Backend API Documentation

## Overview

Backend API untuk Survey2Earn platform yang memungkinkan pengguna membuat survey dan mendapatkan reward berupa token S2E dan XP untuk berpartisipasi dalam survey.

## Features

- ✅ **Survey Management**: Create, update, publish, dan delete surveys
- ✅ **Response System**: Start, submit answers, dan complete surveys  
- ✅ **Authentication**: Wallet-based authentication dengan JWT
- ✅ **Reward System**: Automatic token dan XP distribution
- ✅ **Real-time Progress**: Track survey progress dan time limits
- ✅ **Quality Control**: Answer validation dan quality scoring
- ✅ **Analytics**: Survey statistics dan user performance metrics

## Tech Stack

- **Backend**: Go (Gin framework)
- **Database**: PostgreSQL dengan GORM
- **Authentication**: JWT dengan wallet signature
- **API Design**: RESTful dengan JSON responses

## Quick Start

### Prerequisites

- Go 1.21+
- PostgreSQL 14+
- Environment variables configured

### Installation

```bash
# Clone repository
git clone <repository-url>
cd survey2earn-backend

# Install dependencies
go mod download

# Setup environment variables
cp .env.example .env
# Edit .env with your configuration

# Run migrations
go run cmd/main.go
```

### Environment Variables

```env
# Server Configuration
PORT=8080
ENV=development
API_VERSION=v1

# Database Configuration  
DB_HOST=localhost
DB_PORT=5432
DB_USER=survey2earn
DB_PASSWORD=your_password
DB_NAME=survey2earn_db
DB_SSL_MODE=disable

# JWT Configuration
JWT_SECRET=your-secret-key
JWT_EXPIRATION_HOURS=24

# CORS Configuration
ALLOWED_ORIGINS=http://localhost:3000,http://localhost:3001
ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS
ALLOWED_HEADERS=Content-Type,Authorization

# Logging
LOG_LEVEL=info
LOG_FORMAT=json
```

## API Endpoints

### Base URL
```
http://localhost:8080/api/v1
```

### Authentication

#### Register User
```http
POST /auth/register
Content-Type: application/json

{
  "wallet_address": "0x1234567890123456789012345678901234567890"
}
```

#### Login
```http
POST /auth/login
Content-Type: application/json

{
  "wallet_address": "0x1234567890123456789012345678901234567890",
  "signature": "0x...",
  "message": "Login to Survey2Earn"
}
```

#### Get Profile
```http
GET /user/profile
Authorization: Bearer <token>
```

### Survey Management

#### Create Survey
```http
POST /surveys
Authorization: Bearer <token>
Content-Type: application/json

{
  "title": "DeFi User Experience Research",
  "description": "Help us understand how users interact with DeFi protocols",
  "category": "DeFi",
  "estimatedTime": "5-10 min",
  "rewardAmount": 50.0,
  "maxParticipants": 100,
  "xpReward": 150,
  "questions": [
    {
      "type": "single_choice",
      "title": "How often do you use DeFi protocols?",
      "description": "Select the option that best describes your usage",
      "required": true,
      "options": [
        {
          "id": "opt1",
          "label": "Daily",
          "value": "daily",
          "order": 1
        },
        {
          "id": "opt2", 
          "label": "Weekly",
          "value": "weekly",
          "order": 2
        }
      ],
      "order": 1
    },
    {
      "type": "text",
      "title": "What challenges do you face with DeFi?",
      "required": true,
      "maxLength": 500,
      "order": 2
    },
    {
      "type": "rating",
      "title": "Rate your overall DeFi experience",
      "required": true,
      "minValue": 1,
      "maxValue": 5,
      "order": 3
    }
  ],
  "isAnonymous": true,
  "isPublic": true,
  "requireLogin": true,
  "allowMultiple": false
}
```

#### Get Public Surveys
```http
GET /surveys?page=1&limit=10&category=DeFi&status=published
```

#### Get User's Surveys
```http
GET /surveys/my?status=draft&page=1&limit=10
Authorization: Bearer <token>
```

#### Update Survey (Draft only)
```http
PUT /surveys/{id}
Authorization: Bearer <token>
Content-Type: application/json

{
  "title": "Updated Survey Title",
  "rewardAmount": 75.0
}
```

#### Publish Survey
```http
POST /surveys/{id}/publish
Authorization: Bearer <token>
Content-Type: application/json

{
  "startDate": "2024-01-01T00:00:00Z",
  "endDate": "2024-12-31T23:59:59Z"
}
```

### Survey Responses

#### Start Survey
```http
POST /responses/start
Authorization: Bearer <token>
Content-Type: application/json

{
  "survey_id": 1,
  "timezone": "Asia/Jakarta",
  "language": "en"
}
```

Response:
```json
{
  "success": true,
  "data": {
    "response_id": 123,
    "survey_id": 1,
    "status": "started",
    "started_at": "2024-01-15T10:00:00Z",
    "time_left": 600
  },
  "message": "Survey started successfully"
}
```

#### Submit Answers
```http
POST /responses/{response_id}/answers
Authorization: Bearer <token>
Content-Type: application/json

[
  {
    "question_id": 1,
    "answer": {
      "type": "single_choice",
      "options": ["daily"]
    },
    "time_spent": 15,
    "is_skipped": false
  },
  {
    "question_id": 2,
    "answer": {
      "type": "text",
      "value": "High gas fees and complex interfaces"
    },
    "time_spent": 45,
    "is_skipped": false
  },
  {
    "question_id": 3,
    "answer": {
      "type": "rating",
      "rating": 4
    },
    "time_spent": 10,
    "is_skipped": false
  }
]
```

#### Complete Survey
```http
POST /responses/complete
Authorization: Bearer <token>
Content-Type: application/json

{
  "response_id": 123,
  "answers": [
    // Optional: final answers if not submitted yet
  ],
  "duration": 180
}
```

Response:
```json
{
  "success": true,
  "data": {
    "response_id": 123,
    "status": "completed",
    "completed_at": "2024-01-15T10:03:00Z",
    "duration": 180,
    "reward_earned": 50.0,
    "xp_earned": 150,
    "nft_certificate": "NFT-CERT-1-1-123",
    "transaction_hash": null,
    "message": "Survey completed successfully! Your rewards will be processed shortly."
  }
}
```

#### Get Response Progress
```http
GET /responses/{response_id}/progress
Authorization: Bearer <token>
```

Response:
```json
{
  "success": true,
  "data": {
    "response_id": 123,
    "survey_id": 1,
    "status": "started",
    "progress": 66.67,
    "questions_total": 3,
    "questions_answered": 2,
    "time_spent": 120,
    "time_left": 480,
    "started_at": "2024-01-15T10:00:00Z",
    "last_answered_at": "2024-01-15T10:02:00Z"
  }
}
```

#### Get User Responses
```http
GET /responses?status=completed&page=1&limit=10
Authorization: Bearer <token>
```

#### Update Single Answer
```http
PUT /responses/{response_id}/questions/{question_id}
Authorization: Bearer <token>
Content-Type: application/json

{
  "answer": {
    "type": "rating",
    "rating": 5
  },
  "time_spent": 20,
  "is_skipped": false
}
```

#### Abandon Survey
```http
POST /responses/{response_id}/abandon
Authorization: Bearer <token>
```

## Response Format

### Success Response
```json
{
  "success": true,
  "data": {...},
  "message": "Operation successful"
}
```

### Error Response
```json
{
  "error": "error_code",
  "message": "Human readable error message"
}
```

## Question Types

### Text Input
```json
{
  "type": "text",
  "title": "What is your opinion?",
  "maxLength": 500,
  "minLength": 10
}
```

### Single Choice
```json
{
  "type": "single_choice", 
  "title": "Choose one option",
  "options": [
    {"id": "opt1", "label": "Option 1", "value": "option1", "order": 1},
    {"id": "opt2", "label": "Option 2", "value": "option2", "order": 2}
  ]
}
```

### Multiple Choice
```json
{
  "type": "multiple_choice",
  "title": "Select all that apply",
  "options": [...]
}
```

### Rating
```json
{
  "type": "rating",
  "title": "Rate this experience",
  "minValue": 1,
  "maxValue": 5
}
```

### Scale
```json
{
  "type": "scale", 
  "title": "How likely are you to recommend?",
  "minValue": 1,
  "maxValue": 10
}
```

### Date
```json
{
  "type": "date",
  "title": "When did this happen?"
}
```

## Answer Format

### Text Answer
```json
{
  "type": "text",
  "value": "User's text response"
}
```

### Choice Answer
```json
{
  "type": "single_choice",
  "options": ["selected_option_value"]
}

{
  "type": "multiple_choice", 
  "options": ["option1", "option2", "option3"]
}
```

### Rating/Scale Answer
```json
{
  "type": "rating",
  "rating": 4
}

{
  "type": "scale",
  "scale": 8
}
```

### Date Answer
```json
{
  "type": "date",
  "date": "2024-01-15T00:00:00Z"
}
```

## Error Codes

| Code | Description |
|------|-------------|
| `unauthorized` | Missing or invalid authentication |
| `forbidden` | Insufficient permissions |
| `invalid_request` | Malformed request data |
| `not_found` | Resource not found |
| `creation_failed` | Failed to create resource |
| `update_failed` | Failed to update resource |
| `delete_failed` | Failed to delete resource |
| `start_failed` | Failed to start survey |
| `submit_failed` | Failed to submit answers |
| `completion_failed` | Failed to complete survey |

## Status Codes

### Survey Status
- `draft` - Survey is being created/edited
- `published` - Survey is live and accepting responses
- `paused` - Survey is temporarily paused
- `completed` - Survey has reached max responses or end date
- `cancelled` - Survey was cancelled

### Response Status  
- `started` - User has started the survey
- `completed` - User has completed the survey
- `abandoned` - User abandoned the survey

### Transaction Status
- `pending` - Transaction is waiting to be processed
- `processing` - Transaction is being processed
- `completed` - Transaction completed successfully
- `failed` - Transaction failed
- `cancelled` - Transaction was cancelled

## Rate Limiting

- **General API**: 60 requests per minute per IP
- **Authentication**: 10 requests per minute per IP
- **Survey Creation**: 5 requests per minute per user

## Security Features

- JWT-based authentication with wallet signatures
- Request validation and sanitization
- Rate limiting per endpoint
- CORS protection
- SQL injection prevention via GORM
- Input validation for all endpoints

## Integration dengan Frontend

### Frontend Integration Example

```javascript
// Initialize API client
const API_BASE = 'http://localhost:8080/api/v1';
const token = localStorage.getItem('access_token');

// Create survey
async function createSurvey(surveyData) {
  const response = await fetch(`${API_BASE}/surveys`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    },
    body: JSON.stringify(surveyData)
  });
  
  return response.json();
}

// Start survey
async function startSurvey(surveyId) {
  const response = await fetch(`${API_BASE}/responses/start`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    },
    body: JSON.stringify({
      survey_id: surveyId,
      timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
      language: navigator.language
    })
  });
  
  return response.json();
}

// Submit answers
async function submitAnswers(responseId, answers) {
  const response = await fetch(`${API_BASE}/responses/${responseId}/answers`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json', 
      'Authorization': `Bearer ${token}`
    },
    body: JSON.stringify(answers)
  });
  
  return response.json();
}
```

## Database Schema

Refer to the models package for complete database schema:
- `User` - User accounts dan wallet addresses
- `Survey` - Survey definitions dan metadata
- `Question` - Survey questions dengan options
- `Response` - User survey responses
- `Answer` - Individual question answers  
- `RewardPool` - Survey reward pools
- `RewardTransaction` - Token reward transactions
- `UserBalance` - User token balances

## Future Enhancements

- [ ] Real blockchain integration (smart contracts)
- [ ] Advanced analytics dashboard
- [ ] Survey templates
- [ ] Collaboration features
- [ ] API documentation with Swagger
- [ ] WebSocket for real-time updates
- [ ] File upload support for questions
- [ ] Survey branching logic
- [ ] Multi-language support
- [ ] Email notifications