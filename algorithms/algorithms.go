package algorithms

import (
	"math"

	"github.com/serabur/irtlib/data"
)

func birnbaum(c, delta, theta float64, result bool) float64 {
	const alpha = 1.71
	exponent := math.Exp(alpha * (theta - delta))

	if completionProbability := c + (1-c)*(exponent/1+(exponent)); result {
		return completionProbability
	} else {
		return 1 - completionProbability
	}

}

type ItemAnalysisAlgorithm interface {
	Execute(itemData data.ItemData) data.AnalysisData
}

// * ВЫЧИСЛЕНИЕ ФАКТИЧЕСКОЙ ТРУДНОСТИ

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
			likelihoodRatio: calculateLikelihoodRatio(
				itemData.GuessingProbability,
				delta,
				itemData.CompletingResults,
			),
		})
	}

	analysisData := data.AnalysisData{
		ItemID:       itemData.ItemID,
		AnalysisType: data.ActualDifficultyAT,
		Data:         pickApproximateDelta(pairs),
	}

	return analysisData
}

func calculateLikelihoodRatio(guessingProbability, delta float64, completingResults []data.CompletingResult) float64 {
	sum := 0.0
	for _, completingResult := range completingResults {
		sum += math.Log(birnbaum(guessingProbability, delta, completingResult.PreparednessLevel, completingResult.Result))
	}

	return sum
}

func pickApproximateDelta(pairs []deltaRatioPair) float64 {
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

// * ВЫЯВЛЕНИЕ ПОТЕНЦИАЛЬНО НЕКОРРЕКТНЫХ ЗАДАНИЙ

type IncorrectItemDetector struct{}

// * ВЫЯВЛЕНИЕ ПОТЕНЦИАЛЬНО ВЗЛОМАННЫХ ЗАДАНИЙ

type CompromisedItemDetector struct{}
