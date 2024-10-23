package API

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	wp "hydraulicPress/lib/WorkerPool"
	"hydraulicPress/lib/db_funcs"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

type UploadProfileReq struct {
	GUID    string `json:"GUID"`
	Picture string `json:"Picture"`
}

func UploadPfp(w http.ResponseWriter, r *http.Request, pool *wp.WorkerPool) {
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

	// Unmarshal the JSON body into the LoginRequest struct
	var UploadReq UploadProfileReq
	err = json.Unmarshal(body, &UploadReq)
	if err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}
	if UploadReq.GUID == "" {
		return
	} //no guid case
	/*base64Image := "data:image/png;base64,iVBORw0KGgoAAAANSUhEUg..."

	  // Step 1: Strip the "data:image/png;base64," prefix if present
	  commaIndex := strings.Index(base64Image, ",")
	  if commaIndex != -1 {
	      base64Image = base64Image[commaIndex+1:]
	  }

	  // Step 2: Decode the Base64 string into bytes
	  imageData, err := base64.StdEncoding.DecodeString(base64Image)
	  if err != nil {
	      fmt.Println("Error decoding base64:", err)
	      return
	  }*/
	imgchan := make(chan []byte)

	pool.Enqueue(func() {
		base64Image := UploadReq.Picture
		commaIndex := strings.Index(base64Image, ",")
		if commaIndex != -1 {
			base64Image = base64Image[commaIndex+1:]
		}
		imageData, err := base64.StdEncoding.DecodeString(base64Image)
		if err != nil {
			fmt.Println("Error decoding base64:", err)
		}
		imgchan <- imageData
	}) //image processing

	pool.Enqueue(func() {
		url := db_funcs.GetHashedGUID(UploadReq.GUID)
		filepath := "./data/Players/" + url + "/profile.png"
		file, err := os.Create(filepath)
		if err != nil {
			fmt.Println("failed to create file: ")
		}
		defer file.Close()
		imgdata := <-imgchan //wait for img to be processed
		fmt.Println("image: ", UploadReq.Picture)
		_, err = file.Write(imgdata)
		if err != nil {
			fmt.Println("failed to write to file")
		}
	}) // Is EXTREMELY unlikely there is a write error. but unfortunately im not holding up the server to check for this

	w.WriteHeader(http.StatusOK)
}
