package tm1go

type ObjectService struct {
	rest *RestService
}

func NewObjectService(rest *RestService) *ObjectService {
	return &ObjectService{rest: rest}
}

func (os *ObjectService) Exists(url string) (bool, error) {
	response, err := os.rest.GET(url, nil, 0, nil)
	if err != nil {
		return false, err
	}
	if response.StatusCode == 404 {
		return false, nil
	}
	return true, nil
}
