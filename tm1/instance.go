package tm1

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/cookiejar"
)

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
