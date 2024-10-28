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
type FetchProfileReq struct {
	GUID string `json:"GUID"`
}
type PlayerConfig struct {
	Password     string              `json:"password"`
	Gmail        string              `json:"gmail"`
	DomainName   string              `json:"domain_name"`
	TradingCards []string            `json:"trading_cards"`
	Medals       map[string][]string `json:"medals"`
}
type FetchProfileResp struct {
	Medal_names  []string `json:"Medal_names"`
	Medal_images []string `json:"Medal_images"`
	Medal_desc   []string `json:"Medal_desc"`
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
func FetchProfile(w http.ResponseWriter, r *http.Request, pool *wp.WorkerPool) {
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
	var FetchReq FetchProfileReq
	err = json.Unmarshal(body, &FetchReq)
	if err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}
	h_guid := db_funcs.GetHashedGUID(FetchReq.GUID)
	filepath := "./data/Players/" + h_guid + "/user_config.json"
	var user_config PlayerConfig
	file, err := os.Open(filepath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	// Decode JSON data into the UserConfig struct
	if err := json.NewDecoder(file).Decode(&user_config); err != nil {
		fmt.Println("Error decoding JSON:", err)
		return
	}
	namechan := make(chan []string)
	imgchan := make(chan []string)
	descchan := make(chan []string)

	pool.Enqueue(func() {
		namelist := MedalNameHelper(user_config.Medals)
		namechan <- namelist
	})

	pool.Enqueue(func() {
		imglist := MedalImageHelper(user_config.Medals)
		imgchan <- imglist
	})

	pool.Enqueue(func() {
		desclist := MedalImageHelper(user_config.Medals)
		descchan <- desclist
	})

	medalnames := <-namechan
	medalimgs := <-imgchan
	medaldescs := <-descchan
	response := FetchProfileResp{
		Medal_names:  medalnames,
		Medal_images: medalimgs,
		Medal_desc:   medaldescs,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

}

func MedalNameHelper(Medals map[string][]string) []string {
	var namelist []string
	for regiment, medals := range Medals {
		fmt.Printf("Regiment: %s\n", regiment)

		// For each medal name in the regiment, generate the file path
		for _, medal := range medals {
			namelist = append(namelist, medal)
		}
	}
	return namelist
}
func MedalImageHelper(Medals map[string][]string) []string {
	var encodedImages []string

	for regiment, medals := range Medals {
		fmt.Printf("Regiment: %s\n", regiment)

		// For each medal name in the regiment, generate the file path
		for _, medal := range medals {
			imagePath := fmt.Sprintf("data/medals/%s/%s.png", regiment, medal)
			fmt.Printf("Processing image path: %s\n", imagePath)

			// Read and encode image to base64
			encoded, err := encodeImageToBase64(imagePath)
			if err != nil {
				fmt.Printf("Error encoding image %s: %v\n", imagePath, err)
				encodedImages = append(encodedImages, "")
				continue
			}
			encodedImages = append(encodedImages, encoded)
		}
	}

	return encodedImages
}

func MedalDescHelper(Medals map[string][]string) []string {
	var descriptions []string

	for regiment, medals := range Medals {
		fmt.Printf("Regiment: %s\n", regiment)
		descPath := fmt.Sprintf("data/medals/%s/medal_desc.json", regiment)

		// Read and parse the description file for the current regiment
		medalDescriptions, err := loadDescriptions(descPath)
		if err != nil {
			fmt.Printf("Error reading descriptions for %s: %v\n", regiment, err)
			// Append empty strings for each medal in case of an error
			for range medals {
				descriptions = append(descriptions, "")
			}
			continue
		} else {
			for _, medal := range medals {
				fmt.Printf("Getting description for %s\n", medal)
				desc, ok := medalDescriptions[medal]
				if !ok {
					// If the medal description is missing, append an empty string
					fmt.Printf("Description for %s not found\n", medal)
					descriptions = append(descriptions, "")
				} else {
					// Append the retrieved description
					descriptions = append(descriptions, desc)
				}
			}
		}
	}

	return descriptions
}
func loadDescriptions(descPath string) (map[string]string, error) {
	file, err := os.Open(descPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var descriptions map[string]string
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&descriptions); err != nil {
		return nil, err
	}

	return descriptions, nil
}

func encodeImageToBase64(imagePath string) (string, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Read the entire file
	imageBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return "", err
	}

	// Encode the image bytes to base64
	encoded := base64.StdEncoding.EncodeToString(imageBytes)
	return encoded, nil
}
