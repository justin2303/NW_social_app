package db_funcs

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
	"unicode"
)

func ReadLine(scanner *bufio.Scanner) (string, error) {
	if scanner.Scan() {
		line := scanner.Text() // Get the first line
		return line, nil
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
		return "", err
	}
	return "", nil // Return empty string if no more lines
}

func GetTime(line string) string {
	for i, r := range line {
		if unicode.IsDigit(r) {
			if i+7 <= len(line) {
				return line[i : i+8]
			}
			return line[i:]
		}
	}
	return ""
}

func GetActors(line string) []string {
	// This function returns a list of actor strings: player1, action, player2, GUID
	// Not all fields must be present; sometimes all are empty
	if len(line) < 12 {
		return []string{""}
	}

	var action string
	var player1 string
	var player2 string
	var GUID string
	inAction := false // A flag to track if we're in an action
	inPlayer1 := false
	inPlayer2 := false
	skipNextChar := false
	inGUID := false
	chatbuffer := ""

	// Process line starting at index 11
	for i, r := range line[11:] {
		if skipNextChar {
			skipNextChar = false
			continue
		}

		// If starting an action
		if r == '<' {
			inAction = true
			continue
		}

		if inAction {
			if r != '>' {
				action += string(r)
			} else {
				inAction = false
				skipNextChar = true
				inPlayer2 = true
			}
			continue
		} else if inPlayer1 {
			if r != ' ' {
				player1 += string(r)
			} else {
				inPlayer1 = false
			}
			continue
		} else if inPlayer2 {
			if r != ' ' && r != '.' {
				player2 += string(r)
			} else {
				inPlayer2 = false
			}
			continue
		} else if inGUID {
			if unicode.IsDigit(r) {
				GUID += string(r)
			} else {
				inGUID = false
			}
			continue
		}

		// Skip the first whitespace
		if i == 0 && r == ' ' {
			continue
		}
		// If player1 has been set and we're expecting an action or player2
		if len(player1) == 0 {
			inPlayer1 = true
			player1 += string(r)
			continue
		}

		// If chatbuffer indicates the start of a GUID
		if chatbuffer == "has joined the game with ID:" {
			inGUID = true
			action = "join"
			continue
		} else if chatbuffer == "teamkilled" {
			action = "teamkill"
			inPlayer2 = true
			continue
		} else if chatbuffer == "has left the game with ID:" {
			inGUID = true
			action = "leave"
			continue
		}

		// Accumulate characters into chatbuffer
		chatbuffer += string(r)
	}

	// If the line is from a server message, return an empty string slice
	if player1 == "[SERVER]:" {
		return []string{""}
	}
	// Return the collected player1, action, player2, and GUID as a slice
	player1 = strings.ReplaceAll(player1, " ", "")
	player2 = strings.ReplaceAll(player2, " ", "")
	GUID = strings.ReplaceAll(GUID, " ", "")
	action = strings.ReplaceAll(action, " ", "")
	return []string{player1, action, player2, GUID}
}

func GetFileName() string {
	est, err := time.LoadLocation("America/New_York")
	if err != nil {
		fmt.Println("Error loading timezone:", err)
		return ""
	}

	// Get the current time in EST
	now := time.Now().In(est)

	// Format the date as MM_DD_YY
	formattedDate := now.Format("01_02_06")
	filename := "server_log_" + formattedDate + ".txt"
	return filename
}
func GetFileDate() string {
	est, err := time.LoadLocation("America/New_York")
	if err != nil {
		fmt.Println("Error loading timezone:", err)
		return ""
	}

	// Get the current time in EST
	now := time.Now().In(est)

	// Format the date as MM_DD_YY
	formattedDate := now.Format("01_02_06")
	return formattedDate
}

func GetLastSat() string {
	est, err := time.LoadLocation("America/New_York")
	if err != nil {
		fmt.Println("Error loading timezone:", err)
		return ""
	}
	today := time.Now().In(est)
	// Check if today is Saturday, return today if true
	if today.Weekday() != time.Saturday {
		daysBack := (int(today.Weekday()) + 1) % 7
		// Subtract the number of days to get the last Saturday
		today = today.AddDate(0, 0, -daysBack)
	}

	// Format the date as MM_DD_YY
	formattedDate := today.Format("01_02_06")
	return formattedDate
}

// HasEventStart checks if the line contains specific phrases and starts with "19:"
func HasEventStart(line string) bool {
	return strings.Contains(line, "Has reset the Map.") &&
		strings.Contains(line, "[SERVER]") &&
		strings.HasPrefix(line, " 19:")
}

func ParseLogFile() {
	//fname := GetFileName()

	fname := "logs/server_log_" + GetLastSat() + ".txt"
	file, err := os.Open(fname)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	database := MakeConnection()
	defer database.Close()
	create_event_q := "CREATE TABLE event_" + GetLastSat() + ` (
    GUID1 VARCHAR(255) NULL,
    GUID2 VARCHAR(255) NULL,
    Player_Act VARCHAR(255) NULL,
    Action VARCHAR(255) NULL,
    Player_Receive VARCHAR(255) NULL,
    Time VARCHAR(255) NULL
);`
	create_login_q := "CREATE TABLE login_" + GetLastSat() + ` (
		GUID VARCHAR(255) NULL,
		uname VARCHAR(255) NULL,
		Action VARCHAR(255) NULL,
		Time VARCHAR(255) NULL
	);`

	ExecuteQuery(database, create_event_q)
	ExecuteQuery(database, create_login_q)
	var event_start bool
	for {
		line, err := ReadLine(scanner)
		if err != nil {
			fmt.Println("Error reading line:", err)
			break
		}
		if line == "" { //eof
			break
		}
		if !event_start {
			event_start = HasEventStart(line)
			//check if event started until it is true, then ignore
		}
		time := GetTime(line)
		actors := GetActors(line)
		if len(actors) == 4 {
			err = InsertToEvent(database, actors, time, event_start)
			if err != nil {
				fmt.Println("db error: ", err)
				break
			}
		} else {
			fmt.Println("ignored line: ", line)
		}

	}
	UpdateGUIDs1(database)
	UpdateGUIDs2(database)
	PushNewData()
}

/*
func parse_all(file_name string) {
	file, err := os.Open("yourfile.txt")
    if err != nil {
        fmt.Println("Error opening file:", err)
        return
    }
	for scanner.Scan() {
		line, err = readLine(file)
		if err {
			break
		}

	}

	defer file.Close()

}*/
