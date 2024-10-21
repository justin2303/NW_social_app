package API

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	wp "hydraulicPress/lib/WorkerPool"
	"hydraulicPress/lib/db_funcs"
	"io/ioutil"
	"net/http"
	"os"
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

type UserConfig struct {
	Password     string                 `json:"password"`
	Gmail        string                 `json:"gmail"`
	DomainName   string                 `json:"domain_name"`
	TradingCards []string               `json:"trading_cards"`
	Medals       map[string]interface{} `json:"medals"` // Change the type as needed
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

	if VerifyLogin(loginRequest.GUID, loginRequest.Password) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("valid login"))
	} else {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("invalid login"))
	}

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
		CreateUserConfig(hashed_guid, SignupReq.Password)
	})
	//insert hashed GUID to URL
	//prompt user to setup their account, ask if they want a diff userame, ask for recovery mail
	//create the json with the pass, email, domainname, and emtpy medals and tradin_cards
	//create user_config.json by dumping

}

func CreateUserConfig(hashed_guid string, password string) {
	user_config := UserConfig{
		Password:     password,
		Gmail:        "",
		DomainName:   "",
		TradingCards: []string{},
		Medals:       make(map[string]interface{}), // Initialize the map
	}
	jsonData, err := json.MarshalIndent(user_config, "", "  ")
	// Create a file named guid + ".json"
	filename := hashed_guid + "/user_config.json"
	filename = "./data/Players/" + filename
	dirPath := "./data/Players/" + hashed_guid

	// Create the directory if it doesn't exist
	err = os.MkdirAll(dirPath, os.ModePerm) // Create the directory if it doesn't exist
	if err != nil {
		fmt.Println("Error creating directory:", err)
		return
	}
	file, err := os.Create(filename)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close() // Ensure the file is closed after writing

	// Write JSON data to the file
	_, err = file.Write(jsonData)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}

}
func UpdateUserMail(hashed_guid string, Mail string, Domain string) error {
	filePath := "./data/Players/" + hashed_guid + "/user_config.json"
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Step 2: Parse the file's JSON content into a struct
	var userConfig UserConfig
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&userConfig); err != nil {
		return fmt.Errorf("failed to decode JSON: %w", err)
	}

	// Step 3: Update the "Mail" and "Domain" keys
	userConfig.Gmail = Mail
	userConfig.DomainName = Domain

	// Step 4: Marshal the updated struct back to JSON
	updatedData, err := json.MarshalIndent(userConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Step 5: Open the file for writing (truncate it before writing)
	fileWrite, err := os.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file for writing: %w", err)
	}
	defer fileWrite.Close()

	// Step 6: Write the updated JSON back to the file
	_, err = fileWrite.Write(updatedData)
	if err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	fmt.Println("User config updated successfully!")
	return nil
}
func VerifyLogin(guid string, pass string) bool {
	url_q := "select URL from All_players where GUID = ? limit 1"
	db := db_funcs.MakeConnection()
	// Execute the query
	defer db.Close()
	var URL sql.NullString
	row := db.QueryRow(url_q, guid) // Use QueryRow for single row results
	fmt.Println("GUID: ", guid)

	// Check for error during query execution
	if err := row.Scan(&URL); err != nil {
		if err == sql.ErrNoRows {
			// No rows found
			return false
		}
		// Other scanning errors
		fmt.Println("Error scanning row:", err)
		return false
	}

	// Check if URL is valid
	if !URL.Valid {
		fmt.Println("URL is NULL or invalid.")
		return false
	}

	fmt.Println("URL: ", URL.String)
	filename := "./data/Players/" + URL.String + "/user_config.json"
	file, err := os.OpenFile(filename, os.O_RDONLY, 0644)
	if err != nil {
		fmt.Println("error opening file")
		return false
	}
	defer file.Close()
	user_json := make(map[string]string)
	decoder := json.NewDecoder(file)
	decoder.Decode(&user_json)
	password := user_json["password"]
	return password == pass
}
