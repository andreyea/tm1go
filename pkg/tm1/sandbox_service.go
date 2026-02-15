package tm1

// SandboxService handles operations for TM1 sandboxes.
type SandboxService struct {
	rest *RestService
}

// NewSandboxService creates a new SandboxService instance.
func NewSandboxService(rest *RestService) *SandboxService {
	return &SandboxService{rest: rest}
}
