package main

import (
	"⚛sdl"
	"⚛sdl/ttf"
	"⚛sdl/mixer"
	"math"
	"fmt"
	"time"
	"path"
)

const APP_VERSION = "0.1"
const SCREEN_WIDTH = 224
const SCREEN_HEIGHT = 224
const GFX_DIR = "gfx"

type Point struct {
	x int
	y int
}

func (a Point) add(b Point) Point { return Point{a.x + b.x, a.y + b.y} }

func (a Point) sub(b Point) Point { return Point{a.x - b.x, a.y - b.y} }

func (a Point) length() float64 { return math.Sqrt(float64(a.x*a.x + a.y*a.y)) }

func (a Point) mul(b float64) Point {
	return Point{int(float64(a.x) * b), int(float64(a.y) * b)}
}

// Resource manager
var gfxResources map[string]*sdl.Surface
// Load an image from disk, returning a pointer to the SDL surface
// it was loaded into and true if successful, nil and false otherwise
func GetImage(fileName string) (image *sdl.Surface, success bool) {
	if image, ok := gfxResources[fileName]; !ok {
		// Image hasn't been loaded previously
		fmt.Println("GetImage","Loading image", fileName)
		image = sdl.Load(fileName)
		if image != nil {
			gfxResources[fileName] = image
			success = true
		}
	}
	return image, success
}

func FreeResources() {
	// Free all image resources
	for _, v := range(gfxResources) {
		v.Free()
	}
}

// Game objects
type TileSymbol int

const (
	Empty = iota
	Naught = iota
	Cross = iota
)

var TileSymbols = []string {"Empty", "Naught", "Cross"}
var TileImageNames = []string  {"empty.png", "o.png", "x.png"}
var BackgroundImageName = "grid.png"

func (tile TileSymbol) String() string {
	return TileSymbols[tile]
}

func (tile *TileSymbol) Flip() {
	if *tile == Naught {
		*tile = Cross
	} else {
		*tile = Naught
	}
	return
}

type GameBoard [3][3]*TileSymbol

func NewGameBoard() (board GameBoard) {
	for x := 0; x < len(board); x++ {
		for y := 0; y < len(board[0]); y++ {
			board[x][y] = new(TileSymbol)
		}
	}
	return
}

func (board GameBoard) Winner() (winner TileSymbol, won bool) {
	type scanDef struct {x, y, xDiff, yDiff int}
	scanners := []scanDef{
		{0, 0, 1, 0}, // Across from top
		{0, 1, 1, 0}, // Across from middle
		{0, 2, 1, 0}, // Across from bottom
		{0, 0, 0, 1}, // Down from left
		{1, 0, 0, 1}, // Down from middle
		{2, 0, 0, 1}, // Down from right
		{0, 0, 1, 1}, // Diagonally from top-left
		{0, 2, 1, -1}} // Diagonally from bottom-left
	for _, scanner := range(scanners) {
		func () {
			naughts, crosses := 0, 0
			for x := scanner.x; x < 3; x += scanner.xDiff {
				for y := scanner.y; y < 3; y += scanner.yDiff {
					switch (*board[x][y]) {
					case Naught:
						naughts++
					case Cross:
						crosses++
					}
				}
			}
			switch (3) {
			case naughts:
				winner = Naught
				won = true
			case crosses:
				winner = Cross
				won = true
			default:
				winner = Empty
				won = false
			}
			return
		}()
	}
	return
}

func ScreenToBoard(xScreen, yScreen int) (x, y int) {
	gridDim := SCREEN_WIDTH / 3
	x, y = xScreen / gridDim, yScreen / gridDim
	return
}

func BoardToScreen(x, y int) (xScreen, yScreen int) {
	gridDim := SCREEN_WIDTH / 3
	xScreen, yScreen = x * gridDim, y * gridDim
	return
}

func (board GameBoard) PlaceTile(tile TileSymbol, xScreen, yScreen int) (success bool){
	x, y := ScreenToBoard(xScreen, yScreen)
	fmt.Println("Place",tile.String(),"at:",x,y)
	if *board[x][y] == Empty {
		success = true
		*board[x][y] = tile
	}
	return
}
var background, emptyTile, naughtTile, crossTile *sdl.Surface
func (board GameBoard) Draw(screen *sdl.Surface) {
	screen.Blit(&sdl.Rect{0,0,0,0}, background, nil)
	for x := 0; x < len(board); x++ {
		for y:= 0; y < len(board[0]); y++ {
			xScreen, yScreen := BoardToScreen(x, y)
			screen.Blit(&sdl.Rect{xScreen, yScreen, 0, 0}, board[x][y]
		}
	}
}

func init() {
	// Initialise SDL
	if sdl.Init(sdl.INIT_EVERYTHING) != 0 {
		panic(sdl.GetError())
	}
	if ttf.Init() != 0 {
		panic(sdl.GetError())
	}

	// Form resource paths
	for i, rsrc := range(TileImageNames) {
		TileImageNames[i] = path.Join(GFX_DIR, rsrc)
		fmt.Println("Tile",i,"is",TileImageNames[i])
	}
	BackgroundImageName = path.Join(GFX_DIR, BackgroundImageName)
}

func deinit() {
	sdl.Quit();
	ttf.Quit();
}

func main() {
	// Recall that call to init() is implicit
	// Clean up when we quit
	defer deinit()

	if mixer.OpenAudio(mixer.DEFAULT_FREQUENCY, mixer.DEFAULT_FORMAT,
		mixer.DEFAULT_CHANNELS, 4096) != 0 {
		panic(sdl.GetError())
	}

	var screen = sdl.SetVideoMode(SCREEN_WIDTH, SCREEN_HEIGHT, 32, sdl.RESIZABLE)

	if screen == nil {
		panic(sdl.GetError())
	}

	sdl.EnableUNICODE(1)

	sdl.WM_SetCaption("TinyTanks - "+APP_VERSION, "")
	image := sdl.Load("gfx/o.png")
	if image == nil {
		panic(sdl.GetError())
	}
	defer image.Free()

	sdl.WM_SetIcon(image, nil)

	running := true

	font := ttf.OpenFont("Fontin Sans.otf", 72)
	if font == nil {
		panic(sdl.GetError())
	}
	defer font.Close()

	font.SetStyle(ttf.STYLE_UNDERLINE)
	white := sdl.Color{255, 255, 255, 0}
	text := ttf.RenderText_Blended(font, "Test (with music)", white)
	music := mixer.LoadMUS("test.ogg")

	if music == nil {
		panic(sdl.GetError())
	}
	defer music.Free()

	music.PlayMusic(-1)

	if sdl.GetKeyName(270) != "[+]" {
		panic("GetKeyName broken")
	}
	ticker := time.NewTicker(1e9/50 /*50Hz*/)

	var draw = make(chan Point, 10)
	var out = make(chan Point, 10)

	board := NewGameBoard()
	tile := TileSymbol(Naught)

	// Note: The following SDL code is highly ineffective.
	//       It is eating too much CPU. If you intend to use Go-SDL,
	//       you should to do better than this.
	for running {
		select {
		case <-ticker.C: // Redraw
			screen.FillRect(nil, 0x302019)
			board.Draw(screen)
			//screen.Blit(&sdl.Rect{0, 0, 0, 0}, text, nil)
			/*draw <- Point{100,100}
			loop: for {
				select {
				case p := <-draw:
					screen.Blit(&sdl.Rect{int16(p.x), int16(p.y), 0, 0}, image, nil)
				case <-out:
				default:
					break loop
				}
			}
			var p Point
			sdl.GetMouseState(&p.x, &p.y)
			*/
			screen.Flip()
		case _event := <-sdl.Events: // Handle events
			switch e := _event.(type) {
			case sdl.QuitEvent:
				running = false

			case sdl.KeyboardEvent:
				println("")
				println(e.Keysym.Sym, ": ", sdl.GetKeyName(sdl.Key(e.Keysym.Sym)))

				if e.Keysym.Sym == sdl.K_ESCAPE {
					running = false
				}

				fmt.Printf("%04x ", e.Type)

				for i := 0; i < len(e.Pad0); i++ {
					fmt.Printf("%02x ", e.Pad0[i])
				}
				println()

				fmt.Printf("Type: %02x Which: %02x State: %02x Pad: %02x\n", e.Type, e.Which, e.State, e.Pad0[0])
				fmt.Printf("Scancode: %02x Sym: %08x Mod: %04x Unicode: %04x\n", e.Keysym.Scancode, e.Keysym.Sym, e.Keysym.Mod, e.Keysym.Unicode)

			case sdl.MouseButtonEvent:
				if e.Type == sdl.MOUSEBUTTONDOWN {
					println("Trying to place ", tile, "at:", e.X, e.Y)
					if board.PlaceTile(tile, int(e.X), int(e.Y)) {
						tile.Flip()
					}
				//	in = out
				//	out = make(chan Point)
				//	go worm(in, out, draw)
				}

			case sdl.ResizeEvent:
				println("resize screen ", e.W, e.H)

				screen = sdl.SetVideoMode(int(e.W), int(e.H), 32, sdl.RESIZABLE)

				if screen == nil {
					panic(sdl.GetError())
				}
			}
		}
	}
}
