package data


import (
    "testing"
)


func TestWriteNormalDb(t *testing.T) {
    // We're having problems writing to the database.
    // Now we just want to know if a normal write would work.
    db := getDbConn()
    stmt, err := db.Prepare(`INSERT INTO everpocketcreds (ev_temp_request_token, ev_temp_secret, ev_access_token, ev_access_secret, ev_add_data) VALUES ($1, $2, $3, $4, $5)`)
    if err!=nil {
        t.Fatal("Didn't work1: ", err)
    }

    ev_temp_rt := "abcd"

    res, err := stmt.Exec(ev_temp_rt, "abcd", "", "", "")
    if err != nil || res == nil {
        t.Fatal("Exec didn't work")
    }

    // now process the delete
    defer func() {
        stmt, _ := db.Prepare("DELETE FROM everpocketcreds WHERE ev_temp_request_token=$1")
        t.Log("Prepared statement")
        if _, err := stmt.Exec(ev_temp_rt); err != nil {
            t.Fatal("Could not delete fake everpocketcreds")
        }
        t.Log("Done running")
    }()
}

// TestGetEverpocketCreds: Simple test to ensure our GetEverpocketCreds function works correctly.
func TestGetEverpocketCreds(t *testing.T) {

    creds := &EverpocketCreds{
        EvTempRequestToken: "abcd",
        EvTempSecret: "1234",
    }

    err := creds.Write()
    if err != nil {
        t.Fatal("Did not write: ", err)
    }

    r, err := GetEverpocketCreds(map[string]string{"ev_temp_request_token": "abcd"})
    if r == nil || r.EvTempRequestToken != "abcd" {
        t.Fatal("Did not retrieve: ", r)
    } else {
        r.Delete()
    }

}

// TestWriteEverpocketCreds tests that the Write method
// for an EverpocketCreds struct works.
func TestWritePartialEverpocketCreds(t *testing.T) {

    creds := &EverpocketCreds{
        EvTempRequestToken: "abcd",
        EvTempSecret: "1234",
        // note the rest of the struct is empty
    }
    err := creds.Write()
    if err != nil {
        t.Fatalf("Failed to write to database: %v", err)
    }

    err = creds.Delete()
    if err != nil {
        t.Fatalf("Delete failed: %v", err)
    }
}
