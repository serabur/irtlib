package algorithms

import (
	"math"
	"sort"
	"time"

	"github.com/serabur/irtlib/data"
)

type CompromisedItemDetector struct{}

func (algorithm CompromisedItemDetector) Execute(itemData data.ItemData) data.AnalysisData {
	successRates := algorithm.calculateSuccessRates(itemData)

	sensitivity := 2.5 // Значение чувствительности из тестирования (раздел 4)
	jumps := algorithm.detectJumps(successRates, sensitivity)

	graphPoints := make([]data.CompromiseGraphPoint, len(successRates))
	for i, rate := range successRates {
		graphPoints[i] = data.CompromiseGraphPoint{
			Frequence: rate.successRate,
			Date:      rate.date,
		}
	}

	compromiseData := data.CompromiseAnalysisData{
		JumpDateSpans: jumps,
		GraphPoints:   graphPoints,
	}

	return data.AnalysisData{
		ItemID:       itemData.ItemID,
		AnalysisType: data.CompromiseDetectionAT,
		Data:         compromiseData,
	}
}

type taskSuccessRate struct {
	date        time.Time
	successRate float64
}

func (algorithm CompromisedItemDetector) calculateSuccessRates(itemData data.ItemData) []taskSuccessRate {
	dateResults := make(map[time.Time][]data.CompletingResult)
	for _, result := range itemData.CompletingResults {
		date := time.Date(result.ExecutionDate.Year(), result.ExecutionDate.Month(), result.ExecutionDate.Day(), 0, 0, 0, 0, result.ExecutionDate.Location())
		dateResults[date] = append(dateResults[date], result)
	}

	successRates := make([]taskSuccessRate, 0, len(dateResults))
	for date, results := range dateResults {
		successfulAttempts := 0.0
		for _, result := range results {
			if result.Result {
				successfulAttempts++
			}
		}
		successRate := successfulAttempts / float64(len(results))
		successRates = append(successRates, taskSuccessRate{
			date:        date,
			successRate: successRate,
		})
	}

	sort.Slice(successRates, func(i, j int) bool {
		return successRates[i].date.Before(successRates[j].date)
	})

	return successRates
}

func (algorithm CompromisedItemDetector) detectJumps(rates []taskSuccessRate, sensitivity float64) []data.DateSpan {
	changes := make([]float64, 0, len(rates)-1)
	for i := 1; i < len(rates); i++ {
		change := rates[i].successRate - rates[i-1].successRate
		changes = append(changes, change)
	}

	var sum, sumSq float64
	for _, change := range changes {
		sum += change
		sumSq += change * change
	}
	meanChange := sum / float64(len(changes))
	stdDev := math.Sqrt(sumSq/float64(len(changes)) - meanChange*meanChange)

	threshold := meanChange + sensitivity*stdDev

	jumps := make([]data.DateSpan, 0)
	inJump := false
	var jumpStart time.Time

	for i := 1; i < len(rates); i++ {
		change := rates[i].successRate - rates[i-1].successRate
		if change > threshold {
			if !inJump {
				jumpStart = rates[i-1].date
				inJump = true
			}
		} else if inJump {
			jumps = append(jumps, data.DateSpan{
				Start: jumpStart,
				End:   rates[i-1].date,
			})
			inJump = false
		}
	}

	if inJump {
		jumps = append(jumps, data.DateSpan{
			Start: jumpStart,
			End:   rates[len(rates)-1].date,
		})
	}

	return jumps
}
