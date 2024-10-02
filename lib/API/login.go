package API

import (
	"encoding/json"
	"errors"
	"fmt"
	wp "hydraulicPress/lib/WorkerPool"
	"hydraulicPress/lib/db_funcs"
	"io/ioutil"
	"net/http"
)

type LoginRequest struct {
	GUID     string `json:"GUID"`
	Password string `json:"Password"`
}
type CreateConfigReq struct {
	Email      string `json:"Email"`
	DomainName string `json:"Domain"`
}
type SignupResponse struct {
	Message  string `json:"message"`
	Username string `json:"username,omitempty"`
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Read the request body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	// Unmarshal the JSON body into the LoginRequest struct
	var loginRequest LoginRequest
	err = json.Unmarshal(body, &loginRequest)
	if err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}

	// Print the received body
	fmt.Printf("Received Login Request: GUID=%s, Password=%s\n", loginRequest.GUID, loginRequest.Password)
	//db_funcs. fetch GUID and all stats from All_players, first make parse func for that
	// Send a response back
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Login request received"))
}

func SignupHandler(w http.ResponseWriter, r *http.Request, pool *wp.WorkerPool) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Read the request body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	// Unmarshal the JSON body into the LoginRequest struct
	var SignupReq LoginRequest
	err = json.Unmarshal(body, &SignupReq)
	if err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}
	fmt.Printf("Received Signup Request: GUID=%s, Password=%s\n", SignupReq.GUID, SignupReq.Password)

	// Check if GUID exists
	uname, err := db_funcs.CheckGUIDexists(SignupReq.GUID)
	var ErrUserAlreadySignedUp = errors.New("user has already signed up, login to access your data")
	if errors.Is(err, ErrUserAlreadySignedUp) {
		// Prepare JSON response
		resp := SignupResponse{
			Message: "User has already signed up, login to access your data",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	} else if err != nil {
		// Handle case where user is not found
		resp := SignupResponse{
			Message: "User with that GUID not found in records",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(resp)
		return
	}

	// Return success JSON response
	resp := SignupResponse{
		Message:  "valid Signup request received",
		Username: uname,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp) //once signed in, hash the GUID and then create a folder in data/Players with the hashed GUID,
	pool.Enqueue(func() {
		hashed_guid := HashGUID(SignupReq.GUID) //hash the guid
		ins_url_q := fmt.Sprintf("Update All_players SET URL = '%s' where GUID = '%s'", hashed_guid, SignupReq.GUID)
		database := db_funcs.MakeConnection()
		db_funcs.ExecuteQuery(database, ins_url_q) //set new url
	})
	//insert hashed GUID to URL
	//prompt user to setup their account, ask if they want a diff userame, ask for recovery mail
	//create the json with the pass, email, domainname, and emtpy medals and tradin_cards
	//create user_config.json by dumping

}

func CreateUserConfig(w http.ResponseWriter, r *http.Request) {

}
