package neuralspell

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/unixpickle/num-analysis/linalg"
	"github.com/unixpickle/sgd"
	"github.com/unixpickle/speechrecog/ctc"
)

// A Dictionary is an sgd.SampleSet which can generate
// pronunciation or spelling training samples.
type Dictionary struct {
	Spellings      []string
	Pronunciations []string

	// If InputPhones is true, the samples map phonetic
	// sequences to spellings.
	// Otherwise, the samples map spellings to phonetic
	// sequences.
	InputPhones bool
}

// ReadDictionary reads a dictionary from a file.
func ReadDictionary(file string) (*Dictionary, error) {
	contents, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(contents), "\n")
	res := &Dictionary{}
	for i, l := range lines {
		if l == "" {
			continue
		}
		parts := strings.Split(l, ",")
		if len(parts) != 2 {
			return nil, fmt.Errorf("line %d: expected two columns", i)
		}
		for _, x := range parts[0] {
			if x < 'a' || x > 'z' {
				return nil, fmt.Errorf("line %d: unexpected letter: %c", i, x)
			}
		}
		for _, x := range parts[1] {
			if _, err := phoneIndex(x); err != nil {
				return nil, fmt.Errorf("line %d: unexpected phonetic symbol: %c", i, x)
			}
		}
		res.Spellings = append(res.Spellings, parts[0])
		res.Pronunciations = append(res.Pronunciations, parts[1])
	}
	return res, nil
}

// Len returns the number of entries in the dictionary.
func (d *Dictionary) Len() int {
	return len(d.Spellings)
}

// Swap swaps two indexed entries in the dictionary.
func (d *Dictionary) Swap(i, j int) {
	d.Spellings[i], d.Spellings[j] = d.Spellings[j], d.Spellings[i]
	d.Pronunciations[i], d.Pronunciations[j] = d.Pronunciations[j], d.Pronunciations[i]
}

// GetSample generates a seqtoseq.Sample for the indexed
// entry in the dictionary.
func (d *Dictionary) GetSample(i int) interface{} {
	if d.InputPhones {
		return ctc.Sample{
			Input: d.phoneInput(i),
			Label: d.letterOutput(i),
		}
	} else {
		return ctc.Sample{
			Input: d.letterInput(i),
			Label: d.phoneOutput(i),
		}
	}
}

// Copy generates a copy of the sample set.
func (d *Dictionary) Copy() sgd.SampleSet {
	res := &Dictionary{
		Spellings:      make([]string, len(d.Spellings)),
		Pronunciations: make([]string, len(d.Pronunciations)),
		InputPhones:    d.InputPhones,
	}
	copy(res.Spellings, d.Spellings)
	copy(res.Pronunciations, d.Pronunciations)
	return res
}

// Subset generates a subset of this sample set.
func (d *Dictionary) Subset(start, end int) sgd.SampleSet {
	return &Dictionary{
		Spellings:      d.Spellings[start:end],
		Pronunciations: d.Pronunciations[start:end],
		InputPhones:    d.InputPhones,
	}
}

func (d *Dictionary) phoneInput(i int) []linalg.Vector {
	var phoneSeq []linalg.Vector
	for _, phone := range d.Pronunciations[i] {
		index, err := phoneIndex(phone)
		if err != nil {
			panic(err)
		}
		phoneVec := make(linalg.Vector, len(Phones)+1)
		phoneVec[index] = 1
		phoneSeq = append(phoneSeq, phoneVec)
		repeatVec := make(linalg.Vector, len(phoneVec))
		repeatVec[index] = 1
		repeatVec[len(Phones)] = 1
		for i := 0; i < phoneSequenceSpacing; i++ {
			phoneSeq = append(phoneSeq, repeatVec)
		}
	}
	return phoneSeq
}

func (d *Dictionary) letterInput(i int) []linalg.Vector {
	var spellingSeq []linalg.Vector
	for _, letter := range d.Spellings[i] {
		index := int(letter) - int('a')
		if index < 0 || index >= LetterCount {
			panic("unknown letter: " + string(letter))
		}
		letterVec := make(linalg.Vector, LetterCount+1)
		letterVec[index] = 1
		spellingSeq = append(spellingSeq, letterVec)
		repeatVec := make(linalg.Vector, len(letterVec))
		letterVec[index] = 1
		letterVec[LetterCount] = 1
		for i := 0; i < letterSequenceSpacing; i++ {
			spellingSeq = append(spellingSeq, repeatVec)
		}
	}
	return spellingSeq
}

func (d *Dictionary) phoneOutput(i int) []int {
	var phoneSeq []int
	for _, phone := range d.Pronunciations[i] {
		index, err := phoneIndex(phone)
		if err != nil {
			panic(err)
		}
		phoneSeq = append(phoneSeq, index)
	}
	return phoneSeq
}

func (d *Dictionary) letterOutput(i int) []int {
	var letterSeq []int
	for _, letter := range d.Spellings[i] {
		index := int(letter) - int('a')
		if index < 0 || index >= LetterCount {
			panic("unknown letter: " + string(letter))
		}
		letterSeq = append(letterSeq, index)
	}
	return letterSeq
}
