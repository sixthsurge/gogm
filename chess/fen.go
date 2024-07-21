package chess

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

// FEN for the chess starting position
const StartingPositionFen string = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

func LoadFen(fen string) (*Board, error) {
	board := NewBoard()

	fields := strings.Split(fen, " ")

	// Piece placement
	rankIndex := 0
	fileIndex := 0

	for _, char := range fields[0] {
		square := SquareAt(File(fileIndex), Rank(rankIndex))

		if unicode.IsDigit(char) {
			skipAmount := int(char) - int('0')
			fileIndex += skipAmount
		} else if unicode.IsLetter(char) {
			isBlack := unicode.IsLower(char)

			pieceKind, err := PieceWithAlgebraicLetter(char)
			if err != nil {
				return nil, errors.New(fmt.Sprintf("unexpected letter: %v", char))
			}

			board.SetPiece(square, pieceKind, isBlack)
			fileIndex += 1
		} else if char == '/' {
			rankIndex += 1
			fileIndex = 0
		} else {
			return nil, errors.New(fmt.Sprintf("unexpected character: %v", char))
		}
	}

	// Side to move
	switch fields[1] {
	case "w":
		board.blackToMove = false

	case "b":
		board.blackToMove = true

	default:
		return nil, errors.New(fmt.Sprintf("unexpected side to move: %v", fields[1]))
	}

	// Castling rights
	board.castlingRights.WhiteKingside = strings.ContainsRune(fields[2], 'K')
	board.castlingRights.BlackKingside = strings.ContainsRune(fields[2], 'k')
	board.castlingRights.WhiteQueenside = strings.ContainsRune(fields[2], 'Q')
	board.castlingRights.BlackQueenside = strings.ContainsRune(fields[2], 'q')

	// TODO: en passant target, halfmove clock, fullmove number

	return &board, nil
}

func (board *Board) Fen() string {
	// Piece placement
	var sb strings.Builder

	for rankIndex := 0; rankIndex < 8; rankIndex++ {
		numEmptySquares := 0

		for fileIndex := 0; fileIndex < 8; fileIndex++ {
			square := SquareAt(File(fileIndex), Rank(rankIndex))
			piece := board.GetPiece(square)

			if piece == nil {
				numEmptySquares += 1
			} else {
				if numEmptySquares > 0 {
					sb.WriteString(fmt.Sprint(numEmptySquares))
					numEmptySquares = 0
				}
				pieceLetter := piece.Kind.AlgebraicLetter()

				if !piece.IsBlack {
					pieceLetter = unicode.ToUpper(pieceLetter)
				}

				sb.WriteRune(pieceLetter)
			}
		}

		if numEmptySquares > 0 {
			sb.WriteString(fmt.Sprint(numEmptySquares))
		}

		if rankIndex != 7 {
			sb.WriteRune('/')
		}
	}

	sb.WriteRune(' ')

	// Side to move
	var sideToMoveRune rune
	if board.blackToMove {
		sideToMoveRune = 'b'
	} else {
		sideToMoveRune = 'w'
	}
	sb.WriteRune(sideToMoveRune)
	sb.WriteRune(' ')

	// Castling rights
	if !board.castlingRights.WhiteKingside && !board.castlingRights.WhiteQueenside && !board.castlingRights.BlackKingside && !board.castlingRights.BlackQueenside {
		sb.WriteRune('-')
	} else {
		if board.castlingRights.WhiteKingside {
			sb.WriteRune('K')
		}
		if board.castlingRights.WhiteQueenside {
			sb.WriteRune('Q')
		}
		if board.castlingRights.BlackKingside {
			sb.WriteRune('k')
		}
		if board.castlingRights.BlackQueenside {
			sb.WriteRune('q')
		}
	}


	return sb.String()
}
