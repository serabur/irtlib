package algorithms

import (
	"math"

	"github.com/serabur/irtlib/data"
)

func birnbaum(c, delta, theta float64, result bool) float64 {
	const alpha = 1.71
	exponent := math.Exp(alpha * (theta - delta))

	if completionProbability := c + (1-c)*(exponent/(1+exponent)); result {
		return completionProbability
	} else {
		return 1 - completionProbability
	}

}

type ItemAnalysisAlgorithm interface {
	Execute(itemData data.ItemData) data.AnalysisData
}