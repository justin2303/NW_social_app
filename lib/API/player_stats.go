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

type RegDataReq struct {
	Regiment string `json:"Regiment"`
	GUID     string `json:"GUID"`
}
type RegDataResp struct {
	GUID                []string `json:"GUID"`
	Uname               []string `json:"Uname"`
	Total_kills         []int    `json:"Total_kills"`
	Total_deaths        []int    `json:"Total_deaths"`
	Total_teamkills     []int    `json:"Total_teamkills"`
	Events_Participated []int    `json:"Events_Participated"`
	Last_Event          []string `json:"Last_Event"`
	P_week              []int    `json:"P_week"`
	R_week              []int    `json:"R_week"`
	URL                 []string `json:"URL"`
	Reg                 []string `json:"Reg"`
}

func GetAllRegData(w http.ResponseWriter, r *http.Request, pool *wp.WorkerPool) {
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
	var RegReq RegDataReq
	err = json.Unmarshal(body, &RegReq)
	if err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}
	fmt.Println("Requested Reg: ", RegReq.Regiment)
	db := db_funcs.MakeConnection()
	defer db.Close()
	reg_query := "select * from All_players where Reg = ?"
	var rows *sql.Rows
	if RegReq.Regiment != "pub" {
		rows, err = db.Query(reg_query, RegReq.Regiment)
	} else {
		reg_query = "select * from All_players where Reg is null"
		rows, err = db.Query(reg_query)
	}
	if err != nil {
		http.Error(w, "Error asking DB", http.StatusBadRequest)
		return
	}
	var Reggies []PlayerData
	var GUID, Uname, Last_event, URL, Reg_buffer sql.NullString
	var Kills, Deaths, Teamkills, Attendance, P_week, R_week int

	for rows.Next() {
		fmt.Println("parsing row")
		if err := rows.Scan(&GUID, &Uname, &Kills, &Deaths, &Teamkills, &Attendance, &Last_event, &P_week, &R_week, &URL, &Reg_buffer); err != nil {
			fmt.Println("query error")
		} // no need to parse validity, since null values .String = ""

		curr_reggie := PlayerData{
			GUID:                GUID.String,
			Uname:               Uname.String,
			Total_kills:         Kills,
			Total_deaths:        Deaths,
			Total_teamkills:     Teamkills,
			Events_Participated: Attendance,
			Last_Event:          Last_event.String,
			P_week:              P_week,
			R_week:              R_week,
			URL:                 URL.String,
			Reg:                 Reg_buffer.String,
		}
		Reggies = append(Reggies, curr_reggie)
	} //Should be fast, just 1 pass and only appends, no additional parsing
	resp := RegDataResp{
		GUID:                []string{}, // Initialize slices for each field
		Uname:               []string{},
		Total_kills:         []int{},
		Total_deaths:        []int{},
		Total_teamkills:     []int{},
		Events_Participated: []int{},
		Last_Event:          []string{},
		P_week:              []int{},
		R_week:              []int{},
		URL:                 []string{},
		Reg:                 []string{},
	}

	// Loop through Reggies and append each field to the corresponding slice
	for _, reggie := range Reggies {
		resp.GUID = append(resp.GUID, reggie.GUID)
		resp.Uname = append(resp.Uname, reggie.Uname)
		resp.Total_kills = append(resp.Total_kills, reggie.Total_kills)
		resp.Total_deaths = append(resp.Total_deaths, reggie.Total_deaths)
		resp.Total_teamkills = append(resp.Total_teamkills, reggie.Total_teamkills)
		resp.Events_Participated = append(resp.Events_Participated, reggie.Events_Participated)
		resp.Last_Event = append(resp.Last_Event, reggie.Last_Event)
		resp.P_week = append(resp.P_week, reggie.P_week)
		resp.R_week = append(resp.R_week, reggie.R_week)
		//get Image instead of url
		img64 := ""
		if reggie.URL != "" {
			temp_path := "./data/Players/" + reggie.URL + "/profile.png"
			img64 = FetchImageStr(temp_path)
			fmt.Println("image: ", img64)
		}
		resp.URL = append(resp.URL, img64)
		resp.Reg = append(resp.Reg, reggie.Reg)
	}
	//resp populated
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)

}

func GetWeeklyData(w http.ResponseWriter, r *http.Request, pool *wp.WorkerPool) {
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
	var RegReq RegDataReq
	err = json.Unmarshal(body, &RegReq)
	if err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}
}
