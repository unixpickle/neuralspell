package main

import (
	"flag"
	"fmt"

	"github.com/unixpickle/anynet/anysgd"
	"github.com/unixpickle/essentials"
	"github.com/unixpickle/neuralspell"
	"github.com/unixpickle/serializer"
)

func main() {
	var netPath string
	var dataPath string
	var task string

	flag.StringVar(&netPath, "net", "../train/out_net", "network file path")
	flag.StringVar(&dataPath, "data", "../dict/cmudict-IPA.txt", "dictionary path")
	flag.StringVar(&task, "task", "spell", "task ('spell' or 'pronounce')")

	flag.Parse()

	if task != "spell" && task != "pronounce" {
		essentials.Die("unknown task:", task)
	}

	var net *neuralspell.Network
	if err := serializer.LoadAny(netPath, &net); err != nil {
		essentials.Die(err)
	}

	dict, err := neuralspell.ReadDictionary(dataPath)
	if err != nil {
		essentials.Die(err)
	}

	anysgd.Shuffle(dict)

	inputs := dict.Pronunciations
	outputs := dict.Spellings
	method := net.Spell

	if task != "spell" {
		inputs, outputs = outputs, inputs
		method = net.Pronounce
	}

	var correct, total int
	for i, in := range inputs {
		desired := outputs[i]
		actual, err := method(in)
		if err != nil {
			fmt.Println()
			essentials.Die(err)
		}
		if desired == actual {
			correct++
		}
		total++
		fmt.Printf("\rGot %d/%d (%.2f%%)", correct, total,
			100*float64(correct)/float64(total))
	}
	fmt.Println()
}
