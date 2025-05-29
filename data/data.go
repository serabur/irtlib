package data

import "time"

const (
	ActualDifficultyAT = "ACTUAL_DIFFICULTY_CALCULATION"
	IncorrectDetectionAT = "INCORRECT_ITEM_DETECTION"
	CompromiseDetectionAT = "COMPROMISE_ITEM_DETECTION"
)

type ItemData struct {
	ItemID string
	ActualDifficulty float64
	GuessingProbability float64
	CompletingResults []CompletingResult
}

type CompletingResult struct {
	Result bool
	PreparednessLevel float64
	ExecutionDate time.Time
}

type AnalysisData struct {
	ItemID string
	AnalysisType string
	Data any
}

type CorrectnessAnalysisData struct {
	CriticalityLevel float64
	HypothesesLikelihoodRatios [3]float64
	GraphPoints [][]CorrectnessGraphPoint
}

type CorrectnessGraphPoint struct {
	Frequence float64
	GroupTheta float64
}

type CompromiseAnalysisData struct {
	JumpDateSpans []DateSpan
	GraphPoints []CompromiseGraphPoint
}

type DateSpan struct {
	Start time.Time
	End time.Time 
}

type CompromiseGraphPoint struct {
	Frequence float64
	Date time.Time
}