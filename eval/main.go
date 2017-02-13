package main

import (
	"flag"
	"fmt"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/neuralspell"
	"github.com/unixpickle/serializer"
)

func main() {
	var netPath string
	var phonetics string
	var spelling string

	flag.StringVar(&netPath, "net", "../train/out_net", "network file path")
	flag.StringVar(&phonetics, "phonetics", "", "input phonetics")
	flag.StringVar(&spelling, "spelling", "", "input spelling")

	flag.Parse()

	if phonetics == "" && spelling == "" {
		essentials.Die("Must specify -phonetics or -spelling. See -help.")
	}

	var net *neuralspell.Network
	if err := serializer.LoadAny(netPath, &net); err != nil {
		essentials.Die(err)
	}

	if phonetics != "" {
		sp, err := net.Spell(phonetics)
		if err != nil {
			essentials.Die(err)
		}
		fmt.Println("Spelling:", sp)
	}
	if spelling != "" {
		pron, err := net.Pronounce(spelling)
		if err != nil {
			essentials.Die(err)
		}
		fmt.Println("Pronunciation:", pron)
	}
	if spelling != "" && phonetics != "" {
		spellCost, pronCost, err := net.Costs(spelling, phonetics)
		if err != nil {
			essentials.Die(err)
		}
		fmt.Println("Spell cost:", spellCost)
		fmt.Println("Pronounce cost:", pronCost)
	}
}
