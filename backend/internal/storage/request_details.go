package storage

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// RequestDetail represents a request detail record
// ref: open-sse/handlers/chatCore/requestDetail.js - buildRequestDetail
type RequestDetail struct {
	ID               string          `json:"id"`
	Timestamp        time.Time       `json:"timestamp"`
	Method           string          `json:"method"`
	Path             string          `json:"path"`
	Headers          json.RawMessage `json:"headers,omitempty"`
	Body             json.RawMessage `json:"body,omitempty"`
	Response         json.RawMessage `json:"response,omitempty"`
	StatusCode       *int            `json:"status_code,omitempty"`
	DurationMs       *int            `json:"duration_ms,omitempty"`
	Error            *string         `json:"error,omitempty"`
	ProviderID       *int64          `json:"provider_id,omitempty"`
	AccountID        *int64          `json:"account_id,omitempty"`
	Model            *string         `json:"model,omitempty"`
	TokensPrompt     *int            `json:"tokens_prompt,omitempty"`
	TokensCompletion *int            `json:"tokens_completion,omitempty"`
	CreatedAt        time.Time       `json:"created_at"`
}

// RequestDetailFilters represents filters for querying request details
type RequestDetailFilters struct {
	StartTime   *time.Time `json:"start_time,omitempty"`
	EndTime     *time.Time `json:"end_time,omitempty"`
	ProviderID  *int64     `json:"provider_id,omitempty"`
	Model       *string    `json:"model,omitempty"`
	StatusCode  *int       `json:"status_code,omitempty"`
	Limit       int        `json:"limit"`
	Offset      int        `json:"offset"`
}

// RequestDetailStore handles request detail storage operations
type RequestDetailStore struct {
	db *pgxpool.Pool
}

// NewRequestDetailStore creates a new request detail store
func NewRequestDetailStore(db *pgxpool.Pool) *RequestDetailStore {
	return &RequestDetailStore{db: db}
}

// SaveRequestDetail saves a request detail to the database
// ref: open-sse/handlers/chatCore/requestDetail.js - saveRequestDetail
func (s *RequestDetailStore) SaveRequestDetail(ctx context.Context, detail *RequestDetail) error {
	query := `
		INSERT INTO request_details (
			timestamp, method, path, headers, body, response,
			status_code, duration_ms, error, provider_id, account_id,
			model, tokens_prompt, tokens_completion, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id
	`

	var headers, body, response interface{}
	if detail.Headers != nil {
		headers = detail.Headers
	}
	if detail.Body != nil {
		body = detail.Body
	}
	if detail.Response != nil {
		response = detail.Response
	}

	return s.db.QueryRow(ctx, query,
		detail.Timestamp,
		detail.Method,
		detail.Path,
		headers,
		body,
		response,
		detail.StatusCode,
		detail.DurationMs,
		detail.Error,
		detail.ProviderID,
		detail.AccountID,
		detail.Model,
		detail.TokensPrompt,
		detail.TokensCompletion,
		detail.CreatedAt,
	).Scan(&detail.ID)
}

// GetRequestDetails retrieves request details with filters
func (s *RequestDetailStore) GetRequestDetails(ctx context.Context, filters RequestDetailFilters) ([]RequestDetail, error) {
	query := `
		SELECT 
			id, timestamp, method, path, headers, body, response,
			status_code, duration_ms, error, provider_id, account_id,
			model, tokens_prompt, tokens_completion, created_at
		FROM request_details
		WHERE 1=1
	`

	args := []interface{}{}
	argIndex := 1

	if filters.StartTime != nil {
		query += " AND timestamp >= $" + string(rune('0'+argIndex))
		args = append(args, *filters.StartTime)
		argIndex++
	}

	if filters.EndTime != nil {
		query += " AND timestamp <= $" + string(rune('0'+argIndex))
		args = append(args, *filters.EndTime)
		argIndex++
	}

	if filters.ProviderID != nil {
		query += " AND provider_id = $" + string(rune('0'+argIndex))
		args = append(args, *filters.ProviderID)
		argIndex++
	}

	if filters.Model != nil {
		query += " AND model = $" + string(rune('0'+argIndex))
		args = append(args, *filters.Model)
		argIndex++
	}

	if filters.StatusCode != nil {
		query += " AND status_code = $" + string(rune('0'+argIndex))
		args = append(args, *filters.StatusCode)
		argIndex++
	}

	query += " ORDER BY timestamp DESC"
	
	if filters.Limit > 0 {
		query += " LIMIT $" + string(rune('0'+argIndex))
		args = append(args, filters.Limit)
		argIndex++
	}

	if filters.Offset > 0 {
		query += " OFFSET $" + string(rune('0'+argIndex))
		args = append(args, filters.Offset)
	}

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var details []RequestDetail
	for rows.Next() {
		var detail RequestDetail
		err := rows.Scan(
			&detail.ID,
			&detail.Timestamp,
			&detail.Method,
			&detail.Path,
			&detail.Headers,
			&detail.Body,
			&detail.Response,
			&detail.StatusCode,
			&detail.DurationMs,
			&detail.Error,
			&detail.ProviderID,
			&detail.AccountID,
			&detail.Model,
			&detail.TokensPrompt,
			&detail.TokensCompletion,
			&detail.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		details = append(details, detail)
	}

	return details, nil
}

// DeleteOldRequestDetails deletes request details older than specified time
func (s *RequestDetailStore) DeleteOldRequestDetails(ctx context.Context, olderThan time.Time) error {
	query := `DELETE FROM request_details WHERE created_at < $1`
	_, err := s.db.Exec(ctx, query, olderThan)
	return err
}

// GetRequestDetailByID retrieves a single request detail by ID
func (s *RequestDetailStore) GetRequestDetailByID(ctx context.Context, id string) (*RequestDetail, error) {
	query := `
		SELECT 
			id, timestamp, method, path, headers, body, response,
			status_code, duration_ms, error, provider_id, account_id,
			model, tokens_prompt, tokens_completion, created_at
		FROM request_details
		WHERE id = $1
	`

	var detail RequestDetail
	err := s.db.QueryRow(ctx, query, id).Scan(
		&detail.ID,
		&detail.Timestamp,
		&detail.Method,
		&detail.Path,
		&detail.Headers,
		&detail.Body,
		&detail.Response,
		&detail.StatusCode,
		&detail.DurationMs,
		&detail.Error,
		&detail.ProviderID,
		&detail.AccountID,
		&detail.Model,
		&detail.TokensPrompt,
		&detail.TokensCompletion,
		&detail.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &detail, nil
}
