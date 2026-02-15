package tm1

import (
	"context"

	"github.com/andreyea/tm1go/pkg/models"
)

// MonitoringService combines thread, user, and session monitoring operations.
type MonitoringService struct {
	rest    *RestService
	users   *UserService
	threads *ThreadService
	session *SessionService
}

// NewMonitoringService creates a new MonitoringService instance.
func NewMonitoringService(rest *RestService) *MonitoringService {
	return &MonitoringService{
		rest:    rest,
		users:   NewUserService(rest),
		threads: NewThreadService(rest),
		session: NewSessionService(rest),
	}
}

// GetThreads returns current threads.
func (ms *MonitoringService) GetThreads(ctx context.Context) ([]map[string]interface{}, error) {
	return ms.threads.GetAll(ctx)
}

// GetActiveThreads returns non-idle threads.
func (ms *MonitoringService) GetActiveThreads(ctx context.Context) ([]map[string]interface{}, error) {
	return ms.threads.GetActive(ctx)
}

// CancelThread cancels a thread by ID.
func (ms *MonitoringService) CancelThread(ctx context.Context, threadID int) error {
	return ms.threads.Cancel(ctx, threadID)
}

// CancelAllRunningThreads cancels all running threads.
func (ms *MonitoringService) CancelAllRunningThreads(ctx context.Context) ([]map[string]interface{}, error) {
	return ms.threads.CancelAllRunning(ctx)
}

// GetActiveUsers returns active users.
func (ms *MonitoringService) GetActiveUsers(ctx context.Context) ([]*models.User, error) {
	return ms.users.GetActive(ctx)
}

// UserIsActive checks if a user is active.
func (ms *MonitoringService) UserIsActive(ctx context.Context, userName string) (bool, error) {
	return ms.users.IsActive(ctx, userName)
}

// DisconnectUser disconnects a user.
func (ms *MonitoringService) DisconnectUser(ctx context.Context, userName string) error {
	return ms.users.Disconnect(ctx, userName)
}

// GetActiveSessionThreads retrieves threads for current active session.
func (ms *MonitoringService) GetActiveSessionThreads(ctx context.Context, excludeIdle bool) ([]map[string]interface{}, error) {
	return ms.session.GetThreadsForCurrent(ctx, excludeIdle)
}

// GetSessions returns all sessions.
func (ms *MonitoringService) GetSessions(ctx context.Context, includeUser bool, includeThreads bool) ([]map[string]interface{}, error) {
	return ms.session.GetAll(ctx, includeUser, includeThreads)
}

// DisconnectAllUsers disconnects all users except current one.
func (ms *MonitoringService) DisconnectAllUsers(ctx context.Context) ([]string, error) {
	return ms.users.DisconnectAll(ctx)
}

// CloseSession closes one session.
func (ms *MonitoringService) CloseSession(ctx context.Context, sessionID interface{}) error {
	return ms.session.Close(ctx, sessionID)
}

// CloseAllSessions closes all sessions except current user.
func (ms *MonitoringService) CloseAllSessions(ctx context.Context) ([]map[string]interface{}, error) {
	return ms.session.CloseAll(ctx)
}

// GetCurrentUser returns current authenticated user.
func (ms *MonitoringService) GetCurrentUser(ctx context.Context) (*models.User, error) {
	return ms.users.GetCurrent(ctx)
}
