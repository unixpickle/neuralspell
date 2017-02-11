package neuralspell

import (
	"errors"
	"fmt"

	"github.com/unixpickle/anyvec"
)

const (
	phoneSeqSpacing  = 4
	letterSeqSpacing = 4
)

var Phones = []string{
	"ɹ", "θ", "I", "s", "ʃ", "j", "v", "æ", "ʊ", "ŋ", "u", "o", "h", "l", "a", "g", "ɛ",
	"d", "z", "t", "p", "n", "m", "e", "b", "i", "f", "ð", "ʌ", "ɔ", "ʒ", "w", "ə", "k",
}

const LetterCount = 26

func spacedInputs(c anyvec.Creator, label []int, alphabetSize, spacing int) []anyvec.Vector {
	var res []anyvec.Vector
	for _, x := range label {
		oneHot := make([]float64, alphabetSize)
		oneHot[x] = 1
		for i := 0; i < spacing; i++ {
			v := c.MakeVectorData(c.MakeNumericList(oneHot))
			res = append(res, v)
		}
	}
	return res
}

// phoneLabels produces CTC labels for an IPA string.
func phoneLabels(phonetics string) ([]int, error) {
	var res []int
	for _, x := range phonetics {
		idx, err := phoneIndex(x)
		if err != nil {
			return nil, err
		}
		res = append(res, idx)
	}
	return res, nil
}

// spellingLabels produces CTC labels for a word.
func spellingLabels(spelling string) ([]int, error) {
	var letterSeq []int
	for _, letter := range spelling {
		index := int(letter) - int('a')
		if index < 0 || index >= LetterCount {
			return nil, errors.New("unknown letter: " + string(letter))
		}
		letterSeq = append(letterSeq, index)
	}
	return letterSeq, nil
}

func phoneIndex(ph rune) (int, error) {
	for i, x := range Phones {
		if x == string(ph) {
			return i, nil
		}
	}
	return 0, fmt.Errorf("unknown phone: %s", string(ph))
}
