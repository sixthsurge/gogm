package botv1

import (
	"math"
	"gogm/chess"
)

func evaluate(board *chess.Board) float64 {
    evaluation := float64(0.0)
    black := board.IsBlackToMove()
    legalMoves := board.GetLegalMoves()

    if len(legalMoves) == 0 {
        if board.IsCheck() {
            // Checkmate
            return math.Inf(-1)
        } else {
            // Stalemate
            return 0.0
        }
    }

    evaluation += evaluatePieces(board, black)
    evaluation -= evaluatePieces(board, !black)

    evaluation += evaluateCastlingRights(board, black)
    evaluation -= evaluateCastlingRights(board, !black)

    return evaluation
}

func evaluatePieces(board *chess.Board, black bool) (result float64) {
    const (
        pawnValue   float64 = 1.0
        knightValue float64 = 3.0
        bishopValue float64 = 3.2
        rookValue   float64 = 5.0
        queenValue  float64 = 9.0
    )

    endgameWeight := 0.0

    for _, piece := range board.GetPiecesForSide(black) {
        switch piece.Kind {
        case chess.Pawn:
            result += pawnValue
            result += evaluatePieceSquareTables(piece.Square, endgameWeight, &pieceSquareTablePawnMiddlegame, &pieceSquareTablePawnEndgame)

        case chess.Knight:
            result += knightValue
            result += evaluatePieceSquareTables(piece.Square, endgameWeight, &pieceSquareTableKnightMiddlegame, &pieceSquareTableKnightEndgame)

        case chess.Bishop:
            result += bishopValue
            result += evaluatePieceSquareTables(piece.Square, endgameWeight, &pieceSquareTableBishopMiddlegame, &pieceSquareTableBishopEndgame)

        case chess.Rook:
            result += rookValue
            result += evaluatePieceSquareTables(piece.Square, endgameWeight, &pieceSquareTableRookMiddlegame, &pieceSquareTableRookEndgame)

        case chess.Queen:
            result += queenValue
            result += evaluatePieceSquareTables(piece.Square, endgameWeight, &pieceSquareTableQueenMiddlegame, &pieceSquareTableQueenEndgame)

        case chess.King:
            result += evaluatePieceSquareTables(piece.Square, endgameWeight, &pieceSquareTableKingMiddlegame, &pieceSquareTableKingEndgame)
        }
    }

    return
}

func evaluateCastlingRights(board *chess.Board, black bool) float64 {
    const castlingRightsBonus float64 = 0.5

    canCastleKingside, canCastleQueenside := board.GetCastlingRights(black)

    if canCastleKingside || canCastleQueenside {
        return castlingRightsBonus
    } else {
        return 0.0
    }
}

func evaluatePieceSquareTables(sq chess.Square, endgameWeight float64, middlegameTable *[64]int, endgameTable *[64]int) float64 {
    tableIndex := 63 - uint(sq)
    middlegameSquareValue := float64(middlegameTable[tableIndex]) * 0.01
    endgameSquareValue := float64(endgameTable[tableIndex]) * 0.01
    return mix(middlegameSquareValue, endgameSquareValue, endgameWeight)
}

func mix(a float64, b float64, t float64) float64 {
    return a + (b - a) * t
}
