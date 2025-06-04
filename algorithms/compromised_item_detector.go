package algorithms

import (
	"math"
	"sort"
	"time"
	"github.com/serabur/irtlib/data"
)

type CompromisedItemDetector struct{}

func (algorithm CompromisedItemDetector) Execute(itemData data.ItemData) data.AnalysisData {
	
	
	return data.AnalysisData{}
}

type taskSuccessRate struct {
	date time.Time
	successRate float64
}

func (algorithm CompromisedItemDetector) newTaskSuccessRate(itemData data.ItemData, date time.Time) taskSuccessRate {
	successfulAttempts := 0.0
	for _, item := range(itemData.CompletingResults) {
		if item.Result {
			successfulAttempts += 1.0
		}
	}
	successRate := float64(len(itemData.CompletingResults))/successfulAttempts

	return taskSuccessRate{
		date: date,
		successRate: successRate,
	}
}

func DetectJumps(rates []taskSuccessRate, sensitivity float64) []data.DateSpan {
	// Сортируем значения по дате
	sort.Slice(rates, func(i, j int) bool {
		return rates[i].date.Before(rates[j].date)
	})

	// Вычисляем изменения
	changes := []float64{}
	for i := 1; i < len(rates); i++ {
		change := rates[i].successRate - rates[i-1].successRate
		changes = append(changes, change)
	}

	// Вычисляем среднее изменение и стандартное отклонение
	var sum, sumSq float64
	for _, change := range changes {
		sum += change
		sumSq += change * change
	}
	meanChange := sum / float64(len(changes))
	stdDev := math.Sqrt(sumSq/float64(len(changes)) - meanChange*meanChange)

	threshold := meanChange + sensitivity*stdDev

	jumps := []data.DateSpan{}
	inJump := false

	for i := 1; i < len(rates); i++ {
		change := rates[i].successRate - rates[i-1].successRate
		if change > threshold {
			if !inJump {
				jumps = append(jumps, data.DateSpan{Start: rates[i-1].date, End: rates[i-1].date})
				inJump = true
			}
		} else {
			inJump = false
		}
	}

	return jumps
}