package main

import (
	"log"
	"math/rand"
	"os"
	"slices"
	"strconv"
	"time"

	"github.com/gdamore/tcell/v2"
)

type Piece struct {
	Block     Block
	BaseShape [][]bool
	Rotation  int8
	X         int
	Y         int
	Kick      map[[2]int][5][2]int
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

func (piece Piece) MoveX(board [20][10]*Block, offset int) (Piece, bool) {
	prevX := piece.X

	piece.X += offset
	if checkOverlap(board, piece.GetBlocks()) {
		piece.X = prevX
		return piece, true
	}

	return piece, false
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

func (piece Piece) Rotate(board [20][10]*Block, offset int8) (Piece, bool) {
	testPiece := piece
	testPiece.Rotation += offset
	testPiece.Rotation = testPiece.Rotation % 4
	if testPiece.Rotation < 0 {
		testPiece.Rotation += 4
	}

	kick_tests := piece.Kick[[2]int{int(piece.Rotation), int(testPiece.Rotation)}]

	for _, test := range kick_tests {
		testPiece.X += test[0]
		testPiece.Y -= test[1]

		if !checkOverlap(board, testPiece.GetBlocks()) {
			return testPiece, false
		}
	}

	return piece, true
}

type Block struct {
	Color tcell.Color
}

var blocks = map[string]Block{
	"I": {Color: tcell.NewRGBColor(129, 200, 190)},
	"L": {Color: tcell.NewRGBColor(239, 159, 118)},
	"J": {Color: tcell.NewRGBColor(140, 170, 238)},
	"O": {Color: tcell.NewRGBColor(229, 200, 144)},
	"S": {Color: tcell.NewRGBColor(166, 209, 137)},
	"T": {Color: tcell.NewRGBColor(202, 158, 230)},
	"Z": {Color: tcell.NewRGBColor(231, 130, 132)},
}

var JLSTZKick = map[[2]int][5][2]int{
	{0, 1}: {{0, 0}, {-1, 0}, {-1, 1}, {0, -2}, {-1, -2}}, // 0 -> R
	{1, 0}: {{0, 0}, {1, 0}, {1, -1}, {0, 2}, {1, 2}},     // R -> 0
	{1, 2}: {{0, 0}, {1, 0}, {1, -1}, {0, 2}, {1, 2}},     // R -> 2
	{2, 1}: {{0, 0}, {-1, 0}, {-1, 1}, {0, -2}, {-1, -2}}, // 2 -> R
	{2, 3}: {{0, 0}, {1, 0}, {1, 1}, {0, -2}, {1, -2}},    // 2 -> L
	{3, 2}: {{0, 0}, {-1, 0}, {-1, -1}, {0, 2}, {-1, 2}},  // L -> 2
	{3, 0}: {{0, 0}, {-1, 0}, {-1, -1}, {0, 2}, {-1, 2}},  // L -> 0
	{0, 3}: {{0, 0}, {1, 0}, {1, 1}, {0, -2}, {1, -2}},    // 0 -> L
}

var OKick = map[[2]int][5][2]int{
	{0, 1}: {{0, 0}, {0, 0}, {0, 0}, {0, 0}, {0, 0}},
	{1, 0}: {{0, 0}, {0, 0}, {0, 0}, {0, 0}, {0, 0}},
	{1, 2}: {{0, 0}, {0, 0}, {0, 0}, {0, 0}, {0, 0}},
	{2, 1}: {{0, 0}, {0, 0}, {0, 0}, {0, 0}, {0, 0}},
	{2, 3}: {{0, 0}, {0, 0}, {0, 0}, {0, 0}, {0, 0}},
	{3, 2}: {{0, 0}, {0, 0}, {0, 0}, {0, 0}, {0, 0}},
	{3, 0}: {{0, 0}, {0, 0}, {0, 0}, {0, 0}, {0, 0}},
	{0, 3}: {{0, 0}, {0, 0}, {0, 0}, {0, 0}, {0, 0}},
}

var IKick = map[[2]int][5][2]int{
	{0, 1}: {{0, 0}, {-2, 0}, {1, 0}, {-2, -1}, {1, 2}}, // 0 -> R
	{1, 0}: {{0, 0}, {2, 0}, {-1, 0}, {2, 1}, {-1, -2}}, // R -> 0
	{1, 2}: {{0, 0}, {-1, 0}, {2, 0}, {-1, 2}, {2, -1}}, // R -> 2
	{2, 1}: {{0, 0}, {1, 0}, {-2, 0}, {1, -2}, {-2, 1}}, // 2 -> R
	{2, 3}: {{0, 0}, {2, 0}, {-1, 0}, {2, 1}, {-1, -2}}, // 2 -> L
	{3, 2}: {{0, 0}, {-2, 0}, {1, 0}, {-2, -1}, {1, 2}}, // L -> 2
	{3, 0}: {{0, 0}, {1, 0}, {-2, 0}, {1, -2}, {-2, 1}}, // L -> 0
	{0, 3}: {{0, 0}, {-1, 0}, {2, 0}, {-1, 2}, {2, -1}}, // 0 -> L
}

var pieces = map[string]Piece{
	"I": {Block: blocks["I"], Y: -1, X: 3, Rotation: 0, Kick: IKick,
		BaseShape: [][]bool{
			{false, false, false, false},
			{true, true, true, true},
			{false, false, false, false},
			{false, false, false, false},
		},
	},
	"L": {Block: blocks["L"], Y: -1, X: 3, Rotation: 0, Kick: JLSTZKick,
		BaseShape: [][]bool{
			{false, false, true},
			{true, true, true},
			{false, false, false},
		},
	},
	"J": {Block: blocks["J"], Y: -1, X: 3, Rotation: 0, Kick: JLSTZKick,
		BaseShape: [][]bool{
			{true, false, false},
			{true, true, true},
			{false, false, false},
		},
	},
	"O": {Block: blocks["O"], Y: -1, X: 3, Rotation: 0, Kick: OKick,
		BaseShape: [][]bool{
			{false, false, false, false},
			{false, true, true, false},
			{false, true, true, false},
			{false, false, false, false},
		},
	},
	"S": {Block: blocks["S"], Y: -1, X: 3, Rotation: 0, Kick: JLSTZKick,
		BaseShape: [][]bool{
			{false, true, true},
			{true, true, false},
			{false, false, false},
		}},
	"Z": {Block: blocks["Z"], Y: -1, X: 3, Rotation: 0, Kick: JLSTZKick,
		BaseShape: [][]bool{
			{true, true, false},
			{false, true, true},
			{false, false, false},
		}},
	"T": {Block: blocks["T"], Y: -1, X: 3, Rotation: 0, Kick: JLSTZKick,
		BaseShape: [][]bool{
			{false, true, false},
			{true, true, true},
			{false, false, false},
		}},
}

type GameState struct {
	Board        [20][10]*Block
	ActivePiece  Piece
	HeldPiece    *Piece
	Frame        int64
	HasHeld      bool
	Lost         bool
	LockDelay    [3]int
	LockMove     bool
	Bag          []string
	Preview      []Piece
	Score        int
	ClearedLines int
	Level        int
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

func drawFromBag(bag []string) (Piece, []string) {
	randIndex := rand.Intn(len(bag))
	key := bag[randIndex]
	bag = slices.Delete(bag, randIndex, randIndex)

	return pieces[key], bag
}

func (gameState GameState) reset() GameState {
	gameState.LockMove = false
	gameState.HasHeld = false
	gameState.LockDelay = defaultLockDelay

	return gameState
}

func (gameState GameState) nextPiece() GameState {
	if len(gameState.Bag) == 0 {
		gameState.Bag = []string{"I", "O", "T", "S", "Z", "J", "L"}
	}

	gameState = gameState.reset()

	var newPiece Piece
	newPiece, gameState.Bag = drawFromBag(gameState.Bag)

	if len(gameState.Preview) == 0 {
		for range 7 {
			var piece Piece
			piece, gameState.Bag = drawFromBag(gameState.Bag)
			gameState.Preview = append(gameState.Preview, piece)
		}
	}

	gameState.ActivePiece = gameState.Preview[0]
	gameState.Preview = slices.Delete(gameState.Preview, 0, 1)

	gameState.Preview = append(gameState.Preview, newPiece)

	if checkOverlap(gameState.Board, gameState.ActivePiece.GetBlocks()) {
		gameState.Lost = true
	}

	return gameState
}

func (gameState GameState) clearFilledLines() (GameState, int) {
	clearedRows := 0
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
			clearedRows++
		}
	}
	return gameState, clearedRows
}

const titleHeight = 2

const boardWidth = 21
const boardHeight = 21

var defaultLockDelay = [3]int{30, 120, 1200}

const fps = 60
const frameDuration = time.Second / fps

var boardCoords = [4]int{0, titleHeight + 1, boardWidth, titleHeight + 1 + boardHeight}

var scoreMap = map[int]int{
	1: 100,
	2: 300,
	3: 500,
	4: 800,
}

const previewPaneWidth = 10
const previewPaneHeight = 18

var previewPaneCoords = [4]int{boardWidth + 1, 0, boardWidth + 1 + previewPaneWidth, previewPaneHeight}
var holdPaneCoords = [4]int{boardWidth + 1, titleHeight - 4 + boardHeight, boardWidth + 1 + previewPaneWidth, titleHeight + 1 + boardHeight}

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

	s.SetStyle(defStyle)

	// tetris bounding box
	// drawTextCentered(s, 0, 0, boardWidth, titleHeight, textStyle, "TETRIS")

	drawBox(s, boardCoords[0], boardCoords[1], boardCoords[2], boardCoords[3], boxStyle)

	quit := func() {
		s.Fini()
		os.Exit(0)
	}
	defer quit()

	gameState := GameState{Level: 1}
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

		instantLock := false
		rotated := false
		shifted := false
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
						var b bool
						gameState.ActivePiece, b = gameState.ActivePiece.Rotate(gameState.Board, 1)
						rotated = !b
					case tcell.KeyLeft:
						var b bool
						gameState.ActivePiece, b = gameState.ActivePiece.MoveX(gameState.Board, -1)
						shifted = !b
					case tcell.KeyRight:
						var b bool
						gameState.ActivePiece, b = gameState.ActivePiece.MoveX(gameState.Board, 1)
						shifted = !b
					case tcell.KeyDown:
						// soft drop doesnt lock instantly
						gameState.ActivePiece, _ = gameState.ActivePiece.MoveY(gameState.Board, 1)
					case tcell.KeyRune:
						switch ev.Rune() {
						case 'z':
							var b bool
							gameState.ActivePiece, b = gameState.ActivePiece.Rotate(gameState.Board, 3)
							rotated = !b
						case ' ':
							for !instantLock {
								gameState.ActivePiece, instantLock = gameState.ActivePiece.MoveY(gameState.Board, 1)
							}
						case 'c':
							if gameState.HasHeld {
								break
							}
							if gameState.HeldPiece != nil {
								heldCopy := gameState.HeldPiece
								activeCopy := gameState.ActivePiece

								gameState.ActivePiece, gameState.HeldPiece = *heldCopy, &activeCopy
								gameState = gameState.reset()
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

		var gravityTable = []int{
			48, // Level 0
			43,
			38,
			33,
			28,
			23,
			18,
			13,
			8,
			6,
			5,
			5,
			5,
			4,
			4,
			4,
			3,
			3,
			3,
			2,   // Level 19
			1.0, // Level 20+ (instant)
		}
		var gravity int
		if gameState.Level > 19 {
			gravity = 1
		} else {
			gravity = gravityTable[gameState.Level]
		}

		if gameState.Frame%int64(gravity) == 0 {
			gameState.ActivePiece, _ = gameState.ActivePiece.MoveY(gameState.Board, 1)
		}

		var onGround bool
		_, onGround = gameState.ActivePiece.MoveY(gameState.Board, 1)

		if onGround {
			if shifted {
				gameState.LockMove = true
			}
			if rotated {
				gameState.LockMove = true
				gameState.LockDelay[1] = defaultLockDelay[1]
			}

			gameState.LockDelay[2]--
			gameState.LockDelay[0]--
			gameState.LockDelay[1]--

			if !gameState.LockMove && gameState.LockDelay[0] <= 0 {
				instantLock = true
			} else if gameState.LockMove && gameState.LockDelay[1] <= 0 {
				instantLock = true
			} else if gameState.LockDelay[2] <= 0 {
				instantLock = true
			}
		}

		// drawTextCentered(s, 0, 0, boardWidth, titleHeight, textStyle, "OVERLAP")
		if instantLock {
			gameState.Board = placePiece(gameState.Board, gameState.ActivePiece)

			var clearedLines int
			gameState, clearedLines = gameState.clearFilledLines()
			gameState.Score = gameState.Level * scoreMap[clearedLines]
			gameState.ClearedLines += clearedLines
			gameState.Level = (gameState.ClearedLines / 10) + 1

			gameState = gameState.nextPiece()
		} else {
			// drawTextCentered(s, 0, 0, boardWidth, titleHeight, textStyle, "TETRIS")
		}

		drawBoard(s, gameState)

		elapsed := time.Since(startTime)
		if elapsed < frameDuration {
			time.Sleep(frameDuration - elapsed)
		}
	}
}

func drawBoard(s tcell.Screen, gameState GameState) {
	boxStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorReset)
	for row, rowBlocks := range gameState.Board {
		for col, block := range rowBlocks {
			var style tcell.Style
			if block == nil {
				drawCustomOnBoard(s, row, col, ' ', '.', tcell.StyleDefault)
				continue
			} else {
				style = tcell.StyleDefault.Foreground(block.Color).Background(block.Color)
			}

			drawOnBoard(s, row, col, tcell.RuneBlock, style)
		}
	}

	ghostPiece := gameState.ActivePiece
	hitBottom := false
	for !hitBottom {
		ghostPiece, hitBottom = ghostPiece.MoveY(gameState.Board, 1)
	}
	for _, coords := range ghostPiece.GetBlocks() {
		style := tcell.StyleDefault.Foreground(gameState.ActivePiece.Block.Color)
		drawCustomOnBoard(s, coords[0], coords[1], '0', '0', style)
	}

	drawBox(s, 0, 0, boardWidth, titleHeight, boxStyle)

	// drawTextCentered(s, 0, 0, boardWidth, titleHeight, boxStyle, strings.Trim(strings.Join(strings.Fields(fmt.Sprint(gameState.LockDelay)), ", "), "[]"))
	// drawTextCentered(s, 0, 0, boardWidth, titleHeight, boxStyle, strconv.Itoa(int(gameState.ActivePiece.Rotation)))
	drawTextCentered(s, 0, 0, boardWidth, 1, boxStyle, "LVL "+strconv.Itoa(gameState.Level))
	drawTextCentered(s, 0, 0, boardWidth, titleHeight, boxStyle, strconv.Itoa(gameState.Score))

	drawBox(s, previewPaneCoords[0], previewPaneCoords[1], previewPaneCoords[2], previewPaneCoords[3], boxStyle)

	for i, piece := range gameState.Preview {
		style := tcell.StyleDefault.Foreground(piece.Block.Color)

		if i > 4 {
			continue
		}

		for row, rowBlocks := range piece.BaseShape {
			for col, active := range rowBlocks {
				termCol := previewPaneCoords[0] + col*2 + 2
				termRow := previewPaneCoords[1] + row + 1 + i*3
				if !active {
					s.SetContent(termCol, termRow, ' ', nil, tcell.StyleDefault)
					s.SetContent(termCol+1, termRow, ' ', nil, tcell.StyleDefault)
					continue
				}

				s.SetContent(termCol, termRow, tcell.RuneBlock, nil, style)
				s.SetContent(termCol+1, termRow, tcell.RuneBlock, nil, style)
			}
		}
	}

	drawText(s, previewPaneCoords[0]+1, previewPaneCoords[1], previewPaneCoords[2], previewPaneCoords[0]+1, boxStyle, "PREVIEW")

	drawBox(s, holdPaneCoords[0], holdPaneCoords[1], holdPaneCoords[2], holdPaneCoords[3], boxStyle)

	if gameState.HeldPiece != nil {
		style := tcell.StyleDefault.Foreground(gameState.HeldPiece.Block.Color)
		for row, rowBlocks := range gameState.HeldPiece.BaseShape {
			for col, active := range rowBlocks {
				termCol := holdPaneCoords[0] + col*2 + 2
				termRow := holdPaneCoords[1] + row + 1
				if !active {
					s.SetContent(termCol, termRow, ' ', nil, tcell.StyleDefault)
					s.SetContent(termCol+1, termRow, ' ', nil, tcell.StyleDefault)
					continue
				}

				s.SetContent(termCol, termRow, tcell.RuneBlock, nil, style)
				s.SetContent(termCol+1, termRow, tcell.RuneBlock, nil, style)
			}
		}

	}

	holdTitleStyle := tcell.StyleDefault.Bold(true)
	if gameState.HasHeld {
		holdTitleStyle = holdTitleStyle.Bold(false)

	}
	drawText(s, holdPaneCoords[0]+1, holdPaneCoords[1], holdPaneCoords[2], holdPaneCoords[0]+1, holdTitleStyle, "HOLD")

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

func drawCustomOnBoard(s tcell.Screen, row int, col int, char rune, char2 rune, style tcell.Style) {
	if row >= 20 || col >= 10 || row < 0 || col < 0 {
		return
	}

	termCol := boardCoords[0] + col*2 + 1
	termRow := boardCoords[1] + row + 1

	s.SetContent(termCol, termRow, char, nil, style)
	s.SetContent(termCol+1, termRow, char2, nil, style)
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
