package main

import (
	"log"
	"os"

	"github.com/gdamore/tcell/v2"
)

type Piece struct {
	Block     Block
	BaseShape [][]bool
	Rotation  int8
	X         int
	Y         int
}

func rotateBoolArrayOnce(shape [][]bool) [][]bool {
	sideLength := len(shape)

	newShape := make([][]bool, sideLength)
	for i := range newShape {
		newShape[i] = make([]bool, sideLength)
	}

	for row, rowState := range shape {
		for col, state := range rowState {
			if !state {
				continue
			}

			newShape[col][sideLength-1-row] = true
		}
	}

	return newShape
}

func (p Piece) GetRotated(r int8) [][]bool {
	// nth row becomes len-1th column

	shape := p.BaseShape
	for range r {
		shape = rotateBoolArrayOnce(shape)
	}
	return shape
}

func (p Piece) GetBlocks() [][2]int {
	shape := p.GetRotated(p.Rotation)
	var pieceBlocks [][2]int

	for row, rowState := range shape {
		for col, active := range rowState {
			if active {
				pieceBlocks = append(pieceBlocks, [2]int{p.Y + row, p.X + col})
			}
		}
	}

	return pieceBlocks
}

type Block struct {
	Color tcell.Color
}

var blocks = map[string]Block{
	"I": {Color: tcell.ColorCadetBlue},
	"L": {Color: tcell.ColorBlue},
	"J": {Color: tcell.ColorOrange},
	"O": {Color: tcell.ColorYellow},
	"S": {Color: tcell.ColorGreen},
	"T": {Color: tcell.ColorPurple},
	"Z": {Color: tcell.ColorRed},
}

var pieces = map[string]Piece{
	"I": {Block: blocks["I"], Y: -1, X: 3, Rotation: 0, BaseShape: [][]bool{
		{false, false, false, false},
		{true, true, true, true},
		{false, false, false, false},
		{false, false, false, false},
	}},
}

type GameState struct {
	Board       [20][10]*Block
	ActivePiece Piece
}

const titleHeight = 2

const boardWidth = 21
const boardHeight = 21

var boardCoords = [4]int{0, titleHeight + 1, boardWidth, titleHeight + 1 + boardHeight}

func main() {
	s, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("%+v", err)
	}

	if err := s.Init(); err != nil {
		log.Fatalf("%+v", err)
	}

	defStyle := tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset)
	boxStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorReset)
	textStyle := tcell.StyleDefault.Bold(true).Foreground(tcell.ColorPurple).Background(tcell.ColorReset)

	s.SetStyle(defStyle)

	// tetris bounding box
	drawBox(s, 0, 0, boardWidth, titleHeight, boxStyle)
	drawTextCentered(s, 0, 0, boardWidth, titleHeight, textStyle, "TETRIS")

	drawBox(s, boardCoords[0], boardCoords[1], boardCoords[2], boardCoords[3], boxStyle)

	quit := func() {
		s.Fini()
		os.Exit(0)
	}
	defer quit()

	gameState := GameState{ActivePiece: pieces["I"]}
	// p := blocks["L"]
	// gameState.Board[0][0] = &p

	for {
		s.Show()

		ev := s.PollEvent()
		gameState.ActivePiece.Rotation += 1
		gameState.ActivePiece.Rotation %= 3
		switch ev := ev.(type) {
		case *tcell.EventResize:
			s.Sync()
		case *tcell.EventKey:
			switch key := ev.Key(); key {
			case tcell.KeyEscape, tcell.KeyCtrlC:
				quit()
			}
		}

		drawBoard(s, gameState)
	}
}

func drawBoard(s tcell.Screen, gameState GameState) {
	for row, rowPieces := range gameState.Board {
		for col, piece := range rowPieces {
			var style tcell.Style
			if piece == nil {
				drawOnBoard(s, row, col, ' ', tcell.StyleDefault)
				continue
			} else {
				style = tcell.StyleDefault.Foreground(piece.Color)
			}

			drawOnBoard(s, row, col, tcell.RuneBlock, style)
		}
	}

	for _, coords := range gameState.ActivePiece.GetBlocks() {
		style := tcell.StyleDefault.Foreground(gameState.ActivePiece.Block.Color)
		drawOnBoard(s, coords[0], coords[1], tcell.RuneBlock, style)
	}
}

func drawOnBoard(s tcell.Screen, row int, col int, char rune, style tcell.Style) {
	termCol := boardCoords[0] + col*2 + 1
	termRow := boardCoords[1] + row + 1

	s.SetContent(termCol, termRow, char, nil, style)
	s.SetContent(termCol+1, termRow, char, nil, style)
}

func drawBox(s tcell.Screen, x1, y1, x2, y2 int, style tcell.Style) {
	if y2 < y1 {
		y1, y2 = y2, y1
	}
	if x2 < x1 {
		x1, x2 = x2, x1
	}

	// Fill background
	for row := y1; row <= y2; row++ {
		for col := x1; col <= x2; col++ {
			s.SetContent(col, row, ' ', nil, style)
		}
	}

	// Draw borders
	for col := x1; col <= x2; col++ {
		s.SetContent(col, y1, tcell.RuneHLine, nil, style)
		s.SetContent(col, y2, tcell.RuneHLine, nil, style)
	}
	for row := y1 + 1; row < y2; row++ {
		s.SetContent(x1, row, tcell.RuneVLine, nil, style)
		s.SetContent(x2, row, tcell.RuneVLine, nil, style)
	}

	// Only draw corners if necessary
	if y1 != y2 && x1 != x2 {
		s.SetContent(x1, y1, tcell.RuneULCorner, nil, style)
		s.SetContent(x2, y1, tcell.RuneURCorner, nil, style)
		s.SetContent(x1, y2, tcell.RuneLLCorner, nil, style)
		s.SetContent(x2, y2, tcell.RuneLRCorner, nil, style)
	}
}

func drawTextCentered(s tcell.Screen, x1, y1, x2, y2 int, style tcell.Style, text string) {
	if y2 < y1 {
		y1, y2 = y2, y1
	}
	if x2 < x1 {
		x1, x2 = x2, x1
	}

	textLenOffset := len(text) / -2
	centerX := (x2 - x1) / 2
	centerY := (y2 - y1) / 2

	col := centerX + textLenOffset
	for _, r := range []rune(text) {
		s.SetContent(col, centerY, r, nil, style)
		col++
	}
}

func drawText(s tcell.Screen, x1, y1, x2, y2 int, style tcell.Style, text string) {
	row := y1
	col := x1
	for _, r := range []rune(text) {
		s.SetContent(col, row, r, nil, style)
		col++
		if col >= x2 {
			row++
			col = x1
		}
		if row > y2 {
			break
		}
	}
}
