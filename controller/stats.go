package controller

import (
	"math"
	"net/http"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	perfmetrics "github.com/QuantumNous/new-api/pkg/perf_metrics"

	"github.com/gin-gonic/gin"
)

// GetStatsSummary returns platform-level statistics for the model marketplace header.
// GET /api/stats/summary
// Public endpoint — returns aggregate, non-sensitive metrics only.
func GetStatsSummary(c *gin.Context) {
	// model count from pricing cache (already computed, cheap)
	pricing := model.GetPricing()
	modelCount := len(pricing)

	// today's request count and cost from the log table
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Unix()
	todayEnd := now.Unix()

	stat, statErr := model.SumUsedQuota(model.LogTypeConsume, todayStart, todayEnd, "", "", "", 0, "")

	todayRequests := model.CountConsumedLogs(todayStart, todayEnd)
	todayCostUSD := 0.0
	if statErr == nil && stat.Quota > 0 {
		todayCostUSD = math.Round(float64(stat.Quota)/common.QuotaPerUnit*100) / 100
	}

	// average availability from perf metrics (24h success rate across all models)
	avgAvailability := 0.0
	if perfResult, err := perfmetrics.QuerySummaryAll(24); err == nil && len(perfResult.Models) > 0 {
		var totalRate float64
		for _, m := range perfResult.Models {
			totalRate += m.SuccessRate
		}
		rawAvg := totalRate / float64(len(perfResult.Models)) / 100.0
		avgAvailability = math.Round(rawAvg*1000000) / 1000000
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"model_count":      modelCount,
			"avg_availability": avgAvailability,
			"today_requests":   todayRequests,
			"today_cost_usd":   todayCostUSD,
		},
	})
}
