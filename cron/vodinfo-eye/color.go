package main

const BLACK_MAX_VAL = 10000
const WHITE_MIN_VAL = 60000

func isWhite(r, g, b uint32) bool {
	return r > WHITE_MIN_VAL && g > WHITE_MIN_VAL && b > WHITE_MIN_VAL
}

func isBlue(r, g, b uint32) bool {
	return r < BLACK_MAX_VAL && g < BLACK_MAX_VAL && b > WHITE_MIN_VAL
}

func isBlack(r, g, b uint32) bool {
	return r <= BLACK_MAX_VAL && g <= BLACK_MAX_VAL && b <= BLACK_MAX_VAL
}
