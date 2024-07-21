package main

// https://www.chessprogramming.org/Perft

import (
	"fmt"
	"gogm/chess"
	"log"
)

var red   string = "\033[31m"
var green string = "\033[32m"
var reset string = "\033[0m"

const detectIllegalMoves bool = false

func perft(board *chess.Board, depth int) uint64 {
    if depth == 0 {
        return 1
    }

    nodeCount := uint64(0)

    for _, move := range board.GetLegalMoves() {
        unmove := board.MakeMove(move)

        if detectIllegalMoves {
            if board.DetectIllegalMove() {
                board.UnmakeMove(unmove)
                log.Fatalf("illegal move detected %v in position %v\n", move, board.Fen())
            }
        }

        nodeCount += perft(board, depth - 1)
        board.UnmakeMove(unmove)
    }

    return nodeCount
}

func testPosition(fen string, expectedResults []uint64) {
    fmt.Printf("Testing position %v\n", fen)

    for depth := 0; depth < len(expectedResults); depth++ {
        board, err := chess.LoadFen(fen)

        if err != nil {
            panic(err)
        }

        result := perft(board, depth)

        fmt.Printf("Depth %v Result %v Expected %v ", depth, result, expectedResults[depth])

        if result == expectedResults[depth] {
            fmt.Printf("%vCorrect%v\n", green, reset)
        } else {
            fmt.Printf("%vIncorrect%v\n", red, reset)
        }
    }
}

func divide(fen string, depth int) {
    board, err := chess.LoadFen(fen)
    board.MakeMove(chess.Move{ Source: chess.C3, Destination: chess.B1 })
    board.MakeMove(chess.Move{ Source: chess.B4, Destination: chess.B3 })

    if err != nil {
        panic(err)
    }

    var total uint64 = 0

    for _, move := range board.GetLegalMoves() {
        unmove := board.MakeMove(move)
        perftResult := perft(board, depth)
        board.UnmakeMove(unmove)
        total += perftResult
        fmt.Printf("%v - %v\n", move, perftResult)
    }

    fmt.Printf("Total: %v\n", total)
}

func main() {
    // Starting position
    testPosition(
        "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
        []uint64 {
            1,
            20,
            400,
            8902,
            197281,
            4865609,
            119060324,
        },
    )

    // Kiwipete
    testPosition(
        "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq -",
        []uint64 {
            1,
            48,
            2039,
            97862,
            4085603,
            193690690,
            8031647685,
        },
    )

    // "Position 5"
    testPosition(
        "rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8  ",
        []uint64 {
            1,
            44,
            1486,
            62379,
            2103487,
            89941194,
        },
    )
}
