package filter

import (
	"fmirage/internal/config"
	"fmirage/internal/output"
	"fmirage/internal/utils"
)

type Filter struct {
	MatchCodes map[int]bool
	MatchSizes map[int64]bool
	MatchWords map[int]bool
	MatchLines map[int]bool

	FilterCodes map[int]bool
	FilterSizes map[int64]bool
	FilterWords map[int]bool
	FilterLines map[int]bool
}

func New(cfg *config.Config) *Filter {
	return &Filter{
		MatchCodes: utils.ParseCommaSeparatedToMap[int](cfg.MatchCodes),
		MatchSizes: utils.ParseCommaSeparatedToMap[int64](cfg.MatchSizes),
		MatchWords: utils.ParseCommaSeparatedToMap[int](cfg.MatchWords),
		MatchLines: utils.ParseCommaSeparatedToMap[int](cfg.MatchLines),

		FilterCodes: utils.ParseCommaSeparatedToMap[int](cfg.FilterCodes),
		FilterSizes: utils.ParseCommaSeparatedToMap[int64](cfg.FilterSizes),
		FilterWords: utils.ParseCommaSeparatedToMap[int](cfg.FilterWords),
		FilterLines: utils.ParseCommaSeparatedToMap[int](cfg.FilterLines),
	}
}

func (f *Filter) ShouldMatch(result *output.Result) bool {
	// Check match criteria
	if len(f.MatchCodes) > 0 && !f.MatchCodes[result.StatusCode] {
		return false
	}
	if len(f.MatchSizes) > 0 && !f.MatchSizes[result.Size] {
		return false
	}
	if len(f.MatchWords) > 0 && !f.MatchWords[result.Words] {
		return false
	}
	if len(f.MatchLines) > 0 && !f.MatchLines[result.Lines] {
		return false
	}

	return true
}

func (f *Filter) ShouldDrop(result *output.Result) bool {

	// Check filter criteria
	if f.FilterCodes[result.StatusCode] {
		return true
	}
	if f.FilterSizes[result.Size] {
		return true
	}
	if f.FilterWords[result.Words] {
		return true
	}
	if f.FilterLines[result.Lines] {
		return true
	}

	return false
}
