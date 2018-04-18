package tm1

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"time"
)

//ProcessParameter
type ProcessParameter struct {
	Name  string
	Value string
}

//Tm1Instance describes connection to the server
type Tm1Instance struct {
	Ssl      bool
	Server   string
	HttpPort string
	User     string
	Password string
}

//GetProtocol returns http or https depending on ssl
func (tm1 Tm1Instance) GetProtocol() (protocol string) {
	if tm1.Ssl {
		return "https"
	} else {
		return "http"
	}
}

//CubesResponse response of Cubes request
type CubesResponse struct {
	OdataContext string `json:"@odata.context"`
	Value        []Cube `json:"value"`
}

//Cube definition of cube
type Cube struct {
	OdataEtag         string      `json:"@odata.etag"`
	Name              string      `json:"Name"`
	Rules             string      `json:"Rules"`
	DrillthroughRules interface{} `json:"DrillthroughRules"`
	LastSchemaUpdate  time.Time   `json:"LastSchemaUpdate"`
	LastDataUpdate    time.Time   `json:"LastDataUpdate"`
	Dimensions        []Dimension `json:"Dimensions"`
}

//DimensionsResponse
type DimensionsResponse struct {
	OdataContext string      `json:"@odata.context"`
	Value        []Dimension `json:"value"`
}

//Dimension is describing a single dimension
type Dimension struct {
	OdataEtag              string `json:"@odata.etag"`
	Name                   string `json:"Name"`
	UniqueName             string `json:"UniqueName"`
	AllLeavesHierarchyName string `json:"AllLeavesHierarchyName"`
}

//Cell is a single cell of a cellset
type Cell struct {
	Ordinal        int         `json:"Ordinal"`
	Value          interface{} `json:"Value"`
	FormattedValue string      `json:"FormattedValue"`
}

//Axes of a view/mdx query
type Axis struct {
	Ordinal     int         `json:"Ordinal"`
	Cardinality int         `json:"Cardinality"`
	Hierarchies []Hierarchy `json:"Hierarchies"`
	Tuples      []Tuple     `json:"Tuples"`
}

type Hierarchy struct {
	OdataEtag   string        `json:"@odata.etag"`
	Name        string        `json:"Name"`
	UniqueName  string        `json:"UniqueName"`
	Cardinality int           `json:"Cardinality"`
	Structure   int           `json:"Structure"`
	Visible     bool          `json:"Visible"`
	Attributes  []interface{} `json:"Attributes"`
}

type Member struct {
	Name          string        `json:"Name"`
	UniqueName    string        `json:"UniqueName"`
	Type          string        `json:"Type"`
	Ordinal       int           `json:"Ordinal"`
	IsPlaceholder bool          `json:"IsPlaceholder"`
	Weight        int           `json:"Weight"`
	Attributes    []interface{} `json:"Attributes"`
}

type Tuple struct {
	Ordinal int      `json:"Ordinal"`
	Members []Member `json:"Members"`
}

type CellSet struct {
	OdataContext string `json:"@odata.context"`
	ID           string `json:"ID"`
	Cube         Cube   `json:"Cube"`
	Axes         []Axis `json:"Axes"`
	Cells        []Cell `json:"Cells"`
}

//executeProcessResponse error response from executing process
type executeProcessResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

//create client and cookie jar

var (
	cookieJar *cookiejar.Jar
	client    *http.Client
)

func init() {
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	cookieJar, _ = cookiejar.New(nil)
	client = &http.Client{Transport: tr, Jar: cookieJar}
}

func Login(tm1 Tm1Instance) (e error) {

	req, err := http.NewRequest("GET", tm1.GetProtocol()+"://"+tm1.Server+":"+tm1.HttpPort+"/api/v1/Cubes?$select=Name", nil)
	if err != nil {
		return err
	}

	req.SetBasicAuth(tm1.User, tm1.Password)

	res, err := client.Do(req)

	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode >= 400 {
		return errors.New("Login failed with a status code: " + res.Status)
	}

	return nil
}

func GetCubes(tm1 Tm1Instance) (cubes CubesResponse, e error) {

	cubes = CubesResponse{}

	req, err := http.NewRequest("GET", tm1.GetProtocol()+"://"+tm1.Server+":"+tm1.HttpPort+"/api/v1/Cubes?$select=Name", nil)
	if err != nil {
		return cubes, err
	}
	res, err := client.Do(req)

	if err != nil {
		return cubes, err
	}
	defer res.Body.Close()
	if res.StatusCode >= 400 {
		return cubes, errors.New("Login failed with a status code: " + res.Status)
	}

	content, _ := ioutil.ReadAll(res.Body)

	_ = json.Unmarshal(content, &cubes)

	return cubes, nil
}

func Logout(tm1 Tm1Instance) (e error) {

	var payload = []byte(`{}`)
	req, _ := http.NewRequest("POST", tm1.GetProtocol()+"://"+tm1.Server+":"+tm1.HttpPort+"/api/v1/ActiveSession/tm1.Close", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	return nil
}

func cellGet(tm1 Tm1Instance, cube string, element []string) (val interface{}, e error) {

	var (
		mdx     string
		cellSet CellSet
	)

	dims, _ := GetCubeDimensions(tm1, cube)
	if len(dims.Value) != len(element) {
		return 0, errors.New("Wrong number of elements")
	}

	if len(dims.Value) == 2 {
		mdx = "select {[" + dims.Value[0].Name + "].[" + element[0] + "]} on 0,{[" + dims.Value[1].Name + "].[" + element[1] + "]} on 1 from " + cube
	} else {
		mdx = "select {[" + dims.Value[0].Name + "].[" + element[0] + "]} on 0,{[" + dims.Value[1].Name + "].[" + element[1] + "]} on 1 from " + cube + " WHERE ("

		for i := 2; i < len(dims.Value); i++ {
			mdx += "[" + dims.Value[i].Name + "].[" + element[i] + "]"
			if i < len(dims.Value)-1 {
				mdx += ","
			}
		}
		mdx += ")"
	}
	cellSet, _ = ExecuteMdx(tm1, mdx)
	//fmt.Println(cellSet.Cells[0].Value)
	return cellSet.Cells[0].Value, nil
}

func CellGetN(tm1 Tm1Instance, cube string, element ...string) (val float64, e error) {
	value, _ := cellGet(tm1, cube, element)

	switch x := value.(type) {
	case float64:
		val = x
	case int64:
		val = float64(x)
	case int32:
		val = float64(x)
	case int16:
		val = float64(x)
	case string:
		fmt.Println("the cell is a string. cannot convert to a number")
	default:
		fmt.Println("error converting to float64")
	}

	fmt.Println("CellGetN result:", val)
	return val, nil
}

func CellGetS(tm1 Tm1Instance, cube string, element ...string) (val string, e error) {
	value, _ := cellGet(tm1, cube, element)

	switch x := value.(type) {
	case string:
		val = x
	default:
		fmt.Println("error converting to string")
	}

	fmt.Println("CellGetS result:", val)
	return val, nil
}

func ExecuteMdx(tm1 Tm1Instance, mdx string) (cellset CellSet, e error) {

	cellSet := CellSet{}

	query := `{"MDX":"` + mdx + `"}`

	var payload = []byte(query)
	req, _ := http.NewRequest("POST", tm1.GetProtocol()+"://"+tm1.Server+":"+tm1.HttpPort+"/api/v1/ExecuteMDX?$expand=Axes($expand=Hierarchies,Tuples($expand=Members)),Cells,Cube($select=Name;$expand=Dimensions($select=Name))", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return cellSet, err
	}

	defer res.Body.Close()

	content, _ := ioutil.ReadAll(res.Body)

	_ = json.Unmarshal(content, &cellSet)
	//fmt.Println("mdx response:", cellSet.Cells)
	return cellSet, nil
}

func GetCubeDimensions(tm1 Tm1Instance, cube string) (dims DimensionsResponse, e error) {

	dims = DimensionsResponse{}

	req, err := http.NewRequest("GET", tm1.GetProtocol()+"://"+tm1.Server+":"+tm1.HttpPort+"/api/v1/Cubes('"+cube+"')/Dimensions?$select=Name", nil)
	if err != nil {
		return dims, err
	}

	res, err := client.Do(req)

	if err != nil {
		return dims, err
	}
	defer res.Body.Close()

	content, _ := ioutil.ReadAll(res.Body)
	_ = json.Unmarshal(content, &dims)

	return dims, nil
}

//ExecuteProcess executes tm1 process
func ExecuteProcess(tm1 Tm1Instance, process string, p string) (e error) {

	parameters := `{"Parameters":[{"Name":"pEl","Value":"` + p + `"}]}`

	var payload = []byte(parameters)
	req, _ := http.NewRequest("POST", tm1.GetProtocol()+"://"+tm1.Server+":"+tm1.HttpPort+"/api/v1/Processes('"+process+"')/tm1.Execute", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode >= 400 {
		errorMessage := executeProcessResponse{}
		content, _ := ioutil.ReadAll(res.Body)
		_ = json.Unmarshal(content, &errorMessage)
		return errors.New("ExecuteProcess failed: " + res.Status + ". " + errorMessage.Error.Message)
	}
	return nil
}
