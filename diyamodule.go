package diyamodule

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-rest-framework/core"
	"github.com/jinzhu/gorm"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

var App core.App
var ACQUIRER_TOKEN string
var AUTH_ACQUIRER_TOKEN string
var URL = "api2s.diia.gov.ua"

type DiyaData struct {
	gorm.Model
	Userrequestid string `json:"userrequestid"`
	Deeplink      string `json:"deeplink"`
	Rawdata       string `json:"rawdata"`
}

type TokenRequesData struct {
	Token string `json:"token"`
}

type IDRequesData struct {
	ID string `json:"_id"`
}

type DeepLinkData struct {
	Deeplink string `json:"deeplink"`
}

type DiyaDeepLinkRes struct {
	DeepLink      string `json:"deeplink"`
	UserRequestID string `json:"userrequestid"`
}

func GenMD5Hash() string {
	hash := md5.Sum([]byte(time.Now().Format("2006.01.02 15:04:05")))
	return hex.EncodeToString(hash[:])
}

func Configure(a core.App, acquirer_token string, auth_acquirer_token string) {
	fmt.Println("DIYA module started")
	App = a

	App.DB.AutoMigrate(
		&DiyaData{},
	)

	ACQUIRER_TOKEN = acquirer_token
	AUTH_ACQUIRER_TOKEN = auth_acquirer_token

	App.R.HandleFunc("/diya/test", actionDiyaTest).Methods("GET")
	App.R.HandleFunc("/diya/data", actionDiyaData).Methods("GET")
	App.R.HandleFunc("/diya/point", actionDiyaPoint).Methods("POST")
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
	var (
		model DiyaDeepLinkRes
		rsp   = core.Response{Data: &model, Req: r}
	)

	w.Header().Set("Content-Type", "application/json")
	session_token, err := RequestSessionToken()
	if err != nil {
		log.Println(err)
		rsp.Errors.Add("api", string(err.Error()))

		return
	}
	fmt.Println("SESSION_TOKEN", session_token)
	branch_id, err := CreateBranch("Bearer " + session_token)
	if err != nil {
		log.Println(err)
		rsp.Errors.Add("api", string(err.Error()))

		return
	}
	fmt.Println("BRANCH_ID", branch_id)
	offer_id, err := CreateOffer(branch_id, "Bearer "+session_token)
	if err != nil {
		log.Println(err)
		rsp.Errors.Add("api", string(err.Error()))

		return
	}
	fmt.Println("OFFER_ID", offer_id)

	userRequestId := GenMD5Hash()

	deeplink, err := RequestDeeplink(branch_id, offer_id, "Bearer "+session_token, userRequestId)
	if err != nil {
		log.Println(err)
		rsp.Errors.Add("api", string(err.Error()))

		return
	}

    App.DB.Create(&DiyaData{Userrequestid: userRequestId, Deeplink: deeplink})

	fmt.Println("DEEPLINK", deeplink)

	model.DeepLink = deeplink
	model.UserRequestID = userRequestId

	w.Write(rsp.Make())
}

func actionDiyaPoint(w http.ResponseWriter, r *http.Request) {

	if r.URL.Query().Get("request_id") == "23f25b64bafb0d6c88b1e009b60d527d" {
        w.Header().Set("Content-Type", "application/json")
        body, err := ioutil.ReadAll(r.Body)
        if err != nil {
            log.Fatal(err)
        }
        App.DB.Create(&DiyaData{Rawdata: string(body)})
		w.Write([]byte(`{"success":true}`))
	} else {
		w.WriteHeader(http.StatusForbidden)
	}
}

func actionDiyaData(w http.ResponseWriter, r *http.Request) {
	var (
		data   []DiyaData
		rsp    = core.Response{Data: &data, Req: r}
	)

	//App.DB.Where(&Cdb3Token{El_type: "Bid", Parent: vars["id"]}).Find(&tokens)
	App.DB.Find(&data)

	rsp.Data = data

	w.Write(rsp.Make())
}
