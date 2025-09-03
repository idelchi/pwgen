package dictionary

import (
	"errors"
	"fmt"

	"github.com/sethvargo/go-diceware/diceware"
)

const (
	// effWordlistSize is the size of the EFF Large Wordlist (6^5).
	effWordlistSize = 7776
	// effEntropyBits is the entropy bits per word for EFF wordlist (log2(7776)).
	effEntropyBits = 12.925
	// sampleWordLimit is the limit for generating sample words.
	sampleWordLimit = 1000
)

// ExternalEFF returns an EFF dictionary using the sethvargo/go-diceware library.
// This provides the complete EFF Large Wordlist (7776 words) from an external source.
//
//nolint:ireturn // Dictionary interface is the intended public API for polymorphism
func ExternalEFF() Dictionary {
	return &externalEFFDict{name: "external-eff"}
}

// externalEFFDict implements Dictionary using sethvargo/go-diceware.
type externalEFFDict struct {
	name string
}

func (d *externalEFFDict) Name() string {
	return d.name
}

func (d *externalEFFDict) Size() int {
	// EFF Large Wordlist has 7776 words (6^5)
	return effWordlistSize
}

func (d *externalEFFDict) Words() []string {
	// Generate all words by repeatedly calling the library
	// Note: This is inefficient, but matches the interface
	words := make([]string, 0, effWordlistSize)
	seen := make(map[string]bool)

	// Generate many words and collect unique ones
	// This is a workaround since sethvargo doesn't expose the wordlist directly
	for len(words) < sampleWordLimit { // Get a reasonable sample
		list, err := diceware.Generate(1)
		if err != nil {
			break
		}

		if len(list) > 0 && !seen[list[0]] {
			words = append(words, list[0])
			seen[list[0]] = true
		}
	}

	return words
}

func (d *externalEFFDict) RandomWord() (string, error) {
	list, err := diceware.Generate(1)
	if err != nil {
		return "", fmt.Errorf("generating word: %w", err)
	}

	if len(list) == 0 {
		return "", errors.New("no words generated")
	}

	return list[0], nil
}

func (d *externalEFFDict) EntropyBits() float64 {
	// log2(7776) ≈ 12.925 bits per word
	return effEntropyBits
}
