package chess

type Bot interface {
	Think(*Board) Move
}
