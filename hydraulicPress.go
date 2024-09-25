package main

import (
    "fmt"
	"os"
    "hydraulicPress/lib/db_funcs" // Adjust the import path to your db package
	"bufio"
)

func main() {
	file, err := os.Open("lib/db_funcs/server_log_09_21_24.txt")
    if err != nil {
        fmt.Println("Error opening file:", err)
        return
    }
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for x := 0; x < 4383; x++ {
		line, err := db_funcs.ReadLine(scanner)
		if err != nil {
			return 
		}
		time := db_funcs.GetTime(line)
		actors := db_funcs.GetActors(line)

		if len(actors)==4{
			if actors[3] != ""{
				fmt.Printf("%s %s has joined the game with ID %s\n", time, actors[0],actors[3])
			}else if actors[2]!=""{
				fmt.Printf("%s %s did %s on %s\n", time, actors[0], actors[1], actors[2])
			}else if actors[1] != ""{
				fmt.Printf("%s %s did %s on themselves\n", time, actors[0], actors[1])
			}else if actors[0]!=""{
				fmt.Printf("probably teamhit or leave for %s\n", actors[0])
			}
		}else{
			fmt.Println("Nothing happened. for: ", line)
		}
	}
    /*database := db_funcs.MakeConnection() // Call the function to get the db instance
    defer database.Close() // Ensure the connection is closed when done
	rows, err := db_funcs.ExecuteReadQuery(database, "show tables;")
    if err != nil {
        log.Fatal(err)
    }
    defer rows.Close() // Ensure the rows are closed when done
	for rows.Next() {
        var name string

        // Scan the values into the variables
        if err := rows.Scan( &name); err != nil {
            log.Fatal(err)
        }

        // Do something with the data
        fmt.Printf("Table: %s\n", name)
    }*/
	
}