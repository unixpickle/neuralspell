package neuralspell

import (
	"github.com/unixpickle/anydiff/anyseq"
	"github.com/unixpickle/anynet"
	"github.com/unixpickle/anynet/anyctc"
	"github.com/unixpickle/anynet/anyrnn"
	"github.com/unixpickle/anyvec"
	"github.com/unixpickle/anyvec/anyvec32"
	"github.com/unixpickle/essentials"
	"github.com/unixpickle/serializer"
)

func init() {
	var n Network
	serializer.RegisterTypedDeserializer(n.SerializerType(), DeserializeNetwork)
}

// A Network can spell and pronounce words.
type Network struct {
	Speller    *anyrnn.Bidir
	Pronouncer *anyrnn.Bidir
}

// DeserializeNetwork deserializes a Network.
func DeserializeNetwork(d []byte) (*Network, error) {
	var res Network
	if err := serializer.DeserializeAny(d, &res.Speller, &res.Pronouncer); err != nil {
		return nil, essentials.AddCtx("deserialize Network", err)
	}
	return &res, nil
}

// NewNetwork creates a new, untrained network.
func NewNetwork() *Network {
	c := anyvec32.CurrentCreator()
	return &Network{
		Speller:    newBidir(c, len(Phones), LetterCount),
		Pronouncer: newBidir(c, LetterCount, len(Phones)),
	}
}

// Spell produces a spelling for the pronunciation.
func (n *Network) Spell(phonetics string) (string, error) {
	labels, err := phoneLabels(phonetics)
	if err != nil {
		return "", essentials.AddCtx("spell", err)
	}
	c := n.Speller.Parameters()[0].Vector.Creator()
	in := spacedInputs(c, labels, len(Phones), phoneSeqSpacing)
	out := n.Speller.Apply(anyseq.ConstSeqList(c, [][]anyvec.Vector{in}))
	outLabels := anyctc.BestLabels(out, -1e-3)[0]

	var res string
	for _, x := range outLabels {
		res += string(rune(x) + 'a')
	}
	return res, nil
}

// Pronounce produces phonetics for a spelling.
func (n *Network) Pronounce(spelling string) (string, error) {
	labels, err := spellingLabels(spelling)
	if err != nil {
		return "", essentials.AddCtx("pronounce", err)
	}
	c := n.Pronouncer.Parameters()[0].Vector.Creator()
	in := spacedInputs(c, labels, LetterCount, letterSeqSpacing)
	out := n.Pronouncer.Apply(anyseq.ConstSeqList(c, [][]anyvec.Vector{in}))
	outLabels := anyctc.BestLabels(out, -1e-3)[0]

	var res string
	for _, x := range outLabels {
		res += Phones[x]
	}
	return res, nil
}

// Costs returns the CTC costs both for trying to spell
// the word and for attempting to pronounce it.
func (n *Network) Costs(spelling, phonetics string) (spellCost, pronCost float64, err error) {
	defer essentials.AddCtxTo("compute costs", &err)
	spellLabels, err := spellingLabels(spelling)
	if err != nil {
		return
	}
	pronLabels, err := phoneLabels(phonetics)
	if err != nil {
		return
	}

	c := n.Speller.Parameters()[0].Vector.Creator()

	spellIn := anyseq.ConstSeqList(c, [][]anyvec.Vector{
		spacedInputs(c, spellLabels, LetterCount, letterSeqSpacing),
	})
	cost1 := anyctc.Cost(n.Pronouncer.Apply(spellIn), [][]int{pronLabels})
	pronIn := anyseq.ConstSeqList(c, [][]anyvec.Vector{
		spacedInputs(c, pronLabels, len(Phones), phoneSeqSpacing),
	})
	cost2 := anyctc.Cost(n.Speller.Apply(pronIn), [][]int{spellLabels})

	spellCost = float64(cost2.Output().Data().([]float32)[0])
	pronCost = float64(cost1.Output().Data().([]float32)[0])

	return
}

// SerializerType returns the unique ID used to serialize
// a Network.
func (n *Network) SerializerType() string {
	return "github.com/unixpickle/neuralspell.Network"
}

// Serialize serializes a Network.
func (n *Network) Serialize() ([]byte, error) {
	return serializer.SerializeAny(n.Speller, n.Pronouncer)
}

func newBidir(c anyvec.Creator, inCount, labelCount int) *anyrnn.Bidir {
	inScale := c.MakeNumeric(0x10)
	return &anyrnn.Bidir{
		Forward: anyrnn.Stack{
			anyrnn.NewLSTM(c, inCount, 0x80).ScaleInWeights(inScale),
			anyrnn.NewLSTM(c, 0x80, 0x80),
		},
		Backward: anyrnn.Stack{
			anyrnn.NewLSTM(c, inCount, 0x80).ScaleInWeights(inScale),
			anyrnn.NewLSTM(c, 0x80, 0x80),
		},
		Mixer: &anynet.AddMixer{
			In1: anynet.NewFC(c, 0x80, 0x80),
			In2: anynet.NewFC(c, 0x80, 0x80),
			Out: anynet.Net{
				anynet.NewFC(c, 0x80, labelCount+1),
				anynet.LogSoftmax,
			},
		},
	}
}
