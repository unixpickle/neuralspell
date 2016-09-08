package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/unixpickle/neuralspell"
	"github.com/unixpickle/serializer"
	"github.com/unixpickle/sgd"
	"github.com/unixpickle/weakai/rnn"
)

func main() {
	if len(os.Args) != 4 {
		dieUsage()
	}
	if os.Args[1] != "spell" && os.Args[1] != "pronounce" {
		dieUsage()
	}
	spelling := os.Args[1] == "spell"

	dict, err := neuralspell.ReadDictionary(os.Args[3])
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to read dictionary:", err)
		os.Exit(1)
	}
	sgd.ShuffleSampleSet(dict)

	rnnData, err := ioutil.ReadFile(os.Args[2])
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to read RNN:", err)
		os.Exit(1)
	}
	rnnObj, err := serializer.DeserializeWithType(rnnData)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to deserialize RNN:", err)
		os.Exit(1)
	}
	seqFunc, ok := rnnObj.(rnn.SeqFunc)
	if !ok {
		fmt.Fprintf(os.Stderr, "Invalid deserialized type: %T\n", rnnObj)
		os.Exit(1)
	}

	resChan := make(chan bool, 10)
	if spelling {
		go rateSpellings(dict, seqFunc, resChan)
	} else {
		go ratePronunciations(dict, seqFunc, resChan)
	}

	fmt.Println("Testing on a total of", len(dict.Spellings), "entries...")

	var correct, total int
	for rating := range resChan {
		total++
		if rating {
			correct++
		}
		fmt.Printf("\rGot %d/%d (%.2f%%)    ", correct, total,
			100*float64(correct)/float64(total))
	}
	fmt.Println()
}

func rateSpellings(dict *neuralspell.Dictionary, seqFunc rnn.SeqFunc, res chan<- bool) {
	for i, phonetics := range dict.Pronunciations {
		actual, err := neuralspell.Spell(seqFunc, phonetics)
		if err != nil {
			fmt.Fprintln(os.Stderr, "\nFailed to spell:", err)
			os.Exit(1)
		}
		expected := dict.Spellings[i]
		res <- expected == actual
	}
	close(res)
}

func ratePronunciations(dict *neuralspell.Dictionary, seqFunc rnn.SeqFunc, res chan<- bool) {
	for i, letters := range dict.Spellings {
		actual, err := neuralspell.Pronounce(seqFunc, letters)
		if err != nil {
			fmt.Fprintln(os.Stderr, "\nFailed to pronounce:", err)
			os.Exit(1)
		}
		expected := dict.Pronunciations[i]
		res <- expected == actual
	}
	close(res)
}

func dieUsage() {
	fmt.Fprintln(os.Stderr, "Usage:", os.Args[0], "<spell | pronounce> rnn_file dict_file")
	os.Exit(1)
}
