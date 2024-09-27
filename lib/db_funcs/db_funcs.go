package db_funcs

import (
    "database/sql"
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

func ExecuteReadQuery(db *sql.DB, query string) (*sql.Rows,error) {
    // Execute the query
    rows, err := db.Query(query)
    if err != nil {
        log.Printf("Error executing query: %s\n", err)
        return nil,err // Return the error
    }
    
    log.Println("Query executed successfully.")
    return rows,nil // No error
}
func GetGUID(db *sql.DB, uname string) (string, error){
    query := "SELECT DISTINCT GUID FROM login_09_21_24 WHERE uname = ? LIMIT 1"//always GUID1 cuz that player_act
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


func InsertToEvent(db *sql.DB, actors []string, datetime string) error{
	if strings.Contains(actors[1], "img_ico"){
		switch actors[1] {
		case "img=ico_crossbow":
			actors[1] = "musket"
		case "img=ico_spear":
			actors[1] = "bayonet"
		case "img=ico_swordone":
			actors[1] = "sword"
		case "img=ico_headshot":
			actors[1]= "musket_comp"
		case "img=ico_custom_3":
			actors[1] = "explosive"
		case "img=ico_custom_1": //find out if this is howitzer or round and if custom3 is both howi and tnt
			actors[1] = "arty_round"
		case "img=ico_custom_2":
			actors[1] = "arty_canister"//this and the other customs are all guesses
		default:
			actors[1] = "blunt_damage"//horse bump or fist or some other weid stuff
		}
	}//change action to actual


	if actors[3]!= "" {
		ins_q := "INSERT INTO login_09_21_24 (GUID, uname, Action, Time) VALUES (?, ?, ?, ?)"
		_, err := db.Exec(ins_q, actors[3],actors[0], actors[1], datetime)
		return err
	}
	if actors[2]!=""{
		//player 2 exists
		ins_q := "INSERT INTO event_09_21_24 ( Player_Act, Action, Player_Receive ,Time) VALUES (?, ?, ?, ?)"
		_, err := db.Exec(ins_q, actors[0], actors[1],actors[2], datetime)
		return err
	}
	if actors[1]!=""{

		ins_q := "INSERT INTO event_09_21_24 (Player_Act, Action,Time) VALUES (?, ?, ?)"
		_, err := db.Exec(ins_q, actors[0], actors[1], datetime)
		return err
	}
	//do nothing else, if no receiving player and no guid1, then it means prob suicide (admin actions n caht is ignored)
	return nil
}

func UpdateGUIDs1(db *sql.DB) error {
	fmt.Println("reached here")
    // Step 1: Get distinct Player_Act where GUID1 is null
    query := "SELECT DISTINCT Player_Act FROM event_09_21_24 WHERE GUID1 IS NULL"
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
        updateQuery := "UPDATE event_09_21_24 SET GUID1 = ? WHERE Player_Act = ?"
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
    query := "SELECT DISTINCT Player_Receive FROM event_09_21_24 WHERE GUID2 IS NULL AND Player_Receive is not NULL"
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
        updateQuery := "UPDATE event_09_21_24 SET GUID2 = ? WHERE Player_Receive = ?"
        _, err = db.Exec(updateQuery, guid, Player_Receive)
        if err != nil {
            log.Printf("Error updating GUID2 for %s: %v", Player_Receive, err)
        } else {
            fmt.Printf("Updated GUID2 for Player_Receive: %s with GUID: %s\n", Player_Receive, guid)
        }
    }
    
    return nil
}

