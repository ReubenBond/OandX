package main

import (
	"⚛sdl"
	"⚛sdl/ttf"
	"⚛sdl/mixer"
	"fmt"
	"time"
	"path"
	//	"os"
)

const APP_VERSION = "0.1"
const SCREEN_WIDTH = 224
const SCREEN_HEIGHT = 224
//var WORKING_DIR, _ = os.Getwd()
var GFX_DIR string = "gfx"

// Resource manager
var gfxResources = make(map[string]*sdl.Surface)

// Get an image, returning a pointer to the SDL surface
// it was loaded into and true if successful, nil and false otherwise
func GetImage(fileName string) (image *sdl.Surface, ok bool) {
	// Reuse existing surface if image has been loaded previously
	if image, ok = gfxResources[fileName]; !ok {
		// Image hasn't been loaded previously
		image = sdl.Load(fileName)
		// Optimise surface for display
		image.DisplayFormat()
		if image != nil {
			// Image successfully loaded, store resource in
			// gfxResources for future reference
			gfxResources[fileName] = image
			ok = true
		}
	}
	return image, ok
}

// Get an image or panic
func MustGetImage(fileName string) *sdl.Surface {
	image, ok := GetImage(fileName)
	if !ok {
		panic("MustGetImage:"+" Failed to load "+fileName)
	}
	return image
}

// Free an image by its file name
func FreeImageByName(fileName string) {
	if image, ok := gfxResources[fileName]; ok {
		image.Free()
	}
	gfxResources[fileName] = nil, false
}

// Free an image by its handle
func FreeImage(image *sdl.Surface) {
	for k, v := range gfxResources {
		if v == image {
			v.Free()
			gfxResources[k] = nil, false
			break
		}
	}
}

// Free all managed resources
func FreeResources() {
	// Free all managed image resources
	for k, v := range gfxResources {
		v.Free()
		gfxResources[k] = nil, false
	}
}

// Game objects
type TileSymbol int

const (
	Empty  = iota
	Naught = iota
	Cross  = iota
)

var TileSymbols = []string{"Empty", "Naught", "Cross"}
var TileImageNames = []string{"empty.png", "o.png", "x.png"}
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
	type scanDef struct {
		x, y, xDiff, yDiff int
	}
	scanners := []scanDef{
		{0, 0, 1, 0},  // Across from top
		{0, 1, 1, 0},  // Across from middle
		{0, 2, 1, 0},  // Across from bottom
		{0, 0, 0, 1},  // Down from left
		{1, 0, 0, 1},  // Down from middle
		{2, 0, 0, 1},  // Down from right
		{0, 0, 1, 1},  // Diagonally from top-left
		{0, 2, 1, -1}} // Diagonally from bottom-left
	for _, scanner := range scanners {
		winner, won = func() (TileSymbol, bool) {
			naughts, crosses := 0, 0
			for x, y := scanner.x, scanner.y;
					x < len(board) && y < len(board[0]);
					x, y = x+scanner.xDiff, y+scanner.yDiff {
				switch *board[x][y] {
				case Naught:
					naughts++
				case Cross:
					crosses++
				}
			}
			switch 3 {
			case naughts:
				return Naught, true
			case crosses:
				return Cross, true
			default:
				return Empty, false
			}
			return Empty, false
		}()
		if won {
			return winner, won
		}
	}
	return
}

func ScreenToBoard(xScreen, yScreen int) (x, y int) {
	gridDim := SCREEN_WIDTH / 3
	x, y = xScreen/gridDim, yScreen/gridDim
	return
}

func BoardToScreen(x, y int) (xScreen, yScreen int) {
	gridDim := SCREEN_WIDTH / 3
	xScreen, yScreen = x*gridDim, y*gridDim
	return
}

func (board GameBoard) PlaceTile(tile TileSymbol, xScreen, yScreen int) (success bool) {
	x, y := ScreenToBoard(xScreen, yScreen)
	if *board[x][y] == Empty {
		success = true
		*board[x][y] = tile
	}
	return
}

var screen *sdl.Surface
var background *sdl.Surface
var tiles [3]*sdl.Surface

func (tile *TileSymbol) Sprite() (image *sdl.Surface, ok bool) {
	switch *tile {
	case Empty:
		return tiles[Empty], true
	case Naught:
		return tiles[Naught], true
	case Cross:
		return tiles[Cross], true
	}
	return
}

func (board GameBoard) Draw(screen *sdl.Surface) {
	screen.Blit(&sdl.Rect{0, 0, 0, 0}, background, nil)
	for x := 0; x < len(board); x++ {
		for y := 0; y < len(board[0]); y++ {
			xScreen, yScreen := BoardToScreen(x, y)
			image, ok := board[x][y].Sprite()
			if ok {
				screen.Blit(&sdl.Rect{int16(xScreen), int16(yScreen), 0, 0}, image, nil)
			}
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
	if mixer.OpenAudio(mixer.DEFAULT_FREQUENCY, mixer.DEFAULT_FORMAT,
		mixer.DEFAULT_CHANNELS, 4096) != 0 {
		panic(sdl.GetError())
	}
	sdl.EnableUNICODE(1)
	sdl.WM_SetCaption("OandX - "+APP_VERSION, "")
	screen = sdl.SetVideoMode(SCREEN_WIDTH, SCREEN_HEIGHT, 32, sdl.RESIZABLE)
	if screen == nil {
		panic(sdl.GetError())
	}

	// Form resource paths and load resources
	BackgroundImageName = path.Join(GFX_DIR, BackgroundImageName)
	background = MustGetImage(BackgroundImageName)
	for i, rsrc := range TileImageNames {
		TileImageNames[i] = path.Join(GFX_DIR, rsrc)
		tiles[i] = MustGetImage(TileImageNames[i])
		fmt.Println("Tile", i, "is", TileImageNames[i])
	}
}

func deinit() {
	sdl.Quit()
	ttf.Quit()
}

func main() {
	// Recall that call to init() is implicit
	// Clean up when we quit
	defer deinit()

	/*background = sdl.Load("gfx/grid.png")
	tiles[0] = sdl.Load("gfx/empty.png")
	tiles[1] = sdl.Load("gfx/o.png")
	tiles[2] = sdl.Load("gfx/x.png")
*/
	/*	if image == nil {
			panic(sdl.GetError())
		}
		defer image.Free()

	*/
	sdl.WM_SetIcon(tiles[Cross], nil)
	running := true

	font := ttf.OpenFont("Fontin Sans.otf", 36)
	if font == nil {
		panic(sdl.GetError())
	}
	defer font.Close()

	font.SetStyle(ttf.STYLE_UNDERLINE)
	red := sdl.Color{255, 0, 0, 0}
	text := &sdl.Surface{}
	//ttf.RenderText_Blended(font, "Have fun!", red)
	music := mixer.LoadMUS("test.ogg")

	if music == nil {
		panic(sdl.GetError())
	}
	defer music.Free()

	music.PlayMusic(-1)

	if sdl.GetKeyName(270) != "[+]" {
		panic("GetKeyName broken")
	}
	ticker := time.NewTicker(1e9 / 50 /*50Hz*/ )

	board := NewGameBoard()
	tile := TileSymbol(Naught)

	// Note: The following SDL code is highly ineffective.
	//       It is eating too much CPU. If you intend to use Go-SDL,
	//       you should to do better than this.
	for running {
		select {
		case <-ticker.C: // Redraw
			screen.FillRect(nil, 0xFFFFFF)
			board.Draw(screen)
			screen.Blit(&sdl.Rect{0, 0, 0, 0}, text, nil)
			screen.Flip()
		case _event := <-sdl.Events: // Handle other events
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
					if winner, won := board.Winner(); won {
						text.Free()
						winString := winner.String() + " won!!"
						text = ttf.RenderText_Blended(font, winString, red)
						fmt.Println(winString)
					}
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
