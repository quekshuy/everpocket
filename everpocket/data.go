package everpocket

import (
    "encoding/json"

    "appengine"
    "appengine/datastore"
    "appengine/memcache"
)

type EverpocketData interface {
    Write()
}


// We specify the various structs to store in the GAE Data Store
// here.

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

// TO USE LATER
type EverpocketQueryRecord struct {
    // record the last time we queried for a user
}

func (cred *EverpocketCreds) key() string {
    return "EV_" + cred.EvTempRequestToken
}

func everpocketCredsKey(c *appengine.Context) {
    return datastore.NewKey(c, "EverpocketCreds", "default_creds", 0, nil)
}


// Write will write to the memcache as well as the 
// as DataStore
func (cred *EverpocketCreds) Write(c *appengine.Context) error {

    // write to memcache first
    item := &memcache.Item{
        Key: cred.key(),
        Value: json.Marshal(*cred),
    }

    if err := memcache.Add(c, item); err == memcache.ErrNotStored {
        // then we want to set it instead
        if err = memcache.Set(c, item);  err != nil {
            c.Errorf("Error setting item: %v", err)
        }
    } else if err != nil {
        // another error
        c.Errorf("Unknown memcache error: %v", err)

    }

    if err == nil {
        // everything successful
        // do the write
        key := datastore.NewIncompleteKey(c, "EverpocketCreds", everpocketCredsKey(c))
        _, err := datastore.Put(c, key, cred)
        return err
    }
    return nil
}


