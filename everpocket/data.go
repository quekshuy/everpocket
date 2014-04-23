package everpocket

import (
    "os"
    "log"
    "strings"
    "encoding/json"
    _ "github.com/lib/pq"
    "database/sql"
)

type EverpocketData interface {
    Write()
}

// All the data definitions here. We also need to run migrations
// this way, so we need to version. How to version? 
// Let's create a map. Inside each map is the version number, and
// the migrations from the previous version to the latest version.
// We also need to create a table to store versions, but this is later
// Let's not over-engineer and just write out the definitions now.

const SQL_DDL = `
CREATE TABLE everpocketcreds (
    ev_temp_request_token varchar(256),
    ev_temp_secret varchar(256),
    ev_access_token varchar(256),
    ev_access_secret varchar(256),
    ev_add_data varchar(512)
);
`

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

func (c *EverpocketCreds) String() string {
    return strings.Join([]string{"Cred", c.EvTempRequestToken, c.EvTempSecret}, "_")
}

func getDbConn() *sql.DB {

    db, err := sql.Open("postgres", os.Getenv("EVERPOCKET_PG_URL"))
    if err != nil {
        log.Fatal("Cannot open database connection")
    }
    return db
}

func (c *EverpocketCreds) Write() ( error) {

    jsonEvData, err := json.Marshal(c.EvAddData)
    if err != nil {
        log.Fatal("Error marshalling data")
    }

    db := getDbConn()

    _, err = db.Exec(`INSERT INTO everpocketcreds (
                ev_temp_request_token, 
                ev_temp_secret, 
                ev_access_token, 
                ev_access_secret, 
                ev_add_data
            ) VALUES (?, ?, ?, ?, ?);`,
            c.EvTempRequestToken,
            c.EvTempSecret,
            c.EvAccessToken,
            c.EvAccessSecret,
            jsonEvData,
    )
    if err != nil {
        log.Fatal("Could not write creds. %v", c)
    }
return err
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
        log.Fatalf("Error creating databaset table")
    }

    return nil;
}


