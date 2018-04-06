package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"time"
)

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

func main() {

	//credentials
	user := "admin"
	password := "apple"
	server := "localhost"
	httpPort := "55000"
	ssl := false

	_ = login(ssl, httpPort, server, user, password)
	cubes, _ := getCubes(ssl, httpPort, server)

	//fmt.Println("errors:", err)
	for _, v := range cubes.Value {
		fmt.Println(v.Name)
	}

}

func login(ssl bool, httpPort, server, user, password string) (loginError error) {
	var protocol string
	if ssl {
		protocol = "https"
	} else {
		protocol = "http"
	}

	req, err := http.NewRequest("GET", protocol+"://"+server+":"+httpPort+"/api/v1/Cubes?$select=Name", nil)
	if err != nil {
		return err
	}

	req.SetBasicAuth(user, password)

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

func getCubes(ssl bool, httpPort, server string) (cubes CubesResponse, e error) {
	var protocol string
	cubes = CubesResponse{}
	if ssl {
		protocol = "https"
	} else {
		protocol = "http"
	}

	req, err := http.NewRequest("GET", protocol+"://"+server+":"+httpPort+"/api/v1/Cubes?$select=Name", nil)
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

func logout() (e error) {
	return nil
}
