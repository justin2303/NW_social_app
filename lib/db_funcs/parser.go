package db_funcs

import (
	"bufio"
	"fmt"
	"unicode"
)

func ReadLine(scanner *bufio.Scanner) (string, error) {
	if scanner.Scan() {
        line := scanner.Text() // Get the first line
		return line,nil
	}
	if err := scanner.Err(); err != nil {
        fmt.Println("Error reading file:", err)
		return "", err
    }
	return "", nil // Return empty string if no more lines
}

func GetTime(line string) string{
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
	inAction := false  // A flag to track if we're in an action
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
			if r != ' ' {
				player2 += string(r)
			} else {
				inPlayer2 = false
			}
			continue
		} else if inGUID {
			GUID += string(r)
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
	return []string{player1, action, player2, GUID}
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