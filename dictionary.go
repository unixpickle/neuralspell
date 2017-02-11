package neuralspell

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/unixpickle/anynet/anyctc"
	"github.com/unixpickle/anynet/anysgd"
	"github.com/unixpickle/anyvec/anyvec32"
	"github.com/unixpickle/essentials"
)

// A Dictionary is an anyctc.SampleList that generates
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
		if len(parts[0]) == 0 || len(parts[1]) == 0 {
			return nil, fmt.Errorf("line %d: columns may not be empty", i)
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

// GetSample generates a CTC sample for the entry.
func (d *Dictionary) GetSample(i int) (*anyctc.Sample, error) {
	c := anyvec32.CurrentCreator()
	phones, err := phoneLabels(d.Pronunciations[i])
	if err != nil {
		return nil, essentials.AddCtx("get sample", err)
	}
	spelling, err := spellingLabels(d.Spellings[i])
	if err != nil {
		return nil, essentials.AddCtx("get sample", err)
	}
	if d.InputPhones {
		return &anyctc.Sample{
			Input: spacedInputs(c, phones, len(Phones), phoneSeqSpacing),
			Label: spelling,
		}, nil
	} else {
		return &anyctc.Sample{
			Input: spacedInputs(c, spelling, LetterCount, letterSeqSpacing),
			Label: spelling,
		}, nil
	}
}

// Slice generates a subset of this sample set.
func (d *Dictionary) Slice(start, end int) anysgd.SampleList {
	return &Dictionary{
		Spellings:      append([]string{}, d.Spellings[start:end]...),
		Pronunciations: append([]string{}, d.Pronunciations[start:end]...),
		InputPhones:    d.InputPhones,
	}
}
