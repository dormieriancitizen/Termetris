package main

import (
	"log"
	"math/rand"
	"os"
	"time"

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

func (piece Piece) MoveX(board [20][10]*Block, offset int) Piece {
	prevX := piece.X

	piece.X += offset
	if checkOverlap(board, piece.GetBlocks()) {
		piece.X = prevX
	}

	return piece
}

func (piece Piece) MoveY(board [20][10]*Block, offset int) (Piece, bool) {
	prevY := piece.Y

	piece.Y += offset
	if checkOverlap(board, piece.GetBlocks()) {
		piece.Y = prevY
		return piece, true
	}

	return piece, false
}

func (piece Piece) Rotate(board [20][10]*Block, offset int8) Piece {
	prevRot := piece.Rotation

	piece.Rotation += offset
	piece.Rotation = piece.Rotation % 4

	if checkOverlap(board, piece.GetBlocks()) {
		// wallkick implementation later
		piece.Rotation = prevRot
	}

	return piece
}

type Block struct {
	Color tcell.Color
}

var blocks = map[string]Block{
	"I": {Color: tcell.ColorCadetBlue},
	"L": {Color: tcell.ColorOrange},
	"J": {Color: tcell.ColorBlue},
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
	"L": {Block: blocks["L"], Y: -1, X: 3, Rotation: 0, BaseShape: [][]bool{
		{false, false, true},
		{true, true, true},
		{false, false, false},
	}},
	"J": {Block: blocks["J"], Y: -1, X: 3, Rotation: 0, BaseShape: [][]bool{
		{true, false, false},
		{true, true, true},
		{false, false, false},
	}},
	"O": {Block: blocks["O"], Y: -1, X: 3, Rotation: 0, BaseShape: [][]bool{
		{false, false, false, false},
		{false, true, true, false},
		{false, true, true, false},
		{false, false, false, false},
	}},
	"S": {Block: blocks["S"], Y: -1, X: 3, Rotation: 0, BaseShape: [][]bool{
		{false, true, true},
		{true, true, false},
		{false, false, false},
	}},
	"Z": {Block: blocks["Z"], Y: -1, X: 3, Rotation: 0, BaseShape: [][]bool{
		{true, true, false},
		{false, true, true},
		{false, false, false},
	}},
	"T": {Block: blocks["T"], Y: -1, X: 3, Rotation: 0, BaseShape: [][]bool{
		{false, true, false},
		{true, true, true},
		{false, false, false},
	}},
}

type GameState struct {
	Board       [20][10]*Block
	ActivePiece Piece
	HeldPiece   *Piece
	Frame       int64
	HasHeld     bool
	Lost        bool
}

func checkOverlap(board [20][10]*Block, shape [][2]int) bool {
	for _, block := range shape {
		y, x := block[0], block[1]
		if y >= 20 || x >= 10 || y < -2 || x < 0 {
			return true
		}
		if y >= 0 && board[y][x] != nil {
			return true
		}
	}
	return false
}

func placePiece(board [20][10]*Block, piece Piece) [20][10]*Block {
	for _, coords := range piece.GetBlocks() {
		row, col := coords[0], coords[1]

		if row >= 20 || col >= 10 || row < 0 || col < 0 {
			continue
		}
		p := piece.Block
		board[row][col] = &p
	}
	return board
}

func getRandomPiece() Piece {
	keys := make([]string, 0, len(pieces))
	for k := range pieces {
		keys = append(keys, k)
	}
	randKey := keys[rand.Intn(len(keys))]
	return pieces[randKey]
}

func (gameState GameState) nextPiece() GameState {
	gameState.ActivePiece = getRandomPiece()
	gameState.HasHeld = false

	if checkOverlap(gameState.Board, gameState.ActivePiece.GetBlocks()) {
		gameState.Lost = true
	}

	return gameState
}

func (gameState GameState) clearFilledLines() GameState {
	for rowIndice, row := range gameState.Board {
		rowFull := true
		for _, block := range row {
			if block == nil {
				rowFull = false
				break
			}
		}
		if rowFull {
			// Shift all rows above down by one
			for i := rowIndice; i > 0; i-- {
				gameState.Board[i] = gameState.Board[i-1]
			}
			// Clear top row
			for col := range len(gameState.Board[0]) {
				gameState.Board[0][col] = nil
			}
		}
	}
	return gameState
}

const titleHeight = 2

const boardWidth = 21
const boardHeight = 21

const fps = 60
const frameDuration = time.Second / fps

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

	gameState := GameState{}
	gameState = gameState.nextPiece()
	// p := blocks["L"]
	// gameState.Board[0][0] = &p

	evCh := make(chan tcell.Event)
	quitCh := make(chan struct{})
	go s.ChannelEvents(evCh, quitCh)

	for {
		s.Show()
		startTime := time.Now()
		gameState.Frame += 1

		if gameState.Lost {
			quit()
		}

		hitBottom := false
		select {
		case ev := <-evCh:
			if ev != nil {
				switch ev := ev.(type) {
				case *tcell.EventResize:
					s.Sync()
				case *tcell.EventKey:
					switch key := ev.Key(); key {
					case tcell.KeyEscape, tcell.KeyCtrlC:
						quit()
					case tcell.KeyUp:
						gameState.ActivePiece = gameState.ActivePiece.Rotate(gameState.Board, 1)
					case tcell.KeyLeft:
						gameState.ActivePiece = gameState.ActivePiece.MoveX(gameState.Board, -1)
					case tcell.KeyRight:
						gameState.ActivePiece = gameState.ActivePiece.MoveX(gameState.Board, 1)
					case tcell.KeyDown:
						gameState.ActivePiece, hitBottom = gameState.ActivePiece.MoveY(gameState.Board, 1)
					case tcell.KeyRune:
						switch ev.Rune() {
						case 'z':
							gameState.ActivePiece = gameState.ActivePiece.Rotate(gameState.Board, 3)
						case ' ':
							for !hitBottom {
								gameState.ActivePiece, hitBottom = gameState.ActivePiece.MoveY(gameState.Board, 1)
							}
						case 'c':
							if gameState.HasHeld {
								break
							}
							if gameState.HeldPiece != nil {
								heldCopy := gameState.HeldPiece
								activeCopy := gameState.ActivePiece

								gameState.ActivePiece, gameState.HeldPiece = *heldCopy, &activeCopy
							} else {
								heldCopy := gameState.ActivePiece
								gameState.HeldPiece = &heldCopy
								gameState = gameState.nextPiece()
							}
							gameState.HasHeld = true
						}
					}
				}
			}
		default:

		}

		if gameState.Frame%64 == 0 {
			gameState.ActivePiece, hitBottom = gameState.ActivePiece.MoveY(gameState.Board, 1)
		}

		if hitBottom || checkOverlap(gameState.Board, gameState.ActivePiece.GetBlocks()) {
			drawTextCentered(s, 0, 0, boardWidth, titleHeight, textStyle, "OVERLAP")
			gameState.Board = placePiece(gameState.Board, gameState.ActivePiece)
			gameState = gameState.clearFilledLines()

			gameState = gameState.nextPiece()
		} else {
			drawTextCentered(s, 0, 0, boardWidth, titleHeight, textStyle, "TETRIS")
		}

		drawBoard(s, gameState)

		elapsed := time.Since(startTime)
		if elapsed < frameDuration {
			time.Sleep(frameDuration - elapsed)
		}
	}
}

func drawBoard(s tcell.Screen, gameState GameState) {
	for row, rowBlocks := range gameState.Board {
		for col, block := range rowBlocks {
			var style tcell.Style
			if block == nil {
				drawOnBoard(s, row, col, ' ', tcell.StyleDefault)
				continue
			} else {
				style = tcell.StyleDefault.Foreground(block.Color).Background(block.Color)
			}

			drawOnBoard(s, row, col, tcell.RuneBlock, style)
		}
	}

	prevY := gameState.ActivePiece.Y

	hitBottom := false
	for !hitBottom {
		gameState.ActivePiece, hitBottom = gameState.ActivePiece.MoveY(gameState.Board, 1)
	}

	//ghost
	for _, coords := range gameState.ActivePiece.GetBlocks() {
		style := tcell.StyleDefault.Foreground(gameState.ActivePiece.Block.Color)
		drawOnBoard(s, coords[0], coords[1], 'o', style)
	}

	gameState.ActivePiece.Y = prevY

	if gameState.HeldPiece != nil {
		style := tcell.StyleDefault.Foreground(gameState.HeldPiece.Block.Color)
		for row, rowBlocks := range gameState.HeldPiece.BaseShape {
			for col, active := range rowBlocks {
				termCol := boardCoords[2] + col*2 + 1
				termRow := boardCoords[0] + row + 1
				if !active {
					s.SetContent(termCol, termRow, ' ', nil, tcell.StyleDefault)
					continue
				}

				s.SetContent(termCol, termRow, tcell.RuneBlock, nil, style)
				s.SetContent(termCol+1, termRow, tcell.RuneBlock, nil, style)
			}
		}
	}

	for _, coords := range gameState.ActivePiece.GetBlocks() {
		style := tcell.StyleDefault.Foreground(gameState.ActivePiece.Block.Color)
		drawOnBoard(s, coords[0], coords[1], tcell.RuneBlock, style)
	}
}

func drawOnBoard(s tcell.Screen, row int, col int, char rune, style tcell.Style) {
	if row >= 20 || col >= 10 || row < 0 || col < 0 {
		return
	}

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
