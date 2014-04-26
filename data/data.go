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

// Delete removes a row from the database.
func (c *EverpocketCreds) Delete(tokens map[string]string) (error) {
    params := []string{}
    args := []interface{}{}
    i := 1
    for k, v := range tokens {
        params = append(params, k  + "=$" + strconv.Itoa(i))
        /*params = append(params, "?=?")*/
        args = append(args, v)
        i++
    }

    db := getDbConn()

    formatted := strings.Join([]string{"WHERE", strings.Join(params, " AND ")}, " ")
    sqlStatement := "DELETE FROM " + TABLE_NAME + " " + formatted + ";"
    stmt, err := db.Prepare(sqlStatement)
    if err != nil {
        log.Fatal("Error preparing: ", err)
    }
    _, err = stmt.Exec(args...)
    return err
}

// formatMapIntoQuerySlice will take a map[string]string and 
// turn it into a slice that does arranges pairwise the key and the value.
// Returns that slice (as []interface{} because the DB driver only accepts 
// []interface{}... as argument.
func formatMapIntoQuerySlice(params map[string]string) []interface{} {
    s := make([]interface{}, len(params) * 2)
    values := make([]interface{}, len(params))
    for k, v := range params {
        s = append(s, k)
        values = append(values, v)
    }
    return append(s, values...)
}

func GetEverpocketCreds(tokens map[string]string) *EverpocketCreds {

    // format tokens into a slice, with the following sequence: 
    // column_name1, column_value1, column_name2, ....
    tokensAsSlice := formatMapIntoQuerySlice(tokens)

    db := getDbConn()
    row := db.QueryRow(`SELECT * FROM everpocketcreds WHERE ? = ?;`, tokensAsSlice...)
    creds := EverpocketCreds{}
    var evAddData string;
    row.Scan(
        &creds.EvTempRequestToken,
        &creds.EvTempSecret,
        &creds.EvAccessToken,
        &creds.EvAccessSecret,
        &evAddData,
    )

    err := json.Unmarshal([]byte(evAddData), &creds.EvAddData)
    if err != nil {
        log.Fatalf("Error unmarshaling json data: %v", err)
    }
    return &creds
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


