package tm1go

type SandboxService struct {
	rest *RestService
}

func NewSandboxService(rest *RestService) *SandboxService {
	return &SandboxService{rest: rest}
}
