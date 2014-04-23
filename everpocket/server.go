package main

import (
    "data"
    "github.com/go-martini/martini"
)


func main() {
    m := martini.Classic()

    // Front page --> explains what happens
    // Login page
    // Done page
    // Options page (to cancel the connection, etc)
    // Total pages: 4

    m.Run()
}

