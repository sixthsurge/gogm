package botv1

import (
	"gogm/chess"
	"math"
)

type BotV1 struct {}

func (bot *BotV1) Think(board *chess.Board) chess.Move {
    bestMove, _ := search(4, board, math.Inf(-1), math.Inf(1))
    return bestMove
}

func search(depth int, board *chess.Board, alpha float64, beta float64) (bestMove chess.Move, bestEval float64) {
    if depth == 0 {
        return chess.Move{}, evaluate(board)
    }

    moves := board.GetLegalMoves()

    if len(moves) == 0 {
        return chess.Move{}, evaluate(board)
    }

    bestMove = moves[0]

    for _, move := range moves {
        unmove := board.MakeMove(move)

        _, eval := search(depth - 1, board, -beta, -alpha)
        eval = -eval

        board.UnmakeMove(unmove)

        if eval >= beta {
            return move, beta
        }

        if eval > alpha {
            bestMove = move
            alpha = eval
        }
    }

    return bestMove, alpha
}
