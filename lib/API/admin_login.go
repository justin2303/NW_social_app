package API

import (
	"encoding/json"
	"fmt"
	"sync"
	"io/ioutil"
	"net/http"
	"os"
)

type AdminLoginReq struct {
	Regiment string `json:"Regiment"`
	Uname     string `json:"Uname"`
	Password string `json:"Password"`
}
var (
	session_admin sync.Mutex 
)

func AdminLogin(w http.ResponseWriter, r *http.Request) {
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
	var adminlogin AdminLoginReq
	err = json.Unmarshal(body, &adminlogin)
	if err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}
	if VerifyAdminLogin(adminlogin.Uname, adminlogin.Regiment, adminlogin.Password) {
		tf, session := GenerateSessionCodeAdmin(adminlogin.Uname)
		if tf {
			resp := LoginResponse{
				Session: session,
			}
			fmt.Println("valid admin login", err)
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(resp)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("valid login but server having issues, please try again"))
		}
	}  else{
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("invalid admin login"))
	}

}

func VerifyAdminLogin(Uname string, Regiment string, Password string) bool{
	filename := "data/Admins/" + Regiment + "/admins.json"
	file, err := os.OpenFile(filename, os.O_RDONLY, 0644)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return false
	}
	defer file.Close()

	var Admins map[string]string
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&Admins); err != nil {
		fmt.Println("Error decoding JSON:", err)
		return false
	}
	if Admins[Uname] == Password {
		return true
	} 
	return false
}
func GenerateSessionCodeAdmin(Uname string) (bool, string){
	code := GenerateCode()
	session_admin.Lock()
	defer session_admin.Unlock()
	filename := "data/Admins/session.json"
	file, err := os.OpenFile(filename, os.O_RDWR, 0644)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return false,""
	}
	defer file.Close()
	var sessionCodes map[string]string
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&sessionCodes); err != nil {
		fmt.Println("Error decoding JSON:", err)
		return false, ""
	}
	sessionCodes[Uname] = code
	if _, err := file.Seek(0, 0); err != nil {
		fmt.Println("Error seeking file:", err)
		return false, ""
	}
	if err := file.Truncate(0); err != nil {
		fmt.Println("Error truncating file:", err)
		return false, ""
	}
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(&sessionCodes); err != nil {
		fmt.Println("Error encoding JSON:", err)
		return false, ""
	}
	return true, code
}


func CheckSessionAdmin(Uname string, Code string) bool {
	session_admin.Lock()
	defer session_admin.Unlock()
	filename := "data/Admins/session.json"
	file, err := os.OpenFile(filename, os.O_RDONLY, 0644)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return false
	}
	defer file.Close()
	var sessionCodes map[string]string
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&sessionCodes); err != nil {
		fmt.Println("Error decoding JSON:", err)
		return false
	}
	if sessionCodes[Uname] == Code {
		return true
	}
	return false
}