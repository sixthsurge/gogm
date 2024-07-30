package botv1

import (
	"gogm/chess"
	"math"
)

type BotV1 struct {}

// Depth to search all legal moves to (ply)
const searchDepth int = 4

// Depth to search captures only at the end of the main search (ply)
const quiescenceSearchDepth int = 4

func (bot *BotV1) Think(board *chess.Board) chess.Move {
    bestMove, _ := search(searchDepth, board, math.Inf(-1), math.Inf(1))
    return bestMove
}

// Negamax search with alpha-beta pruning
// Alpha and beta are used to prune large portions of the game tree using the observation that
// if we have already evaluated one option and are currently evaluating another, if any of the
// opponent's options lead to an evaluation that is worse than that of the move already evaluated,
// there is no need to continue exploring that option as the original option is guaranteed to be
// better
// Alpha: the minimum evaluation that we can be assured of
// Beta: the maximum evaluation that our opponent can be assured of
func search(depth int, board *chess.Board, alpha float64, beta float64) (bestMove chess.Move, bestEval float64) {
    if depth <= 0 {
        // At the end of the main search, perform a quiescence search to avoid the horizon effect
        eval := quiescenceSearch(quiescenceSearchDepth, board, alpha, beta)
        return chess.Move{}, eval
    }

    // Get legal moves from the current position
    moves := board.GetLegalMoves(false)

    if len(moves) == 0 {
        // Checkmate or stalemate
        return chess.Move{}, evaluate(board)
    }

    bestMove = moves[0]

    for _, move := range moves {
        unmove := board.MakeMove(move)

        // Continue the search from the opponent's perspective
        _, eval := search(depth - 1, board, -beta, -alpha)
        eval = -eval

        board.UnmakeMove(unmove)

        if eval >= beta {
            // "Fail hard": the evaluation of this node is greater than or equal to the maximum (worst)
            // evaluation the opponent is already assured of by another branch. This means that the
            // opponent will never play into this line so there is no point continuing to explore
            // it
            return bestMove, beta // what move is returned here should not matter
        }

        if eval > alpha {
            // The evaluation returned by this node is greater than our best evaluation attained
            // so far, meaning we have found a new best move
            alpha = eval
            bestMove = move
        }
    }

    return bestMove, alpha
}

// A second search performed at the end of the main search intended to only evaluate "quiet"
// positions with no tension between pieces. This is needed to avoid the horizon effect
func quiescenceSearch(depth int, board *chess.Board, alpha float64, beta float64) float64 {
    // Current evaluation used to establish a lower bound for the score
    standPat := evaluate(board)

    // If the evaluation of the current position is better than the maximum (worst) evaluation our
    // opponent is already assured of, we can fail hard here as the opponent will never play into this line
    if (standPat >= beta) {
        return beta
    }

    // If the evaluation of the current position is better than the maximum (best) evaluation we
    // are already assured of, update our maximum evaluation
    if (standPat > alpha) {
        alpha = standPat
    }

    // Get all legal capturing moves
    legalCaptures := board.GetLegalMoves(true)

    if len(legalCaptures) == 0 || depth <= 0 {
        // Checkmate or stalemate
        return evaluate(board)
    }

    for _, move := range legalCaptures {
        unmove := board.MakeMove(move)

        // Continue the search from the opponent's perspective
        eval := -quiescenceSearch(depth - 1, board, -beta, -alpha)

        board.UnmakeMove(unmove)

        if eval >= beta {
            // Fail hard
            return beta
        }

        if eval > alpha {
            // New best move
            alpha = eval
        }
    }

    return alpha
}

