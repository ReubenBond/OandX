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
var GFX_DIR string = "gfx" // Graphics
var FONT_DIR string = "font" // Fonts
var MUSIC_DIR string = "music" // Background music
var SFX_DIR string = "sfx" // Sound effects

//TODO: Separate resource manager into separate package
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
	return
}

// Get an image or panic
func MustGetImage(fileName string) (image *sdl.Surface) {
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

// Describes a font uniquely
var fontResources = make(map[string]map[int]*ttf.Font)
func GetFont(fileName string, size int) (font *ttf.Font, ok bool) {
	if font, ok = fontResources[fileName][size]; !ok {
		font = ttf.OpenFont(fileName, size)
		if font != nil {
			fontResources[fileName][size] = font
			ok = true
		}
	}
	return
}

func MustGetFont(fileName string, size int) (font *ttf.Font) {
	font, ok := GetFont(fileName, size);
	if !ok {
		panic("MustGetFont"+" Failed to load "+fileName+" with specified size")
	}
	return
}

//TODO: FreeFont*() functions

// Free all managed resources
func FreeResources() {
	// Free all managed image resources
	for k, v := range gfxResources {
		v.Free()
		gfxResources[k] = nil, false
	}
	// Close all fonts
	for _, sizes := range fontResources {
		for size, font := range sizes {
			font.Close()
			sizes[size] = nil, false
		}
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
	screen = sdl.SetVideoMode(SCREEN_WIDTH, SCREEN_HEIGHT, 32, sdl.RESIZABLE & sdl.OPENGL)
	if screen == nil {
		panic(sdl.GetError())
	}

	// Form resource paths and load resources
	BackgroundImageName = path.Join(GFX_DIR, BackgroundImageName)
	background = MustGetImage(BackgroundImageName)
	for i, rsrc := range TileImageNames {
		// Resource comes from graphics directory
		TileImageNames[i] = path.Join(GFX_DIR, rsrc)
		// Load resource or die
		tiles[i] = MustGetImage(TileImageNames[i])
	}
	// Set the Window Manager icon for the game window
	sdl.WM_SetIcon(tiles[Cross], nil)
}

func deinit() {
	sdl.Quit()
	ttf.Quit()
}

func main() {
	// Recall that call to init() is implicit
	// Clean up when we quit
	defer deinit()
	running := true

	font := ttf.OpenFont(path.Join(FONT_DIR, "Fontin Sans.otf"), 36)
	if font == nil {
		panic(sdl.GetError())
	}
	defer font.Close()

	font.SetStyle(ttf.STYLE_UNDERLINE)
	red := sdl.Color{255, 0, 0, 0}
	text := &sdl.Surface{}
	//ttf.RenderText_Blended(font, "Have fun!", red)
	music := mixer.LoadMUS(path.Join(MUSIC_DIR, "bgm.ogg"))

	if music == nil {
		panic(sdl.GetError())
	}
	defer music.Free()

	music.PlayMusic(-1)
	ticker := time.NewTicker(1e9 / 50 /*50Hz*/ )

	board := NewGameBoard()
	tile := TileSymbol(Naught) // Naught plays first

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
				if e.Keysym.Sym == sdl.K_ESCAPE {
					running = false
				}
			case sdl.MouseButtonEvent:
				if e.Type == sdl.MOUSEBUTTONDOWN {
					// User clicked somewhere... do we place a tile?
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

				screen = sdl.SetVideoMode(int(e.W), int(e.H), 32, sdl.RESIZABLE & sdl.OPENGL)

				if screen == nil {
					panic(sdl.GetError())
				}
			}
		}
	}
}
