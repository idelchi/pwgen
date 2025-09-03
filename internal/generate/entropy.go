// Package generate provides passphrase generation and entropy calculation functionality.
package generate

import (
	"math"
	"regexp"
	"strings"
	"unicode"
)

const (
	// Character set sizes.
	digitCharsetSize  = 10
	letterCharsetSize = 26
	symbolCharsetSize = 32
	spaceCharsetSize  = 1

	// Pattern detection constants.
	minSequenceLength           = 3
	minRepetitionLength         = 3
	sequencePenaltyMultiplier   = 2.0
	repetitionPenaltyMultiplier = 3.0
	dictionaryPenaltyMultiplier = 1.5
	yearPatternPenalty          = 8.0

	// Word detection constants.
	minWordLikeRatio = 0.6
	minLetterRatio   = 0.7
	minWordsRequired = 2
	avgWordLength    = 5.5

	// Word length constraints.
	minWordLength = 2
	maxWordLength = 15
)

// EntropyCalculator calculates entropy for existing passphrases.
type EntropyCalculator struct{}

// NewEntropyCalculator creates a new entropy calculator.
func NewEntropyCalculator() *EntropyCalculator {
	return &EntropyCalculator{}
}

// CharsetInfo represents information about a character set.
type CharsetInfo struct {
	Name  string
	Size  int
	Chars string
}

// Common character sets for entropy calculation.
//
//nolint:gochecknoglobals // Package-level constants for character set definitions
var (
	DigitCharset = CharsetInfo{
		Name:  "digits",
		Size:  digitCharsetSize,
		Chars: "0123456789",
	}
	LowerCharset = CharsetInfo{
		Name:  "lowercase",
		Size:  letterCharsetSize,
		Chars: "abcdefghijklmnopqrstuvwxyz",
	}
	UpperCharset = CharsetInfo{
		Name:  "uppercase",
		Size:  letterCharsetSize,
		Chars: "ABCDEFGHIJKLMNOPQRSTUVWXYZ",
	}
	SymbolCharset = CharsetInfo{
		Name:  "symbols",
		Size:  symbolCharsetSize,
		Chars: "!@#$%^&*()_+-=[]{}|;:,.<>?",
	}
	SpaceCharset = CharsetInfo{
		Name:  "space",
		Size:  spaceCharsetSize,
		Chars: " ",
	}
)

// AnalysisResult contains detailed entropy analysis of a passphrase.
type AnalysisResult struct {
	Passphrase     string         `json:"passphrase"`
	Length         int            `json:"length"`
	Entropy        float64        `json:"entropy"`
	CharsetSize    int            `json:"charsetSize"`
	Charsets       []string       `json:"charsets"`
	Strength       string         `json:"strength"`
	CrackTime      string         `json:"crackTime"`
	Patterns       []PatternMatch `json:"patterns,omitempty"`
	WordBased      bool           `json:"wordBased"`
	EstimatedWords int            `json:"estimatedWords,omitempty"`
}

// PatternMatch represents a detected pattern in the passphrase.
type PatternMatch struct {
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Position    int     `json:"position"`
	Length      int     `json:"length"`
	Penalty     float64 `json:"entropyPenalty"`
}

// CalculateEntropy performs detailed entropy analysis of a passphrase.
func (ec *EntropyCalculator) CalculateEntropy(passphrase string) AnalysisResult {
	if passphrase == "" {
		return AnalysisResult{
			Passphrase: "",
			Length:     0,
			Entropy:    0,
			Strength:   "None",
			CrackTime:  "Instant",
		}
	}

	// Determine character sets used
	charsets := ec.detectCharsets(passphrase)
	charsetSize := ec.calculateCharsetSize(charsets)

	// Calculate base entropy
	baseEntropy := float64(len(passphrase)) * math.Log2(float64(charsetSize))

	// Detect patterns and apply penalties
	patterns := ec.detectPatterns(passphrase)
	adjustedEntropy := ec.applyPatternPenalties(baseEntropy, patterns)

	// Check if it looks word-based
	wordBased, estimatedWords := ec.analyzeWordStructure(passphrase)

	result := AnalysisResult{
		Passphrase:     passphrase,
		Length:         len(passphrase),
		Entropy:        adjustedEntropy,
		CharsetSize:    charsetSize,
		Charsets:       ec.charsetNames(charsets),
		Strength:       calculateStrength(adjustedEntropy),
		CrackTime:      estimateCrackTime(adjustedEntropy),
		Patterns:       patterns,
		WordBased:      wordBased,
		EstimatedWords: estimatedWords,
	}

	return result
}

// detectCharsets determines which character sets are present.
func (ec *EntropyCalculator) detectCharsets(str string) []CharsetInfo {
	var charsets []CharsetInfo

	hasDigits := false
	hasLower := false
	hasUpper := false
	hasSymbols := false
	hasSpace := false

	for _, r := range str { //nolint:varnamelen // r is standard Go idiom for rune iteration
		switch {
		case unicode.IsDigit(r):
			hasDigits = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsSpace(r):
			hasSpace = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSymbols = true
		}
	}

	if hasDigits {
		charsets = append(charsets, DigitCharset)
	}

	if hasLower {
		charsets = append(charsets, LowerCharset)
	}

	if hasUpper {
		charsets = append(charsets, UpperCharset)
	}

	if hasSymbols {
		charsets = append(charsets, SymbolCharset)
	}

	if hasSpace {
		charsets = append(charsets, SpaceCharset)
	}

	return charsets
}

// calculateCharsetSize sums up the sizes of all character sets.
func (ec *EntropyCalculator) calculateCharsetSize(charsets []CharsetInfo) int {
	total := 0

	for _, charset := range charsets {
		total += charset.Size
	}

	return total
}

// charsetNames returns the names of the character sets.
func (ec *EntropyCalculator) charsetNames(charsets []CharsetInfo) []string {
	names := make([]string, len(charsets))
	for i, charset := range charsets {
		names[i] = charset.Name
	}

	return names
}

// detectPatterns finds common patterns that reduce entropy.
func (ec *EntropyCalculator) detectPatterns(str string) []PatternMatch {
	var patterns []PatternMatch

	// Sequential patterns (123, abc, etc.)
	patterns = append(patterns, ec.findSequentialPatterns(str)...)

	// Repetition patterns (aaa, 111, etc.)
	patterns = append(patterns, ec.findRepetitionPatterns(str)...)

	// Dictionary word patterns (common words)
	patterns = append(patterns, ec.findDictionaryPatterns(str)...)

	// Date patterns (1234, 2023, etc.)
	patterns = append(patterns, ec.findDatePatterns(str)...)

	return patterns
}

// findSequentialPatterns finds sequential character patterns.
func (ec *EntropyCalculator) findSequentialPatterns(str string) []PatternMatch {
	var patterns []PatternMatch

	// Look for sequences of 3+ characters.
	for i := 0; i < len(str)-minSequenceLength+1; i++ { //nolint:varnamelen // i is standard loop var
		seqLen := ec.getSequenceLength(str, i)
		if seqLen >= minSequenceLength {
			patterns = append(patterns, PatternMatch{
				Type:        "sequential",
				Description: "Sequential pattern: " + str[i:i+seqLen],
				Position:    i,
				Length:      seqLen,
				Penalty:     float64(seqLen) * sequencePenaltyMultiplier, // Reduce entropy significantly
			})

			i += seqLen - 1 // Skip past this pattern
		}
	}

	return patterns
}

// getSequenceLength returns the length of a sequential pattern starting at pos.
func (ec *EntropyCalculator) getSequenceLength(str string, pos int) int {
	if pos+minSequenceLength-1 >= len(str) {
		return 0
	}

	length := minSequenceLength - 1
	for pos+length < len(str) {
		// Check if characters are sequential
		if str[pos+length] != str[pos+length-1]+1 {
			break
		}

		length++
	}

	return length
}

// findRepetitionPatterns finds repeated character patterns.
func (ec *EntropyCalculator) findRepetitionPatterns(str string) []PatternMatch {
	var patterns []PatternMatch

	// Find sequences of 3+ identical characters.
	for i := 0; i < len(str)-minRepetitionLength+1; i++ { //nolint:varnamelen // i is standard loop var
		if str[i] == str[i+1] && str[i+1] == str[i+minRepetitionLength-1] {
			// Found start of repetition, find end
			j := i + minRepetitionLength - 1 //nolint:varnamelen // j is temp var
			for j < len(str) && str[j] == str[i] {
				j++
			}

			length := j - i

			patterns = append(patterns, PatternMatch{
				Type:        "repetition",
				Description: "Repeated character: " + str[i:j],
				Position:    i,
				Length:      length,
				Penalty:     float64(length-1) * repetitionPenaltyMultiplier, // Heavy penalty for repetition
			})

			i = j - 1 // Skip past this repetition
		}
	}

	return patterns
}

// findDictionaryPatterns finds common dictionary words.
func (ec *EntropyCalculator) findDictionaryPatterns(str string) []PatternMatch {
	var patterns []PatternMatch

	// Simple check for common words (this could be expanded)
	commonWords := []string{
		"password", "admin", "user", "login", "welcome",
		"hello", "world", "test", "demo", "sample",
		"the", "and", "you", "that", "was", "for", "are",
	}

	lower := strings.ToLower(str)

	for _, word := range commonWords {
		if idx := strings.Index(lower, word); idx != -1 {
			patterns = append(patterns, PatternMatch{
				Type:        "dictionary",
				Description: "Common word: " + word,
				Position:    idx,
				Length:      len(word),
				Penalty:     float64(len(word)) * dictionaryPenaltyMultiplier, // Moderate penalty
			})
		}
	}

	return patterns
}

// findDatePatterns finds date-like patterns.
func (ec *EntropyCalculator) findDatePatterns(str string) []PatternMatch {
	// Look for 4-digit years
	re := regexp.MustCompile(`\b(19|20)\d{2}\b`)
	matches := re.FindAllStringIndex(str, -1)

	patterns := make([]PatternMatch, 0, len(matches))

	for _, match := range matches {
		patterns = append(patterns, PatternMatch{
			Type:        "date",
			Description: "Year pattern: " + str[match[0]:match[1]],
			Position:    match[0],
			Length:      match[1] - match[0],
			Penalty:     yearPatternPenalty, // Years have limited entropy
		})
	}

	return patterns
}

// applyPatternPenalties reduces entropy based on detected patterns.
func (ec *EntropyCalculator) applyPatternPenalties(baseEntropy float64, patterns []PatternMatch) float64 {
	totalPenalty := 0.0

	for _, pattern := range patterns {
		totalPenalty += pattern.Penalty
	}

	adjusted := baseEntropy - totalPenalty
	if adjusted < 0 {
		adjusted = 0
	}

	return adjusted
}

// analyzeWordStructure determines if the passphrase appears to be word-based.
func (ec *EntropyCalculator) analyzeWordStructure(str string) (bool, int) {
	// Look for common separators
	separators := []string{"-", "_", ".", " ", ""}

	for _, sep := range separators {
		if ec.looksLikeWordPattern(str, sep) {
			words := ec.countWords(str, sep)

			return true, words
		}
	}

	return false, 0
}

// looksLikeWordPattern checks if the string follows a word+separator pattern.
func (ec *EntropyCalculator) looksLikeWordPattern(str, sep string) bool {
	if sep == "" {
		// For concatenated words, look for alternating case patterns
		return ec.hasAlternatingCase(str)
	}

	// Check if separators divide the string into word-like segments
	parts := strings.Split(str, sep)
	if len(parts) < minWordsRequired {
		return false
	}

	wordLikeCount := 0

	for _, part := range parts {
		if ec.isWordLike(part) {
			wordLikeCount++
		}
	}

	// Most parts should be word-like
	return float64(wordLikeCount)/float64(len(parts)) > minWordLikeRatio
}

// hasAlternatingCase checks for camelCase or similar patterns.
func (ec *EntropyCalculator) hasAlternatingCase(str string) bool {
	hasUpper := false
	hasLower := false

	for _, r := range str {
		if unicode.IsUpper(r) {
			hasUpper = true
		}

		if unicode.IsLower(r) {
			hasLower = true
		}
	}

	return hasUpper && hasLower
}

// isWordLike checks if a string segment resembles a word.
func (ec *EntropyCalculator) isWordLike(str string) bool {
	if len(str) < minWordLength || len(str) > maxWordLength {
		return false
	}

	// Should be mostly letters
	letterCount := 0

	for _, r := range str {
		if unicode.IsLetter(r) {
			letterCount++
		}
	}

	return float64(letterCount)/float64(len(str)) > minLetterRatio
}

// countWords estimates the number of words in a word-based passphrase.
func (ec *EntropyCalculator) countWords(str, sep string) int {
	if sep == "" {
		// For concatenated words, this is harder to estimate
		// Use a simple heuristic based on length and case changes
		return ec.estimateConcatenatedWords(str)
	}

	parts := strings.Split(str, sep)
	wordCount := 0

	for _, part := range parts {
		if ec.isWordLike(part) {
			wordCount++
		}
	}

	return wordCount
}

// estimateConcatenatedWords estimates word count in concatenated strings.
func (ec *EntropyCalculator) estimateConcatenatedWords(str string) int {
	// Simple heuristic: assume average word length of 5-6 characters
	estimatedWords := int(float64(len(str)) / avgWordLength)

	if estimatedWords < 1 {
		estimatedWords = 1
	}

	return estimatedWords
}
