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

// CellState is a kind of enum
type CellState string

const (
	cellStateUnknown = CellState("?")
	cellStateBomb    = CellState("*")
	cellStateEmpty   = CellState(" ")
)

var knownGames = make(map[string]*Game)

// Game describes board state and dimensions
type Game struct {
	ID                 string      `json:"game_id"`
	Status             string      `json:"status,omitempty"`
	BoardWidth         int      `json:"board_width"`
	BoardHeight        int      `json:"board_height"`
	MinesCount         int      `json:"mines_count"`
	RevealedBoardState []CellState `json:"board_state"`
	trueBoardState     []CellState

	PrettyBoardState string `json:"pretty_board_state"`
}

func (game *Game) initRevealedBoardState() {
	game.RevealedBoardState = make([]CellState, game.BoardWidth*game.BoardHeight)

	for idx := range game.RevealedBoardState {
		game.RevealedBoardState[idx] = cellStateUnknown
	}

	game.prettyPrintBoard()
}

func (game *Game) initTrueBoardState(firstClickX, firstClickY int) {
	game.trueBoardState = make([]CellState, game.BoardWidth*game.BoardHeight)
	b := 0
	for  {
		if b >= game.MinesCount {
			break
		}

		x := rand.Intn(int(game.BoardWidth))
		y := rand.Intn(int(game.BoardHeight))
		if areCellsAdjacent(x, y, int(firstClickX), int(firstClickY)) { // guaranteed empty space at first click
			continue
		}

		offset := y*int(game.BoardWidth) + x
		if game.trueBoardState[offset] != cellStateBomb {
			game.trueBoardState[offset] = cellStateBomb
			b++
		}
	}
	for idx := range game.trueBoardState {
		if game.trueBoardState[idx] != cellStateBomb {
			game.trueBoardState[idx] = cellStateEmpty
		}
	}
}

func areCellsAdjacent(x1 int, y1 int, x2 int, y2 int) bool {
	return absInt(x1-x2) <= 1 && absInt(y1-y2) <= 1
}

func absInt(num int) int {
	if num < 0 {
		return -num
	}
	return num
}

func (game *Game) prettyPrintBoard() {
	buf := bytes.Buffer{}
	printBoardState(&buf, game, game.RevealedBoardState)
	game.PrettyBoardState = buf.String()
}

// Open reveals a cell on the board
func (game *Game) Open(x, y int) error {
	if game.trueBoardState == nil {
		game.initTrueBoardState(x, y)
	}

	if game.Status != "" {
		return errors.New("can't play a finished game")
	}
	offset := y*game.BoardWidth + x
	fmt.Printf("open: x %d, y %d\n", x, y)

	cellState := game.trueBoardState[offset]
	if cellState == cellStateBomb {
		game.Status = "loss"
		game.revealBombs()
	}

	if cellState == cellStateEmpty {
		game.revealEmptyAt(x, y)
	}

	if game.onlyBombsRemainHidden() {
		game.Status = "win"
	}

	game.prettyPrintBoard()
	return nil
}

// DebugPrint pretty-prints game state (revealed/true state, etc.)
func (game *Game) DebugPrint() {
	fmt.Println("<<<<<<<<<<<<<<<<")
	fmt.Println("Game", game.ID)
	fmt.Printf("board %d x %d, with %d mines\n", game.BoardWidth, game.BoardHeight, game.MinesCount)

	if game.RevealedBoardState != nil {
		fmt.Println("Revealed state")
		printBoardState(os.Stdout, game, game.RevealedBoardState)
	}

	if game.trueBoardState != nil {
		fmt.Println("True state")
		printBoardState(os.Stdout, game, game.trueBoardState)
	}

	fmt.Println(">>>>>>>>>>>>>>>>")
}

func (game *Game) revealBombs() {
	for idx, cellState := range game.trueBoardState {
		if cellState == cellStateBomb {
			game.RevealedBoardState[idx] = cellState
		}
	}
}

func (game *Game) revealEmptyAt(x, y int) {
	if x < 0 || y < 0 || x >= game.BoardWidth || y >= game.BoardHeight {
		return
	}
	offset := y*game.BoardWidth + x

	if game.RevealedBoardState[offset] != cellStateUnknown {
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
		game.RevealedBoardState[offset] = cellStateEmpty

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

func (game *Game) bombsAt(x, y int) int {
	if x < 0 || y < 0 || x >= game.BoardWidth || y >= game.BoardHeight {
		return 0
	}
	offset := y*game.BoardWidth + x

	if game.trueBoardState[offset] == cellStateBomb {
		return 1
	}
	return 0
}

func (game *Game) onlyBombsRemainHidden() bool {
	for idx, cellState := range game.RevealedBoardState {
		if cellState == cellStateUnknown {
			if game.trueBoardState[idx] != cellStateBomb {
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
			idx := j + i*int(game.BoardWidth)
			fmt.Fprint(w, states[idx])
			fmt.Fprint(w, " ")
		}
		fmt.Fprintln(w, verticalLine)
	}
	fmt.Fprintf(w, "%s%s%s\n", leftBottomCorner, strings.Repeat(horizontalLine, int(game.BoardWidth)*2), rightBottomCorner)

}

func newGameID() string {
	randomID, err := uuid.NewUUID()
	if err != nil {
		panic(err)
	}
	return randomID.String()
}

// NewGame creates a new board
func NewGame() *Game {
	id := newGameID()
	game := &Game{
		ID:          id,
		BoardWidth: 8,
		BoardHeight: 8,
		MinesCount:  10,
	}
	game.initRevealedBoardState()
	return game
}
