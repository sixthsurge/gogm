package chess

type Bitboard uint64

const EmptyBitboard Bitboard = Bitboard(0)

func (bitboard Bitboard) Get(sq Square) bool {
	return bitboard & (1 << Bitboard(sq)) != EmptyBitboard
}

func (bitboard Bitboard) Set(sq Square) Bitboard {
	return bitboard | 1 << Bitboard(sq)
}

func (bitboard Bitboard) Unset(sq Square) Bitboard {
	return bitboard & ^(1 << Bitboard(sq))
}

func (bitboard Bitboard) IntersectsSquare(sq Square) bool {
	return bitboard & EmptyBitboard.Set(sq) != EmptyBitboard
}
