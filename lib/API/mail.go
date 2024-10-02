package API

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"gopkg.in/gomail.v2"
)

type EmailRequest struct {
	Email  string `json:"Email"`
	Domain string `json:"Domain"`
}

func SendEmail(w http.ResponseWriter, r *http.Request) {
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

	// Set up the message
	msg := gomail.NewMessage()
	msg.SetHeader("From", from)
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", "Test Email from Go")
	msg.SetBody("text/html", "Hello, this is a test email from Go using Gmail SMTP!")

	// Set up the SMTP dialer
	d := gomail.NewDialer(smtpHost, smtpPort, from, password)

	// Send the email
	if err := d.DialAndSend(msg); err != nil {
		fmt.Println("Error sending email:", err)
	} else {
		fmt.Println("Email sent successfully!")
	}
	//success
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Email sent successfully!"))
	fmt.Println("Email sent successfully!")
}
