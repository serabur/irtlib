package algorithms

import (
	"math"

	"github.com/serabur/irtlib/data"
)

type DifficultyCalculator struct{}

type deltaRatioPair struct {
	delta           float64
	likelihoodRatio float64
}

func (algorithm DifficultyCalculator) Execute(itemData data.ItemData) data.AnalysisData {
	const step = 0.5
	pairs := []deltaRatioPair{}
	for delta := -3.0; delta <= 3.0; delta += step {
		pairs = append(pairs, deltaRatioPair{
			delta: delta,
			likelihoodRatio: algorithm.calculateLikelihoodRatio(
				itemData.GuessingProbability,
				delta,
				itemData.CompletingResults,
			),
		})
	}

	return data.AnalysisData{
		ItemID:       itemData.ItemID,
		AnalysisType: data.ActualDifficultyAT,
		Data:         algorithm.pickApproximateDelta(pairs),
	}
}

func (algorithm DifficultyCalculator) calculateLikelihoodRatio(guessingProbability, delta float64, completingResults []data.CompletingResult) float64 {
	sum := 0.0
	for _, completingResult := range completingResults {
		sum += math.Log(birnbaum(guessingProbability, delta, completingResult.PreparednessLevel, completingResult.Result))
	}

	return sum
}

func (algorithm DifficultyCalculator) pickApproximateDelta(pairs []deltaRatioPair) float64 {
	maxL := pairs[0].likelihoodRatio
	delta := pairs[0].delta

	for _, pair := range pairs {
		if pair.likelihoodRatio > maxL {
			maxL = pair.likelihoodRatio
			delta = pair.delta
		}
	}

	return delta
}
