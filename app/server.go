package main

import (
    "os"
    "fmt"
    "log"
    "strings"
    "strconv"
    "net/http"
    "github.com/lib/pq"
    "github.com/go-martini/martini"

    "github.com/quekshuy/everpocket/data"
    "github.com/quekshuy/evernote-golang-sdk/auth"
    pauth "github.com/quekshuy/pocket-golang-sdk/auth"
)


const EVERNOTE_HOST = "https://sandbox.evernote.com"
const ENV_PO_SVC_CALLBACK = "POCKET_SERVICE_CALLBACK"

func InitiateEvernoteOauth() (string, error) {
    token, secret, url, _, err := auth.GetEvernoteTempRequestToken(EVERNOTE_HOST)
    if err == nil {
        creds := &data.EverpocketCreds{
            EvTempRequestToken: token,
            EvTempSecret: secret,
        }
        if err = creds.Write(); err != nil {
            log.Fatal("Failed to write EverpocketCreds: %v", err)
        }
        return url, nil
    }
    return "", err
}

func EvernoteCallbackHandler(res http.ResponseWriter, req *http.Request) {

    if err := req.ParseForm(); err != nil {
        log.Println("Error parsing form: " + req.URL.Host + "/" + req.URL.Path)
    }

    token := req.Form.Get("oauth_token")
    verifier := req.Form.Get("oauth_verifier")

    log.Printf("Token = %v, verifier = %v", token, verifier)

    if token != "" && verifier != "" {
        // get the matching EverpocketCreds
        creds, _ := data.GetEverpocketCreds(map[string]string{
            "ev_temp_request_token": token,
        })

        log.Printf("Creds = %v", creds)
        if creds.EvTempRequestToken != token {
            log.Fatal("Error querying DB for token ", token, ".")
        }

        // use the verifier to exchange for the access token
        accessToken, accessSecret, addData, err := auth.GetEvernoteAccessToken(
            EVERNOTE_HOST,
            creds.EvTempRequestToken,
            creds.EvTempSecret,
            verifier,
            true,
        )

        if err != nil {
            log.Fatal("Error getting evernote access token: %v", err)
        }

        // TODO: once we have this set of info, we can write a goroutine to 
        // write this to the database
        // and we just return OK so that we can do the next thing
        go func() {
            creds.EvAccessSecret = accessSecret
            creds.EvAccessToken = accessToken
            creds.EvAddData = addData
            err := creds.Write()
            if err != nil {
                log.Fatal("Failed to write evernote data: %v", err)
            }
        }()

        donePageUrl := fmt.Sprintf(DONEPAGE, "/connect_pocket?cid=" + strconv.Itoa(creds.Id))

        res.Write([]byte(donePageUrl))
    }

}

func PocketCallbackHandler(res http.ResponseWriter, req *http.Request) {
}

func ConnectPocket(res http.ResponseWriter, req *http.Request) {

    // Get the credentials Id 
    var credId string

    if err := req.ParseForm(); err != nil {
        log.Fatal("Error parsing form")
    } else {
        credId := req.Form.Get("cid")
    }

    // Get the Pocket Request Token
    if redirectUri := os.Getenv(ENV_PO_SVC_CALLBACK); redirectUri != "" {
        token := pauth.GetPocketRequestToken(redirectUri)
        /*intCredId, err := strconv.Atoi(credId)*/
        /*if err != nil {*/
            creds, err := data.GetEverpocketCreds(map[string]string{ "creds_id": credId })
            if err != nil {
                // no error, query success
                creds.PoRequestToken = token
                creds.Write()
            }
        /*}*/
    }
}

func Login(res http.ResponseWriter, req *http.Request) {
    // Get temporary credentials for Evernote
    url, err := InitiateEvernoteOauth()
    if err != nil {
        log.Fatal("Could not get temp request token")
    }
    res.Header().Set("Location", url)
    res.WriteHeader(http.StatusFound)
}

// Home returns the page that gives users 
func Home() string {
    return fmt.Sprintf(FRONTPAGE, "/login")
}

func Done() string {
    return DONEPAGE
}

func SetupServer() {

    m := martini.Classic()

    m.Get("/", Home)
    // Login page
    m.Get("/login", Login)
    //Redirect page: the page that evernote redirects to after success
    m.Get("/evernote_done", EvernoteCallbackHandler)

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

        // if our table already exists, ignore. otherwise throw
        // big error.
        if !strings.Contains(errMsg, "already exists") {
            // big error
            log.Fatalf("Database error: %s", errMsg)
        }
    }

    SetupServer()

}

