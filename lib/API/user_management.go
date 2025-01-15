package API

import (
	"encoding/json"
	"fmt"
	"hydraulicPress/lib/db_funcs"
	"io/ioutil"
	"net/http"
)

type EnlistReq struct {
	Regiment string `json:"Regiment"`
	GUID	 string `json:"GUID"`
	Uname	 string `json:"Uname"`
	Session  string `json:"Session"`
}
type RankReq struct {
	Regiment string `json:"Regiment"`
	GUID	 string `json:"GUID"`
	Uname	 string `json:"Uname"`
	Session  string `json:"Session"`
	Rank	 string `json:"Rank"`
}
type RoleReq struct {
	Regiment string `json:"Regiment"`
	GUID	 string `json:"GUID"`
	Uname	 string `json:"Uname"`
	Session  string `json:"Session"`
	Role	 string `json:"Role"`
}
type OrderReq struct {
	Regiment string `json:"Regiment"`
	GUID	 string `json:"GUID"`
	Uname	 string `json:"Uname"`
	Session  string `json:"Session"`
	Order	 string `json:"Order"`
}


func EnlistPlayer(w http.ResponseWriter, r *http.Request){
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
	var EnlistRequest EnlistReq
	err = json.Unmarshal(body, &EnlistRequest)
	if err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}
	if !CheckSessionAdmin(EnlistRequest.Uname, EnlistRequest.Session){
		http.Error(w, "invalid session id", http.StatusBadRequest)
		return
	}
	CreateRegTable(EnlistRequest.Regiment)
	db := db_funcs.MakeConnection()
	enlist_q := `update All_players set Reg = ? where GUID = ?;`
	_, err = db.Exec(enlist_q, EnlistRequest.Regiment, EnlistRequest.GUID)
	insertQuery := fmt.Sprintf(`
    INSERT INTO reg_%s (GUID, PoW)
    SELECT ?, P_week 
    FROM All_players
    WHERE GUID = ?
	`, EnlistRequest.Regiment)
	_, err = db.Exec(insertQuery, EnlistRequest.GUID, EnlistRequest.GUID)
	if err != nil {
		fmt.Println("Error  enlisting player", err)
		http.Error(w, "failed to enlist player",http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func GivePlayerRank(w http.ResponseWriter, r *http.Request){
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
	var request RankReq
	err = json.Unmarshal(body, &request)
	if err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}
	if !CheckSessionAdmin(request.Uname, request.Session){
		http.Error(w, "invalid session id", http.StatusBadRequest)
		return
	}
	if CreateRegTable(request.Regiment) {
		if SetRank(request.Regiment, request.Rank, request.GUID) {
			w.WriteHeader(http.StatusOK)
		} else {
			http.Error(w, "error writing rank to DB", http.StatusBadRequest)
		}
	} else{
		http.Error(w, "Error accessing reg table", http.StatusBadRequest)
		return
	}
}

func GivePlayerRole(w http.ResponseWriter, r *http.Request){
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
	var request RoleReq
	err = json.Unmarshal(body, &request)
	if err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}
	if !CheckSessionAdmin(request.Uname, request.Session){
		http.Error(w, "invalid session id", http.StatusBadRequest)
		return
	}
	if CreateRegTable(request.Regiment) {
		if SetRole(request.Regiment, request.Role, request.GUID) {
			w.WriteHeader(http.StatusOK)
		} else {
			http.Error(w, "error writing rank to DB", http.StatusBadRequest)
		}
	} else{
		http.Error(w, "Error accessing reg table", http.StatusBadRequest)
		return
	} 
}

func GivePlayerOrder(w http.ResponseWriter, r *http.Request){
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
	var request OrderReq
	err = json.Unmarshal(body, &request)
	if err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}
	if !CheckSessionAdmin(request.Uname, request.Session){
		http.Error(w, "invalid session id", http.StatusBadRequest)
		return
	}
	if CreateRegTable(request.Regiment) {
		if SetOrder(request.Regiment, request.Order, request.GUID) {
			w.WriteHeader(http.StatusOK)
		} else {
			http.Error(w, "error writing rank to DB", http.StatusBadRequest)
		}
	} else{
		http.Error(w, "Error accessing reg table", http.StatusBadRequest)
		return
	} 
}


func CreateRegTable(Reg string) bool{
	check_q := fmt.Sprintf(`SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'NW_Players' and table_name = 
	'reg_%s'`, Reg)
	db:= db_funcs.MakeConnection()
	defer db.Close()
	var count int
	err := db.QueryRow(check_q).Scan(&count)
	if err != nil {
		fmt.Println("Error checking db:", err)
		return false
	}
	if count != 0 {
		return true
	}
	createTableQuery := fmt.Sprintf(`
	CREATE TABLE reg_%s (
		GUID VARCHAR(255),
		Grade VARCHAR(255),
		Unit VARCHAR(255),
		PoW int,
		Ordre VARCHAR(255)
	)`, Reg)
	_, err = db.Exec(createTableQuery)
	if err != nil {
		fmt.Println("Error creating table:", err)
		return false
	}
	fmt.Printf("Table reg_%s created successfully\n", Reg)
	return true
}

func SetRank(Reg string, Rank string, GUID string) bool {
	set_q := fmt.Sprintf(`
	Update reg_%s set Grade = ? where GUID = ?`, Reg)
	db:= db_funcs.MakeConnection()
	defer db.Close()
	_, err := db.Exec(set_q, Rank, GUID)
	if err != nil {
		fmt.Println("Error updating table", err)
		return false
	}
	return true
} 
func SetRole(Reg string, Role string, GUID string) bool {
	set_q := fmt.Sprintf(`
	Update reg_%s set Unit = ? where GUID = ?`, Reg)
	db:= db_funcs.MakeConnection()
	defer db.Close()
	_, err := db.Exec(set_q, Role, GUID)
	if err != nil {
		fmt.Println("Error updating table", err)
		return false
	}
	return true
} 
func SetOrder(Reg string, Order string, GUID string) bool {
	set_q := fmt.Sprintf(`
	Update reg_%s set Ordre = ? where GUID = ?`, Reg)
	db:= db_funcs.MakeConnection()
	defer db.Close()
	_, err := db.Exec(set_q, Order, GUID)
	if err != nil {
		fmt.Println("Error updating table", err)
		return false
	}
	return true
} 