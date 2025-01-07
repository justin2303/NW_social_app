package API

import (
	"encoding/json"
	"hydraulicPress/lib/db_funcs"
	"io/ioutil"
	"net/http"
)

type ResetReq struct {
	GUID     string `json:"GUID"`
}
type VerifyResetReq struct {
	GUID     string `json:"GUID"`
	Code	 string `json:"Code"`
}
type ChangeReq struct {
	GUID     string `json:"GUID"`
	Password	 string `json:"Password"`
}

func ResetPassHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()
	var Resreq ResetReq
	err = json.Unmarshal(body, &Resreq)
	if err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}
	URL := db_funcs.GetHashedGUID(Resreq.GUID)
	if URL== "" {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("No existing player (signed up with that GUID)"))
	} else {
		//call email to user_config.json email, then frontend goes to verification code page
		gmail, domain_name := FetchEmail(URL)
		if gmail == "" || domain_name == "" {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("No existing player (signed up with that GUID)"))
		} else if SendCode(Resreq.GUID, gmail, domain_name ) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("sending verification code to email"))
		} else {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("error sending mail"))
		}
	}
}

func VerifyReset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()
	var VerifyReq VerifyResetReq
	err = json.Unmarshal(body, &VerifyReq)
	if err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}
	actual_code := FetchVCode(VerifyReq.GUID)
	if actual_code == VerifyReq.Code {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Correct Code"))
	} else {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Wrong Verification code"))
	}
	
}

func ChangePassReq(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()
	var changereq ChangeReq
	err = json.Unmarshal(body, &changereq)
	if err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}
	URL := db_funcs.GetHashedGUID(changereq.GUID)
	if ChangePass(URL, changereq.Password) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Pass changed"))
	} else {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Error changing password"))
	}
	
}