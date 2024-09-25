package db_funcs

import (
    "database/sql"
    "fmt"
    "log"
	
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
