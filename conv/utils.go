package conv

import (
	"github.com/rs/zerolog/log"

	"github.com/ericlagergren/decimal"
)

// Divide two uint64 numbers with a 10^prec precision and return the result in the same format
func Divide(x, y uint64, xprec, yprec, prec int) uint64 {
	xDec := new(decimal.Big).SetUint64(x)
	xDec.Quo(xDec, new(decimal.Big).SetUint64(y))
	xDec.Mul(xDec, decimal.New(10, -1*(xprec-yprec+prec-1))).RoundToInt()
	z, _ := xDec.Uint64()
	return z
}

// Multiply two uint64 numbers with a 10^prec precision and return the result in the same format
func Multiply(x, y uint64, xprec, yprec, prec int) uint64 {
	xDec := new(decimal.Big).SetUint64(x)
	xDec.Mul(xDec, new(decimal.Big).SetUint64(y))
	xDec.Quo(xDec, decimal.New(10, -1*(xprec+yprec-prec-1))).RoundToInt()
	z, ok := xDec.Uint64()
	if !ok {
		log.Warn().
			Str("section", "convert").
			Uint64("x", x).
			Uint64("y", y).
			Int("xprec", xprec).
			Int("yprec", yprec).
			Int("prec", prec).
			Str("rez", xDec.String()).
			Msg("Unable to properly multiply two numbers as uint64. Number probably exceeds alowed bounds")
	}
	return z
}

// Max returns the maximum value between two numbers
func Max(x, y uint64) uint64 {
	if x > y {
		return x
	}
	return y
}

// Min returns the minimum value between two numbers
func Min(x, y uint64) uint64 {
	if x < y {
		return x
	}
	return y
}
