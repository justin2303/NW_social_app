package API

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	wp "hydraulicPress/lib/WorkerPool"
	"hydraulicPress/lib/db_funcs"
	"io/ioutil"
	"net/http"
)

type SimplePlayer struct {
	GUID                string
	Uname               string
	URL					string
	Reg					string
}
type CommendReq struct {
	Regiment string `json:"Regiment"`
	GUID     string `json:"GUID"`
	ToCommend string `json:"ToCommend"`
}
type CommendsResp struct {
	Commends_left         	int `json:"Commends_left"`
	GUID                []string `json:"GUID"`
	Uname               []string `json:"Uname"`
	URL                 []string `json:"URL"`
	Reg                 []string `json:"Reg"`
	Uncommendables 		[]string `json:"Uncommendables"`
}

func GetCommendationsData(w http.ResponseWriter, r *http.Request, pool *wp.WorkerPool){
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
	var CommendsReq RegDataReq
	err = json.Unmarshal(body, &CommendsReq)
	if err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}
	//get top kills
	// get best kd
	// get most deaths without tks
	//select 2 random ppl withh at least 20 kills who isnt topkills
	
	topkills := make(chan SimplePlayer)
	bestKD := make(chan SimplePlayer)
	mostDeaths := make(chan SimplePlayer)
	commends_left := make(chan int)
	noncommends := make(chan []string)
	//random_high := make(chan []SimplePlayer)
	pool.Enqueue(func() {
		resp :=  GetTopKills(CommendsReq.Regiment)
		if resp.URL != "" {
			temp_path := "./data/Players/" + resp.URL + "/profile.png"
			img64 := FetchImageStr(temp_path)
			resp.URL = img64
		}
		topkills <- resp
		fmt.Println("get topkills done")
	}) 
	pool.Enqueue(func() {
		resp :=  GetTopDeaths(CommendsReq.Regiment)
		if resp.URL != "" {
			temp_path := "./data/Players/" + resp.URL + "/profile.png"
			img64 := FetchImageStr(temp_path)
			resp.URL = img64
		}
		mostDeaths <- resp
		fmt.Println("get TopDeaths done")
	}) 
	pool.Enqueue(func() {
		resp :=  GetBestKD(CommendsReq.Regiment)
		if resp.URL != "" {
			temp_path := "./data/Players/" + resp.URL + "/profile.png"
			img64 := FetchImageStr(temp_path)
			resp.URL = img64
		}
		bestKD <- resp
		fmt.Println("get TopKD done")
	}) 
	pool.Enqueue(func() {
		resp := GetNumCommends(CommendsReq.GUID)
		commends_left <- resp
	}) 
	pool.Enqueue(func() {
		resp := GetNonCommendables(CommendsReq.GUID)
		noncommends <- resp
	})

	var persons []SimplePlayer
	person1 := <-topkills
	person2 := <-bestKD
	person3 := <-mostDeaths
	uncommendables := <- noncommends
	commends_to_give := <-commends_left
	excludeGUIDs := []string{person1.GUID, person2.GUID, person3.GUID}
	persons45 := GetTwoRand(CommendsReq.Regiment, excludeGUIDs)
	persons = append(persons, person1)
	persons = append(persons, person2)
	persons = append(persons, person3)
	for _, person := range persons45 {
		persons = append(persons, person)
	}
	var response CommendsResp
	for _, person := range persons {
		response.GUID = append(response.GUID, person.GUID)
		response.Uname = append(response.Uname, person.Uname)
		response.URL = append(response.URL, person.URL)
		response.Reg = append(response.Reg, person.Reg)
	}
	response.Commends_left = commends_to_give
	response.Uncommendables = uncommendables
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}


func GetTopKills(regiment string) SimplePlayer {
	long_query := `SELECT 
    c.GUID, 
    c.Uname, 
    a.URL, 
    a.Reg 
FROM 
    commends_` + db_funcs.GetLastSat() + `  c
JOIN 
    All_players a ON c.GUID = a.GUID
WHERE
	a.Reg = ?
ORDER BY 
    c.Total_kills DESC
LIMIT 1;
`
	db := db_funcs.MakeConnection()
	defer db.Close()
	rows, err := db.Query(long_query,regiment)
	if err != nil {
		fmt.Println("erro1")
		return SimplePlayer{}
	}
	defer rows.Close()
	var response SimplePlayer
    var url sql.NullString
    var reg sql.NullString
    if rows.Next() {
        err := rows.Scan(&response.GUID, &response.Uname, &url, &reg)
        if err != nil {
            fmt.Println("Scan error:", err)
            return SimplePlayer{}
        }
    }
    if url.Valid {
        response.URL = url.String
    } else {
        response.URL = ""
    }
    if reg.Valid {
        response.Reg = reg.String
    } else {
        response.Reg = ""
    }
    return response
}

func GetTopDeaths(regiment string) SimplePlayer {
	long_query := `SELECT 
    c.GUID, 
    c.Uname, 
    a.URL, 
    a.Reg 
FROM 
    commends_` + db_funcs.GetLastSat() + `  c
JOIN 
    All_players a ON c.GUID = a.GUID
WHERE
	a.Reg = ?
ORDER BY 
    c.Total_deaths DESC
LIMIT 1;
`
	db := db_funcs.MakeConnection()
	defer db.Close()
	rows, err := db.Query(long_query,regiment)
	if err != nil {
		fmt.Println("erro1")
		return SimplePlayer{}
	}
	defer rows.Close()
	var response SimplePlayer
    var url sql.NullString
    var reg sql.NullString
    if rows.Next() {
        err := rows.Scan(&response.GUID, &response.Uname, &url, &reg)
        if err != nil {
            fmt.Println("Scan error:", err)
            return SimplePlayer{}
        }
    }
    if url.Valid {
        response.URL = url.String
    } else {
        response.URL = ""
    }
    if reg.Valid {
        response.Reg = reg.String
    } else {
        response.Reg = ""
    }
    return response
}


func GetBestKD(regiment string) SimplePlayer {
    long_query := `SELECT 
        c.GUID, 
        c.Uname, 
        a.URL, 
        a.Reg 
    FROM 
        commends_` + db_funcs.GetLastSat() + `  c
    JOIN 
        All_players a ON c.GUID = a.GUID
	WHERE 
		a.Reg = ?
    ORDER BY 
        c.Total_kills / CASE WHEN c.Total_deaths = 0 THEN 1 ELSE c.Total_deaths END DESC
    LIMIT 1;
    `
    db := db_funcs.MakeConnection()
    defer db.Close()
    rows, err := db.Query(long_query, regiment)
    if err != nil {
        fmt.Println("error1")
        return SimplePlayer{}
    }
    defer rows.Close()

    var response SimplePlayer
    var url sql.NullString
    var reg sql.NullString

    if rows.Next() {
        err := rows.Scan(&response.GUID, &response.Uname, &url, &reg)
        if err != nil {
            fmt.Println("Scan error:", err)
            return SimplePlayer{}
        }
    }

    if url.Valid {
        response.URL = url.String
    } else {
        response.URL = ""
    }

    if reg.Valid {
        response.Reg = reg.String
    } else {
        response.Reg = ""
    }

    return response
}

func GetTwoRand(regiment string, excludeGUIDs []string) []SimplePlayer {
	long_query := `SELECT 
		c.GUID, 
		c.Uname, 
		a.URL, 
		a.Reg 
	FROM 
		commends_` + db_funcs.GetLastSat() + `  c
	JOIN 
		All_players a ON c.GUID = a.GUID
	WHERE 
		a.Reg = ? AND a.GUID NOT IN (?, ?, ?) AND a.Total_kills > 11
	ORDER BY 
		RAND()
	LIMIT 2;
	`

	db := db_funcs.MakeConnection()
	defer db.Close()

	rows, err := db.Query(long_query, regiment, excludeGUIDs[0], excludeGUIDs[1], excludeGUIDs[2])
	if err != nil {
		fmt.Println("Query error:", err)
		return nil
	}
	defer rows.Close()

	var responses []SimplePlayer

	for rows.Next() {
		var response SimplePlayer
		var url sql.NullString
		var reg sql.NullString

		err := rows.Scan(&response.GUID, &response.Uname, &url, &reg)
		if err != nil {
			fmt.Println("Scan error:", err)
			continue
		}

		if url.Valid {
			temp_path := "./data/Players/" + url.String + "/profile.png"
			img64 := FetchImageStr(temp_path)
			response.URL = img64
		} else {
			response.URL = ""
		}

		if reg.Valid {
			response.Reg = reg.String
		} else {
			response.Reg = ""
		}

		responses = append(responses, response)

		// Break the loop if we've collected 2 players
		if len(responses) == 2 {
			break
		}
	}

	return responses
}

func GetNumCommends(GUID string) int {
	commend_query := `SELECT Commends_left from commends_` + db_funcs.GetLastSat() + `  where GUID =  ?`
	db := db_funcs.MakeConnection()
	defer db.Close()
	rows, err := db.Query(commend_query, GUID)
	if err != nil {
		fmt.Println("You weren't in  event", err)
		return 0
	}
	var commends_left int
	defer rows.Close()
	if rows.Next() {
		err := rows.Scan(&commends_left)
		if err != nil {
			fmt.Println("Scan error:", err)
			return 0
		}
	}
	return commends_left
}

func GetNonCommendables(GUID string) []string {
	db := db_funcs.MakeConnection()
	get_q := `select * from weekly_commendations where Commender = ?`
	rows, err := db.Query(get_q, GUID)
	var empty string
	var person1 string
	var person2 string
	var person3 string
	if err != nil {
		fmt.Println("You weren't in the event", err)
		return []string{}
	}
	defer rows.Close()
	if rows.Next() {
		err := rows.Scan(&empty, &person1, &person2, &person3)
		if err != nil {
			fmt.Println("Scan error:", err)
			return []string{}
		}
	}
	return []string{person1, person2, person3}
}

func CommendPlayer(w http.ResponseWriter, r *http.Request, pool *wp.WorkerPool){
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
	var CommendsReq CommendReq
	err = json.Unmarshal(body, &CommendsReq)
	if err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}
	fmt.Println("Commending Player", CommendsReq.ToCommend )
	var wg sync.WaitGroup
	wg.Add(2) 
	pool.Enqueue(func() {
		defer wg.Done()
		UpdateCommends(CommendsReq.GUID, CommendsReq.ToCommend)
	}) 
	pool.Enqueue(func() {
		defer wg.Done()
		UpdateRelations(CommendsReq.GUID, CommendsReq.ToCommend)
	})
	wg.Wait()
	w.WriteHeader(http.StatusOK)
}

func UpdateCommends(GUID1 string, GUID2 string) {
	update_q := `UPDATE commends_` + db_funcs.GetLastSat() + ` 
	SET Commends_left = Commends_left - 1
	WHERE GUID = ?;`
	db := db_funcs.MakeConnection()
	defer db.Close()
	_, err := db.Exec(update_q,GUID1)
	if err != nil {
		fmt.Println("Error updating N_commends:", err)
	}
	update_q = `UPDATE commends_` + db_funcs.GetLastSat() + ` 
	SET N_commends = N_commends + 1
	WHERE GUID = ?;`
	_, err = db.Exec(update_q, GUID2)
	if err != nil {
		fmt.Println("Error updating N_commends:", err)
	}
}

func UpdateRelations(GUID1 string, GUID2 string) {
	db := db_funcs.MakeConnection()
	defer db.Close()

	// Attempt to update Commendee_a first
	queryA := `UPDATE weekly_commendations
		SET Commendee_a = ?
		WHERE Commender = ? AND Commendee_a = '';`
	res, err := db.Exec(queryA, GUID2, GUID1)
	if err != nil {
		fmt.Println("Error updating Commendee_a:", err)
		return
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected > 0 {
		fmt.Println("Successfully updated Commendee_a")
		return
	}

	// If Commendee_a is filled, attempt to update Commendee_b
	queryB := `UPDATE weekly_commendations
		SET Commendee_b = ?
		WHERE Commender = ? AND Commendee_b = '';`
	res, err = db.Exec(queryB, GUID2, GUID1)
	if err != nil {
		fmt.Println("Error updating Commendee_b:", err)
		return
	}
	rowsAffected, _ = res.RowsAffected()
	if rowsAffected > 0 {
		fmt.Println("Successfully updated Commendee_b")
		return
	}

	// If Commendee_b is filled, attempt to update Commendee_c
	queryC := `UPDATE weekly_commendations
		SET Commendee_c = ?
		WHERE Commender = ? AND Commendee_c = '';`
	res, err = db.Exec(queryC, GUID2, GUID1)
	if err != nil {
		fmt.Println("Error updating Commendee_c:", err)
		return
	}
	rowsAffected, _ = res.RowsAffected()
	if rowsAffected > 0 {
		fmt.Println("Successfully updated Commendee_c")
		return
	}

	// If all slots are filled, no update is performed
	fmt.Println("No available slots to update for Commender:", GUID1)
}