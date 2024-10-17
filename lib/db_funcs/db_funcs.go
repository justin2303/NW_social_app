package db_funcs

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

func MakeConnection() *sql.DB {
	dsn := "root:WarlordJetch1488@tcp(127.0.0.1:6161)/NW_Players"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err) // If there's an error, the connection won't be closed yet.
	}
	fmt.Println("Connection established successfully.")
	return db
}

func ExecuteQuery(db *sql.DB, query string) error {
	// Execute the query
	_, err := db.Exec(query)
	if err != nil {
		log.Printf("Error executing query: %s\n", err)
		return err // Return the error
	}

	log.Println("Query executed successfully.")
	return nil // No error
}

func ExecuteReadQuery(db *sql.DB, query string) (*sql.Rows, error) {
	// Execute the query
	rows, err := db.Query(query)
	if err != nil {
		log.Printf("Error executing query: %s\n", err)
		return nil, err // Return the error
	}

	log.Println("Query executed successfully.")
	return rows, nil // No error
}
func GetGUID(db *sql.DB, uname string) (string, error) {
	date := GetFileDate()
	query := fmt.Sprintf("SELECT DISTINCT GUID FROM login_%s WHERE uname = ? LIMIT 1", date) //always GUID1 cuz that player_act
	row := db.QueryRow(query, uname)
	var guid string

	// Scan the result into the guid variable
	err := row.Scan(&guid)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("no GUID found for player: %s", uname)
		}
		return "", fmt.Errorf("error executing query: %w", err)
	}

	// Return the found GUID
	return guid, nil
}

func InsertToEvent(db *sql.DB, actors []string, datetime string, event_flag bool) error {
	date := GetFileDate()
	if strings.Contains(actors[1], "img=ico") {
		switch actors[1] {
		case "img=ico_crossbow":
			actors[1] = "musket"
		case "img=ico_spear":
			actors[1] = "bayonet"
		case "img=ico_swordone":
			actors[1] = "sword"
		case "img=ico_headshot":
			actors[1] = "musket_comp"
		case "img=ico_custom_3":
			actors[1] = "explosive"
		case "img=ico_custom_1": //find out if this is howitzer or round and if custom3 is both howi and tnt
			actors[1] = "arty_round"
		case "img=ico_custom_2":
			actors[1] = "arty_canister" //this and the other customs are all guesses
		default:
			actors[1] = "blunt_damage" //horse bump or fist or some other weid stuff
		}
	} //change action to actual

	if actors[3] != "" {
		ins_q := fmt.Sprintf("INSERT INTO login_%s (GUID, uname, Action, Time) VALUES (?, ?, ?, ?)", date)
		_, err := db.Exec(ins_q, actors[3], actors[0], actors[1], datetime)
		return err
	} else if !event_flag {
		return nil //skip appending kills suicides and tks and all that ifnot started, just do logins.
	}
	if actors[2] != "" {
		//player 2 exists
		ins_q := fmt.Sprintf("INSERT INTO event_%s ( Player_Act, Action, Player_Receive ,Time) VALUES (?, ?, ?, ?)", date)
		_, err := db.Exec(ins_q, actors[0], actors[1], actors[2], datetime)
		return err
	}
	if actors[1] != "" {

		ins_q := fmt.Sprintf("INSERT INTO event_%s (Player_Act, Action,Time) VALUES (?, ?, ?)", date)
		_, err := db.Exec(ins_q, actors[0], actors[1], datetime)
		return err
	}
	//do nothing else, if no receiving player and no guid1, then it means prob suicide (admin actions n caht is ignored)
	return nil
}

func UpdateGUIDs1(db *sql.DB) error {
	fmt.Println("reached here")
	date := GetFileDate()
	// Step 1: Get distinct Player_Act where GUID1 is null
	query := fmt.Sprintf("SELECT DISTINCT Player_Act FROM event_%s WHERE GUID1 IS NULL", date)
	rows, err := db.Query(query)
	if err != nil {
		return fmt.Errorf("error querying Player_Act: %w", err)
	}
	defer rows.Close()

	// Step 2: Iterate over each distinct Player_Act
	for rows.Next() {
		var playerAct string
		if err := rows.Scan(&playerAct); err != nil {
			return fmt.Errorf("error scanning Player_Act: %w", err)
		}

		// Step 3: Get GUID for Player_Act
		guid, err := GetGUID(db, playerAct)
		if err != nil {
			log.Printf("Error getting GUID for %s: %v", playerAct, err)
			continue // Skip to the next Player_Act on error
		}

		// Step 4: Update GUID1 in event_09_21_24
		updateQuery := fmt.Sprintf("UPDATE event_%s SET GUID1 = ? WHERE Player_Act = ?", date)
		_, err = db.Exec(updateQuery, guid, playerAct)
		if err != nil {
			log.Printf("Error updating GUID1 for %s: %v", playerAct, err)
		} else {
			fmt.Printf("Updated GUID1 for Player_Act: %s with GUID: %s\n", playerAct, guid)
		}
	}

	return nil
}

func UpdateGUIDs2(db *sql.DB) error {
	fmt.Println("reached here")
	// Step 1: Get distinct Player_Receive where GUID2 is null
	date := GetFileDate()
	query := fmt.Sprintf("SELECT DISTINCT Player_Receive FROM event_%s WHERE GUID2 IS NULL AND Player_Receive is not NULL", date)
	rows, err := db.Query(query)
	if err != nil {
		return fmt.Errorf("error querying Player_Receive: %w", err)
	}
	defer rows.Close()

	// Step 2: Iterate over each distinct Player_Receive
	for rows.Next() {
		var Player_Receive string
		if err := rows.Scan(&Player_Receive); err != nil {
			return fmt.Errorf("error scanning Player_Receive: %w", err)
		}

		// Step 3: Get GUID for Player_Receive
		guid, err := GetGUID(db, Player_Receive)
		if err != nil {
			log.Printf("Error getting GUID for %s: %v", Player_Receive, err)
			continue // Skip to the next Player_Receive on error
		}

		// Step 4: Update GUID2 in event_09_21_24
		updateQuery := fmt.Sprintf("UPDATE event_%s SET GUID2 = ? WHERE Player_Receive = ?", date)
		_, err = db.Exec(updateQuery, guid, Player_Receive)
		if err != nil {
			log.Printf("Error updating GUID2 for %s: %v", Player_Receive, err)
		} else {
			fmt.Printf("Updated GUID2 for Player_Receive: %s with GUID: %s\n", Player_Receive, guid)
		}
	}

	return nil
}

func PushNewData() error {
	date := GetFileDate()
	query := "SELECT GUID, MIN(uname) AS uname FROM login_" + date + " GROUP BY GUID;"
	database := MakeConnection()
	defer database.Close()
	rows, err := ExecuteReadQuery(database, query)
	if err != nil {
		fmt.Printf("logs for %s not found\n", date)
		return err
	}
	defer rows.Close() // Close rows after scanning

	// Loop through the rows
	for rows.Next() {
		var guid, uname string
		// Scan the columns into variables
		if err := rows.Scan(&guid, &uname); err != nil {
			fmt.Println("Error scanning row:", err)
			return err
		}
		ins_q := `
		INSERT INTO All_players (GUID, Uname, Last_event, Events_Participated) 
		VALUES (?, ?, ?, 1) 
		ON DUPLICATE KEY UPDATE Events_Participated = Events_Participated + 1;`
		database.Exec(ins_q, guid, uname, date)
	} //all unique GUIDs inserted with A name and events participated updated

	//now get * from event_date and then for each row, if guid2 nil, continue (suicide)
	//if Action = "teamkill" totalkills guid1 -1, totaldeaths guid2 -1 else
	//if guid1 & 2 not null, totalKills +=1 guid1, totaldeaths +1 for guid2
	//
	query2 := "SELECT * FROM event_" + date
	rows2, err2 := ExecuteReadQuery(database, query2)
	if err2 != nil {
		fmt.Printf("logs for %s not found\n", date)
		return err
	}
	defer rows2.Close()
	for rows2.Next() {
		var guid1, guid2, Player_Act, Action, Player_Receive, Time sql.NullString
		// Scan the columns into variables
		if err := rows2.Scan(&guid1, &guid2, &Player_Act, &Action, &Player_Receive, &Time); err != nil {
			fmt.Println("Error scanning row:", err)
			return err
		}
		if Action.Valid {
			if Action.String == "teamkill" {
				tk_q := "Update All_players SET Total_kills = Total_kills - 1,  Total_teamkills = Total_teamkills + 1   where GUID = ?"
				database.Exec(tk_q, guid1)
				tk_q = "Update All_players SET Total_deaths = Total_deaths - 1 where GUID = ?"
				database.Exec(tk_q, guid2)
				continue
			}
		} else { //if action null, mistake in parsing
			fmt.Println("Null action")
			continue
		}
		if !guid1.Valid || !guid2.Valid {
			//nothing to record, suicide or no player act
			fmt.Println("player ID missing")
			continue
		}
		//normal case, not TK, all players valid
		// Before updating Total_kills and Total_deaths
		check_q := "SELECT Total_kills, Total_deaths FROM All_players WHERE GUID = ?;"
		rows3, _ := database.Query(check_q, guid1.String)
		rows3.Close()
		var currentKills, currentDeaths int
		rows3.Scan(&currentKills, &currentDeaths)
		fmt.Printf("Before Update: Total_kills = %d, Total_deaths = %d for GUID: %s\n", currentKills, currentDeaths, guid1.String)
		up_q := "Update All_players SET Total_kills = Total_kills + 1 where GUID = ?"
		database.Exec(up_q, guid1.String)
		up_q = "Update All_players SET Total_deaths = Total_deaths + 1 where GUID = ?"
		database.Exec(up_q, guid2.String)
		fmt.Println("update executed successfully! for time: ", Time)
	}
	return nil

}

func CheckGUIDexists(guid string) (string, error) {
	database := MakeConnection() // Assuming you have a function to make the DB connection
	defer database.Close()

	// Prepare the query to check if GUID exists
	query := "SELECT Uname, URL FROM All_players WHERE GUID = ?"

	// Use QueryRow for a single row result
	var uname, url sql.NullString
	err := database.QueryRow(query, guid).Scan(&uname, &url)

	// Check if the query returned any rows or error
	if err == sql.ErrNoRows || err != nil {
		// No matching GUID found
		return "", err
	}
	if url.Valid {
		err1 := errors.New("user has already signed up, login to access your data")
		return "", err1
	} else if uname.Valid {
		username := uname.String
		return username, nil //return user name so react can say, we found your records with a username of .. do you want to keep this name?
	}

	// GUID found
	return "", nil
}
func GetRegiment(guid string) string {
	db := MakeConnection()
	defer db.Close()
	query := "Select Reg from All_players where GUID = ?"
	var Regiment sql.NullString

	err := db.QueryRow(query, guid).Scan(&Regiment)
	if err != nil {
		fmt.Println("what the hel?? how did you call get reg for a non-existing GUID???")
		return ""
	}
	if Regiment.Valid {
		return Regiment.String
	}
	return "pub" //this means public player. short for pubbie

}

func GetHashedGUID(guid string) string {
	db := MakeConnection()
	defer db.Close()
	query := "Select URL from All_players where GUID = ?"
	var URL sql.NullString

	err := db.QueryRow(query, guid).Scan(&URL)
	if err != nil {
		fmt.Println("what the hel?? how did you call get reg for a non-existing GUID???")
		return ""
	}
	if URL.Valid {
		return URL.String
	}
	return guid

}
