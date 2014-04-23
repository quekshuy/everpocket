package main

import (
    "fmt"
    "log"
    "strings"
    "github.com/lib/pq"
    "github.com/go-martini/martini"

    "github.com/quekshuy/everpocket/data"
    "github.com/quekshuy/evernote-golang-sdk/auth"
)

const EVERNOTE_HOST = "https://sandbox.evernote.com"

func InitiateEvernoteOauth() (string, error) {
    token, secret, url, err, _ := auth.GetEvernoteTempRequestToken(EVERNOTE_HOST)
    if err == nil {
        creds := &data.EverpocketCreds{
            EvTempRequestToken: token,
            EvTempSecret: secret,
        }
        creds.Write()
        return url, nil
    }
    return "", err
}

// Home returns the page that gives users 
func Home() string {

    // Get temporary credentials for Evernote
    url, err := InitiateEvernoteOauth()
    if err != nil {
        log.Fatal("Could not get temp request token")
    }
    return fmt.Sprintf(FRONTPAGE, url)
}

func Done() string {
    return DONEPAGE
}

func SetupServer() {

    m := martini.Classic()

    // Login page
    m.Get("/", Home)
    // Done page
    // Options page (to cancel the connection, etc) *not that important for now
    // Total pages: 3

    m.Run()

}

func main() {

    // Create Database Tables first if necessary
    err := data.CreateDataStore()

    // check to see if it's a postgres error
    if err, ok := err.(*pq.Error); ok {

        errMsg := err.Message
        log.Println("Error msg: ", errMsg)

        // if our table already exists, ignore
        if !strings.Contains(errMsg, "already exists") {
            // big error
            log.Fatalf("Database error: %s", errMsg)
        }
    }

    SetupServer()

}

