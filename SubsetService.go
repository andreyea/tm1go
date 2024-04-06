package tm1go

type SubsetService struct {
	rest *RestService
}

func NewSubsetService(rest *RestService) *SubsetService {
	return &SubsetService{rest: rest}
}
