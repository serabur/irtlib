package algorithms

import (
	"math"

	"github.com/serabur/irtlib/data"
)

type IncorrectItemDetector struct{}

func (algorithm IncorrectItemDetector) Execute(itemData data.ItemData) data.AnalysisData {
	pairGroups := algorithm.formPairGroups(itemData.CompletingResults)
	actualPoints := algorithm.formActualPoints(pairGroups)
	birnbaumPoints := algorithm.formBirnbaumPoints(itemData.GuessingProbability, itemData.ActualDifficulty, pairGroups)
	positiveCIPoints, negativeCIPoints := algorithm.formCIPoints(itemData.ActualDifficulty, itemData.GuessingProbability, pairGroups)

	criticalityLevel := algorithm.criticalityLevelGM(itemData.GuessingProbability, itemData.ActualDifficulty, actualPoints, birnbaumPoints, pairGroups)

	correctnessAnalysisData := data.CorrectnessAnalysisData{
		CriticalityLevel: criticalityLevel,
		HypothesesLikelihoodRatios: [3]float64{0.0, 0.0, 0.0},
		GraphPoints: [][]data.CorrectnessGraphPoint{
			actualPoints,
			birnbaumPoints,
			positiveCIPoints,
			negativeCIPoints,
		},
	}

	return data.AnalysisData{
		ItemID: itemData.ItemID,
		AnalysisType: data.IncorrectDetectionAT,
		Data: correctnessAnalysisData,
	}
}

func (algorithm IncorrectItemDetector) formPairGroups(pairs []data.CompletingResult) [][]data.CompletingResult {
	const (
		minGroupCount = 5
		maxGroupCount = 8
	)

	resultsCount := len(pairs)
	minRemainder := resultsCount % maxGroupCount
	groupCount := maxGroupCount

	for i := maxGroupCount; i >= minGroupCount; i-- {
		if resultsCount % i < minRemainder {
			minRemainder = resultsCount % i
			groupCount = i
		}
	}

	groups := [][]data.CompletingResult{}
	startIndex := 0
	groupSize := int(math.Floor(float64(resultsCount) / float64(groupCount)))

	for i := 0; i < groupCount; i++ {
		currentSize := groupSize
		if i == 0 || i == (groupCount - 1) {
			currentSize += int(math.Floor(float64(minRemainder) / 2.0))
		}

		groups = append(groups, pairs[startIndex:startIndex + currentSize])
		startIndex += currentSize
	}

	return groups
}

func (algorithm IncorrectItemDetector) formActualPoints(pairGroups [][]data.CompletingResult) []data.CorrectnessGraphPoint {
	actualPoints := []data.CorrectnessGraphPoint{}
	previousPoint := data.CorrectnessGraphPoint{Frequence: 0.0, GroupTheta: 0.0}

	for _, group := range pairGroups {
		if len(group) != 0  {
			currentPoint := data.CorrectnessGraphPoint {
				Frequence: float64(algorithm.correctAnswersAmount(group)) / float64(len(group)),
				GroupTheta: group[0].PreparednessLevel,
			}

			previousPoint = currentPoint
			actualPoints = append(actualPoints, currentPoint)
		} else {
			actualPoints = append(actualPoints, previousPoint)
		}
	}

	return actualPoints
}

func (algorithm IncorrectItemDetector) correctAnswersAmount(pairs []data.CompletingResult) int {
	sum := 0
	for _, item := range pairs {
		if item.Result {
			sum += 1
		}
	}

	return sum
}

func (algorithm IncorrectItemDetector) formBirnbaumPoints(guessingProbability, actualDifficulty float64, pairGroups [][]data.CompletingResult) []data.CorrectnessGraphPoint {
	birnbaumPoints := []data.CorrectnessGraphPoint{}
	previousPoint := data.CorrectnessGraphPoint{Frequence: 0.0, GroupTheta: 0.0}

	for _, group := range pairGroups {
		if len(group) != 0 {
			currentPoint := data.CorrectnessGraphPoint{
				Frequence: birnbaum(guessingProbability, actualDifficulty, group[0].PreparednessLevel, true),
				GroupTheta: group[0].PreparednessLevel,
			}

			previousPoint = currentPoint
			birnbaumPoints = append(birnbaumPoints, currentPoint)
		} else {
			birnbaumPoints = append(birnbaumPoints, previousPoint)
		}
	}

	return birnbaumPoints
}

func (algorithm IncorrectItemDetector) formCIPoints(actualDifficulty, guessingProbability float64, pairGroups [][]data.CompletingResult) ([]data.CorrectnessGraphPoint, []data.CorrectnessGraphPoint) {
	positiveCIPoints := []data.CorrectnessGraphPoint{}
	negativeCIPoints := []data.CorrectnessGraphPoint{}

	previousPositiveCIPoint := data.CorrectnessGraphPoint{Frequence: 0.0, GroupTheta: 0.0}
	previousNegativeCIPoint := data.CorrectnessGraphPoint{Frequence: 0.0, GroupTheta: 0.0}

	for _, group := range pairGroups {
		if len(group) != 0 {
			studentsAmount := len(group)
			frequence := birnbaum(guessingProbability, actualDifficulty, group[0].PreparednessLevel, true)

			currentPositivePoint := data.CorrectnessGraphPoint{
				GroupTheta:     group[0].PreparednessLevel,
				Frequence: frequence + algorithm.standartDeviation(studentsAmount, frequence),
			}

			currentNegativePoint := data.CorrectnessGraphPoint{
				GroupTheta:     group[0].PreparednessLevel,
				Frequence: frequence - algorithm.standartDeviation(studentsAmount, frequence),
			}

			previousPositiveCIPoint = currentPositivePoint
			previousNegativeCIPoint = currentNegativePoint

			positiveCIPoints = append(positiveCIPoints, currentPositivePoint)
			negativeCIPoints = append(negativeCIPoints, currentNegativePoint)
		} else {
			positiveCIPoints = append(positiveCIPoints, previousPositiveCIPoint)
			negativeCIPoints = append(negativeCIPoints, previousNegativeCIPoint)
		}
	}

	return positiveCIPoints, negativeCIPoints
}

func (algorithm IncorrectItemDetector) groupSigmas(pairGroups [][]data.CompletingResult, guessingProbability, actualDifficulty float64) []float64 {
	groupSigmas := []float64{}
	previousSigma := 0.0

	for _, group := range pairGroups {
		if len(group) != 0 {
			studentsAmount := len(group)
			frequence := birnbaum(guessingProbability, actualDifficulty, group[0].PreparednessLevel, true)
			previousSigma = algorithm.standartDeviation(studentsAmount, frequence)
			groupSigmas = append(groupSigmas, algorithm.standartDeviation(studentsAmount, frequence))
		} else {
			groupSigmas = append(groupSigmas, previousSigma)
		}
	}

	return groupSigmas
}

func (algorithm IncorrectItemDetector) standartDeviation(studentCount int, frequence float64) float64 {
	return math.Sqrt(float64(studentCount) * frequence * (1.0 - frequence))
}

func (algorithm IncorrectItemDetector) criticalityLevelGM(guessingProbability, actualDifficulty float64, actualPoints, birnbaumPoints []data.CorrectnessGraphPoint, pairGroups [][]data.CompletingResult) float64 {
	actualSquare := 0.0
	for i := 0; i < len(actualPoints); i++ {
		groupWidth := pairGroups[i][len(pairGroups[i]) - 1].PreparednessLevel - pairGroups[i][0].PreparednessLevel
		actualSquare += math.Abs(actualPoints[i].Frequence - birnbaumPoints[i].Frequence) * groupWidth
	}

	confidenceSquare := 0.0
	sigmas := algorithm.groupSigmas(pairGroups, guessingProbability, actualDifficulty)

	for i := 0; i < len(sigmas); i++ {
		groupWidth := pairGroups[i][len(pairGroups) - 1].PreparednessLevel - pairGroups[i][0].PreparednessLevel
		confidenceSquare += sigmas[i] * groupWidth
	}

	return actualSquare / confidenceSquare
}