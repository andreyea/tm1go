package tm1

import (
	"encoding/json"
	"strconv"
)

//SessionsResponse
type SessionsResponse struct {
	OdataContext string    `json:"@odata.context"`
	Value        []Session `json:value`
}

//Session describes a tm1 session
type Session struct {
	ID      int    `json:"ID"`
	Context string `json:"Context"`
}

//ThreadsResponse response from threads request
type ThreadsResponse struct {
	OdataContext string   `json:"@odata.context"`
	Value        []Thread `json:"value"`
}

//Thread describes a thread in tm1
type Thread struct {
	ID          int      `json:"ID"`
	Type        string   `json:"Type"`
	Name        string   `json:"Name"`
	Context     string   `json:"Context"`
	State       string   `json:"State"`
	Function    string   `json:"Function"`
	ObjectType  string   `json:"ObjectType"`
	ObjectName  string   `json:"ObjectName"`
	RLocks      int      `json:"RLocks"`
	IXLocks     int      `json:"IXLocks"`
	WLocks      int      `json:"WLocks"`
	ElapsedTime string   `json:"ElapsedTime"`
	WaitTime    string   `json:"WaitTime"`
	Info        string   `json:"Info"`
	Session     *Session `json:"Session"`
}

//GetThreads gets all threads in tm1
func (s Tm1Session) GetThreads() ([]Thread, error) {

	threads := ThreadsResponse{}
	req, err := s.NewRequest("GET", "/Threads?$expand=Session", nil)
	if err != nil {
		return nil, err
	}

	content, err := s.Do(req)
	_ = json.Unmarshal(content, &threads)

	return threads.Value, nil
}

//Cancel a thread bound
func (t Thread) Cancel(s *Tm1Session) error {
	_, err := s.Tm1SendHttpRequest("POST", "/Threads('"+strconv.Itoa(t.ID)+"')/tm1.CancelOperation", "{}")

	if err != nil {
		return err
	}
	return nil
}

//ThreadCancel cancels a thread (unbound)
func (s Tm1Session) ThreadCancel(thread Thread) error {
	_, err := s.Tm1SendHttpRequest("POST", "/Threads('"+string(thread.ID)+"')/tm1.CancelOperation", "{}")

	if err != nil {

		return err
	}
	return nil
}
