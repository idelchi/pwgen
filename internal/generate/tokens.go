package generate

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/idelchi/pwgen/internal/dictionary"
)

const (
	// digitBase is the base for decimal digits.
	digitBase = 10
	// binaryChoiceRange is the range for binary random choices.
	binaryChoiceRange = 2
)

// Token represents a single element in a passphrase.
type Token interface {
	// Generate produces a random value for this token.
	Generate() (string, error)

	// EntropyBits returns the entropy contributed by this token.
	EntropyBits() float64

	// Type returns a description of this token type.
	Type() string
}

// WordToken generates random words from a dictionary.
type WordToken struct {
	Dict   dictionary.Dictionary
	Casing CaseStyle
}

// Generate produces a random word with the specified casing.
func (w *WordToken) Generate() (string, error) {
	word, err := w.Dict.RandomWord()
	if err != nil {
		return "", err
	}

	return ApplyCasing(word, w.Casing)
}

// EntropyBits returns the entropy contributed by this word token.
func (w *WordToken) EntropyBits() float64 {
	return w.Dict.EntropyBits() + w.Casing.EntropyBits()
}

// Type returns a description of this token type.
func (w *WordToken) Type() string {
	return fmt.Sprintf("word(%s)", w.Casing)
}

// DigitToken generates random digits.
type DigitToken struct {
	Count int // Number of digits to generate
}

// Generate produces random digits.
func (d *DigitToken) Generate() (string, error) {
	if d.Count <= 0 {
		return "", errors.New("digit count must be positive")
	}

	var result strings.Builder
	result.Grow(d.Count)

	for range d.Count {
		digit, err := rand.Int(rand.Reader, big.NewInt(digitBase))
		if err != nil {
			return "", fmt.Errorf("generating random digit: %w", err)
		}

		result.WriteString(strconv.FormatInt(digit.Int64(), digitBase))
	}

	return result.String(), nil
}

// EntropyBits returns the entropy contributed by this digit token.
func (d *DigitToken) EntropyBits() float64 {
	if d.Count <= 0 {
		return 0
	}
	// log2(10^count) = count * log2(10)
	return float64(d.Count) * math.Log2(digitBase)
}

// Type returns a description of this token type.
func (d *DigitToken) Type() string {
	if d.Count == 1 {
		return "digit"
	}

	return fmt.Sprintf("digits(%d)", d.Count)
}

// SymbolToken generates random symbols from a character set.
type SymbolToken struct {
	Count   int    // Number of symbols to generate
	Charset string // Available symbols
}

// DefaultSymbolCharset contains the default symbol set.
const DefaultSymbolCharset = "!@#$%^&*()_+-=[]{}|;:,.<>?"

// Generate produces random symbols.
func (s *SymbolToken) Generate() (string, error) {
	if s.Count <= 0 {
		return "", errors.New("symbol count must be positive")
	}

	charset := s.Charset
	if charset == "" {
		charset = DefaultSymbolCharset
	}

	if len(charset) == 0 {
		return "", errors.New("symbol charset cannot be empty")
	}

	var result strings.Builder
	result.Grow(s.Count)

	maxIdx := big.NewInt(int64(len(charset)))

	for range s.Count {
		idx, err := rand.Int(rand.Reader, maxIdx)
		if err != nil {
			return "", fmt.Errorf("generating random symbol: %w", err)
		}

		result.WriteByte(charset[idx.Int64()])
	}

	return result.String(), nil
}

// EntropyBits returns the entropy contributed by this symbol token.
func (s *SymbolToken) EntropyBits() float64 {
	if s.Count <= 0 {
		return 0
	}

	charset := s.Charset
	if charset == "" {
		charset = DefaultSymbolCharset
	}

	if len(charset) == 0 {
		return 0
	}

	// log2(charset_size^count) = count * log2(charset_size)
	return float64(s.Count) * math.Log2(float64(len(charset)))
}

// Type returns a description of this token type.
func (s *SymbolToken) Type() string {
	if s.Count == 1 {
		return "symbol"
	}

	return fmt.Sprintf("symbols(%d)", s.Count)
}

// SeparatorToken represents a fixed separator.
type SeparatorToken struct {
	Value string
}

// Generate returns the separator value (no randomness).
func (s *SeparatorToken) Generate() (string, error) {
	return s.Value, nil
}

// EntropyBits returns zero (separators add no entropy).
func (s *SeparatorToken) EntropyBits() float64 {
	return 0
}

// Type returns a description of this token type.
func (s *SeparatorToken) Type() string {
	return fmt.Sprintf("sep(%q)", s.Value)
}

// CaseStyle represents different casing styles.
type CaseStyle int

const (
	// CaseLower represents lowercase-only casing style.
	CaseLower CaseStyle = iota
	// CaseUpper represents uppercase-only casing style.
	CaseUpper
	// CaseTitle represents title case (first letter capitalized).
	CaseTitle
	// CaseMixed represents randomly mixed case style.
	CaseMixed
)

func (c CaseStyle) String() string {
	switch c {
	case CaseLower:
		return "lower"
	case CaseUpper:
		return "upper"
	case CaseTitle:
		return "title"
	case CaseMixed:
		return "mixed"
	default:
		return "unknown"
	}
}

// EntropyBits returns additional entropy bits contributed by this casing style.
func (c CaseStyle) EntropyBits() float64 {
	switch c {
	case CaseLower, CaseUpper, CaseTitle:
		// Fixed casing adds no additional entropy
		return 0
	case CaseMixed:
		// Mixed case adds entropy by randomizing case per character
		// This is a simplified calculation
		return 1.0
	default:
		// Fixed casing adds no additional entropy
		return 0
	}
}

// ApplyCasing applies the specified casing style to a word.
func ApplyCasing(word string, style CaseStyle) (string, error) {
	switch style {
	case CaseLower:
		return strings.ToLower(word), nil
	case CaseUpper:
		return strings.ToUpper(word), nil
	case CaseTitle:
		return cases.Title(language.English).String(strings.ToLower(word)), nil
	case CaseMixed:
		return applyMixedCase(word)
	default:
		return word, nil
	}
}

// applyMixedCase applies random casing to each alphabetic character.
func applyMixedCase(word string) (string, error) {
	runes := []rune(word)
	changes := 0

	// First pass: randomize case
	for i, r := range runes { //nolint:varnamelen // i is standard loop var, r is standard for rune
		if unicode.IsLetter(r) {
			bit, err := rand.Int(rand.Reader, big.NewInt(binaryChoiceRange))
			if err != nil {
				return "", fmt.Errorf("generating random bit for casing: %w", err)
			}

			if bit.Int64() == 1 {
				runes[i] = unicode.ToUpper(r)
				changes++
			} else {
				runes[i] = unicode.ToLower(r)
			}
		}
	}

	// Ensure at least one change from all-lowercase
	if changes == 0 {
		// Find first letter and make it uppercase
		for i, r := range runes {
			if unicode.IsLetter(r) {
				runes[i] = unicode.ToUpper(r)

				break
			}
		}
	}

	return string(runes), nil
}

// ParseCaseStyle converts a string to a CaseStyle.
func ParseCaseStyle(str string) (CaseStyle, error) {
	switch strings.ToLower(strings.TrimSpace(str)) {
	case "lower":
		return CaseLower, nil
	case "upper":
		return CaseUpper, nil
	case "title":
		return CaseTitle, nil
	case "mixed":
		return CaseMixed, nil
	default:
		return CaseLower, fmt.Errorf("unknown case style: %q", str)
	}
}
