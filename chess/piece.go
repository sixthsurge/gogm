package chess

import (
	"errors"
	"fmt"
	"log"
	"unicode"
)

type PieceKind uint8

const (
	King PieceKind = iota
	Queen
	Bishop
	Knight
	Rook
	Pawn
)

// Returns the piece represented by the given letter in English algebraic notation
func PieceWithAlgebraicLetter(letter rune) (PieceKind, error) {
	switch unicode.ToLower(letter) {
	case 'p':
		return Pawn, nil
	case 'n':
		return Knight, nil
	case 'b':
		return Bishop, nil
	case 'r':
		return Rook, nil
	case 'q':
		return Queen, nil
	case 'k':
		return King, nil
	}

	return Pawn, errors.New(fmt.Sprintf("unknown piece letter: %v", letter))
}

// Returns the lowercase letter representing the piece in English algebraic notation
func (piece PieceKind) AlgebraicLetter() rune {
	switch piece {
	case Pawn:
		return 'p'
	case Knight:
		return 'n'
	case Bishop:
		return 'b'
	case Rook:
		return 'r'
	case Queen:
		return 'q'
	case King:
		return 'k'
	}

	log.Fatalf("unknown piece kind: %v", piece)
	return '_'
}
