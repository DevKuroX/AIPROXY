package chat

import (
	"context"
	"encoding/json"
	"time"

	"github.com/DevKuroX/AIPROXY/internal/storage"
)

type RequestDetailLogger struct {
	storage *storage.RequestDetailStore
}

func NewRequestDetailLogger(store *storage.RequestDetailStore) *RequestDetailLogger {
	return &RequestDetailLogger{
		storage: store,
	}
}

func (l *RequestDetailLogger) LogRequest(ctx context.Context, detail *RequestDetail) error {
	dbDetail := &storage.RequestDetail{
		ID:        detail.ID,
		Timestamp: detail.Timestamp,
		Method:    detail.Method,
		Path:      detail.Path,
		Model:     stringPtr(detail.Model),
	}

	if detail.StatusCode > 0 {
		dbDetail.StatusCode = intPtr(detail.StatusCode)
	}
	if detail.Duration > 0 {
		dbDetail.DurationMs = intPtr(int(detail.Duration.Milliseconds()))
	}
	if detail.Error != "" {
		dbDetail.Error = stringPtr(detail.Error)
	}
	if detail.ProviderID > 0 {
		dbDetail.ProviderID = int64Ptr(detail.ProviderID)
	}
	if detail.AccountID > 0 {
		dbDetail.AccountID = int64Ptr(detail.AccountID)
	}
	if detail.TokensPrompt > 0 {
		dbDetail.TokensPrompt = intPtr(detail.TokensPrompt)
	}
	if detail.TokensCompletion > 0 {
		dbDetail.TokensCompletion = intPtr(detail.TokensCompletion)
	}

	if detail.Headers != nil {
		if headersJSON, err := json.Marshal(detail.Headers); err == nil {
			dbDetail.Headers = headersJSON
		}
	}

	if detail.Body != nil {
		if bodyJSON, err := json.Marshal(detail.Body); err == nil {
			dbDetail.Body = bodyJSON
		}
	}

	if detail.Response != nil {
		if responseJSON, err := json.Marshal(detail.Response); err == nil {
			dbDetail.Response = responseJSON
		}
	}

	return l.storage.SaveRequestDetail(ctx, dbDetail)
}

func (l *RequestDetailLogger) GetRequestDetails(ctx context.Context, filters RequestDetailFilters) ([]RequestDetail, error) {
	dbFilters := storage.RequestDetailFilters{
		Limit:  filters.Limit,
		Offset: filters.Offset,
	}
	if !filters.StartTime.IsZero() {
		dbFilters.StartTime = &filters.StartTime
	}
	if !filters.EndTime.IsZero() {
		dbFilters.EndTime = &filters.EndTime
	}
	if filters.ProviderID > 0 {
		dbFilters.ProviderID = int64Ptr(filters.ProviderID)
	}
	if filters.Model != "" {
		dbFilters.Model = stringPtr(filters.Model)
	}
	if filters.StatusCode > 0 {
		dbFilters.StatusCode = intPtr(filters.StatusCode)
	}

	dbDetails, err := l.storage.GetRequestDetails(ctx, dbFilters)
	if err != nil {
		return nil, err
	}

	details := make([]RequestDetail, len(dbDetails))
	for i, db := range dbDetails {
		details[i] = RequestDetail{
			ID:        db.ID,
			Timestamp: db.Timestamp,
			Method:    db.Method,
			Path:      db.Path,
			Model:     ptrToString(db.Model),
		}
		if db.StatusCode != nil {
			details[i].StatusCode = *db.StatusCode
		}
		if db.DurationMs != nil {
			details[i].Duration = time.Duration(*db.DurationMs) * time.Millisecond
		}
		if db.Error != nil {
			details[i].Error = *db.Error
		}
		if db.ProviderID != nil {
			details[i].ProviderID = *db.ProviderID
		}
		if db.AccountID != nil {
			details[i].AccountID = *db.AccountID
		}
		if db.TokensPrompt != nil {
			details[i].TokensPrompt = *db.TokensPrompt
		}
		if db.TokensCompletion != nil {
			details[i].TokensCompletion = *db.TokensCompletion
		}
	}

	return details, nil
}

func intPtr(v int) *int {
	return &v
}

func int64Ptr(v int64) *int64 {
	return &v
}

func stringPtr(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}

func ptrToString(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}
