package algorithms

import (
	"math"
	"sort"

	"github.com/serabur/irtlib/data"
)

type IncorrectItemDetector struct{}

func (algorithm IncorrectItemDetector) Execute(itemData data.ItemData) data.AnalysisData {
	pairGroups := algorithm.formPairGroups(itemData.CompletingResults)

	actualPoints := algorithm.formActualPoints(pairGroups)
	birnbaumPoints := algorithm.formBirnbaumPoints(itemData.GuessingProbability, itemData.ActualDifficulty, pairGroups)
	positiveCIPoints, negativeCIPoints := algorithm.formCIPoints(itemData.ActualDifficulty, itemData.GuessingProbability, pairGroups)

	criticalityGM := algorithm.criticalityLevelGM(itemData.GuessingProbability, itemData.ActualDifficulty, actualPoints, birnbaumPoints, pairGroups)

	L1 := algorithm.calculateCorrectHypothesis(itemData.GuessingProbability, itemData.ActualDifficulty, itemData.CompletingResults)
	L2 := algorithm.calculateIndifferentHypothesis(itemData.CompletingResults)
	L3 := algorithm.calculateIncorrectHypothesis(itemData.GuessingProbability, itemData.ActualDifficulty, itemData.CompletingResults)

	criticalityHM := algorithm.criticalityLevelHM(L1, L2, L3)

	criticalityLevel := criticalityGM + criticalityHM

	correctnessAnalysisData := data.CorrectnessAnalysisData{
		CriticalityLevel:           criticalityLevel,
		HypothesesLikelihoodRatios: [3]float64{L1, L2, L3},
		GraphPoints: [][]data.CorrectnessGraphPoint{
			actualPoints,
			birnbaumPoints,
			positiveCIPoints,
			negativeCIPoints,
		},
	}

	return data.AnalysisData{
		ItemID:       itemData.ItemID,
		AnalysisType: data.IncorrectDetectionAT,
		Data:         correctnessAnalysisData,
	}
}

func (algorithm IncorrectItemDetector) formPairGroups(pairs []data.CompletingResult) [][]data.CompletingResult {
	const (
		minGroupCount = 5
		maxGroupCount = 8
	)

	if len(pairs) == 0 {
		return [][]data.CompletingResult{}
	}

	sortedPairs := make([]data.CompletingResult, len(pairs))
	copy(sortedPairs, pairs)
	sort.Slice(sortedPairs, func(i, j int) bool {
		return sortedPairs[i].PreparednessLevel < sortedPairs[j].PreparednessLevel
	})

	resultsCount := len(sortedPairs)
	minRemainder := resultsCount % maxGroupCount
	groupCount := maxGroupCount

	for i := maxGroupCount; i >= minGroupCount; i-- {
		if resultsCount%i < minRemainder {
			minRemainder = resultsCount % i
			groupCount = i
		}
	}

	groups := make([][]data.CompletingResult, groupCount)
	startIndex := 0
	groupSize := resultsCount / groupCount
	remainder := resultsCount % groupCount

	for i := 0; i < groupCount; i++ {
		currentSize := groupSize
		if i < remainder {
			currentSize++
		}
		groups[i] = sortedPairs[startIndex : startIndex+currentSize]
		startIndex += currentSize
	}

	return groups
}

func (algorithm IncorrectItemDetector) formActualPoints(pairGroups [][]data.CompletingResult) []data.CorrectnessGraphPoint {
	actualPoints := make([]data.CorrectnessGraphPoint, 0, len(pairGroups))

	for _, group := range pairGroups {
		if len(group) == 0 {
			continue
		}

		var thetaSum float64
		for _, result := range group {
			thetaSum += result.PreparednessLevel
		}
		avgTheta := thetaSum / float64(len(group))

		successful := algorithm.correctAnswersAmount(group)
		frequence := float64(successful) / float64(len(group))

		actualPoints = append(actualPoints, data.CorrectnessGraphPoint{
			Frequence:  frequence,
			GroupTheta: avgTheta,
		})
	}

	return actualPoints
}

func (algorithm IncorrectItemDetector) correctAnswersAmount(pairs []data.CompletingResult) int {
	sum := 0
	for _, item := range pairs {
		if item.Result {
			sum++
		}
	}
	return sum
}

func (algorithm IncorrectItemDetector) formBirnbaumPoints(guessingProbability, actualDifficulty float64, pairGroups [][]data.CompletingResult) []data.CorrectnessGraphPoint {
	birnbaumPoints := make([]data.CorrectnessGraphPoint, 0, len(pairGroups))

	for _, group := range pairGroups {
		if len(group) == 0 {
			continue
		}

		var thetaSum float64
		for _, result := range group {
			thetaSum += result.PreparednessLevel
		}
		avgTheta := thetaSum / float64(len(group))

		frequence := birnbaum(guessingProbability, actualDifficulty, avgTheta, true)

		birnbaumPoints = append(birnbaumPoints, data.CorrectnessGraphPoint{
			Frequence:  frequence,
			GroupTheta: avgTheta,
		})
	}

	return birnbaumPoints
}

func (algorithm IncorrectItemDetector) formCIPoints(actualDifficulty, guessingProbability float64, pairGroups [][]data.CompletingResult) ([]data.CorrectnessGraphPoint, []data.CorrectnessGraphPoint) {
	positiveCIPoints := make([]data.CorrectnessGraphPoint, 0, len(pairGroups))
	negativeCIPoints := make([]data.CorrectnessGraphPoint, 0, len(pairGroups))

	for _, group := range pairGroups {
		if len(group) == 0 {
			continue
		}

		var thetaSum float64
		for _, result := range group {
			thetaSum += result.PreparednessLevel
		}
		avgTheta := thetaSum / float64(len(group))

		frequence := birnbaum(guessingProbability, actualDifficulty, avgTheta, true)
		stdDev := algorithm.standartDeviation(len(group), frequence)

		positiveCIPoints = append(positiveCIPoints, data.CorrectnessGraphPoint{
			GroupTheta: avgTheta,
			Frequence:  frequence + stdDev,
		})
		negativeCIPoints = append(negativeCIPoints, data.CorrectnessGraphPoint{
			GroupTheta: avgTheta,
			Frequence:  frequence - stdDev,
		})
	}

	return positiveCIPoints, negativeCIPoints
}

func (algorithm IncorrectItemDetector) groupSigmas(pairGroups [][]data.CompletingResult, guessingProbability, actualDifficulty float64) []float64 {
	groupSigmas := make([]float64, 0, len(pairGroups))

	for _, group := range pairGroups {
		if len(group) == 0 {
			continue
		}

		var thetaSum float64
		for _, result := range group {
			thetaSum += result.PreparednessLevel
		}
		avgTheta := thetaSum / float64(len(group))

		frequence := birnbaum(guessingProbability, actualDifficulty, avgTheta, true)
		groupSigmas = append(groupSigmas, algorithm.standartDeviation(len(group), frequence))
	}

	return groupSigmas
}

func (algorithm IncorrectItemDetector) standartDeviation(studentCount int, frequence float64) float64 {
	return math.Sqrt((frequence * (1.0 - frequence))/float64(studentCount))
}

func (algorithm IncorrectItemDetector) calculateCorrectHypothesis(guessingProbability, actualDifficulty float64, results []data.CompletingResult) float64 {
	var L1 float64
	for _, result := range results {
		Pi := birnbaum(guessingProbability, actualDifficulty, result.PreparednessLevel, result.Result)
		L1 += math.Log(Pi)
	}
	return L1
}

func (algorithm IncorrectItemDetector) calculateIndifferentHypothesis(results []data.CompletingResult) float64 {
	n := len(results)
	return -float64(n) * math.Log(2)
}

func (algorithm IncorrectItemDetector) calculateIncorrectHypothesis(guessingProbability, actualDifficulty float64, results []data.CompletingResult) float64 {
	var L3 float64
	for _, result := range results {
		Pi := algorithm.birnbaumWithAlpha(guessingProbability, actualDifficulty, result.PreparednessLevel, result.Result, -1.71)
		L3 += math.Log(Pi)
	}
	return L3
}

func (algorithm IncorrectItemDetector) birnbaumWithAlpha(c, delta, theta float64, result bool, alpha float64) float64 {
	exponent := math.Exp(alpha * (theta - delta))
	completionProbability := c + (1-c)*(exponent/(1+exponent))
	if result {
		return completionProbability
	}
	return 1 - completionProbability
}

func (algorithm IncorrectItemDetector) criticalityLevelGM(guessingProbability, actualDifficulty float64, actualPoints, birnbaumPoints []data.CorrectnessGraphPoint, pairGroups [][]data.CompletingResult) float64 {
	actualSquare := 0.0
	for i := 0; i < len(actualPoints); i++ {
		if len(pairGroups[i]) == 0 {
			continue
		}
		var thetaMin, thetaMax float64
		for j, result := range pairGroups[i] {
			if j == 0 || result.PreparednessLevel < thetaMin {
				thetaMin = result.PreparednessLevel
			}
			if j == 0 || result.PreparednessLevel > thetaMax {
				thetaMax = result.PreparednessLevel
			}
		}
		groupWidth := thetaMax - thetaMin
		actualSquare += math.Abs(actualPoints[i].Frequence-birnbaumPoints[i].Frequence) * groupWidth
	}

	confidenceSquare := 0.0
	sigmas := algorithm.groupSigmas(pairGroups, guessingProbability, actualDifficulty)
	for i := 0; i < len(sigmas); i++ {
		if len(pairGroups[i]) == 0 {
			continue
		}
		var thetaMin, thetaMax float64
		for j, result := range pairGroups[i] {
			if j == 0 || result.PreparednessLevel < thetaMin {
				thetaMin = result.PreparednessLevel
			}
			if j == 0 || result.PreparednessLevel > thetaMax {
				thetaMax = result.PreparednessLevel
			}
		}
		groupWidth := thetaMax - thetaMin
		confidenceSquare += sigmas[i] * groupWidth
	}

	if confidenceSquare == 0 {
		return 0
	}
	return actualSquare / confidenceSquare
}

func (algorithm IncorrectItemDetector) criticalityLevelHM(L1, L2, L3 float64) float64 {
	sum := L3 + L2 + L1
	if sum == 0 {
		return 0
	}
	return (L3 + 0.1*L2 - L1) / sum
}
