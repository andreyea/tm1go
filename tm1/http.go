package tm1

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

//NewRequest method
func (s Tm1Session) NewRequest(method, path string, body interface{}) (*http.Request, error) {

	payload := new(bytes.Buffer)

	if body != nil {
		payload = bytes.NewBuffer([]byte(fmt.Sprintf("%v", body)))
	}

	req, err := http.NewRequest(method, s.baseUrl+path, payload)

	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	req.Header.Set("Authorization", s.token)

	return req, nil
}

//Do method sends http request to tm1 instance
func (s Tm1Session) Do(req *http.Request) ([]byte, error) {
	resp, err := s.httpClient.Do(req)

	if resp.StatusCode >= 400 {
		//fix this in the future. add proper error handling and message output
		content, _ := ioutil.ReadAll(resp.Body)
		fmt.Println(string(content))
		return content, errors.New("Request failed with a status code: " + resp.Status)
	}

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	content, _ := ioutil.ReadAll(resp.Body)

	return content, nil
}

//Tm1SendHttpRequest combines NewRequest and Do functions
func (s *Tm1Session) Tm1SendHttpRequest(method, path string, body interface{}) ([]byte, error) {
	req, err := s.NewRequest(method, path, body)
	if err != nil {
		return nil, err
	}
	res, err := s.Do(req)
	if err != nil {
		return nil, err
	}
	return res, nil
}
