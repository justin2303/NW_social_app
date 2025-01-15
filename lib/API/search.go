package API

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"hydraulicPress/lib/db_funcs"
	"io/ioutil"
	"net/http"
)
type SearchReq struct {
	Regiment string `json:"Regiment"`
	Uname 	 string `json:"Uname"`
	GUID	 string `json:"GUID"` //this is player not ADMIN guid, so it's independent of the Uname which refers to the admin asking
	Session  string `json:"Session"`
}
type UnameSearchReq struct {
	Regiment string `json:"Regiment"`
	Uname 	 string `json:"Uname"`
	Uname2	 string `json:"Uname2"` //this is player not ADMIN Uname
	Session  string `json:"Session"`
}
type SearchResponse struct {
	Regiment []string `json:"Regiment"`
	Uname 	 []string `json:"Uname"`
	R_week	 []int `json:"R_week"`
	URL  []string `json:"URL"`
	GUID []string `json:"GUID"`
}

func AdminSearchbyGUID(w http.ResponseWriter, r *http.Request){
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
	var Searchrequest SearchReq
	err = json.Unmarshal(body, &Searchrequest)
	if err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}
	if !CheckSessionAdmin(Searchrequest.Uname, Searchrequest.Session){
		http.Error(w, "invalid session id", http.StatusBadRequest)
		return
	}
	response := GetPlayerbyGUID(Searchrequest.GUID)
	if len(response.Uname) == 0{
		http.Error(w, "Unexisting player", http.StatusBadRequest)
		return
	}else{
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}
func AdminSearchbyUname(w http.ResponseWriter, r *http.Request){
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
	var Searchrequest UnameSearchReq
	err = json.Unmarshal(body, &Searchrequest)
	if err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}
	if !CheckSessionAdmin(Searchrequest.Uname, Searchrequest.Session){
		http.Error(w, "invalid session id", http.StatusBadRequest)
		return
	}
	response := GetPlayerbyUname(Searchrequest.Uname2)
	if len(response.Uname) == 0{
		http.Error(w, "Unexisting player", http.StatusBadRequest)
		return
	}else{
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

func GetPlayerbyGUID(GUID string) SearchResponse {
	player_q := `SELECT Uname, R_week, URL, Reg FROM All_players 
	WHERE GUID=?;`
	db := db_funcs.MakeConnection()
	defer db.Close()
	var Uname string
	var R_week int
	var URL sql.NullString
	var Regiment sql.NullString
	resp := SearchResponse{}
	resp.GUID = append(resp.GUID, GUID)
	err := db.QueryRow(player_q, GUID).Scan(&Uname, &R_week, &URL, &Regiment)
	if err != nil {
		fmt.Println("Error finding player with that GUID: ", err)
		return resp
	}
	Reg:= ""
	if URL.Valid {
		pfp_path := "./data/Players/" + URL.String + "/profile.png"
		resp.URL = append(resp.URL,FetchImageStr(pfp_path))
	} else {
		resp.URL = append(resp.URL,"")
	}
	if Regiment.Valid {
		Reg = Regiment.String
	}
	resp.Regiment = append(resp.Regiment,Reg)
	resp.Uname = append(resp.Uname,Uname)
	resp.R_week = append(resp.R_week,R_week)
	return resp
}

func GetPlayerbyUname(Uname string) SearchResponse {
    player_q := `SELECT GUID, Uname, R_week, URL, Reg FROM All_players 
                 WHERE Uname LIKE ?;`
    db := db_funcs.MakeConnection()
    defer db.Close()
    rows, err := db.Query(player_q, "%"+Uname+"%")
    if err != nil {
        fmt.Println("Error executing query:", err)
        return SearchResponse{}
    }
    defer rows.Close()

    var resp SearchResponse
    for rows.Next() {
        var guid string
        var uname string
        var rWeek int
        var url sql.NullString
        var regiment sql.NullString

        err := rows.Scan(&guid, &uname, &rWeek, &url, &regiment)
        if err != nil {
            fmt.Println("Error scanning row:", err)
            continue
        }

        // Handle nullable fields
        Reg:= ""
		if url.Valid {
			pfp_path := "./data/Players/" + url.String + "/profile.png"
			resp.URL = append(resp.URL,FetchImageStr(pfp_path))
		} else {
			resp.URL = append(resp.URL,"")
		}
		if regiment.Valid {
			Reg = regiment.String
		}
        resp.GUID = append(resp.GUID, guid)
        resp.Uname = append(resp.Uname, uname)
        resp.R_week = append(resp.R_week, rWeek)
        resp.Regiment = append(resp.Regiment, Reg)
    }

    // Check for errors after iterating through rows
    if err = rows.Err(); err != nil {
        fmt.Println("Error during row iteration:", err)
    }

    return resp
}