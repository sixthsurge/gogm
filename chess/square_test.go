package chess_test

import (
	"github.com/stretchr/testify/assert"
	"gogm/chess"
	"testing"
)

func TestSquareAt(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(board.SquareAt(board.FileA, board.Rank1), board.A1)
	assert.Equal(board.SquareAt(board.FileB, board.Rank1), board.B1)
	assert.Equal(board.SquareAt(board.FileA, board.Rank8), board.A8)
}

func TestSquareWithAlgebraicName(t *testing.T) {
	assert := assert.New(t)

	a1, err := board.SquareWithAlgebraicName("a1")
	assert.Equal(a1, board.A1)
	assert.Nil(err)

	a2, err := board.SquareWithAlgebraicName("a2")
	assert.Equal(a2, board.A2)
	assert.Nil(err)

	b1, err := board.SquareWithAlgebraicName("b1")
	assert.Equal(b1, board.B1)
	assert.Nil(err)
}

func TestFile(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(board.A1.File(), board.File(0))
	assert.Equal(board.B1.File(), board.File(1))
	assert.Equal(board.A2.File(), board.File(0))
	assert.Equal(board.B2.File(), board.File(1))
}

func TestRank(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(board.A1.Rank(), board.Rank(7))
	assert.Equal(board.B1.Rank(), board.Rank(7))
	assert.Equal(board.A2.Rank(), board.Rank(6))
	assert.Equal(board.B2.Rank(), board.Rank(6))
}

func TestAlgebraicName(t *testing.T) {
	assert := assert.New(t)

	a1Name, err := board.A1.AlgebraicName()
	assert.Equal(a1Name, "a1")
	assert.Nil(err)

	a2Name, err := board.A2.AlgebraicName()
	assert.Equal(a2Name, "a2")
	assert.Nil(err)

	b1Name, err := board.B1.AlgebraicName()
	assert.Equal(b1Name, "b1")
	assert.Nil(err)

	_, err = board.Square(64).AlgebraicName()
	assert.NotNil(err)
}
