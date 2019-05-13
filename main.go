package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	http.HandleFunc("/newgame", newgameHandler)
	http.HandleFunc("/move", moveHandler)
	http.HandleFunc("/", readmeHandler)

	envPort := os.Getenv("PORT")
	if len(envPort) == 0 {
		envPort = "3000"
	}

	err := http.ListenAndServe(":" + envPort, nil)
	if err != nil {
		panic(err)
	}
}

func newgameHandler(writer http.ResponseWriter, request *http.Request) {
	fmt.Fprintln(writer, "new game")
}

func readmeHandler(writer http.ResponseWriter, request *http.Request) {
	fmt.Fprintln(writer, "make a POST to /newgame")

}

func moveHandler(writer http.ResponseWriter, request *http.Request) {
	fmt.Fprintln(writer, "here's your new board state")
}
