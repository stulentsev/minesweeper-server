package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type apiError struct {
	Error string `json:"error"`
}

type MoveInfo struct {
	GameId string `json:"game_id"`
	X      uint32 `json:"x"`
	Y      uint32 `json:"y"`
}

func main() {
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
	knownGames[game.Id] = game
	renderJson(writer, game)
	game.DebugPrint()
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
		renderJson(writer, apiError{err.Error()})
		return
	}
	fmt.Printf("%+v\n", moveInfo)

	var gameId string

	gameId = moveInfo.GameId
	if gameId == "" {
		writer.WriteHeader(http.StatusBadRequest)
		renderJson(writer, apiError{"must provide a valid game_id"})
		return
	}

	game, exists := knownGames[gameId]
	if !exists {
		writer.WriteHeader(http.StatusNotFound)
		renderJson(writer, apiError{"game is not found"})
		return
	}
	err = game.Open(moveInfo.X, moveInfo.Y)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		renderJson(writer, apiError{err.Error()})
		return
	}

	fmt.Println("finish open")
	game.DebugPrint()

	renderJson(writer, game)
}

func renderJson(w http.ResponseWriter, payload interface{}) {
	encoder := json.NewEncoder(w)
	_ = encoder.Encode(payload)
}
