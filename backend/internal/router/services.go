package router

import (
	"log/slog"
	"net/http"

	"github.com/DevKuroX/AIPROXY/internal/services"
	"github.com/DevKuroX/AIPROXY/internal/storage"
)

type Services struct {
	AccountFallback *services.AccountFallbackService
	Combo           *services.ComboService
	Compact         *services.CompactService
	Model           *services.ModelService
	ProjectID       *services.ProjectIDService
	Provider        *services.ProviderService
	TokenRefresh    *services.TokenRefreshService
	Usage           *services.UsageService
}

var globalServices *Services

func InitServices(db *storage.DB, logger *slog.Logger, httpClient *http.Client, encryptionKey []byte) *Services {
	globalServices = &Services{
		AccountFallback: services.NewAccountFallbackService(logger),
		Combo:           services.NewComboService(),
		Compact:         services.NewCompactService(logger),
		Model:           services.NewModelService(db),
		ProjectID:       services.NewProjectIDService(httpClient, logger),
		Provider:        services.NewProviderService(db, logger),
		TokenRefresh:    services.NewTokenRefreshService(db, encryptionKey, logger),
		Usage:           services.NewUsageService(db, logger),
	}
	return globalServices
}

func GetServices() *Services {
	return globalServices
}

func (s *Services) StartBackgroundTasks() {
	s.ProjectID.StartCacheCleanup()
}

func (s *Services) StopBackgroundTasks() {
	s.ProjectID.StopCacheCleanup()
}
