package neuralspell

import (
	"github.com/unixpickle/weakai/neuralnet"
	"github.com/unixpickle/weakai/rnn"
)

var Phones = []string{
	"ɹ", "θ", "I", "s", "ʃ", "j", "v", "æ", "ʊ", "ŋ", "u", "o", "h", "l", "a", "g", "ɛ",
	"d", "z", "t", "p", "n", "m", "e", "b", "i", "f", "ð", "ʌ", "ɔ", "ʒ", "w", "ə", "k",
}

const LetterCount = 26

const (
	hiddenDropout  = 0.5
	outHidden      = 50
	forwardHidden  = 30
	backwardHidden = 10
)

func NewSpeller() rnn.SeqFunc {
	return newNetwork(len(Phones), LetterCount)
}

func NewPronouncer() rnn.SeqFunc {
	return newNetwork(LetterCount, len(Phones))
}

func newNetwork(inSymbols, outSymbols int) rnn.SeqFunc {
	outNet := neuralnet.Network{
		&neuralnet.DropoutLayer{
			KeepProbability: hiddenDropout,
			Training:        false,
		},
		&neuralnet.DenseLayer{
			InputCount:  forwardHidden + backwardHidden,
			OutputCount: outHidden,
		},
		&neuralnet.HyperbolicTangent{},
		&neuralnet.DenseLayer{
			InputCount:  outHidden,
			OutputCount: outSymbols + 1,
		},
		&neuralnet.LogSoftmaxLayer{},
	}
	outNet.Randomize()

	forwardBlock := rnn.NewLSTM(inSymbols+1, forwardHidden)
	backwardBlock := rnn.NewLSTM(inSymbols+1, backwardHidden)

	return &rnn.Bidirectional{
		Forward:  &rnn.BlockSeqFunc{Block: forwardBlock},
		Backward: &rnn.BlockSeqFunc{Block: backwardBlock},
		Output:   &rnn.NetworkSeqFunc{Network: outNet},
	}
}
