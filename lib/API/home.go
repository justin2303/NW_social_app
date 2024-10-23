package API

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	wp "hydraulicPress/lib/WorkerPool"
	"hydraulicPress/lib/db_funcs"
	"io/ioutil"
	"net/http"
	"os"
)

type PlayerData struct {
	GUID                string
	Uname               string
	Total_kills         int
	Total_deaths        int
	Total_teamkills     int
	Events_Participated int
	Last_Event          string
	P_week              int
	R_week              int
	URL                 string
	Reg                 string
}

type HomePageReq struct {
	GUID string `json:"GUID"`
}
type HomePageResp1 struct {
	Kills [][]string `json:"Kills"`
}
type HomePageResp2 struct {
	Deaths [][]string `json:"Deaths"`
}
type HomePageCombinedResp struct {
	Kills  [][]string `json:"Kills"`
	Deaths [][]string `json:"Deaths"`
	Logins [][]string `json:"Logins"`
	Uname  string     `json:"Uname"`
}
type NavResponse struct {
	Regiment string `json:"Regiment"`
	Pfp      string `json:"Pfp"`
}

func HomePageHandler(w http.ResponseWriter, r *http.Request, pool *wp.WorkerPool) {
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
	var HomePageReq HomePageReq
	err = json.Unmarshal(body, &HomePageReq)
	if err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}
	killschan := make(chan HomePageResp1)
	deathschan := make(chan HomePageResp2)
	loginschan := make(chan [][]string)
	uname := ""
	//define channels
	fmt.Println("channel created")
	pool.Enqueue(func() {
		resp := GetHPKills(HomePageReq.GUID)
		killschan <- resp
		fmt.Println("homepage kills goroutine finished")
	}) //handle the kills/tks

	pool.Enqueue(func() {
		resp := GetHPDeaths(HomePageReq.GUID)
		deathschan <- resp
		fmt.Println("homepage deaths goroutine finished")
	})
	pool.Enqueue(func() {
		resp := GetHPJoinLeave(HomePageReq.GUID)
		fmt.Println("homepage logins goroutine finished")
		db := db_funcs.MakeConnection()
		err := db.QueryRow("SELECT Uname FROM All_players WHERE GUID = ?", HomePageReq.GUID).Scan(&uname)
		if err != nil {
			if err == sql.ErrNoRows {
				// Handle the case where no rows were returned
				fmt.Println("No user found with the given GUID.")
			} else {
				// Handle any other errors
				fmt.Println("Error querying the database:", err)
			}
		} else {
			// Successfully retrieved the Uname
			fmt.Println("Retrieved Uname:", uname)
		}
		loginschan <- resp
	})

	fmt.Println("goroutine started")
	killresp := <-killschan
	deathresp := <-deathschan
	jlresp := <-loginschan
	fmt.Println("channels returned")
	fmt.Println("amount of login data: ", len(jlresp))
	combined_resp := HomePageCombinedResp{
		Kills:  killresp.Kills,
		Deaths: deathresp.Deaths,
		Logins: jlresp,
		Uname:  uname,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(combined_resp)
	//get username and reg (from guid
	//get last satuday date
	//get all GUID1 = guid player, if teamkill next line, then call it teamkill and remove previous kill
	//get all GUID2 = guid player, deaths
	//get join and leave history
	//get profile picture, and chat logs (implement later), and profile data
	//if not commended yet
	//get list of players to commend (start with people in regiment) most kills,at least 10 kills + 0 teamkills
	//+ event duration <1hr (choose1 rand), and other, other will be another api call
	//if not pub commended yet, just show pub most kills, pub stayed longest, and other, also another api call

}

func GetHPKills(guid_to_fetch string) HomePageResp1 {
	lastSat := db_funcs.GetLastSat()
	fetch_query := fmt.Sprintf("select Action, Player_Receive, Time from event_%s where GUID1 = ? AND GUID2 is not null order by Time", lastSat)
	var Action, Player_Receive, Time sql.NullString
	var kills [][]string
	resp := HomePageResp1{
		Kills: kills,
	}
	db := db_funcs.MakeConnection()
	rows, err := db.Query(fetch_query, guid_to_fetch)
	if err != nil {
		fmt.Println("query error for guid: ", guid_to_fetch)
		return resp
	}
	for rows.Next() {
		fmt.Println("parsing row")
		if err := rows.Scan(&Action, &Player_Receive, &Time); err != nil {
			fmt.Println("query error")
			return resp
		}
		if !Player_Receive.Valid {
			continue //skip if player not valid, should not be triggerd tbh
		}
		if Action.String == "teamkill" {
			if len(kills) > 0 && kills[len(kills)-1][2] == Time.String {
				kills[len(kills)-1][0] = "teamkill"

				continue //skip the rest, we just wanna remove thelast if it was recorded.
			}
		}

		kills = append(kills, []string{Action.String, Player_Receive.String, Time.String})
	}
	resp.Kills = kills

	return resp
}
func GetHPDeaths(guid_to_fetch string) HomePageResp2 {
	lastSat := db_funcs.GetLastSat()
	fetch_query := fmt.Sprintf("select Action, Player_Act, Time from event_%s where GUID2 = ? AND GUID1 is not null order by Time", lastSat)
	var Action, Player_Act, Time sql.NullString
	var deaths [][]string
	db := db_funcs.MakeConnection()
	rows, err := db.Query(fetch_query, guid_to_fetch)
	if err != nil {
		fmt.Println("query error for guid: ", guid_to_fetch)
		resp := HomePageResp2{
			Deaths: deaths,
		}
		return resp //empty response
	}
	for rows.Next() {
		fmt.Println("parsing row")
		if err := rows.Scan(&Action, &Player_Act, &Time); err != nil {
			fmt.Println("query error")
			resp := HomePageResp2{
				Deaths: deaths,
			}

			return resp
		}
		if Action.String == "teamkill" {
			if len(deaths) > 0 && deaths[len(deaths)-1][2] == Time.String {
				deaths[len(deaths)-1][0] = "teamkill" //update prev with tk
				continue                              //skip the rest, we just wanna remove thelast if it was recorded.
			}
		}
		deaths = append(deaths, []string{Action.String, Player_Act.String, Time.String})
	}
	resp := HomePageResp2{
		Deaths: deaths,
	}
	return resp
}

func GetHPJoinLeave(guid_to_fetch string) [][]string {
	lastSat := db_funcs.GetLastSat()
	fetch_query := fmt.Sprintf("select Action, Time from login_%s where GUID = ?", lastSat)
	var Action, Time sql.NullString
	var logins [][]string
	db := db_funcs.MakeConnection()
	defer db.Close()
	rows, err := db.Query(fetch_query, guid_to_fetch)
	if err != nil {
		fmt.Println("query error for guid: ", guid_to_fetch)
		return logins //empty response
	}
	for rows.Next() {
		fmt.Println("parsing row")
		if err := rows.Scan(&Action, &Time); err != nil {
			fmt.Println("query error")
			return logins
		}
		logins = append(logins, []string{Action.String, Time.String})
	}
	return logins
}

func Navigation(w http.ResponseWriter, r *http.Request, pool *wp.WorkerPool) {
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
	var NavReq HomePageReq
	err = json.Unmarshal(body, &NavReq)
	if err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}
	regchan := make(chan string)
	imgchan := make(chan string)
	pool.Enqueue(func() {
		reg := db_funcs.GetRegiment(NavReq.GUID)
		regchan <- reg
		fmt.Println("reg finder routine returned")
	})
	pool.Enqueue(func() {
		h_guid := db_funcs.GetHashedGUID(NavReq.GUID)
		pfp_path := "./data/Players/" + h_guid + "/profile.png"
		str_img := FetchImageStr(pfp_path)
		imgchan <- str_img
		fmt.Println("pfp fetcher routine returned")
	})
	Reg := <-regchan
	Pfp := <-imgchan
	response := NavResponse{
		Regiment: Reg,
		Pfp:      Pfp,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
	//fetch regiment, profile picture

}
func FetchImageStr(path string) string {

	// Open the image file
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return ""
	}
	defer file.Close() // Ensure the file is closed when we are done

	// Get the file size
	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Println("Error getting file info:", err)
		return ""
	}
	fileSize := fileInfo.Size()

	// Create a byte slice to hold the image data
	imageData := make([]byte, fileSize)

	// Read the file's contents into the byte slice
	_, err = file.Read(imageData)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return ""
	}

	// Encode the image data to base64
	base64Data := base64.StdEncoding.EncodeToString(imageData)

	// Create the Data URL
	dataURL := fmt.Sprintf("data:image/png;base64,%s", base64Data)
	// Print the base64 encoded string
	fmt.Println("image encoded")
	return dataURL
}

