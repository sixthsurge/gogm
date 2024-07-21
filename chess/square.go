package chess

import "errors"

// A square of the board
type Square uint8

// A file (column) of the board
type File int

// A rank (row) of the board
type Rank int

// Returns the square with the given rank and file
func SquareAt(file File, rank Rank) Square {
	return Square(uint8(rank)<<3 + uint8(file))
}

// Returns the square with the given name in algebraic notation
func SquareWithAlgebraicName(name string) (Square, error) {
	if len(name) != 2 {
		return A1, errors.New("name must be two bytes")
	}

	fileIndex := int(name[0] - 'a')
	rankIndex := int('8' - name[1])

	if fileIndex < 0 || fileIndex >= 8 {
		return A1, errors.New("bad file")
	}
	if rankIndex < 0 || rankIndex >= 8 {
		return A1, errors.New("bad rank")
	}

	return SquareAt(File(fileIndex), Rank(rankIndex)), nil
}

// Returns an integer representing the file of the given square
// The A file is 0 and the H file is 7
func (sq Square) File() File {
	return File(uint8(sq) & 7)
}

// Returns an integer representing the rank of the given square
// The 8th rank is 0 and the 1st rank is 7
func (sq Square) Rank() Rank {
	return Rank(uint8(sq) >> 3)
}

// / Returns the name of the square in algebraic notation
func (sq Square) AlgebraicName() (string, error) {
	if uint8(sq) >= 64 {
		return "", errors.New("bad square")
	}

	file := sq.File()
	rank := sq.Rank()

	runes := []rune{
		rune(int('a') + int(file)),
		rune(int('8') - int(rank)),
	}

	return string(runes), nil
}

func (sq Square) String() string {
	name, err := sq.AlgebraicName()

	if err == nil {
		return name
	} else {
		return "??"
	}
}

// Returns the square offset from the current square by the given amount, and whether that square
// on the board
func (sq Square) Offset(horizontal int, vertical int) (Square, bool) {
	fileIndex := int(sq.File()) + horizontal
	rankIndex := int(sq.Rank()) + vertical

	if fileIndex >= 0 && fileIndex < 8 && rankIndex >= 0 && rankIndex < 8 {
		return SquareAt(File(fileIndex), Rank(rankIndex)), true
	} else {
		return A1, false
	}
}

const (
	A8 Square = iota
	B8
	C8
	D8
	E8
	F8
	G8
	H8
	A7
	B7
	C7
	D7
	E7
	F7
	G7
	H7
	A6
	B6
	C6
	D6
	E6
	F6
	G6
	H6
	A5
	B5
	C5
	D5
	E5
	F5
	G5
	H5
	A4
	B4
	C4
	D4
	E4
	F4
	G4
	H4
	A3
	B3
	C3
	D3
	E3
	F3
	G3
	H3
	A2
	B2
	C2
	D2
	E2
	F2
	G2
	H2
	A1
	B1
	C1
	D1
	E1
	F1
	G1
	H1
)

const (
	FileA File = iota
	FileB
	FileC
	FileD
	FileE
	FileF
	FileG
	FileH
)

const (
	Rank8 Rank = iota
	Rank7
	Rank6
	Rank5
	Rank4
	Rank3
	Rank2
	Rank1
)
