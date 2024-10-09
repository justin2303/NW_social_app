package API

import (
	"database/sql"
	"encoding/json"
	"fmt"
	wp "hydraulicPress/lib/WorkerPool"
	"hydraulicPress/lib/db_funcs"
	"io/ioutil"
	"net/http"
)

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
		loginschan <- resp
		fmt.Println("homepage logins goroutine finished")
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
	lastSat = "09_21_24"
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
	lastSat = "09_21_24"
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
	lastSat = "09_21_24"
	fetch_query := fmt.Sprintf("select Action, Time from login_%s where GUID = ?", lastSat)
	var Action, Time sql.NullString
	var logins [][]string
	db := db_funcs.MakeConnection()
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
