// Probably could have been done in grpc spec, but that wasn't our main focus

package proto

const InvalidMove int64 = -2
const Lost int64 = -1
const ValidMove int64 = 0
const Draw int64 = 1
const Won int64 = 2

const RegularTicTacToe int64 = 0
const UltimateTicTacToe int64 = 1
