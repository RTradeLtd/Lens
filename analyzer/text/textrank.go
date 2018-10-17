package text

import (
	"github.com/DavidBelicza/TextRank/rank"

	"github.com/DavidBelicza/TextRank/convert"
	"github.com/DavidBelicza/TextRank/parse"

	"github.com/DavidBelicza/TextRank"
)

// TextAnalyzer is used to analyze and extract meta data from text
type TextAnalyzer struct {
	TR        *textrank.TextRank
	Rule      parse.Rule
	Language  convert.Language
	Algorithm rank.Algorithm
}

// NewTextAnalyzer is used to generate our text analyzer
func NewTextAnalyzer(useChainAlgorithm bool) *TextAnalyzer {
	// create our text rank object
	tr := textrank.NewTextRank()
	// generate our ruleset for parsing
	rule := textrank.NewDefaultRule()
	// generate our language filter of stop words
	language := textrank.NewDefaultLanguage()
	var algo rank.Algorithm
	if useChainAlgorithm {
		algo = textrank.NewChainAlgorithm()
	} else {
		algo = textrank.NewDefaultAlgorithm()
	}
	return &TextAnalyzer{
		TR:        tr,
		Rule:      rule,
		Language:  language,
		Algorithm: algo,
	}
}

// Clear is used to reset the data that textrank is parsing
func (ta *TextAnalyzer) Clear() {
	newTR := textrank.NewTextRank()
	ta.TR = newTR
}

// RetrievePhrases is a short wrapper around the FindPhrases function
func (ta *TextAnalyzer) RetrievePhrases(text string) []rank.Phrase {
	ta.TR.Populate(text, ta.Language, ta.Rule)
	ta.TR.Ranking(ta.Algorithm)
	return textrank.FindPhrases(ta.TR)
}

// Summarize is used to summary a given piece of text
func (ta *TextAnalyzer) Summarize(text string, minWeight float32) []string {
	phrases := ta.RetrievePhrases(text)
	pairs := []string{}
	for _, v := range phrases {
		if v.Weight >= minWeight {
			pairs = append(pairs, v.Left)
			pairs = append(pairs, v.Right)
		}
	}
	return pairs
}
