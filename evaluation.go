package neuralspell

import (
	"fmt"
	"strings"

	"github.com/unixpickle/autofunc"
	"github.com/unixpickle/num-analysis/linalg"
	"github.com/unixpickle/speechrecog/ctc"
	"github.com/unixpickle/weakai/rnn"
)

const phoneSequenceSpacing = 3
const letterSequenceSpacing = 3

// Spell generates a spelling for a phonetic string using
// a trained speller network.
// It returns an error if the phonetic sequence contains
// an unrecognized symbol.
func Spell(net rnn.SeqFunc, phonetics string) (string, error) {
	phoneRunes := []rune(phonetics)
	inSeq := make([]autofunc.Result, 0, len(phoneRunes)*(phoneSequenceSpacing+1))
	for _, r := range phoneRunes {
		phone, err := phoneIndex(r)
		if err != nil {
			return "", err
		}
		inVec := make(linalg.Vector, len(Phones)+1)
		inVec[phone] = 1
		repeatVec := make(linalg.Vector, len(Phones)+1)
		repeatVec[phone] = 1
		repeatVec[len(Phones)] = 1
		inSeq = append(inSeq, &autofunc.Variable{Vector: inVec})
		for j := 0; j < phoneSequenceSpacing; j++ {
			inSeq = append(inSeq, &autofunc.Variable{Vector: repeatVec})
		}
	}
	resSeq := net.BatchSeqs([][]autofunc.Result{inSeq}).OutputSeqs()[0]
	symbolSet := ctc.PrefixSearch(resSeq, 1e-4)

	var res string
	for _, x := range symbolSet {
		res += string('a' + rune(x))
	}
	return res, nil
}

// Pronounce generates a spelling for a phonetic string
// using a trained pronouncer network.
// It returns an error if the English string contains
// characters which are not letters.
func Pronounce(net rnn.SeqFunc, english string) (string, error) {
	englishRunes := []rune(strings.ToLower(english))
	inSeq := make([]autofunc.Result, 0, len(englishRunes)*(letterSequenceSpacing+1))
	for _, r := range englishRunes {
		letter := int(r) - int('a')
		if letter < 0 || letter >= LetterCount {
			return "", fmt.Errorf("unknown letter: %s", string(r))
		}
		inVec := make(linalg.Vector, LetterCount+1)
		inVec[letter] = 1
		repeatVec := make(linalg.Vector, LetterCount+1)
		repeatVec[letter] = 1
		repeatVec[LetterCount] = 1
		inSeq = append(inSeq, &autofunc.Variable{Vector: inVec})
		for j := 0; j < phoneSequenceSpacing; j++ {
			inSeq = append(inSeq, &autofunc.Variable{Vector: repeatVec})
		}
	}
	resSeq := net.BatchSeqs([][]autofunc.Result{inSeq}).OutputSeqs()[0]
	symbolSet := ctc.PrefixSearch(resSeq, 1e-4)

	var res string
	for _, x := range symbolSet {
		res += string(Phones[x])
	}
	return res, nil
}

func phoneIndex(ph rune) (int, error) {
	for i, x := range Phones {
		if x == string(ph) {
			return i, nil
		}
	}
	return 0, fmt.Errorf("unknown phonetic symbol: %s", string(ph))
}
