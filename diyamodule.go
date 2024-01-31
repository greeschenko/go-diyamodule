package diyamodule

import (
	"fmt"
	"github.com/go-rest-framework/core"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

var App core.App
var ACQUIRER_TOKEN string
var AUTH_ACQUIRER_TOKEN string
var URL = "api2s.diia.gov.ua"

func Configure(a core.App, acquirer_token string, auth_acquirer_token string) {
	fmt.Println("DIYA module started")
	App = a
	ACQUIRER_TOKEN = acquirer_token
	AUTH_ACQUIRER_TOKEN = auth_acquirer_token

	App.R.HandleFunc("/diya/test", actionDiyaTest).Methods("GET")
}

func doRequest(url, proto, userJson string) *http.Response {
	reader := strings.NewReader(userJson)
	request, err := http.NewRequest(proto, url, reader)

	request.Header.Set("Authorization", "Basic " + AUTH_ACQUIRER_TOKEN)

	resp, err := http.DefaultClient.Do(request)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("REQUEST", url, proto, resp.StatusCode, "RESP HEADER \r\n", resp.Header, "REQ HEADER\r\n", request.Header, userJson)

	return resp
}

func actionDiyaTest(w http.ResponseWriter, r *http.Request) {

	//get session_token "/api/v1/auth/acquirer/{acquirer_token}"

	resp := doRequest("https://"+URL+"/api/v1/auth/acquirer/"+ACQUIRER_TOKEN, "GET", "")
	if resp.StatusCode != 201 && resp.StatusCode != 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		log.Fatalln(string(body))
		defer resp.Body.Close()
	} else {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("request body is: %+v\n", string(body))
		//fmt.Printf("request header is: %+v\n", r.Header)
		//json.Unmarshal([]byte(body), &u)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"success":true}`))
}
