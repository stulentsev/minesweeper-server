package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"io"
	"math/rand"
	"os"
	"strconv"
	"strings"
)

type CellState string

const (
	CellStateUnknown = CellState("?")
	CellStateBomb    = CellState("*")
	CellStateEmpty   = CellState(" ")
)

var knownGames = make(map[string]*Game)

type Game struct {
	Id                 string      `json:"game_id"`
	Status             string      `json:"status,omitempty"`
	BoardWidth         uint32      `json:"board_width"`
	BoardHeight        uint32      `json:"board_height"`
	MinesCount         uint32      `json:"mines_count"`
	RevealedBoardState []CellState `json:"board_state"`
	trueBoardState     []CellState

	PrettyBoardState string `json:"pretty_board_state"`
}

func (game *Game) InitBoardState() {
	game.trueBoardState = make([]CellState, game.BoardWidth*game.BoardHeight)
	for b := uint32(0); b < game.MinesCount; {
		if b >= game.MinesCount {
			break
		}

		x := rand.Intn(int(game.BoardWidth))
		y := rand.Intn(int(game.BoardHeight))

		offset := y*int(game.BoardWidth) + x
		if game.trueBoardState[offset] != CellStateBomb {
			game.trueBoardState[offset] = CellStateBomb
			b++
		}
	}
	for idx := range game.trueBoardState {
		if game.trueBoardState[idx] != CellStateBomb {
			game.trueBoardState[idx] = CellStateEmpty
		}
	}

	game.RevealedBoardState = make([]CellState, len(game.trueBoardState))

	for idx := range game.RevealedBoardState {
		game.RevealedBoardState[idx] = CellStateUnknown
	}

	game.prettyPrintBoard()

}

func (game *Game) prettyPrintBoard() {
	buf := bytes.Buffer{}
	printBoardState(&buf, game, game.RevealedBoardState)
	game.PrettyBoardState = buf.String()
}
func (game *Game) Open(x, y uint32) error {
	fmt.Printf("open: x %d, y %d\n", x, y)
	if game.Status != "" {
		return errors.New("can't play a finished game")
	}
	offset := y*game.BoardWidth + x
	cellState := game.trueBoardState[int(offset)]
	if cellState == CellStateBomb {
		game.Status = "loss"
		game.revealBombs()
	}

	if cellState == CellStateEmpty {
		game.revealEmptyAt(x, y)
	}

	if game.onlyBombsRemainHidden() {
		game.Status = "win"
	}

	game.prettyPrintBoard()
	return nil
}

func (game *Game) DebugPrint() {
	fmt.Println("<<<<<<<<<<<<<<<<")
	fmt.Println("Game", game.Id)
	fmt.Printf("board %d x %d, with %d mines\n", game.BoardWidth, game.BoardHeight, game.MinesCount)

	fmt.Println("Revealed state")
	printBoardState(os.Stdout, game, game.RevealedBoardState)

	fmt.Println("True state")
	printBoardState(os.Stdout, game, game.trueBoardState)

	fmt.Println(">>>>>>>>>>>>>>>>")
}

func (game *Game) revealBombs() {
	for idx, cellState := range game.trueBoardState {
		if cellState == CellStateBomb {
			game.RevealedBoardState[idx] = cellState
		}
	}
}

func (game *Game) revealEmptyAt(x, y uint32) {
	if x < 0 || y < 0 || x >= game.BoardWidth || y >= game.BoardHeight {
		return
	}
	offset := y*game.BoardWidth + x

	if game.RevealedBoardState[offset] != CellStateUnknown {
		return
	}

	bombCount := game.bombsAt(x-1, y-1) +
		game.bombsAt(x, y-1) +
		game.bombsAt(x+1, y-1) +

		game.bombsAt(x-1, y) +
		game.bombsAt(x+1, y) +

		game.bombsAt(x-1, y+1) +
		game.bombsAt(x, y+1) +
		game.bombsAt(x+1, y+1)

	if bombCount == 0 { // spread
		game.RevealedBoardState[offset] = CellStateEmpty

		game.revealEmptyAt(x-1, y-1)
		game.revealEmptyAt(x, y-1)
		game.revealEmptyAt(x+1, y-1)

		game.revealEmptyAt(x-1, y)
		game.revealEmptyAt(x+1, y)

		game.revealEmptyAt(x-1, y+1)
		game.revealEmptyAt(x, y+1)
		game.revealEmptyAt(x+1, y+1)
	} else {
		game.RevealedBoardState[offset] = CellState(strconv.Itoa(bombCount))
	}
}

func (game *Game) bombsAt(x, y uint32) int {
	if x < 0 || y < 0 || x >= game.BoardWidth || y >= game.BoardHeight {
		return 0
	}
	offset := y*game.BoardWidth + x

	if game.trueBoardState[offset] == CellStateBomb {
		return 1
	}
	return 0
}

func (game *Game) onlyBombsRemainHidden() bool {
	for idx, cellState := range game.RevealedBoardState {
		if cellState == CellStateUnknown {
			if game.trueBoardState[idx] != CellStateBomb {
				return false
			}
		}
	}
	return true
}

func printBoardState(w io.Writer, game *Game, states []CellState) {
	leftTopCorner := "\u250c"
	rightTopCorner := "\u2510"
	leftBottomCorner := "\u2514"
	rightBottomCorner := "\u2518"

	horizontalLine := "\u2500"
	verticalLine := "\u2502"

	fmt.Fprintf(w, "%s%s%s\n", leftTopCorner, strings.Repeat(horizontalLine, int(game.BoardWidth)*2), rightTopCorner)
	for i := 0; i < int(game.BoardHeight); i++ {
		fmt.Fprint(w, verticalLine)
		for j := 0; j < int(game.BoardWidth); j++ {
			idx := j + i*int(game.BoardHeight)
			fmt.Fprint(w, states[idx])
			fmt.Fprint(w, " ")
		}
		fmt.Fprintln(w, verticalLine)
	}
	fmt.Fprintf(w, "%s%s%s\n", leftBottomCorner, strings.Repeat(horizontalLine, int(game.BoardWidth)*2), rightBottomCorner)

}

func newGameId() string {
	randomId, err := uuid.NewUUID()
	if err != nil {
		panic(err)
	}
	return randomId.String()
}

func NewGame() *Game {
	id := newGameId()
	game := &Game{
		Id:          id,
		BoardWidth:  8,
		BoardHeight: 8,
		MinesCount:  10,
	}
	game.InitBoardState()
	return game
}
