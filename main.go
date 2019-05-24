package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"
)

type apiError struct {
	Error string `json:"error"`
}

// MoveInfo describes data about player's move
type MoveInfo struct {
	GameID string `json:"game_id"`
	X      uint32 `json:"x"`
	Y      uint32 `json:"y"`
}

func main() {
	rand.Seed(time.Now().Unix())

	http.HandleFunc("/newgame", newgameHandler)
	http.HandleFunc("/move", moveHandler)
	http.HandleFunc("/", readmeHandler)

	envPort := os.Getenv("PORT")
	if len(envPort) == 0 {
		envPort = "3000"
	}

	err := http.ListenAndServe(":"+envPort, nil)
	if err != nil {
		panic(err)
	}
}

func newgameHandler(writer http.ResponseWriter, request *http.Request) {
	game := NewGame()
	knownGames[game.ID] = game
	renderJSON(writer, game)
	//game.DebugPrint()
}

func readmeHandler(writer http.ResponseWriter, request *http.Request) {
	fmt.Fprintln(writer, "make a POST to /newgame")

}

func moveHandler(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	decoder := json.NewDecoder(request.Body)
	moveInfo := MoveInfo{}
	err := decoder.Decode(&moveInfo)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		renderJSON(writer, apiError{err.Error()})
		return
	}
	fmt.Printf("%+v\n", moveInfo)

	var gameID string

	gameID = moveInfo.GameID
	if gameID == "" {
		writer.WriteHeader(http.StatusBadRequest)
		renderJSON(writer, apiError{"must provide a valid game_id"})
		return
	}

	game, exists := knownGames[gameID]
	if !exists {
		writer.WriteHeader(http.StatusNotFound)
		renderJSON(writer, apiError{"game is not found"})
		return
	}
	err = game.Open(int(moveInfo.X), int(moveInfo.Y))
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		renderJSON(writer, apiError{err.Error()})
		return
	}

	//game.DebugPrint()

	renderJSON(writer, game)
}

func renderJSON(w http.ResponseWriter, payload interface{}) {
	js, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}
