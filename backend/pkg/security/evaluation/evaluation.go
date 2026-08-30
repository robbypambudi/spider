package evaluation

import (
	"fmt"
	"math"
	"sort"

	"github.com/spider/spider/pkg/apis"
)

func ConfusionCounts(labels []bool, scores []float64, threshold float64) apis.ConfusionCounts {
	var counts apis.ConfusionCounts
	for i, label := range labels {
		predicted := scores[i] >= threshold
		switch {
		case label && predicted:
			counts.TP++
		case !label && predicted:
			counts.FP++
		case !label && !predicted:
			counts.TN++
		case label && !predicted:
			counts.FN++
		}
	}
	return counts
}

func rate(num, den int) float64 {
	if den == 0 {
		return 0
	}
	return float64(num) / float64(den)
}

func MetricsFromCounts(c apis.ConfusionCounts) (tpr, fpr, tnr, fnr, precision, recall, f1 float64) {
	tpr = rate(c.TP, c.TP+c.FN)
	fpr = rate(c.FP, c.FP+c.TN)
	tnr = rate(c.TN, c.FP+c.TN)
	fnr = rate(c.FN, c.TP+c.FN)
	precision = rate(c.TP, c.TP+c.FP)
	recall = tpr
	if precision+recall > 0 {
		f1 = 2 * precision * recall / (precision + recall)
	}
	return
}

type rocPoint struct {
	tpr float64
	fpr float64
}

func ROCCurve(labels []bool, scores []float64) []rocPoint {
	type pair struct {
		score float64
		label bool
	}
	pairs := make([]pair, len(labels))
	for i := range labels {
		pairs[i] = pair{score: scores[i], label: labels[i]}
	}
	sort.Slice(pairs, func(i, j int) bool { return pairs[i].score > pairs[j].score })

	pos := 0
	neg := 0
	for _, l := range labels {
		if l {
			pos++
		} else {
			neg++
		}
	}

	points := []rocPoint{{tpr: 0, fpr: 0}}
	tp, fp := 0, 0
	for _, p := range pairs {
		if p.label {
			tp++
		} else {
			fp++
		}
		points = append(points, rocPoint{
			tpr: rate(tp, pos),
			fpr: rate(fp, neg),
		})
	}
	return points
}

func AUROC(labels []bool, scores []float64) float64 {
	points := ROCCurve(labels, scores)
	area := 0.0
	for i := 1; i < len(points); i++ {
		width := points[i].fpr - points[i-1].fpr
		height := (points[i].tpr + points[i-1].tpr) / 2
		area += width * height
	}
	return math.Max(0, math.Min(1, area))
}

func TPRAtFPR(labels []bool, scores []float64, targetFPR float64) float64 {
	points := ROCCurve(labels, scores)
	best := 0.0
	for _, p := range points {
		if p.fpr <= targetFPR && p.tpr >= best {
			best = p.tpr
		}
	}
	return best
}

func EvaluateScores(labels []bool, scores []float64, threshold float64, targetFPRs []float64) apis.EvaluationReport {
	counts := ConfusionCounts(labels, scores, threshold)
	tpr, fpr, tnr, fnr, precision, recall, f1 := MetricsFromCounts(counts)
	tprAt := make(map[string]float64)
	for _, tf := range targetFPRs {
		tprAt[fmt.Sprintf("%g", tf)] = TPRAtFPR(labels, scores, tf)
	}
	return apis.EvaluationReport{
		Counts:         counts,
		TPR:            tpr,
		FPR:            fpr,
		TNR:            tnr,
		FNR:            fnr,
		Precision:      precision,
		Recall:         recall,
		F1:             f1,
		AUC:            AUROC(labels, scores),
		TPRAtTargetFPR: tprAt,
		Threshold:      threshold,
		Samples:        len(labels),
	}
}

func SecurityOverhead(baselineMs, protectedMs float64) (overheadMs, overheadPercent float64) {
	overheadMs = protectedMs - baselineMs
	if baselineMs > 0 {
		overheadPercent = (overheadMs / baselineMs) * 100
	}
	return
}
