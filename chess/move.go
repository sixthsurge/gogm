package chess

import "fmt"

// Information necessary to make a move
type Move struct {
	Source        Square
	Destination   Square
	PromotedPiece PieceKind
	IsPromotion   bool
}

// Information necessary to undo a move
type Unmove struct {
	source             Square
	destination        Square
	capturedPiece      PieceKind
	oldEnPassantTarget Square
	oldCastlingRights  CastlingRights
	isCapture          bool
	isPromotion        bool
	hadEnPassantTarget bool
}

// Format a move in UCI notation
func (move Move) String() string {
	if move.IsPromotion {
		src, _ := move.Source.AlgebraicName()
		dst, _ := move.Destination.AlgebraicName()
		promoted := move.PromotedPiece.AlgebraicLetter()

		return fmt.Sprintf("%v%v%v", src, dst, promoted)
	} else {
		src, _ := move.Source.AlgebraicName()
		dst, _ := move.Destination.AlgebraicName()

		return fmt.Sprintf("%v%v", src, dst)

	}
}
