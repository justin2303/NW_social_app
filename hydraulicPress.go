package main

import (
    "fmt"
    "hydraulicPress/lib/db_funcs" // Adjust the import path to your db package
)

func main() {

	db_funcs.Parse_log_file()
	fmt.Println("all good")
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