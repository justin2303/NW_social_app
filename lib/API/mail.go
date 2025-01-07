package API

import (
	"encoding/json"
	"fmt"
	wp "hydraulicPress/lib/WorkerPool"
	"hydraulicPress/lib/db_funcs"
	"io/ioutil"
	"net/http"
	"os"
	"sync"

	"gopkg.in/gomail.v2"
)

var (
	v_file sync.Mutex // Define the mutex

)

type EmailRequest struct {
	GUID   string `json:"GUID"`
	Email  string `json:"Email"`
	Domain string `json:"Domain"`
}
type VerifyRequest struct {
	GUID         string `json:"GUID"`
	Email        string `json:"Email"`
	Domain       string `json:"Domain"`
	Verification string `json:"Verification"`
	Password 	 string `json:"Password"`
}

func SendEmail(w http.ResponseWriter, r *http.Request, pool *wp.WorkerPool) {
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
	var Email_req EmailRequest
	err = json.Unmarshal(body, &Email_req)
	if err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}
	fmt.Printf("Auth req from %s@%s\n", Email_req.Email, Email_req.Domain)
	// Define the SMTP server and authentication details
	smtpHost := "smtp.gmail.com"
	smtpPort := 587

	// Sender data
	from := "61e.Hussars.NA@gmail.com"
	password := "qdyl mrjt hwvs vkdo" // App password generated from Google

	to := Email_req.Email + "@" + Email_req.Domain
	v_code := GenerateCode()
	pool.Enqueue(func() {
		// Set up the message
		v_file.Lock()
		verificationCodes := make(map[string]string)
		file, _ := os.OpenFile("./data/verification/codes.json", os.O_RDWR|os.O_CREATE, 0644)
		// Decode the JSON data into the map
		decoder := json.NewDecoder(file)
		decoder.Decode(&verificationCodes)
		verificationCodes[Email_req.GUID] = v_code
		jsonData, _ := json.MarshalIndent(verificationCodes, "", "    ")
		file.Truncate(0)
		file.Seek(0, 0)
		file.Write(jsonData)
		file.Close()
		v_file.Unlock()
		msg := gomail.NewMessage()
		msg.SetHeader("From", from)
		msg.SetHeader("To", to)
		msg.SetHeader("Subject", "NW Email Verification")
		message_str := v_code + " is your verification code, enter it within the next 2 minutes to verify your email."
		msg.SetBody("text/html", message_str)
		// Set up the SMTP dialer
		d := gomail.NewDialer(smtpHost, smtpPort, from, password)

		// Send the email
		if err := d.DialAndSend(msg); err != nil {
			fmt.Println("Error sending email:", err)
		} else {
			fmt.Println("Email sent successfully!")
		}
	})
	//success
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Email sent successfully!"))
}

func VerifyEmail(w http.ResponseWriter, r *http.Request, pool *wp.WorkerPool) {
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
	var Verify_req VerifyRequest
	err = json.Unmarshal(body, &Verify_req)
	if err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}
	v_file.Lock()
	verificationCodes := make(map[string]string)
	file, _ := os.OpenFile("./data/verification/codes.json", os.O_RDWR|os.O_CREATE, 0644)
	// Decode the JSON data into the map
	decoder := json.NewDecoder(file)
	decoder.Decode(&verificationCodes)
	actual_code := verificationCodes[Verify_req.GUID]
	file.Close()
	v_file.Unlock()
	if actual_code == Verify_req.Verification {
		fmt.Println("correct code! for", Verify_req.GUID)
		w.WriteHeader(http.StatusOK)
		fmt.Println("verified!")
		pool.Enqueue(func() {
			fmt.Println("trying to hashGUID for ", Verify_req.GUID)
			hashed_guid := HashGUID(Verify_req.GUID) //hash the guid
			ins_url_q := fmt.Sprintf("Update All_players SET URL = '%s' where GUID = '%s'", hashed_guid, Verify_req.GUID)
			database := db_funcs.MakeConnection()
			db_funcs.ExecuteQuery(database, ins_url_q) //set new url
			CreateUserConfig(hashed_guid, Verify_req.Password)
			UpdateUserMail(hashed_guid, Verify_req.Email, Verify_req.Domain)
			fmt.Println("created userconfig")
		})
	} else {
		w.WriteHeader(http.StatusUnauthorized) //in the future verification errors should be unauthorized status.
		fmt.Println("wrong code!, actual code is: ", actual_code)
		fmt.Println("and not: ", Verify_req.Verification)
	}
}

func FetchEmail(URL string) (string, string){
	fmt.Println("URL: ", URL)
	filename := "./data/Players/" + URL + "/user_config.json"
	file, err := os.OpenFile(filename, os.O_RDONLY, 0644)
	if err != nil {
		fmt.Println("error opening file")
		return "", ""
	}
	defer file.Close()
	user_json := make(map[string]string)
	decoder := json.NewDecoder(file)
	decoder.Decode(&user_json)
	if err != nil {
		fmt.Println("Error decoding JSON:", err)
		return "", ""
	}
	gmail := user_json["gmail"]
	domainName := user_json["domain_name"]

	return gmail, domainName
}
func SendCode(GUID, gmail string, domain string) bool {
	smtpHost := "smtp.gmail.com"
	smtpPort := 587
	from := "61e.Hussars.NA@gmail.com"
	password := "qdyl mrjt hwvs vkdo" // App password generated from Google

	to := gmail + "@" + domain
	v_code := GenerateCode()
	v_file.Lock()
	verificationCodes := make(map[string]string)
	file, _ := os.OpenFile("./data/verification/codes.json", os.O_RDWR|os.O_CREATE, 0644)
	// Decode the JSON data into the map
	decoder := json.NewDecoder(file)
	decoder.Decode(&verificationCodes)
	verificationCodes[GUID] = v_code
	jsonData, _ := json.MarshalIndent(verificationCodes, "", "    ")
	file.Truncate(0)
	file.Seek(0, 0)
	file.Write(jsonData)
	file.Close()
	v_file.Unlock()
	msg := gomail.NewMessage()
	msg.SetHeader("From", from)
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", "NW Email Verification")
	message_str := v_code + " is your verification code, enter it within the next 2 minutes to verify your email."
	msg.SetBody("text/html", message_str)
	// Set up the SMTP dialer
	d := gomail.NewDialer(smtpHost, smtpPort, from, password)
	if err := d.DialAndSend(msg); err != nil {
		fmt.Println("Error sending email:", err)
		return false
	} else {
		fmt.Println("Email sent successfully!")
		return true
	}
}

func FetchVCode(GUID string) string{
	v_file.Lock()
	verificationCodes := make(map[string]string)
	file, _ := os.OpenFile("./data/verification/codes.json", os.O_RDWR|os.O_CREATE, 0644)
	// Decode the JSON data into the map
	decoder := json.NewDecoder(file)
	decoder.Decode(&verificationCodes)
	actual_code := verificationCodes[GUID]
	file.Close()
	v_file.Unlock()
	return actual_code
}

func ChangePass(URL string, Password string) bool {
	filename := "./data/Players/" + URL + "/user_config.json"
	file, err := os.OpenFile(filename, os.O_RDWR, 0644)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return false
	}
	defer file.Close()
	user_json := make(map[string]string)
	decoder := json.NewDecoder(file)
	decoder.Decode(&user_json)
	if err != nil {
		fmt.Println("Error decoding JSON:", err)
		return false
	}
	user_json["password"] = Password
	file.Seek(0, 0)
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(&user_json); err != nil {
		fmt.Println("Error encoding JSON:", err)
		return false
	}
	return true
}