package main

import (
	"gogm/botv1"
	"gogm/chess"
	"gogm/chessgui"
)

func main() {
    bot := botv1.BotV1 {}

    board, err := chess.LoadFen(chess.StartingPositionFen)
    if err != nil {
        panic(err)
    }

    chessgui.Run(board, nil, &bot)
}
