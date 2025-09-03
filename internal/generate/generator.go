package generate

import (
	"fmt"

	"github.com/idelchi/pwgen/internal/dictionary"
)

const (
	// WeakEntropyThreshold is the entropy threshold below which passphrases are considered weak.
	WeakEntropyThreshold = 45
	// OkayEntropyThreshold is the entropy threshold below which passphrases are considered okay.
	OkayEntropyThreshold = 65
	// StrongEntropyThreshold is the entropy threshold below which passphrases are considered strong.
	StrongEntropyThreshold = 80

	// Entropy thresholds for crack time estimation.
	secondsEntropyThreshold   = 30
	minutesEntropyThreshold   = 40
	hoursEntropyThreshold     = 50
	daysEntropyThreshold      = 60
	yearsEntropyThreshold     = 70
	centuriesEntropyThreshold = 80
)

// Generator is the main passphrase generation engine.
type Generator struct {
	dict           dictionary.Dictionary
	patternBuilder *PatternBuilder
}

// NewGenerator creates a new passphrase generator.
func NewGenerator(dict dictionary.Dictionary, defaultSep string) *Generator {
	return &Generator{
		dict:           dict,
		patternBuilder: NewPatternBuilder(dict, defaultSep),
	}
}

// Options represents configuration for passphrase generation.
type Options struct {
	Words      int
	Digits     int
	Symbols    int
	Separator  string
	Casing     string
	Pattern    string
	Kebab      bool
	Snake      bool
	Camel      bool
	Count      int
	MinEntropy int
	MinLength  int
}

// Result represents a generated passphrase with metadata.
type Result struct {
	Passphrase string  `json:"passphrase"`
	Entropy    float64 `json:"entropy"`
	Length     int     `json:"length"`
	Pattern    string  `json:"pattern"`
	Strength   string  `json:"strength"`
	CrackTime  string  `json:"crackTime"`
	PolicyPass bool    `json:"policyPass"`
}

// Generate creates one or more passphrases based on the given options.
func (g *Generator) Generate(opts Options) ([]Result, error) {
	// Build pattern from options or DSL
	var (
		pattern *Pattern
		err     error
	)

	if opts.Pattern != "" {
		pattern, err = g.patternBuilder.BuildFromDSL(opts.Pattern)
	} else {
		pattern, err = g.patternBuilder.BuildFromOptions(
			opts.Words, opts.Digits, opts.Symbols,
			opts.Casing, opts.Separator,
			opts.Kebab, opts.Snake, opts.Camel,
		)
	}

	if err != nil {
		return nil, fmt.Errorf("building pattern: %w", err)
	}

	// Generate requested number of passphrases
	count := opts.Count
	if count <= 0 {
		count = 1
	}

	results := make([]Result, 0, count)

	for i := range count {
		passphrase, err := pattern.Generate()
		if err != nil {
			return nil, fmt.Errorf("generating passphrase %d: %w", i+1, err)
		}

		result := Result{
			Passphrase: passphrase,
			Entropy:    pattern.EntropyBits(),
			Length:     len(passphrase),
			Pattern:    pattern.String(),
			Strength:   calculateStrength(pattern.EntropyBits()),
			CrackTime:  estimateCrackTime(pattern.EntropyBits()),
			PolicyPass: checkPolicy(passphrase, pattern.EntropyBits(), opts.MinLength, opts.MinEntropy),
		}

		results = append(results, result)
	}

	return results, nil
}

// calculateStrength returns a human-readable strength assessment.
func calculateStrength(entropy float64) string {
	switch {
	case entropy < WeakEntropyThreshold:
		return "Weak"
	case entropy < OkayEntropyThreshold:
		return "Okay"
	case entropy < StrongEntropyThreshold:
		return "Strong"
	default:
		return "Excellent"
	}
}

// estimateCrackTime estimates time to crack with 1e10 guesses/sec.
func estimateCrackTime(entropy float64) string {
	if entropy <= 0 {
		return "Instant"
	}

	// 2^(entropy-1) / 1e10 seconds (average time to crack)
	// This is a simplified calculation
	switch {
	case entropy < secondsEntropyThreshold:
		return "Seconds"
	case entropy < minutesEntropyThreshold:
		return "Minutes"
	case entropy < hoursEntropyThreshold:
		return "Hours"
	case entropy < daysEntropyThreshold:
		return "Days"
	case entropy < yearsEntropyThreshold:
		return "Years"
	case entropy < centuriesEntropyThreshold:
		return "Centuries"
	default:
		return "Millennia"
	}
}

// checkPolicy checks if the passphrase meets minimum requirements.
func checkPolicy(passphrase string, entropy float64, minLength, minEntropy int) bool {
	if minLength > 0 && len(passphrase) < minLength {
		return false
	}

	if minEntropy > 0 && entropy < float64(minEntropy) {
		return false
	}

	return true
}

// SetDictionary changes the dictionary used by the generator.
func (g *Generator) SetDictionary(dict dictionary.Dictionary) {
	g.dict = dict
	g.patternBuilder.defaultDict = dict
}
