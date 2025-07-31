package models

import (
	"errors"
	"time"
)

// TransactionStatus represents the status of a reward transaction
type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "pending"
	TransactionStatusProcessing TransactionStatus = "processing"
	TransactionStatusCompleted TransactionStatus = "completed"
	TransactionStatusFailed    TransactionStatus = "failed"
	TransactionStatusCancelled TransactionStatus = "cancelled"
)

// TransactionType represents the type of transaction
type TransactionType string

const (
	TransactionTypeReward     TransactionType = "reward"
	TransactionTypeWithdrawal TransactionType = "withdrawal"
	TransactionTypeRefund     TransactionType = "refund"
	TransactionTypeFee        TransactionType = "fee"
)

type RewardPool struct {
	BaseModel
	SurveyID          uint      `json:"survey_id" gorm:"unique;not null;index"`
	TotalAmount       float64   `json:"total_amount" gorm:"not null"`
	RewardPerResponse float64   `json:"reward_per_response" gorm:"not null"`
	MaxResponses      int       `json:"max_responses" gorm:"not null"`
	
	CurrentResponses  int       `json:"current_responses" gorm:"default:0"`
	PaidOut           float64   `json:"paid_out" gorm:"default:0"`
	RemainingAmount   float64   `json:"remaining_amount" gorm:"not null"`
	IsActive          bool      `json:"is_active" gorm:"default:true"`
	
	ContractAddress   *string   `json:"contract_address"`
	TxHash            *string   `json:"tx_hash"`
	BlockNumber       *int64    `json:"block_number"`
	
	// Relationships
	Survey            Survey    `json:"survey" gorm:"foreignKey:SurveyID"`
	Transactions      []RewardTransaction `json:"transactions,omitempty" gorm:"foreignKey:PoolID"`
}

// RewardTransaction represents a reward transaction
type RewardTransaction struct {
	BaseModel
	UserID      uint                `json:"user_id" gorm:"not null;index"`
	SurveyID    uint                `json:"survey_id" gorm:"not null;index"`
	ResponseID  *uint               `json:"response_id" gorm:"index"`
	PoolID      *uint               `json:"pool_id" gorm:"index"`
	
	Type        TransactionType     `json:"type" gorm:"not null"`
	Amount      float64             `json:"amount" gorm:"not null"`
	Status      TransactionStatus   `json:"status" gorm:"default:'pending';index"`
	
	TxHash      *string             `json:"tx_hash"`
	BlockNumber *int64              `json:"block_number"`
	GasUsed     *int64              `json:"gas_used"`
	GasFee      *float64            `json:"gas_fee"`
	
	ProcessedAt *time.Time          `json:"processed_at"`
	FailureReason *string           `json:"failure_reason"`
	RetryCount  int                 `json:"retry_count" gorm:"default:0"`
	
	User        User                `json:"user" gorm:"foreignKey:UserID"`
	Survey      Survey              `json:"survey" gorm:"foreignKey:SurveyID"`
	Response    *Response           `json:"response,omitempty" gorm:"foreignKey:ResponseID"`
	Pool        *RewardPool         `json:"pool,omitempty" gorm:"foreignKey:PoolID"`
}

type UserBalance struct {
	UserID          uint      `json:"user_id" gorm:"primaryKey"`
	TotalEarned     float64   `json:"total_earned" gorm:"default:0"`
	TotalWithdrawn  float64   `json:"total_withdrawn" gorm:"default:0"`
	AvailableBalance float64  `json:"available_balance" gorm:"default:0"`
	PendingBalance  float64   `json:"pending_balance" gorm:"default:0"`
	LastUpdatedAt   time.Time `json:"last_updated_at"`
	
	User            User      `json:"user" gorm:"foreignKey:UserID"`
}

type WithdrawalRequest struct {
	BaseModel
	UserID          uint              `json:"user_id" gorm:"not null;index"`
	Amount          float64           `json:"amount" gorm:"not null"`
	WalletAddress   string            `json:"wallet_address" gorm:"not null"`
	Status          TransactionStatus `json:"status" gorm:"default:'pending'"`
	
	TransactionID   *uint             `json:"transaction_id"`
	ProcessedAt     *time.Time        `json:"processed_at"`
	FailureReason   *string           `json:"failure_reason"`
	
	User            User              `json:"user" gorm:"foreignKey:UserID"`
	Transaction     *RewardTransaction `json:"transaction,omitempty" gorm:"foreignKey:TransactionID"`
}

func (rp *RewardPool) IsAvailable() bool {
	return rp.IsActive && rp.RemainingAmount > 0 && rp.CurrentResponses < rp.MaxResponses
}

func (rp *RewardPool) CanProcessReward() bool {
	return rp.IsAvailable() && rp.RemainingAmount >= rp.RewardPerResponse
}

func (rp *RewardPool) ProcessReward() error {
	if !rp.CanProcessReward() {
		return errors.New("cannot process reward: insufficient funds or pool inactive")
	}
	
	rp.CurrentResponses++
	rp.PaidOut += rp.RewardPerResponse
	rp.RemainingAmount -= rp.RewardPerResponse
	
	if rp.CurrentResponses >= rp.MaxResponses || rp.RemainingAmount < rp.RewardPerResponse {
		rp.IsActive = false
	}
	
	return nil
}

// IsCompleted checks if the transaction is completed
func (rt *RewardTransaction) IsCompleted() bool {
	return rt.Status == TransactionStatusCompleted
}

// IsPending checks if the transaction is pending
func (rt *RewardTransaction) IsPending() bool {
	return rt.Status == TransactionStatusPending
}

// MarkAsCompleted marks the transaction as completed
func (rt *RewardTransaction) MarkAsCompleted(txHash string, blockNumber int64) {
	now := time.Now()
	rt.Status = TransactionStatusCompleted
	rt.TxHash = &txHash
	rt.BlockNumber = &blockNumber
	rt.ProcessedAt = &now
}

// MarkAsFailed marks the transaction as failed
func (rt *RewardTransaction) MarkAsFailed(reason string) {
	rt.Status = TransactionStatusFailed
	rt.FailureReason = &reason
	rt.RetryCount++
}

// CanRetry checks if the transaction can be retried
func (rt *RewardTransaction) CanRetry() bool {
	return rt.Status == TransactionStatusFailed && rt.RetryCount < 3
}

// UpdateBalance updates the user balance
func (ub *UserBalance) UpdateBalance(earnedAmount, withdrawnAmount float64) {
	ub.TotalEarned += earnedAmount
	ub.TotalWithdrawn += withdrawnAmount
	ub.AvailableBalance = ub.TotalEarned - ub.TotalWithdrawn - ub.PendingBalance
	ub.LastUpdatedAt = time.Now()
}

// CanWithdraw checks if the user can withdraw the specified amount
func (ub *UserBalance) CanWithdraw(amount float64) bool {
	return ub.AvailableBalance >= amount && amount > 0
}

// TableName returns the table name for RewardPool
func (RewardPool) TableName() string {
	return "reward_pools"
}

// TableName returns the table name for RewardTransaction
func (RewardTransaction) TableName() string {
	return "reward_transactions"
}

// TableName returns the table name for UserBalance
func (UserBalance) TableName() string {
	return "user_balances"
}

// TableName returns the table name for WithdrawalRequest
func (WithdrawalRequest) TableName() string {
	return "withdrawal_requests"
}