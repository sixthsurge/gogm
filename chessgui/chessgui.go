package main

import (
	"gogm/chess"

	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
)

const windowWidth int32 = 800
const windowHeight int32 = 800
const boardWidth int32 = windowHeight
const boardHeight int32 = windowHeight
const squareWidth int32 = boardWidth / 8
const squareHeight int32 = boardHeight / 8
const piecesImagePath string = "assets/pieces.png"

type Bot interface {
	Think(*chess.Board) chess.Move
}

type guiState struct {
	board                   *chess.Board
	whiteBot                *Bot
	blackBot                *Bot
	window                  *sdl.Window
	renderer                *sdl.Renderer
	piecesTexture           *sdl.Texture
	piecesTextureW          int32
	piecesTextureH          int32
	mouseX                  int32
	mouseY                  int32
	movingPiece             bool
	pieceSourceSquare       chess.Square
	pieceDestinationSquares []chess.Square
	lastMove                *chess.Move
	unmoveHistory           []chess.Unmove
}

// Open a window displaying the match between `whiteBot` and `blackBot`
// If either `whiteBot` or `blackBot` or both are `nil`, then that side will be
// played by the user
func Run(board *chess.Board, whiteBot *Bot, blackBot *Bot) {
	state := setup(board, whiteBot, blackBot)
	defer state.destroy()

	exit := false
	for !exit {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event := event.(type) {
			case *sdl.QuitEvent:
				exit = true
				break

			case *sdl.MouseMotionEvent:
				state.mouseX = event.X
				state.mouseY = event.Y

			case *sdl.MouseButtonEvent:
				if event.Button == sdl.BUTTON_LEFT {
					if event.Type == sdl.MOUSEBUTTONDOWN {
						state.onLeftMouseButtonDown()
					}
				}

			case *sdl.KeyboardEvent:
				if event.Keysym.Sym == sdl.GetKeyFromName("b") && event.Type == sdl.KEYDOWN {
					state.onBKeyDown()
				}
			}
		}

		state.render()
	}
}

func setup(board *chess.Board, whiteBot *Bot, blackBot *Bot) (state guiState) {
	// Initialize SDL
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}

	// Create window
	window, err := sdl.CreateWindow("Chess", sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED, windowWidth, windowHeight, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}

	// Create renderer
	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		panic(err)
	}

	// Load pieces texture
	piecesTexture, piecesTextureW, piecesTextureH := loadPiecesTexture(renderer)

	state.board = board
	state.whiteBot = whiteBot
	state.blackBot = blackBot
	state.window = window
	state.renderer = renderer
	state.piecesTexture = piecesTexture
	state.piecesTextureW = piecesTextureW
	state.piecesTextureH = piecesTextureH

	return
}

func (state *guiState) destroy() {
	state.piecesTexture.Destroy()
	state.renderer.Destroy()
	state.window.Destroy()
	sdl.Quit()
}

func (state *guiState) render() {
	state.renderer.SetDrawBlendMode(sdl.BLENDMODE_BLEND)
	state.renderer.SetDrawColor(0, 0, 0, 255)
	state.renderer.Clear()
	state.drawBoard()
	state.renderer.Present()
}

func (state *guiState) drawBoard() {
	lightSquareColor := []uint8{240, 217, 181, 255}
	darkSquareColor := []uint8{181, 136, 99, 255}
	lastMoveColor := []uint8{255, 240, 0, 150}
	destinationSquareColor := []uint8{0, 0, 255, 100}
	checkColor := []uint8{255, 0, 0, 100}

	for x := 0; x < 8; x++ {
		for y := 0; y < 8; y++ {
			square := chess.SquareAt(chess.File(x), chess.Rank(y))
			squareRect := sdl.Rect{X: int32(x) * squareWidth, Y: int32(y) * squareHeight, W: squareWidth, H: squareHeight}

			// Draw square
			isLightSquare := (x+y)&1 == 0
			if isLightSquare {
				state.renderer.SetDrawColorArray(lightSquareColor...)
			} else {
				state.renderer.SetDrawColorArray(darkSquareColor...)
			}
			state.renderer.FillRect(&squareRect)

			// When moving a piece, highlight source and destination square
			isDestinationSquare := false
			if state.movingPiece {
				for _, destinationSquare := range state.pieceDestinationSquares {
					if square == destinationSquare {
						isDestinationSquare = true
					}
				}

				if square == state.pieceSourceSquare || isDestinationSquare {
					state.renderer.SetDrawColorArray(destinationSquareColor...)
					state.renderer.FillRect(&squareRect)
				}
			}

			// Highlight last move
			if state.lastMove != nil && !isDestinationSquare {
				if square == state.lastMove.Source || square == state.lastMove.Destination {
					state.renderer.SetDrawColorArray(lastMoveColor...)
					state.renderer.FillRect(&squareRect)
				}
			}

			// Draw piece on square
			if piece := state.board.GetPiece(square); piece != nil {
				// Highlight king square in check
				if piece.Kind == chess.King && state.board.IsCheck() && piece.IsBlack == state.board.IsBlackToMove() && !isDestinationSquare {
					state.renderer.SetDrawColorArray(checkColor...)
					state.renderer.FillRect(&squareRect)
				}

				var sourceRect sdl.Rect
				sourceRect.W = state.piecesTextureW / 6
				sourceRect.H = state.piecesTextureH / 2
				sourceRect.X = sourceRect.W * (int32(piece.Kind))
				if piece.IsBlack {
					sourceRect.Y = sourceRect.H
				}

				state.renderer.SetDrawColor(255, 255, 255, 255)
				state.renderer.Copy(state.piecesTexture, &sourceRect, &squareRect)
			}
		}
	}
}

func (state *guiState) onLeftMouseButtonDown() {
	var userControl bool
	if state.board.IsBlackToMove() {
		userControl = state.blackBot == nil
	} else {
		userControl = state.whiteBot == nil
	}

	if !userControl {
		return
	}

	hoverSquare := chess.SquareAt(chess.File(state.mouseX/squareWidth), chess.Rank(state.mouseY/squareHeight))

	if state.movingPiece {
		// Finish moving piece
		for _, destinationSquare := range state.pieceDestinationSquares {
			if destinationSquare == hoverSquare {
				// TODO: underpromotions
				var promotionRank chess.Rank
				if state.board.IsBlackToMove() {
					promotionRank = chess.Rank1
				} else {
					promotionRank = chess.Rank8
				}

				isPromotion := state.board.GetPiece(state.pieceSourceSquare).Kind == chess.Pawn && destinationSquare.Rank() == promotionRank

				move := chess.Move{
					Source:      state.pieceSourceSquare,
					Destination: destinationSquare,
					IsPromotion: isPromotion,
					PromotedPiece: chess.Queen,
				}

				unmove := state.board.MakeMove(move)
				state.lastMove = &move
				state.unmoveHistory = append(state.unmoveHistory, unmove)
				break
			}
		}

		state.movingPiece = false
	} else {
		// Start moving piece
		hoverPiece := state.board.GetPiece(hoverSquare)
		if hoverPiece != nil {
			if hoverPiece.IsBlack == state.board.IsBlackToMove() {
				state.movingPiece = true
				state.pieceSourceSquare = hoverSquare
				state.pieceDestinationSquares = []chess.Square{}

				for _, move := range state.board.GetLegalMovesFromSquare(hoverSquare) {
					state.pieceDestinationSquares = append(state.pieceDestinationSquares, move.Destination)
				}
			}
		}
	}
}

func (state *guiState) onBKeyDown() {
	if len(state.unmoveHistory) == 0 {
		return
	}

	// Undo last move
	unmove := state.unmoveHistory[len(state.unmoveHistory) - 1]
	state.unmoveHistory = state.unmoveHistory[:len(state.unmoveHistory) - 1]

	state.board.UnmakeMove(unmove)
}

func loadPiecesTexture(renderer *sdl.Renderer) (piecesTexture *sdl.Texture, piecesTextureW int32, piecesTextureH int32) {
	piecesImage, err := img.Load(piecesImagePath)
	if err != nil {
		panic(err)
	}
	defer piecesImage.Free()

	piecesTextureW = piecesImage.W
	piecesTextureH = piecesImage.H

	piecesTexture, err = renderer.CreateTextureFromSurface(piecesImage)
	if err != nil {
		panic(err)
	}

	return
}

func main() {
	board, err := chess.LoadFen("r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq")

	if err != nil {
		panic(err)
	}

	Run(board, nil, nil)
}
