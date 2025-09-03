// Package dictionary provides word dictionaries for passphrase generation.
package dictionary

import (
	"bufio"
	"crypto/rand"
	"errors"
	"fmt"
	"math"
	"math/big"
	"os"
	"strings"
)

// Dictionary represents a word dictionary for passphrase generation.
type Dictionary interface {
	// Name returns the name of this dictionary.
	Name() string

	// Size returns the number of words in the dictionary.
	Size() int

	// Words returns all words in the dictionary.
	Words() []string

	// RandomWord returns a cryptographically random word from the dictionary.
	RandomWord() (string, error)

	// EntropyBits returns the entropy bits per word for this dictionary.
	EntropyBits() float64
}

// wordDict implements Dictionary with a slice of words.
type wordDict struct {
	name  string
	words []string
}

// NewFromWords creates a new dictionary from a slice of words.
//
//nolint:ireturn // Dictionary interface is the intended public API for polymorphism
func NewFromWords(name string, words []string) Dictionary {
	// Create a copy to avoid external modification.
	wordsCopy := make([]string, len(words))
	copy(wordsCopy, words)

	return &wordDict{
		name:  name,
		words: wordsCopy,
	}
}

// NewFromFile creates a new dictionary from a file path.
// The file should contain one word per line.
//
//nolint:ireturn // Dictionary interface is the intended public API for polymorphism
func NewFromFile(path string) (Dictionary, error) {
	file, err := os.Open(path) //nolint:gosec // User-provided file path is intentional
	if err != nil {
		return nil, fmt.Errorf("opening dictionary file: %w", err)
	}
	defer file.Close()

	var words []string

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			words = append(words, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading dictionary file: %w", err)
	}

	if len(words) == 0 {
		return nil, errors.New("dictionary file contains no words")
	}

	return NewFromWords(path, words), nil
}

// Name returns the name of this dictionary.
func (d *wordDict) Name() string {
	return d.name
}

// Size returns the number of words in the dictionary.
func (d *wordDict) Size() int {
	return len(d.words)
}

// Words returns all words in the dictionary.
func (d *wordDict) Words() []string {
	// Return a copy to prevent external modification.
	result := make([]string, len(d.words))
	copy(result, d.words)

	return result
}

// RandomWord returns a cryptographically random word from the dictionary.
func (d *wordDict) RandomWord() (string, error) {
	if len(d.words) == 0 {
		return "", errors.New("dictionary is empty")
	}

	// Use uniform distribution to avoid modulo bias.
	maxVal := big.NewInt(int64(len(d.words)))

	n, err := rand.Int(rand.Reader, maxVal)
	if err != nil {
		return "", fmt.Errorf("generating random number: %w", err)
	}

	return d.words[n.Int64()], nil
}

// EntropyBits returns the entropy bits per word for this dictionary.
func (d *wordDict) EntropyBits() float64 {
	if len(d.words) == 0 {
		return 0
	}
	// log2(dictionary_size)
	return math.Log2(float64(len(d.words)))
}
