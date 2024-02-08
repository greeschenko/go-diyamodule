package diyamodule

import (
	"encoding/json"
	"errors"
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

type TokenRequesData struct {
	Token string `json:"token"`
}

type IDRequesData struct {
	ID string `json:"_id"`
}

type DeepLinkData struct {
	Deeplink string `json:"deeplink"`
}

func Configure(a core.App, acquirer_token string, auth_acquirer_token string) {
	fmt.Println("DIYA module started")
	App = a
	ACQUIRER_TOKEN = acquirer_token
	AUTH_ACQUIRER_TOKEN = auth_acquirer_token

	App.R.HandleFunc("/diya/test", actionDiyaTest).Methods("GET")
}

func doRequest(url, proto, userJson, auth string) *http.Response {
	reader := strings.NewReader(userJson)
	request, err := http.NewRequest(proto, url, reader)

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", auth)

	resp, err := http.DefaultClient.Do(request)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("REQUEST", url, proto, resp.StatusCode, "RESP HEADER \r\n", resp.Header, "REQ HEADER\r\n", request.Header, userJson)

	return resp
}

func RequestSessionToken() (string, error) {
	var tokenData TokenRequesData
	resp := doRequest("https://"+URL+"/api/v1/auth/acquirer/"+ACQUIRER_TOKEN, "GET", "", "Basic "+AUTH_ACQUIRER_TOKEN)
	if resp.StatusCode != 201 && resp.StatusCode != 200 {
		body, err := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()
		if err != nil {
			return "", errors.New("ERROR session token request " + string(body))
		}
	} else {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		json.Unmarshal([]byte(body), &tokenData)
	}

	return tokenData.Token, nil
}

func CreateBranch(session_token string) (string, error) {
	var idData IDRequesData
	var data = `{
		    "name": "ТОВ ПОЛОНЕКС",
		    "email": "info@polonex.com.ua",
		    "region": "Київ",
		    "district": "Київ",
		    "location": "Київ",
		    "street": "Вулиця",
		    "house": "Будинок",
		    "deliveryTypes": ["api"],
		    "offerRequestType": "dynamic",
		    "scopes": {
		        "sharing": [
	                "internal-passport",
	                "taxpayer-card"
		        ],
		        "identification": [],
		        "documentIdentification": [
		            "internal-passport"
		        ]
	        }
	    }`
	//var data = "{\"customFullName\":\"Повна назва запитувача документа\", \"customFullAddress\":\"Повна адреса відділення\", \"name\":\"Назва відділення\", \"email\":\"test@email.com\", \"region\":\"Київська обл.\", \"district\":\"Києво-Святошинський р-н\", \"location\":\"м. Вишневе\", \"street\":\"вул. Київська\", \"house\":\"2г\", \"deliveryTypes\": [\"api\"], \"offerRequestType\": \"dynamic\", \"scopes\":{\"sharing\":[\"passport\",\"internal-passport\",\"foreign-passport\"], \"identification\":[], \"documentIdentification\":[\"internal-passport\",\"foreign-passport\"]}}"
	resp := doRequest("https://"+URL+"/api/v2/acquirers/branch", "POST", data, session_token)

	if resp.StatusCode != 201 && resp.StatusCode != 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error body read", err)
		}
		defer resp.Body.Close()
		return "", errors.New("ERROR branch create " + string(body))
	} else {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("request body is: %+v\n", string(body))
		json.Unmarshal([]byte(body), &idData)
	}

	return idData.ID, nil
}

func CreateOffer(branch_id, session_token string) (string, error) {
	var idData IDRequesData
	var data = `{
	     "name": "Назва послуги",
	     "returnLink": "https://polonex.com.ua",
	     "scopes": {
	         "sharing": [
	            "internal-passport",
                "taxpayer-card"
            ]
	     }
    }`
	resp := doRequest("https://"+URL+"/api/v1/acquirers/branch/"+branch_id+"/offer", "POST", data, session_token)

	if resp.StatusCode != 201 && resp.StatusCode != 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error body read", err)
		}
		defer resp.Body.Close()
		return "", errors.New("ERROR offer create " + string(body))
	} else {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		json.Unmarshal([]byte(body), &idData)
	}

	return idData.ID, nil
}

func RequestDeeplink(branch_id, offer_id, session_token, request_id string) (string, error) {
	var DLData DeepLinkData
	var data = `{
	    "offerId": "` + offer_id + `",
	    "requestId": "` + request_id + `",
	    "useDiiaId": true
    }`
	resp := doRequest("https://"+URL+"/api/v2/acquirers/branch/"+branch_id+"/offer-request/dynamic", "POST", data, session_token)

	if resp.StatusCode != 201 && resp.StatusCode != 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error body read", err)
		}
		defer resp.Body.Close()
		return "", errors.New("ERROR request deeplink " + string(body))
	} else {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		json.Unmarshal([]byte(body), &DLData)
	}

	return DLData.Deeplink, nil
}

func actionDiyaTest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	session_token, err := RequestSessionToken()
	if err != nil {
		fmt.Println(err)
		w.Write([]byte(`{"error":true}`))
		return
	}
	fmt.Println("SESSION_TOKEN", session_token)
	branch_id, err := CreateBranch("Bearer " + session_token)
	if err != nil {
		fmt.Println(err)
		w.Write([]byte(`{"error":true}`))
		return
	}
	fmt.Println("BRANCH_ID", branch_id)
	offer_id, err := CreateOffer(branch_id, "Bearer "+session_token)
	if err != nil {
		fmt.Println(err)
		w.Write([]byte(`{"error":true}`))
		return
	}
	fmt.Println("OFFER_ID", offer_id)

	deeplink, err := RequestDeeplink(branch_id, offer_id, "Bearer "+session_token, "9174eadaca8d1f1dbe2e8f0685c6753f")
	if err != nil {
		fmt.Println(err)
		w.Write([]byte(`{"error":true}`))
		return
	}
	fmt.Println("DEEPLINK", deeplink)

	w.Write([]byte(`{"success":"` + deeplink + `"}`))
}
