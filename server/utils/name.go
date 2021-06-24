package utils

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func GenerateName(length int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyz")

	s := make([]rune, length)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
	/*
		return fmt.Sprintf(
			"%s-%s-%s",
			generateWord(),
			generateWord(),
			generateWord(),
		)
	*/
}

/*
func generateWord() string {
	initialConsonants := []string{
		"b", "d", "f", "g",
		"h", "j", "k", "l",
		"m", "n", "p", "r",
		"s", "t", "v", "w",
		"z", "bl", "br", "cl",
		"cr", "dr", "fl", "fr",
		"gl", "gr", "pl", "pr",
		"sk", "sl", "sm", "sn",
		"sp", "st", "str", "sw", "tr",
	}
	vowels := []string{
		"a", "e", "i", "o", "u",
	}
	finalConsonants := []string{
		"b", "d", "f", "g",
		"h", "l", "m", "n",
		"p", "r", "s", "t",
		"v", "w", "z", "ck",
		"ct", "ft", "mp", "nd",
		"ng", "nk", "nt", "pt",
		"sk", "sp", "ss", "st",
	}
	return fmt.Sprintf(
		"%s%s%s",
		initialConsonants[rand.Intn(len(initialConsonants))],
		vowels[rand.Intn(len(vowels))],
		finalConsonants[rand.Intn(len(finalConsonants))],
	)
}
*/
