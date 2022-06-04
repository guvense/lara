package generator

import (
	"math/rand"
	"time"

	"github.com/google/uuid"

	"github.com/lucasjones/reggen"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandomString(n int) string {
	rand.Seed(time.Now().UnixNano())

	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func GenerateUuid() string {
	return uuid.New().String()
}

func GenerateNumber(from, to int64) int64{

	rand.Seed(time.Now().UnixNano())
	rng := to - from
	return rand.Int63n(rng) + from
}

func GenerateStringFromRegex(regex string, length int) string {
	str, err := reggen.Generate(regex, 10)
	if err != nil {
		return ""
	}
	return str
}


