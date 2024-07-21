package chess

import (
	"fmt"
	"math/bits"
)

// Implements the "magic bitboards" approach to sliding piece move generation
type SlidingAttackTable struct {
	attackSetBitboards     [64][]Bitboard
	magics                 [64]Bitboard
	relevantBits           [64]uint
	relevantOccupancyMasks [64]Bitboard
}

func (table SlidingAttackTable) GetAttackSet(sq Square, allPiecesBitboard Bitboard) Bitboard {
	relevantOccupancyBitboard := allPiecesBitboard & table.relevantOccupancyMasks[uint(sq)]

	key := (relevantOccupancyBitboard * table.magics[uint(sq)]) >> (64 - table.relevantBits[uint(sq)])

	return table.attackSetBitboards[uint(sq)][key]
}

func CreateSlidingAttackTable(
	magics                 [64]Bitboard,
	relevantBits           [64]uint,
	relevantOccupancyMasks [64]Bitboard,
	attackSetGenerator     func(Square, Bitboard) Bitboard,
) SlidingAttackTable {
	var attackSetBitboards [64][]Bitboard

	for squareIndex := 0; squareIndex < 64; squareIndex++ {
		tableSize := 1 << relevantBits[squareIndex]
		attackSetBitboards[squareIndex] = make([]Bitboard, tableSize, tableSize)

		for _, relevantOccupancyBitboard := range allRelevantOccupacyBitboards(relevantOccupancyMasks[squareIndex]) {
			key := (relevantOccupancyBitboard * magics[squareIndex]) >> (64 - relevantBits[squareIndex])
			attackSetBitboards[squareIndex][key] = attackSetGenerator(Square(squareIndex), relevantOccupancyBitboard)
		}
	}

	return SlidingAttackTable{ attackSetBitboards, magics, relevantBits, relevantOccupancyMasks }
}

func CreateBishopAttackTable() SlidingAttackTable {
	return CreateSlidingAttackTable(bishopMagicNumbers, bishopRelevantBits, bishopRelevantOccupancyMasks, bishopAttacks)
}

func CreateRookAttackTable() SlidingAttackTable {
	return CreateSlidingAttackTable(rookMagicNumbers, rookRelevantBits, rookRelevantOccupancyMasks, rookAttacks)
}

func PawnMoveSet(
	sq                     Square,
	isBlack                bool,
	friendlyPiecesBitboard Bitboard,
	enemyPiecesBitboard    Bitboard,
	kingSquare             Square,
	board                  *Board,
) (result Bitboard) {
	allPiecesBitboard := friendlyPiecesBitboard | enemyPiecesBitboard

	var forwardsDirection int
	var initialRank Rank

	if isBlack {
		forwardsDirection = 1
		initialRank = Rank7
	} else {
		forwardsDirection = -1
		initialRank = Rank2
	}

	// One-square advance
	advanceSq, advanceSqValid := sq.Offset(0, forwardsDirection)
	if advanceSqValid {
		result |= EmptyBitboard.Set(advanceSq) & ^allPiecesBitboard
	}

	// Two-square thrust
	if sq.Rank() == initialRank {
		thrustSq, thrustSqValid := sq.Offset(0, 2*forwardsDirection)
		advanceSqEmpty := EmptyBitboard.Set(advanceSq) & allPiecesBitboard == EmptyBitboard
		if thrustSqValid && advanceSqEmpty {
			result |= EmptyBitboard.Set(thrustSq) & ^allPiecesBitboard
		}
	}

	// Captures
	var attackSetBitboard Bitboard
	var enPassantBitboard Bitboard

	if isBlack {
		attackSetBitboard = blackPawnAttackSets[uint(sq)]
	} else {
		attackSetBitboard = whitePawnAttackSets[uint(sq)]
	}

	if board.hasEnPassantTarget {
		// Handle the annoying en passant pin, when the capturing pawn and the captured pawn
		// are the only two pieces blocking an attack by a rook against the king on the same rank
		enPassantPin := detectEnPassantPin(sq, kingSquare, board)

		if !enPassantPin {
			enPassantBitboard = EmptyBitboard.Set(board.enPassantTarget)
		}
	}

	result |= attackSetBitboard & (enemyPiecesBitboard | enPassantBitboard)

	return
}

func GetPieceAttackSet(piece Piece, allPiecesBitboard Bitboard, board *Board) Bitboard {
	switch piece.Kind {
	case Pawn:
		if piece.IsBlack {
			return blackPawnAttackSets[uint(piece.Square)]
		} else {
			return whitePawnAttackSets[uint(piece.Square)]
		}

	case Knight:
		return knightAttackSets[uint(piece.Square)]

	case Bishop:
		return board.bishopAttackTable.GetAttackSet(piece.Square, allPiecesBitboard)

	case Rook:
		return board.rookAttackTable.GetAttackSet(piece.Square, allPiecesBitboard)

	case Queen:
		return board.rookAttackTable.GetAttackSet(piece.Square, allPiecesBitboard) |
			board.bishopAttackTable.GetAttackSet(piece.Square, allPiecesBitboard)

	case King:
		return kingAttackSets[uint(piece.Square)]
	}

	return EmptyBitboard
}

func (board *Board) GetLegalMoves() []Move {
	moves := make([]Move, 0, 256)

	friendlyPieces := board.GetPiecesForSide(board.blackToMove)
	enemyPieces := board.GetPiecesForSide(!board.blackToMove)

	friendlyPiecesBitboard := board.GetPiecesBitboard(board.blackToMove)
	enemyPiecesBitboard := board.GetPiecesBitboard(!board.blackToMove)
	allPiecesBitboard := friendlyPiecesBitboard | enemyPiecesBitboard

	kingSquare := board.GetKingSquare(board.blackToMove)
	kingDangerMask, checkingPieces := getKingDangerMaskAndCheckingPieces(
		enemyPieces,
		allPiecesBitboard,
		kingSquare,
		board,
	)
	pinMask := getPinMask(
		enemyPieces,
		allPiecesBitboard,
		kingSquare,
		board,
	)
	isCheck := kingDangerMask.IntersectsSquare(kingSquare)

	var promotionRank Rank
	if board.blackToMove {
		promotionRank = Rank1
	} else {
		promotionRank = Rank8
	}

	for _, piece := range friendlyPieces {
		// Get pseudolegal move set
		var moveSet Bitboard

		// If we are in double check, filter only for king moves
		if len(checkingPieces) > 1 && piece.Kind != King {
			continue
		}

		// Get bitboard of pseudolegal destination squares
		if piece.Kind == Pawn {
			moveSet = PawnMoveSet(piece.Square, piece.IsBlack, friendlyPiecesBitboard, enemyPiecesBitboard, kingSquare, board)
		} else {
			moveSet = GetPieceAttackSet(piece, friendlyPiecesBitboard | enemyPiecesBitboard, board) & ^friendlyPiecesBitboard
		}

		// Prevent king from walking into danger
		if piece.Kind == King {
			moveSet &= ^kingDangerMask
		}

		// Handle pins
		isPinned := pinMask.IntersectsSquare(piece.Square)
		if isPinned {
			kingToPieceH := int(piece.Square.File()) - int(kingSquare.File())
			kingToPieceV := int(piece.Square.Rank()) - int(kingSquare.Rank())
			pinRayBitboard := rayBitboard(kingSquare, EmptyBitboard, signum(kingToPieceH), signum(kingToPieceV))
			moveSet &= pinRayBitboard
		}

		// If we are in single check, filter only for king moves or moves that capture the checking
		// piece
		if len(checkingPieces) == 1 {
			if piece.Kind != King {
				checkingPiece := checkingPieces[0]

				// Capturing the checking piece
				validMovesMask := EmptyBitboard.Set(checkingPiece.Square)

				// Interpositions
				if checkingPiece.Kind == Queen || checkingPiece.Kind == Rook || checkingPiece.Kind == Bishop {
					kingToPieceH := int(checkingPiece.Square.File()) - int(kingSquare.File())
					kingToPieceV := int(checkingPiece.Square.Rank()) - int(kingSquare.Rank())
					rayBitboard := rayBitboard(kingSquare, allPiecesBitboard, signum(kingToPieceH), signum(kingToPieceV))
					validMovesMask |= rayBitboard
				}

				moveSet &= validMovesMask
			}
		}

		// Add all moves in move set to the list of legal moves
		bitIndex := 0
		for v := uint64(moveSet); v != 0; {
			trailingZeroes := bits.TrailingZeros64(v)
			bitIndex += trailingZeroes
			v >>= trailingZeroes + 1

			destinationSquare := Square(bitIndex)

			// Handle promotions
			if piece.Kind == Pawn && destinationSquare.Rank() == promotionRank {
				moves = append(moves, Move{
					Source: piece.Square,
					Destination: destinationSquare,
					IsPromotion: true,
					PromotedPiece: Queen,
				});
				moves = append(moves, Move{
					Source: piece.Square,
					Destination: destinationSquare,
					IsPromotion: true,
					PromotedPiece: Rook,
				});
				moves = append(moves, Move{
					Source: piece.Square,
					Destination: destinationSquare,
					IsPromotion: true,
					PromotedPiece: Bishop,
				});
				moves = append(moves, Move{
					Source: piece.Square,
					Destination: destinationSquare,
					IsPromotion: true,
					PromotedPiece: Knight,
				});
			} else {
				moves = append(moves, Move{
					Source: piece.Square,
					Destination: destinationSquare,
				});
			}

			bitIndex += 1
		}
	}

	// Consider castling
	var backRank Rank
	if board.blackToMove {
		backRank = Rank8
	} else {
		backRank = Rank1
	}

	kingsideCastlingRight, queensideCastlingRight := board.GetCastlingRights(board.blackToMove)

	kingsideCastleOccupancyMask := EmptyBitboard.Set(SquareAt(FileF, backRank)).Set(SquareAt(FileG, backRank))
	queensideCastleDangerMask := EmptyBitboard.Set(SquareAt(FileD, backRank)).Set(SquareAt(FileC, backRank))
	queensideCastleOccupancyMask := queensideCastleDangerMask.Set(SquareAt(FileB, backRank))

	// Make sure there is still a black rook on A1 (i.e. it wasn't captured)
	var aRook, hRook bool
	if aPiece := board.GetPiece(SquareAt(FileA, backRank)); aPiece != nil {
		aRook = aPiece.Kind == Rook && aPiece.IsBlack == board.blackToMove
	}
	if hPiece := board.GetPiece(SquareAt(FileH, backRank)); hPiece != nil {
		hRook = hPiece.Kind == Rook && hPiece.IsBlack == board.blackToMove
	}

	canCastleKingside := kingsideCastlingRight &&
		kingsideCastleOccupancyMask & allPiecesBitboard == EmptyBitboard &&
		kingsideCastleOccupancyMask & kingDangerMask == EmptyBitboard &&
		hRook && !isCheck

	canCastleQueenside := queensideCastlingRight &&
		queensideCastleOccupancyMask & allPiecesBitboard == EmptyBitboard &&
		queensideCastleDangerMask & kingDangerMask == EmptyBitboard &&
		aRook && !isCheck

	if canCastleKingside {
		moves = append(moves, Move {
			Source: SquareAt(FileE, backRank),
			Destination: SquareAt(FileG, backRank),
		})
	}

	if canCastleQueenside {
		moves = append(moves, Move {
			Source: SquareAt(FileE, backRank),
			Destination: SquareAt(FileC, backRank),
		})
	}

	return moves
}

func (board *Board) GetLegalMovesFromSquare(sq Square) (result []Move) {
	legalMoves := board.GetLegalMoves()

	for _, move := range legalMoves {
		if move.Source == sq {
			result = append(result, move)
		}
	}

	return
}

func (board *Board) IsCheck() bool {
	friendlyPiecesBitboard := board.GetPiecesBitboard(board.blackToMove)
	enemyPiecesBitboard := board.GetPiecesBitboard(!board.blackToMove)

	kingSquare := board.GetKingSquare(board.blackToMove)
	kingDangerMask, _ := getKingDangerMaskAndCheckingPieces(
		board.GetPiecesForSide(!board.blackToMove),
		friendlyPiecesBitboard | enemyPiecesBitboard,
		kingSquare,
		board,
	)

	return kingDangerMask.IntersectsSquare(kingSquare)
}

// Returns the bitboard of squares on which the king would be placed in check
// This is the set of squares attacked by enemy pieces, with the king excluded as a blocker -
// the king cannot block an attack against itself
func getKingDangerMaskAndCheckingPieces(
	enemyPieces       []Piece,
	allPiecesBitboard Bitboard,
	kingSquare        Square,
	board             *Board,
) (result Bitboard, checkingPieces []Piece) {
	allPiecesExceptKingBitboard := allPiecesBitboard.Unset(kingSquare)

	for _, piece := range enemyPieces {
		attackSet := GetPieceAttackSet(piece, allPiecesExceptKingBitboard, board)
		result |= attackSet

		if attackSet.IntersectsSquare(kingSquare) {
			checkingPieces = append(checkingPieces, piece)
		}
	}

	return
}

// Returns the bitboard of squares containing pieces that are pinned to the king
func getPinMask(
	enemyPieces       []Piece,
	allPiecesBitboard Bitboard,
	kingSquare        Square,
	board             *Board,
) Bitboard {
	// Approach: a piece is pinned if it is attacked by an enemy rook/bishop and both the pinned
	// piece and the pinning piece would be attacked by our king if it were an enemy rook/bishop

	// Bitboard of all squares attacked by enemy rooks and queens
	var enemyRookAttacks Bitboard

	// Bitboard of all squares attacked by enemy bishops
	var enemyBishopAttacks Bitboard

	// Bitboard of squares that would be attacked by a rook on the same square as our king
	kingRookAttacks := board.rookAttackTable.GetAttackSet(kingSquare, allPiecesBitboard)

	// Bitboard of squares that would be attacked by a bishop on the same square as our king
	kingBishopAttacks := board.bishopAttackTable.GetAttackSet(kingSquare, allPiecesBitboard)

	for _, enemyPiece := range enemyPieces {
		if (enemyPiece.Kind == Rook || enemyPiece.Kind == Queen) && unobstructedRookAttacks[uint(kingSquare)].IntersectsSquare(enemyPiece.Square) {
			enemyRookAttacks |= board.rookAttackTable.GetAttackSet(enemyPiece.Square, allPiecesBitboard)
		}
		if (enemyPiece.Kind == Bishop || enemyPiece.Kind == Queen) && unobstructedBishopAttacks[uint(kingSquare)].IntersectsSquare(enemyPiece.Square) {
			enemyBishopAttacks |= board.bishopAttackTable.GetAttackSet(enemyPiece.Square, allPiecesBitboard)
		}
	}

	return (enemyRookAttacks & kingRookAttacks) | (enemyBishopAttacks & kingBishopAttacks)
}

// Given a bitboard, returns the bitboards for all combinations of pieces
// occupying only the marked squares
func allRelevantOccupacyBitboards(bitboard Bitboard) []Bitboard {
	// Find the total number of combinations
	numSetBits := bits.OnesCount64(uint64(bitboard))
	numCombinations := 1 << uint(numSetBits);

	// Find the indices of all ones in the set
	setBitsIndices := make([]int, numSetBits, numSetBits)
	bitIndex := 0
	for i, v := 0, uint64(bitboard); v != 0; i++ {
		trailingZeroes := bits.TrailingZeros64(v)
		bitIndex += trailingZeroes
		v >>= trailingZeroes + 1
		setBitsIndices[i] = bitIndex
		bitIndex += 1
	}

	// For each combination, project the bits onto those at the marked square indiceso
	relevantOccupancyBitboards := make([]Bitboard, numCombinations, numCombinations)
	for combination := 0; combination < numCombinations; combination++ {
		relevantOccupancyBitboard := EmptyBitboard

		for index, squareIndex := range setBitsIndices {
			bit := 1 << index

			if (combination & bit) != 0 {
				relevantOccupancyBitboard = relevantOccupancyBitboard.Set(Square(squareIndex))
			}
		}

		relevantOccupancyBitboards[combination] = relevantOccupancyBitboard
	}

	return relevantOccupancyBitboards
}

func rayBitboard(square Square, occupancyBitboard Bitboard, offsetH int, offsetV int) (result Bitboard) {
	for {
		sq, sqValid := square.Offset(offsetH, offsetV)
		if !sqValid {
			break
		}

		result = result.Set(sq)
		if occupancyBitboard.Get(sq) {
			break
		}

		square = sq
	}

	return
}

func bishopAttacks(square Square, occupancyBitboard Bitboard) Bitboard {
	return rayBitboard(square, occupancyBitboard, -1, -1) |
		rayBitboard(square, occupancyBitboard,  1, -1) |
		rayBitboard(square, occupancyBitboard, -1,  1) |
		rayBitboard(square, occupancyBitboard,  1,  1)
}

func rookAttacks(square Square, occupancyBitboard Bitboard) Bitboard {
	return rayBitboard(square, occupancyBitboard, -1, 0) |
		rayBitboard(square, occupancyBitboard, 1,  0) |
		rayBitboard(square, occupancyBitboard, 0,  1) |
		rayBitboard(square, occupancyBitboard, 0, -1)
}

// Detect the annoying en passant pin, when the capturing pawn and the captured pawn
// are the only two pieces blocking an attack by a rook against the king on the same rank
func detectEnPassantPin(pawnSquare Square, kingSquare Square, board *Board) bool {
	var rank Rank
	if board.blackToMove {
		rank = Rank4
	} else {
		rank = Rank5
	}

	if kingSquare.Rank() != rank || pawnSquare.Rank() != rank {
		return false
	}

	var potentialPinningRooks [2]Square
	numPotentialPinningRooks := 0

	for _, enemyPiece := range board.GetPiecesForSide(!board.blackToMove) {
		if enemyPiece.Kind == Rook && enemyPiece.Square.Rank() == rank {
			potentialPinningRooks[numPotentialPinningRooks] = enemyPiece.Square
			numPotentialPinningRooks += 1
		}
	}

	if numPotentialPinningRooks == 0 {
		return false
	}

	// March from king square to rook square, loooking for intervening pieces that aren't the capturing
	// pawn or the captured pawn
	for i := 0; i < numPotentialPinningRooks; i++ {
		rookSquare := potentialPinningRooks[i]
		marchDirection := signum(int(rookSquare.File()) - int(kingSquare.File()))
		capturedPawnSquare := SquareAt(board.enPassantTarget.File(), rank)

		for fileIndex := int(kingSquare.File()) + marchDirection; fileIndex != int(rookSquare.File()); fileIndex += marchDirection {
			if fileIndex < 0 || fileIndex >= 8 {
				break
			}

			piece := board.GetPiece(SquareAt(File(fileIndex), rank))
			if piece == nil {
				continue
			}

			isCapturingPawn := piece.Square == pawnSquare
			isCapturedPawn := piece.Square == capturedPawnSquare

			if !isCapturedPawn && !isCapturingPawn {
				return false
			}
		}
	}
	fmt.Print(board.Fen())


	return true
}

// Magic numbers, relevant bit counts and relevant occupancy masks for rook and bishop move lookup tables
var rookMagicNumbers           [64]Bitboard = [64]Bitboard {9259400972386469971, 378302682768609280, 432363709392882176, 792669819509149696, 72066390400761862, 3170696865660274184, 1297037800784303112, 4647724711070220416, 4621115431220971552, 9223935263835750912, 36451421809283073, 288371182435059713, 18155170357837952, 4630263409842586640, 577023710847092737, 45317473469333539, 4611827305877078112, 2904830830706688580, 1152992973133185794, 342560544483450920, 108228778483778560, 282574823891480, 2891324155045679234, 1126037350074400, 18155146737369096, 18049583955431424, 1153308569208094848, 72066392286826496, 1776900986372352, 9223409422399963264, 5764609739238410520, 9224515531128766592, 5875016085073297442, 1585302321930175552, 704374661713920, 5084146078588940, 11745466994378413313, 5512828164964881408, 9225711969732920610, 9232942392284283008, 72239288342839296, 306315145569697824, 576479445611774016, 6953593077768454216, 287006911430672, 180781701872517248, 180706952246067208, 142010875916, 337840929260044800, 9403691945961718400, 2328362177773175424, 140771849142400, 5630049374437760, 10450040056571232768, 4611967510617523456, 585471421900161088, 72077798662996233, 882706631853350929, 4900479379968131474, 1450198702706655489, 1234550484871155714, 36310289176725537, 8804951458436, 9224570659152134278}
var rookRelevantBits           [64]uint = [64]uint {12, 11, 11, 11, 11, 11, 11, 12, 11, 10, 10, 10, 10, 10, 10, 12, 11, 10, 10, 10, 10, 10, 10, 12, 11, 10, 10, 10, 10, 10, 10, 12, 11, 10, 10, 10, 10, 10, 10, 12, 11, 10, 10, 10, 10, 10, 10, 12, 11, 10, 10, 10, 10, 10, 10, 12, 12, 11, 11, 11, 11, 11, 11, 12}
var rookRelevantOccupancyMasks [64]Bitboard = [64]Bitboard {282578800148862, 565157600297596, 1130315200595066, 2260630401190006, 4521260802379886, 9042521604759646, 18085043209519166, 36170086419038334, 282578800180736, 565157600328704, 1130315200625152, 2260630401218048, 4521260802403840, 9042521604775424, 18085043209518592, 36170086419037696, 282578808340736, 565157608292864, 1130315208328192, 2260630408398848, 4521260808540160, 9042521608822784, 18085043209388032, 36170086418907136, 282580897300736, 565159647117824, 1130317180306432, 2260632246683648, 4521262379438080, 9042522644946944, 18085043175964672, 36170086385483776, 283115671060736, 565681586307584, 1130822006735872, 2261102847592448, 4521664529305600, 9042787892731904, 18085034619584512, 36170077829103616, 420017753620736, 699298018886144, 1260057572672512, 2381576680245248, 4624614895390720, 9110691325681664, 18082844186263552, 36167887395782656, 35466950888980736, 34905104758997504, 34344362452452352, 33222877839362048, 30979908613181440, 26493970160820224, 17522093256097792, 35607136465616896, 9079539427579068672, 8935706818303361536, 8792156787827803136, 8505056726876686336, 7930856604974452736, 6782456361169985536, 4485655873561051136, 9115426935197958144}

var bishopMagicNumbers           [64]Bitboard = [64]Bitboard {1197960816612343840, 4617882883208794114, 653093423084995074, 4617386739290521616, 14989110974533862472, 572914420221952, 4908998412701533200, 2306124760412063760, 585476782154056016, 18017998161904128, 72392550023438368, 3458804182147399905, 9259420651123378193, 2306970077755607104, 435029482951155712, 648887928310865920, 1162529866810458368, 633971801333888, 650770163337013762, 423690235822100, 577587822632894465, 10376857041717168152, 4648841011882624096, 141841941139536, 38316060780400640, 13537191561528131, 288318389024809024, 9800396290612297738, 216317917724155908, 290764750756774016, 6923163202255728640, 1688987399946496, 1171222463469592576, 577596583250494016, 1165332809042758657, 157637533978067456, 2323859615337422976, 4503883095228672, 293583075781183497, 1209219799029448834, 4693331422714283008, 79184970449920, 13792415643078688, 1105014558976, 450364374104806400, 2314859073283040320, 9008307524117124, 4756366364089260544, 3458909666573025296, 4622140733466347521, 3479059935378539538, 18157336128520192, 2449958335278022658, 35872145809408, 2287265841741824, 2308097025236406560, 9264187854190821376, 577950457856, 288230453478983712, 7512461579728978944, 292734285302071809, 220711703620370826, 1152925937684251661, 666541566856396826}
var bishopRelevantBits           [64]uint   = [64]uint {6, 5, 5, 5, 5, 5, 5, 6, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 7, 7, 7, 7, 5, 5, 5, 5, 7, 9, 9, 7, 5, 5, 5, 5, 7, 9, 9, 7, 5, 5, 5, 5, 7, 7, 7, 7, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 6, 5, 5, 5, 5, 5, 5, 6}
var bishopRelevantOccupancyMasks [64]Bitboard = [64]Bitboard {18049651735527936, 70506452091904, 275415828992, 1075975168, 38021120, 8657588224, 2216338399232, 567382630219776, 9024825867763712, 18049651735527424, 70506452221952, 275449643008, 9733406720, 2216342585344, 567382630203392, 1134765260406784, 4512412933816832, 9024825867633664, 18049651768822272, 70515108615168, 2491752130560, 567383701868544, 1134765256220672, 2269530512441344, 2256206450263040, 4512412900526080, 9024834391117824, 18051867805491712, 637888545440768, 1135039602493440, 2269529440784384, 4539058881568768, 1128098963916800, 2256197927833600, 4514594912477184, 9592139778506752, 19184279556981248, 2339762086609920, 4538784537380864, 9077569074761728, 562958610993152, 1125917221986304, 2814792987328512, 5629586008178688, 11259172008099840, 22518341868716544, 9007336962655232, 18014673925310464, 2216338399232, 4432676798464, 11064376819712, 22137335185408, 44272556441600, 87995357200384, 35253226045952, 70506452091904, 567382630219776, 1134765260406784, 2832480465846272, 5667157807464448, 11333774449049600, 22526811443298304, 9024825867763712, 18049651735527936}

// Attack set bitboards for non-sliding pieces
var whitePawnAttackSets [64]Bitboard = [64]Bitboard {0, 0, 0, 0, 0, 0, 0, 0, 2, 5, 10, 20, 40, 80, 160, 64, 512, 1280, 2560, 5120, 10240, 20480, 40960, 16384, 131072, 327680, 655360, 1310720, 2621440, 5242880, 10485760, 4194304, 33554432, 83886080, 167772160, 335544320, 671088640, 1342177280, 2684354560, 1073741824, 8589934592, 21474836480, 42949672960, 85899345920, 171798691840, 343597383680, 687194767360, 274877906944, 2199023255552, 5497558138880, 10995116277760, 21990232555520, 43980465111040, 87960930222080, 175921860444160, 70368744177664, 562949953421312, 1407374883553280, 2814749767106560, 5629499534213120, 11258999068426240, 22517998136852480, 45035996273704960, 18014398509481984}
var blackPawnAttackSets [64]Bitboard = [64]Bitboard {512, 1280, 2560, 5120, 10240, 20480, 40960, 16384, 131072, 327680, 655360, 1310720, 2621440, 5242880, 10485760, 4194304, 33554432, 83886080, 167772160, 335544320, 671088640, 1342177280, 2684354560, 1073741824, 8589934592, 21474836480, 42949672960, 85899345920, 171798691840, 343597383680, 687194767360, 274877906944, 2199023255552, 5497558138880, 10995116277760, 21990232555520, 43980465111040, 87960930222080, 175921860444160, 70368744177664, 562949953421312, 1407374883553280, 2814749767106560, 5629499534213120, 11258999068426240, 22517998136852480, 45035996273704960, 18014398509481984, 144115188075855872, 360287970189639680, 720575940379279360, 1441151880758558720, 2882303761517117440, 5764607523034234880, 11529215046068469760, 4611686018427387904, 0, 0, 0, 0, 0, 0, 0, 0}
var knightAttackSets    [64]Bitboard = [64]Bitboard {132096, 329728, 659712, 1319424, 2638848, 5277696, 10489856, 4202496, 33816580, 84410376, 168886289, 337772578, 675545156, 1351090312, 2685403152, 1075839008, 8657044482, 21609056261, 43234889994, 86469779988, 172939559976, 345879119952, 687463207072, 275414786112, 2216203387392, 5531918402816, 11068131838464, 22136263676928, 44272527353856, 88545054707712, 175990581010432, 70506185244672, 567348067172352, 1416171111120896, 2833441750646784, 5666883501293568, 11333767002587136, 22667534005174272, 45053588738670592, 18049583422636032, 145241105196122112, 362539804446949376, 725361088165576704, 1450722176331153408, 2901444352662306816, 5802888705324613632, 11533718717099671552, 4620693356194824192, 288234782788157440, 576469569871282176, 1224997833292120064, 2449995666584240128, 4899991333168480256, 9799982666336960512, 1152939783987658752, 2305878468463689728, 1128098930098176, 2257297371824128, 4796069720358912, 9592139440717824, 19184278881435648, 38368557762871296, 4679521487814656, 9077567998918656}
var kingAttackSets      [64]Bitboard = [64]Bitboard {770, 1797, 3594, 7188, 14376, 28752, 57504, 49216, 197123, 460039, 920078, 1840156, 3680312, 7360624, 14721248, 12599488, 50463488, 117769984, 235539968, 471079936, 942159872, 1884319744, 3768639488, 3225468928, 12918652928, 30149115904, 60298231808, 120596463616, 241192927232, 482385854464, 964771708928, 825720045568, 3307175149568, 7718173671424, 15436347342848, 30872694685696, 61745389371392, 123490778742784, 246981557485568, 211384331665408, 846636838289408, 1975852459884544, 3951704919769088, 7903409839538176, 15806819679076352, 31613639358152704, 63227278716305408, 54114388906344448, 216739030602088448, 505818229730443264, 1011636459460886528, 2023272918921773056, 4046545837843546112, 8093091675687092224, 16186183351374184448, 13853283560024178688, 144959613005987840, 362258295026614272, 724516590053228544, 1449033180106457088, 2898066360212914176, 5796132720425828352, 11592265440851656704, 4665729213955833856}

// Attack sets for sliding pieces on an empty board
var unobstructedBishopAttacks[64]Bitboard = [64]Bitboard{9241421688590303744, 36099303471056128,141012904249856,550848566272,6480472064,1108177604608,283691315142656,72624976668147712,4620710844295151618,9241421688590368773,36099303487963146,141017232965652,1659000848424,283693466779728,72624976676520096,145249953336262720,2310355422147510788,4620710844311799048,9241421692918565393,36100411639206946,424704217196612,72625527495610504,145249955479592976,290499906664153120,1155177711057110024,2310355426409252880,4620711952330133792,9241705379636978241,108724279602332802,145390965166737412,290500455356698632,580999811184992272,577588851267340304,1155178802063085600,2310639079102947392,4693335752243822976,9386671504487645697,326598935265674242,581140276476643332,1161999073681608712,288793334762704928,577868148797087808,1227793891648880768,2455587783297826816,4911175566595588352,9822351133174399489,1197958188344280066,2323857683139004420,144117404414255168,360293502378066048,720587009051099136,1441174018118909952,2882348036221108224,5764696068147249408,11529391036782871041,4611756524879479810,567382630219904,1416240237150208,2833579985862656,5667164249915392,11334324221640704,22667548931719168,45053622886727936,18049651735527937}
var unobstructedRookAttacks [64]Bitboard  = [64]Bitboard{72340172838076926, 144680345676153597, 289360691352306939, 578721382704613623, 1157442765409226991, 2314885530818453727, 4629771061636907199, 9259542123273814143, 72340172838141441, 144680345676217602, 289360691352369924, 578721382704674568, 1157442765409283856, 2314885530818502432, 4629771061636939584, 9259542123273813888, 72340172854657281, 144680345692602882, 289360691368494084, 578721382720276488, 1157442765423841296, 2314885530830970912, 4629771061645230144, 9259542123273748608, 72340177082712321, 144680349887234562, 289360695496279044, 578721386714368008, 1157442769150545936, 2314885534022901792, 4629771063767613504, 9259542123257036928, 72341259464802561, 144681423712944642, 289361752209228804, 578722409201797128, 1157443723186933776, 2314886351157207072, 4629771607097753664, 9259542118978846848, 72618349279904001, 144956323094725122, 289632270724367364, 578984165983651848, 1157687956502220816, 2315095537539358752, 4629910699613634624, 9259541023762186368, 143553341945872641, 215330564830528002, 358885010599838724, 645993902138460168, 1220211685215703056, 2368647251370188832, 4665518383679160384, 9259260648297103488, 18302911464433844481, 18231136449196065282, 18087586418720506884, 17800486357769390088, 17226286235867156496, 16077885992062689312, 13781085504453754944, 9187484529235886208}
