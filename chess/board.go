package chess

import "log"

// Information about the state of the board
type Board struct {
	whitePieces        []Piece
	blackPieces        []Piece
	squareContents     [64]*Piece
	enPassantTarget    Square
	hasEnPassantTarget bool
	blackToMove        bool
	castlingRights     CastlingRights
	bishopAttackTable  SlidingAttackTable
	rookAttackTable    SlidingAttackTable
}

// Information about a piece on the board
type Piece struct {
	Kind         PieceKind
	Square       Square
	IsBlack      bool
	HasAttackSet bool
	AttackSet    Bitboard
}

type CastlingRights struct {
	WhiteKingside  bool
	WhiteQueenside bool
	BlackKingside  bool
	BlackQueenside bool
}

func NewBoard() (board Board) {
	board.bishopAttackTable = CreateBishopAttackTable()
	board.rookAttackTable = CreateRookAttackTable()
	return
}

func (board *Board) IsBlackToMove() bool {
	return board.blackToMove
}

func (board *Board) GetCastlingRights(isBlack bool) (kingside bool, queenside bool) {
	if isBlack {
		kingside = board.castlingRights.BlackKingside
		queenside = board.castlingRights.BlackQueenside
	} else {
		kingside = board.castlingRights.WhiteKingside
		queenside = board.castlingRights.WhiteQueenside
	}

	return
}

func (board *Board) GetPiecesForSide(isBlack bool) []Piece {
	if isBlack {
		return board.blackPieces
	} else {
		return board.whitePieces
	}
}

func (board *Board) GetPiece(sq Square) *Piece {
	return board.squareContents[uint32(sq)]
}

func (board *Board) HasPiece(sq Square) bool {
	return board.squareContents[uint32(sq)] != nil
}

func (board *Board) SetPiece(sq Square, kind PieceKind, isBlack bool) {
	// Remove existing piece
	board.SetEmpty(sq)

	// Set square content
	board.squareContents[uint32(sq)] = &Piece{Square: sq, Kind: kind, IsBlack: isBlack}

	// Add piece to piece list
	if isBlack {
		board.blackPieces = append(board.blackPieces, *board.squareContents[uint32(sq)])
	} else {
		board.whitePieces = append(board.whitePieces, *board.squareContents[uint32(sq)])
	}
}

func (board *Board) SetEmpty(sq Square) {
	existingPiece := board.squareContents[uint32(sq)]
	if existingPiece != nil {
		if existingPiece.IsBlack {
			updatedBlackPieces, removed := removeElement(board.blackPieces, *existingPiece)
			if !removed {
				panic("failed to remove piece")
			}
			board.blackPieces = updatedBlackPieces
		} else {
			updatedWhitePieces, removed := removeElement(board.whitePieces, *existingPiece)
			if !removed {
				panic("failed to remove piece")
			}
			board.whitePieces = updatedWhitePieces
		}
	}

	board.squareContents[uint32(sq)] = nil
}

func (board *Board) GetPiecesBitboard(isBlack bool) (result Bitboard) {
	var pieces []Piece
	if isBlack {
		pieces = board.blackPieces
	} else {
		pieces = board.whitePieces
	}

	for _, piece := range pieces {
		result = result.Set(piece.Square)
	}

	return
}

// Update the board state by making the given move
func (board *Board) MakeMove(move Move) (unmove Unmove) {
	pieceMoved := board.GetPiece(move.Source).Kind
	isCapture := board.GetPiece(move.Destination) != nil

	// Setup information to unmake move
	unmove.source             = move.Source
	unmove.destination        = move.Destination
	unmove.oldEnPassantTarget = board.enPassantTarget
	unmove.isCapture          = isCapture
	unmove.isPromotion        = move.IsPromotion
	unmove.hadEnPassantTarget = board.hasEnPassantTarget
	unmove.oldCastlingRights  = board.castlingRights

	if isCapture {
		unmove.capturedPiece = board.GetPiece(move.Destination).Kind
	}

	// In the case of promotion, change the piece moved to the promoted piece
	if pieceMoved == Pawn && move.IsPromotion {
		pieceMoved = move.PromotedPiece
	}

	board.SetEmpty(move.Source)
	board.SetPiece(move.Destination, pieceMoved, board.blackToMove)

	// In the case of en passant, remove the captured pawn
	isEnPassantCapture := pieceMoved == Pawn && !isCapture && move.Source.File() != move.Destination.File()
	if isEnPassantCapture {
		board.SetEmpty(SquareAt(move.Destination.File(), move.Source.Rank()))
	}

	// In case of castling, move the rook
	if pieceMoved == King {
		var backRank Rank
		if board.blackToMove {
			backRank = Rank8
		} else {
			backRank = Rank1
		}

		sourceFile := move.Source.File()
		destinationFile := move.Destination.File()

		if sourceFile == FileE && destinationFile == FileC { // Queenside castle
			board.SetEmpty(SquareAt(FileA, backRank))
			board.SetPiece(SquareAt(FileD, backRank), Rook, board.blackToMove)
		} else if sourceFile == FileE && destinationFile == FileG {
			board.SetEmpty(SquareAt(FileH, backRank))
			board.SetPiece(SquareAt(FileF, backRank), Rook, board.blackToMove)
		}
	}

	// Update en passant target
	board.hasEnPassantTarget = false
	if pieceMoved == Pawn {
		// Detect en passant
		var pawnStartingRank Rank
		var pawnThrustRank Rank
		var enPassantRank Rank

		if board.blackToMove {
			pawnThrustRank = Rank5
			pawnStartingRank = Rank7
			enPassantRank = Rank6
		} else {
			pawnStartingRank = Rank2
			pawnThrustRank = Rank4
			enPassantRank = Rank3
		}

		if move.Source.Rank() == pawnStartingRank && move.Destination.Rank() == pawnThrustRank {
			board.hasEnPassantTarget = true
			board.enPassantTarget = SquareAt(move.Source.File(), enPassantRank)
		}
	}

	// Update castling rights
	var kingsideCastlingRight *bool
	var queensideCastlingRight *bool

	if board.blackToMove {
		kingsideCastlingRight = &board.castlingRights.BlackKingside
		queensideCastlingRight = &board.castlingRights.BlackQueenside
	} else {
		kingsideCastlingRight = &board.castlingRights.WhiteKingside
		queensideCastlingRight = &board.castlingRights.WhiteQueenside
	}

	if pieceMoved == King {
		*kingsideCastlingRight = false
		*queensideCastlingRight = false
	} else if pieceMoved == Rook && move.Source.File() == FileA {
		*queensideCastlingRight = false
	} else if pieceMoved == Rook && move.Source.File() == FileH {
		*kingsideCastlingRight = false
	}

	// Update side to move
	board.blackToMove = !board.blackToMove

	return
}

// Update the board state by unmaking the given move
func (board *Board) UnmakeMove(unmove Unmove) {
	// Update side to move
	board.blackToMove = !board.blackToMove

	var pieceMoved PieceKind
	if unmove.isPromotion {
		pieceMoved = Pawn
	} else {
		if destinationPiece := board.GetPiece(unmove.destination); destinationPiece != nil {
			pieceMoved = destinationPiece.Kind
		} else {
			log.Fatalf("UnmakeMove - Piece on destination square %v is nil\n", unmove.destination)
		}
	}

	// Replace piece on old square
	board.SetPiece(unmove.source, pieceMoved, board.blackToMove)

	if unmove.isCapture {
		// Restore captured piece
		board.SetPiece(unmove.destination, unmove.capturedPiece, !board.blackToMove)
	} else {
		board.SetEmpty(unmove.destination)
	}

	// In the case of en passant, restore the captured pawn
	isEnPassantCapture := pieceMoved == Pawn && !unmove.isCapture && unmove.source.File() != unmove.destination.File()
	if isEnPassantCapture {
		board.SetPiece(SquareAt(unmove.destination.File(), unmove.source.Rank()), Pawn, !board.blackToMove)
	}

	// In the case of castling, move the rook back
	if pieceMoved == King {
		var backRank Rank
		if board.blackToMove {
			backRank = Rank8
		} else {
			backRank = Rank1
		}

		sourceFile := unmove.source.File()
		destinationFile := unmove.destination.File()

		if sourceFile == FileE && destinationFile == FileC { // Queenside castle
			board.SetEmpty(SquareAt(FileD, backRank))
			board.SetPiece(SquareAt(FileA, backRank), Rook, board.blackToMove)
		} else if sourceFile == FileE && destinationFile == FileG { // Kingside castle
			board.SetEmpty(SquareAt(FileF, backRank))
			board.SetPiece(SquareAt(FileH, backRank), Rook, board.blackToMove)
		}
	}

	// Restore en passant target
	board.hasEnPassantTarget = unmove.hadEnPassantTarget
	board.enPassantTarget = unmove.oldEnPassantTarget

	// Restore castling rights
	board.castlingRights = unmove.oldCastlingRights
}

// Returns the square containing the king
func (board *Board) GetKingSquare(isBlack bool) Square {
	for _, piece := range board.GetPiecesForSide(isBlack) {
		if piece.Kind == King {
			return piece.Square
		}
	}

	return A1
}

// True if the previous move left the king in check
func (board *Board) DetectIllegalMove() bool {
	board.blackToMove = !board.blackToMove
	isCheck := board.IsCheck()
	board.blackToMove = !board.blackToMove
	return isCheck
}
