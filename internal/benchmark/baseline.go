package benchmark

import (
	"encoding/json"
	"fmt"
	"os"
)

type Result struct {
	Name        string  `json:"name"`
	NsPerOp     float64 `json:"nsPerOp"`
	AllocsPerOp float64 `json:"allocsPerOp"`
	BytesPerOp  float64 `json:"bytesPerOp"`
}

type Baseline struct {
	Commit  string   `json:"commit"`
	Date    string   `json:"date"`
	Results []Result `json:"results"`
}

type Regression struct {
	Name  string  `json:"name"`
	Old   float64 `json:"old"`
	New   float64 `json:"new"`
	Delta float64 `json:"delta"`
}

func LoadBaseline(path string) (*Baseline, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read baseline: %w", err)
	}
	var b Baseline
	if err := json.Unmarshal(data, &b); err != nil {
		return nil, fmt.Errorf("unmarshal baseline: %w", err)
	}
	return &b, nil
}

func (b *Baseline) SaveBaseline(path string) error {
	data, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal baseline: %w", err)
	}
	return os.WriteFile(path, data, 0o644)
}

func (b *Baseline) Compare(current Baseline) ([]Regression, error) {
	baselineMap := make(map[string]Result)
	for _, r := range b.Results {
		baselineMap[r.Name] = r
	}

	var regressions []Regression
	for _, cr := range current.Results {
		br, exists := baselineMap[cr.Name]
		if !exists {
			continue
		}
		if br.NsPerOp > 0 {
			delta := ((cr.NsPerOp - br.NsPerOp) / br.NsPerOp) * 100
			if delta > 10 {
				regressions = append(regressions, Regression{
					Name:  cr.Name + "/nsPerOp",
					Old:   br.NsPerOp,
					New:   cr.NsPerOp,
					Delta: delta,
				})
			}
		}
		if br.AllocsPerOp > 0 {
			delta := ((cr.AllocsPerOp - br.AllocsPerOp) / br.AllocsPerOp) * 100
			if delta > 10 {
				regressions = append(regressions, Regression{
					Name:  cr.Name + "/allocsPerOp",
					Old:   br.AllocsPerOp,
					New:   cr.AllocsPerOp,
					Delta: delta,
				})
			}
		}
	}
	return regressions, nil
}
