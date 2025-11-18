package tm1

// SubsetService handles operations for TM1 Subsets
type SubsetService struct {
	rest *RestService
}

// NewSubsetService creates a new SubsetService instance
func NewSubsetService(rest *RestService) *SubsetService {
	return &SubsetService{
		rest: rest,
	}
}

// TODO: Implement subset operations
