package data

import (
    "os"
    "strings"
    "strconv"
    "log"
    "encoding/json"
    _ "github.com/lib/pq"
    "database/sql"
)

type EverpocketData interface {
    Write() error
    Delete() error
}

// All the data definitions here. We also need to run migrations
// this way, so we need to version. How to version? 
// Let's create a map. Inside each map is the version number, and
// the migrations from the previous version to the latest version.
// We also need to create a table to store versions, but this is later
// Let's not over-engineer and just write out the definitions now.

const TABLE_NAME = "everpocketcreds"

const SQL_DDL = `
CREATE TABLE `  + TABLE_NAME + ` (
    ev_temp_request_token varchar(256),
    ev_temp_secret varchar(256),
    ev_access_token varchar(256),
    ev_access_secret varchar(256),
    ev_add_data varchar(512)
);
`

// End of data definitions

// EverpocketCreds represents all the oauth tokens
// and whatever not
type EverpocketCreds struct {
    // Evernote oauth interim data below
    EvTempRequestToken string
    EvTempSecret string
    EvAccessToken string
    EvAccessSecret string
    EvAddData map[string]string
}


func getDbConn() *sql.DB {

    // EVERPOCKET_PG_URL should be 
    // postgres://syquek:@localhost/tmp_everpocket?sslmode=disable
    db, err := sql.Open("postgres", os.Getenv("EVERPOCKET_PG_URL"))
    if err != nil {
        log.Fatal("Cannot open database connection")
    }
    return db
}

// Write() is for writing to the database.
func (c *EverpocketCreds) Write() (error) {

    jsonEvData, err := json.Marshal(c.EvAddData)
    if err != nil {
        log.Fatal("Error marshalling data")
    }

    db := getDbConn()

    _, err = db.Exec("INSERT INTO " + TABLE_NAME + " ("+
                "ev_temp_request_token, " +
                "ev_temp_secret, "+
                "ev_access_token, "+
                "ev_access_secret, "+
                "ev_add_data"+
            ") VALUES ($1, $2, $3, $4, $5)",
            c.EvTempRequestToken,
            c.EvTempSecret,
            c.EvAccessToken,
            c.EvAccessSecret,
            jsonEvData,
    )
    //if err != nil {
    //    log.Fatalf("Could not write creds. %v", c)
    //}
    return err
}


// whereClause generates the "WHERE x=y AND ..." part of an 
// SQL query.
func whereClause(tokens map[string]string) (stmt string, args []interface{}) {

    places := []string{}
    i := 1

    for k, v := range tokens {
        args = append(args, v)
        places = append(places, k+"=$"+strconv.Itoa(i))
        i++
    }

    // only if there are tokens
    if i != 1 {
        return " WHERE " + strings.Join(places, " AND "), args
    }
    return "", args
}

// Delete removes a row from the database.
func (c *EverpocketCreds) Delete(tokens map[string]string) (error) {

    db := getDbConn()

    formatted, args := whereClause(tokens)
    sqlStatement := "DELETE FROM " + TABLE_NAME  + formatted + ";"
    //log.Print("delete sql statement: ", sqlStatement, ", args: ", args)
    stmt, err := db.Prepare(sqlStatement)
    if err != nil {
        log.Fatal("Error preparing: ", err)
    }
    _, err = stmt.Exec(args...)
    return err
}

func GetEverpocketCreds(tokens map[string]string) (*EverpocketCreds, error) {

    // format tokens into a slice, with the following sequence: 
    // column_name1, column_value1, column_name2, ....

    db := getDbConn()
    where, args := whereClause(tokens)

    row := db.QueryRow("SELECT * FROM " + TABLE_NAME + " " + where, args...)
    creds := EverpocketCreds{}
    var evAddData string;
    row.Scan(
        &creds.EvTempRequestToken,
        &creds.EvTempSecret,
        &creds.EvAccessToken,
        &creds.EvAccessSecret,
        &evAddData,
    )

    if evAddData  != "" {

        err := json.Unmarshal([]byte(evAddData), &creds.EvAddData)
        if err != nil {
            log.Fatalf("Error unmarshaling json data: %v", err)
        }
    }

    return &creds, nil
}

// CreateDataStore runs the SQL required to create the table in
// the database. Database URL representation is taken from the 
// environment variable EVERPOCKET_PG_URL.
func CreateDataStore() (error) {

    db, err := sql.Open("postgres", os.Getenv("EVERPOCKET_PG_URL"))
    if err != nil {
        log.Fatalf("Error opening database: %v", err)
    }

    if _, err := db.Exec(SQL_DDL); err != nil {
        log.Printf("Error creating database table: %v", err)
        return err
    }

    return nil;
}


