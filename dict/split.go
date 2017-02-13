package main

import (
	"io/ioutil"
	"sort"
	"strings"

	"github.com/unixpickle/anynet/anysgd"
	"github.com/unixpickle/essentials"
	"github.com/unixpickle/neuralspell"
)

func main() {
	samples, err := neuralspell.ReadDictionary("cmudict-IPA.txt")
	if err != nil {
		essentials.Die(err)
	}

	validationSet, trainingSet := anysgd.HashSplit(samples, 0.1)
	validation := dictionaryLines(validationSet)
	training := dictionaryLines(trainingSet)

	writeFile("validation.txt", validation)
	writeFile("training.txt", training)
}

func dictionaryLines(s anysgd.SampleList) []string {
	dict := s.(*neuralspell.Dictionary)
	var res []string
	for i, spell := range dict.Spellings {
		pron := dict.Pronunciations[i]
		res = append(res, spell+","+pron)
	}
	return res
}

func writeFile(outFile string, lines []string) {
	sort.Strings(lines)
	joined := strings.Join(lines, "\n") + "\n"
	if err := ioutil.WriteFile(outFile, []byte(joined), 0644); err != nil {
		essentials.Die(err)
	}
}
