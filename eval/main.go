package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/unixpickle/neuralspell"
	"github.com/unixpickle/serializer"
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

	if spelling {
		out, err := neuralspell.Spell(seqFunc, os.Args[3])
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to spell:", err)
			os.Exit(1)
		}
		fmt.Println("Spelling:", out)
	} else {
		out, err := neuralspell.Pronounce(seqFunc, os.Args[3])
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to pronounce:", err)
			os.Exit(1)
		}
		fmt.Println("Pronunciation:", out)
	}
}

func dieUsage() {
	fmt.Fprintln(os.Stderr, "Usage:", os.Args[0], "<spell | pronounce> rnn_file input_text")
	os.Exit(1)
}
