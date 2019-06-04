package tm1

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"strconv"
	"time"
)

//ProcessParameter
type ProcessParameter struct {
	Name  string
	Value string
}

//DimensionsResponse
type DimensionsResponse struct {
	OdataContext string      `json:"@odata.context"`
	Value        []Dimension `json:"value"`
}

//Dimension is describing a single dimension
type Dimension struct {
	OdataEtag              string      `json:"@odata.etag"`
	Name                   string      `json:"Name"`
	Hierarchies            []Hierarchy `json:"Hierarchies"`
	UniqueName             string      `json:"UniqueName"`
	AllLeavesHierarchyName string      `json:"AllLeavesHierarchyName"`
}

//Cell is a single cell of a cellset
type Cell struct {
	Ordinal        int         `json:"Ordinal"`
	Value          interface{} `json:"Value"`
	FormattedValue string      `json:"FormattedValue"`
}

//Axis of a view/mdx query
type Axis struct {
	Ordinal     int         `json:"Ordinal"`
	Cardinality int         `json:"Cardinality"`
	Hierarchies []Hierarchy `json:"Hierarchies"`
	Tuples      []Tuple     `json:"Tuples"`
}

//Hierarchy
type Hierarchy struct {
	OdataEtag   string    `json:"@odata.etag"`
	Name        string    `json:"Name"`
	UniqueName  string    `json:"UniqueName"`
	Cardinality int       `json:"Cardinality"`
	Elements    []Element `json:"Elements"`
	Structure   int       `json:"Structure"`
	Visible     bool      `json:"Visible"`
}

//Member
type Member struct {
	Name          string `json:"Name"`
	UniqueName    string `json:"UniqueName"`
	Type          string `json:"Type"`
	Ordinal       int    `json:"Ordinal"`
	IsPlaceholder bool   `json:"IsPlaceholder"`
	Weight        int    `json:"Weight"`
}

//Tuple
type Tuple struct {
	Ordinal int      `json:"Ordinal"`
	Members []Member `json:"Members"`
}

//Cellset
type Cellset struct {
	OdataContext string `json:"@odata.context"`
	ID           string `json:"ID"`
	Cube         Cube   `json:"Cube"`
	Axes         []Axis `json:"Axes"`
	Cells        []Cell `json:"Cells"`
}

type Element struct {
	OdataContext string            `json:"@odata.context"`
	OdataEtag    string            `json:"@odata.etag"`
	Name         string            `json:"Name"`
	UniqueName   string            `json:"UniqueName"`
	Type         string            `json:"Type"`
	Level        int               `json:"Level"`
	Index        int               `json:"Index"`
	Attributes   map[string]string `json:"Attributes"`
}

//ExecuteProcessResponse error response from executing process
type ExecuteProcessResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

//Cube
type Cube struct {
	OdataEtag         string      `json:"@odata.etag"`
	Name              string      `json:"Name"`
	Dimensions        []Dimension `json:"Dimensions"`
	Rules             string      `json:"Rules"`
	DrillthroughRules interface{} `json:"DrillthroughRules"`
	LastSchemaUpdate  time.Time   `json:"LastSchemaUpdate"`
	LastDataUpdate    time.Time   `json:"LastDataUpdate"`
}

//CubesResponse
type CubesResponse struct {
	OdataContext string `json:"@odata.context"`
	Value        []Cube `json:"value"`
}

//Configuration struct
type Configuration struct {
	OdataContext                                      string        `json:"@odata.context"`
	ServerName                                        string        `json:"ServerName"`
	AdminHost                                         string        `json:"AdminHost"`
	ProductVersion                                    string        `json:"ProductVersion"`
	PortNumber                                        int           `json:"PortNumber"`
	ClientMessagePortNumber                           int           `json:"ClientMessagePortNumber"`
	HTTPPortNumber                                    int           `json:"HTTPPortNumber"`
	IntegratedSecurityMode                            bool          `json:"IntegratedSecurityMode"`
	SecurityMode                                      string        `json:"SecurityMode"`
	PrincipalName                                     string        `json:"PrincipalName"`
	SecurityPackageName                               string        `json:"SecurityPackageName"`
	ClientCAMURIs                                     []interface{} `json:"ClientCAMURIs"`
	WebCAMURI                                         string        `json:"WebCAMURI"`
	ClientPingCAMPassport                             int           `json:"ClientPingCAMPassport"`
	ServerCAMURI                                      string        `json:"ServerCAMURI"`
	AllowSeparateNandCRules                           bool          `json:"AllowSeparateNandCRules"`
	DistributedOutputDir                              string        `json:"DistributedOutputDir"`
	DisableSandboxing                                 bool          `json:"DisableSandboxing"`
	JobQueuing                                        bool          `json:"JobQueuing"`
	ForceReevaluationOfFeedersForFedCellsOnDataChange bool          `json:"ForceReevaluationOfFeedersForFedCellsOnDataChange"`
	DataBaseDirectory                                 string        `json:"DataBaseDirectory"`
	UnicodeUpperLowerCase                             bool          `json:"UnicodeUpperLowerCase"`
}

//Tm1Session struct
type Tm1Session struct {
	baseUrl    string
	httpClient *http.Client
	cookie     string
	connected  bool
	cam        bool
	token      string
}

//GetCubes function
func (s Tm1Session) GetCubes() ([]Cube, error) {
	cubes := CubesResponse{}
	req, err := s.NewRequest("GET", "/Cubes", nil)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	content, err := s.Do(req)
	_ = json.Unmarshal(content, &cubes)
	return cubes.Value, nil
}

//ExecuteMdx runs mdx query
func (s Tm1Session) ExecuteMdx(mdx string) (cellset Cellset, err error) {
	payload := `{"MDX":"` + mdx + `"}`

	cont, err := s.Tm1SendHttpRequest("POST", "/ExecuteMDX?$expand=Axes($expand=Hierarchies,Tuples($expand=Members)),Cells,Cube($select=Name;$expand=Dimensions($select=Name))", payload)
	if err != nil {
		return cellset, err
	}
	err = json.Unmarshal(cont, &cellset)
	if err != nil {
		return cellset, err
	}
	return cellset, nil
}

//Login method
func (s Tm1Session) Login() error {
	fmt.Println("logging in")
	req, err := s.NewRequest("GET", "/Configuration", nil)
	if err != nil {
		fmt.Println(err)
		return err
	}
	//fmt.Println(req)

	_, err = s.Do(req)
	//cubes := CubesResponse{}

	//json.Unmarshal(content, &cubes)

	//fmt.Println(cubes.Value)

	if err != nil {
		fmt.Println("Login failed")
		return err
	}
	s.connected = true
	return nil
}

//Logout method
func (s Tm1Session) Logout() error {
	fmt.Println("logging out")
	req, err := s.NewRequest("POST", "/ActiveSession/tm1.Close", "{}")
	if err != nil {
		fmt.Println(err)
		return err
	}

	_, err = s.Do(req)

	if err != nil {
		fmt.Println("Logout failed")
		return err
	}
	s.connected = false
	return nil
}

//NewSession creates new session
func NewSession(url, user, password, cam string) *Tm1Session {
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	cookieJar, _ := cookiejar.New(nil)
	token := ""
	if cam != "" {
		token = "CAMNamespace " + base64.StdEncoding.EncodeToString([]byte(user+":"+password+":"+cam))
	} else {
		token = "Basic " + base64.StdEncoding.EncodeToString([]byte(user+":"+password))
	}

	s := Tm1Session{
		baseUrl: url,
		httpClient: &http.Client{
			Transport: tr,
			Jar:       cookieJar,
		},

		token: token,
	}

	return &s
}

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

//Do method
func (s Tm1Session) Do(req *http.Request) ([]byte, error) {
	resp, err := s.httpClient.Do(req)
	if resp.StatusCode >= 400 {
		return nil, errors.New("Login failed with a status code: " + resp.Status)
	}
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	content, _ := ioutil.ReadAll(resp.Body)

	return content, nil
}

//Tm1SendHttpRequest
func (s *Tm1Session) Tm1SendHttpRequest(method, path string, body interface{}) ([]byte, error) {
	req, _ := s.NewRequest(method, path, body)
	res, _ := s.Do(req)
	return res, nil
}

//SessionsResponse
type SessionsResponse struct {
	OdataContext string    `json:"@odata.context"`
	Value        []Session `json:value`
}

//Session
type Session struct {
	ID      int    `json:"ID"`
	Context string `json:"Context"`
}

//ThreadsResponse
type ThreadsResponse struct {
	OdataContext string   `json:"@odata.context"`
	Value        []Thread `json:"value"`
}

//Thread
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

func (s Tm1Session) GetThreads() ([]Thread, error) {

	threads := ThreadsResponse{}
	req, err := s.NewRequest("GET", "/Threads?$expand=Session", nil)
	if err != nil {
		fmt.Println(err)
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
func (s Tm1Session) ThreadCancel(id string) error {
	_, err := s.Tm1SendHttpRequest("POST", "/Threads('"+id+"')/tm1.CancelOperation", "{}")

	if err != nil {
		return err
	}
	return nil
}

//CubeCreate creates new cube
func (s Tm1Session) CubeCreate(cube Cube) error {

	var dims string

	for i, v := range cube.Dimensions {
		if len(cube.Dimensions) == i {
			dims = dims + `"Dimensions('` + v.Name + `')"`
		} else {
			dims = dims + `"Dimensions('` + v.Name + `')",`
		}

	}

	payload := `
	{
		"Name": "` + cube.Name + `",
		"Dimensions@odata.bind": [` + dims + `]
	}
	`

	_, err := s.Tm1SendHttpRequest("POST", "/Cubes", payload)

	if err != nil {
		return err
	}
	return nil
}

//DimensionCreate creates new dimension
func (s Tm1Session) DimensionCreate(dim Dimension) error {

	p1,_ := json.Marshal(dim)
	payload:=string(p1)

	_, err := s.Tm1SendHttpRequest("POST", "/Dimensions", payload)

	if err != nil {
		return err
	}
	return nil

}
