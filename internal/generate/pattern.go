package generate

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/idelchi/pwgen/internal/dictionary"
)

// Pattern represents a passphrase generation pattern.
type Pattern struct {
	Tokens []Token
}

// Generate creates a passphrase from this pattern.
func (p *Pattern) Generate() (string, error) {
	if len(p.Tokens) == 0 {
		return "", errors.New("pattern is empty")
	}

	parts := make([]string, 0, len(p.Tokens))

	for _, token := range p.Tokens {
		part, err := token.Generate()
		if err != nil {
			return "", fmt.Errorf("generating token %s: %w", token.Type(), err)
		}

		parts = append(parts, part)
	}

	return strings.Join(parts, ""), nil
}

// EntropyBits calculates the total entropy of this pattern.
func (p *Pattern) EntropyBits() float64 {
	total := 0.0

	for _, token := range p.Tokens {
		total += token.EntropyBits()
	}

	return total
}

// String returns a human-readable description of the pattern.
func (p *Pattern) String() string {
	parts := make([]string, 0, len(p.Tokens))

	for _, token := range p.Tokens {
		parts = append(parts, token.Type())
	}

	return strings.Join(parts, " ")
}

// PatternBuilder helps construct patterns from various inputs.
type PatternBuilder struct {
	defaultDict dictionary.Dictionary
	defaultSep  string
}

// NewPatternBuilder creates a new pattern builder.
func NewPatternBuilder(defaultDict dictionary.Dictionary, defaultSep string) *PatternBuilder {
	return &PatternBuilder{
		defaultDict: defaultDict,
		defaultSep:  defaultSep,
	}
}

// BuildFromOptions creates a pattern from CLI options.
func (pb *PatternBuilder) BuildFromOptions(
	words, digits, symbols int,
	casing, sep string,
	kebab, snake, camel bool,
) (*Pattern, error) {
	if words == 0 && digits == 0 && symbols == 0 {
		return nil, errors.New("must specify at least one type of token")
	}

	caseStyle, err := ParseCaseStyle(casing)
	if err != nil {
		return nil, err
	}

	// Determine separator
	actualSep := pb.defaultSep

	if sep != "" {
		actualSep = sep
	}

	switch {
	case kebab:
		actualSep = "-"
	case snake:
		actualSep = "_"
	case camel:
		actualSep = ""
	}

	// Estimate capacity: words + separators + digits + symbols
	capacity := words + max(words-1, 0) + digits + symbols
	tokens := make([]Token, 0, capacity)

	// Add words with separators
	for i := range words {
		if i > 0 && actualSep != "" {
			tokens = append(tokens, &SeparatorToken{Value: actualSep})
		}

		tokens = append(tokens, &WordToken{Dict: pb.defaultDict, Casing: caseStyle})
	}

	// Add digits with separators
	for range digits {
		if len(tokens) > 0 && actualSep != "" {
			tokens = append(tokens, &SeparatorToken{Value: actualSep})
		}

		tokens = append(tokens, &DigitToken{Count: 1})
	}

	// Add symbols with separators
	for range symbols {
		if len(tokens) > 0 && actualSep != "" {
			tokens = append(tokens, &SeparatorToken{Value: actualSep})
		}

		tokens = append(tokens, &SymbolToken{Count: 1, Charset: DefaultSymbolCharset})
	}

	return &Pattern{Tokens: tokens}, nil
}

// BuildFromDSL creates a pattern from a DSL string.
// Example DSL: "W:title SEP W:lower SEP DD{2} SEP S".
func (pb *PatternBuilder) BuildFromDSL(dsl string) (*Pattern, error) {
	if strings.TrimSpace(dsl) == "" {
		return nil, errors.New("pattern DSL cannot be empty")
	}

	tokens, err := pb.parseDSL(dsl)
	if err != nil {
		return nil, fmt.Errorf("parsing DSL pattern: %w", err)
	}

	return &Pattern{Tokens: tokens}, nil
}

// parseDSL parses a DSL string into tokens.
func (pb *PatternBuilder) parseDSL(dsl string) ([]Token, error) {
	// Split on whitespace and process each element
	elements := strings.Fields(dsl)

	tokens := make([]Token, 0, len(elements))

	for _, element := range elements {
		token, err := pb.parseElement(element)
		if err != nil {
			return nil, fmt.Errorf("parsing element %q: %w", element, err)
		}

		tokens = append(tokens, token)
	}

	return tokens, nil
}

// parseElement parses a single DSL element.
//
//nolint:ireturn // Token interface is required for polymorphism in pattern parsing
func (pb *PatternBuilder) parseElement(element string) (Token, error) {
	element = strings.TrimSpace(element)

	// Handle SEP (separator)
	if element == "SEP" {
		return &SeparatorToken{Value: pb.defaultSep}, nil
	}

	// Handle W (word) with optional casing
	if strings.HasPrefix(element, "W") {
		return pb.parseWordToken(element)
	}

	// Handle D (digit) with optional count
	if strings.HasPrefix(element, "D") {
		return pb.parseDigitToken(element)
	}

	// Handle S (symbol) with optional count
	if strings.HasPrefix(element, "S") {
		return pb.parseSymbolToken(element)
	}

	return nil, fmt.Errorf("unknown pattern element: %q", element)
}

// parseWordToken parses word tokens like "W", "W:title", "W:mixed".
//
//nolint:ireturn // Token interface is required for polymorphism in pattern parsing
func (pb *PatternBuilder) parseWordToken(element string) (Token, error) {
	parts := strings.Split(element, ":")

	if len(parts) == 1 {
		// Just "W"
		return &WordToken{Dict: pb.defaultDict, Casing: CaseLower}, nil
	}

	if len(parts) != minWordsRequired {
		return nil, fmt.Errorf("invalid word token format: %q", element)
	}

	caseStyle, err := ParseCaseStyle(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid casing in word token %q: %w", element, err)
	}

	return &WordToken{Dict: pb.defaultDict, Casing: caseStyle}, nil
}

// parseDigitToken parses digit tokens like "D", "DD{2}", "D{3}".
//
//nolint:ireturn // Token interface is required for polymorphism in pattern parsing
func (pb *PatternBuilder) parseDigitToken(element string) (Token, error) {
	// Match patterns like "D", "DD", "DDD", "D{n}", "DD{n}"
	re := regexp.MustCompile(`^D+(?:\{(\d+)\})?$`)
	matches := re.FindStringSubmatch(element)

	if matches == nil {
		return nil, fmt.Errorf("invalid digit token format: %q", element)
	}

	count := strings.Count(element, "D")

	// If there's a {n} modifier, use that count
	if matches[1] != "" {
		var err error

		count, err = strconv.Atoi(matches[1])
		if err != nil {
			return nil, fmt.Errorf("invalid digit count in %q: %w", element, err)
		}
	}

	if count <= 0 {
		return nil, fmt.Errorf("digit count must be positive in %q", element)
	}

	return &DigitToken{Count: count}, nil
}

// parseSymbolToken parses symbol tokens like "S", "S{2}".
//
//nolint:ireturn // Token interface is required for polymorphism in pattern parsing
func (pb *PatternBuilder) parseSymbolToken(element string) (Token, error) {
	if element == "S" {
		return &SymbolToken{Count: 1, Charset: DefaultSymbolCharset}, nil
	}

	// Match pattern like "S{n}"
	re := regexp.MustCompile(`^S\{(\d+)\}$`)
	matches := re.FindStringSubmatch(element)

	if matches == nil {
		return nil, fmt.Errorf("invalid symbol token format: %q", element)
	}

	count, err := strconv.Atoi(matches[1])
	if err != nil {
		return nil, fmt.Errorf("invalid symbol count in %q: %w", element, err)
	}

	if count <= 0 {
		return nil, fmt.Errorf("symbol count must be positive in %q", element)
	}

	return &SymbolToken{Count: count, Charset: DefaultSymbolCharset}, nil
}
