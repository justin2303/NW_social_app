package API

import (
	"encoding/json"
	"fmt"
	wp "hydraulicPress/lib/WorkerPool"
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
	Email  string `json:"Email"`
	Domain string `json:"Domain"`
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

	to := "justin.ypc@gmail.com"
	v_code := GenerateCode()
	pool.Enqueue(func() {
		// Set up the message
		v_file.Lock()
		verificationCodes := make(map[string]string)
		file, _ := os.OpenFile("./data/verification/codes.json", os.O_RDWR|os.O_CREATE, 0644)
		// Decode the JSON data into the map
		decoder := json.NewDecoder(file)
		decoder.Decode(&verificationCodes)
		verificationCodes[Email_req.Email] = v_code
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
