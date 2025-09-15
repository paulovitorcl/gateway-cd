package metrics

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	gatewaycdv1alpha1 "gateway-cd/pkg/api/v1alpha1"
)

// Provider defines the interface for metrics collection
type Provider interface {
	RunAnalysis(ctx context.Context, canary *gatewaycdv1alpha1.CanaryDeployment) (*AnalysisResult, error)
	GetMetric(ctx context.Context, query string) (float64, error)
}

// AnalysisResult represents the result of running canary analysis
type AnalysisResult struct {
	Phase          string                                   `json:"phase"`
	SuccessRate    float64                                  `json:"successRate"`
	AverageLatency int32                                    `json:"averageLatency"`
	MetricResults  []gatewaycdv1alpha1.MetricResult         `json:"metricResults"`
	StartedAt      *metav1.Time                             `json:"startedAt"`
	CompletedAt    *metav1.Time                             `json:"completedAt"`
	Passed         bool                                     `json:"passed"`
}

// PrometheusProvider implements metrics collection using Prometheus
type PrometheusProvider struct {
	baseURL string
	client  *http.Client
}

// NewPrometheusProvider creates a new Prometheus metrics provider
func NewPrometheusProvider(prometheusURL string) Provider {
	provider := &PrometheusProvider{
		baseURL: strings.TrimSuffix(prometheusURL, "/"),
		client: &http.Client{
			Timeout: time.Second * 30,
		},
	}
	return provider
}

// PrometheusResponse represents a Prometheus query response
type PrometheusResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric map[string]string `json:"metric"`
			Value  []interface{}     `json:"value"`
		} `json:"result"`
	} `json:"data"`
}

// RunAnalysis performs canary analysis using Prometheus metrics
func (p *PrometheusProvider) RunAnalysis(ctx context.Context, canary *gatewaycdv1alpha1.CanaryDeployment) (*AnalysisResult, error) {
	startTime := time.Now()
	result := &AnalysisResult{
		Phase:       "Running",
		StartedAt:   &metav1.Time{Time: startTime},
		Passed:      true,
	}

	// Run analysis for configured metrics
	for _, metric := range canary.Spec.Analysis.Metrics {
		metricResult, err := p.evaluateMetric(ctx, canary, metric)
		if err != nil {
			result.Phase = "Failed"
			result.Passed = false
			return result, fmt.Errorf("failed to evaluate metric %s: %w", metric.Name, err)
		}

		result.MetricResults = append(result.MetricResults, *metricResult)
		if !metricResult.Passed {
			result.Passed = false
		}
	}

	// Check success rate if configured
	if canary.Spec.Analysis.SuccessRate > 0 {
		successRate, err := p.getSuccessRate(ctx, canary)
		if err != nil {
			result.Phase = "Failed"
			result.Passed = false
			return result, fmt.Errorf("failed to get success rate: %w", err)
		}

		result.SuccessRate = successRate
		if successRate < canary.Spec.Analysis.SuccessRate {
			result.Passed = false
		}
	}

	// Check latency if configured
	if canary.Spec.Analysis.MaxLatency > 0 {
		latency, err := p.getAverageLatency(ctx, canary)
		if err != nil {
			result.Phase = "Failed"
			result.Passed = false
			return result, fmt.Errorf("failed to get latency: %w", err)
		}

		result.AverageLatency = latency
		if latency > canary.Spec.Analysis.MaxLatency {
			result.Passed = false
		}
	}

	result.Phase = "Completed"
	if result.Passed {
		result.Phase = "Successful"
	} else {
		result.Phase = "Failed"
	}

	result.CompletedAt = &metav1.Time{Time: time.Now()}
	return result, nil
}

// evaluateMetric evaluates a single metric against its threshold
func (p *PrometheusProvider) evaluateMetric(ctx context.Context, canary *gatewaycdv1alpha1.CanaryDeployment, metric gatewaycdv1alpha1.AnalysisMetric) (*gatewaycdv1alpha1.MetricResult, error) {
	// Replace placeholders in the query
	query := p.replaceQueryPlaceholders(metric.Query, canary)

	value, err := p.GetMetric(ctx, query)
	if err != nil {
		return nil, err
	}

	passed := p.compareValues(value, metric.Threshold, metric.Operator)

	return &gatewaycdv1alpha1.MetricResult{
		Name:      metric.Name,
		Value:     value,
		Threshold: metric.Threshold,
		Passed:    passed,
	}, nil
}

// getSuccessRate calculates the success rate for canary traffic
func (p *PrometheusProvider) getSuccessRate(ctx context.Context, canary *gatewaycdv1alpha1.CanaryDeployment) (float64, error) {
	// Example query for success rate (customize based on your metrics)
	query := fmt.Sprintf(`
		sum(rate(http_requests_total{service="%s-canary",code!~"5.."}[5m])) /
		sum(rate(http_requests_total{service="%s-canary"}[5m]))
	`, canary.Spec.Service.Name, canary.Spec.Service.Name)

	return p.GetMetric(ctx, query)
}

// getAverageLatency calculates the average latency for canary traffic
func (p *PrometheusProvider) getAverageLatency(ctx context.Context, canary *gatewaycdv1alpha1.CanaryDeployment) (int32, error) {
	// Example query for latency (customize based on your metrics)
	query := fmt.Sprintf(`
		histogram_quantile(0.95,
			sum(rate(http_request_duration_seconds_bucket{service="%s-canary"}[5m])) by (le)
		) * 1000
	`, canary.Spec.Service.Name)

	value, err := p.GetMetric(ctx, query)
	if err != nil {
		return 0, err
	}

	return int32(value), nil
}

// GetMetric executes a Prometheus query and returns the first result value
func (p *PrometheusProvider) GetMetric(ctx context.Context, query string) (float64, error) {
	// Build the query URL
	u, err := url.Parse(fmt.Sprintf("%s/api/v1/query", p.baseURL))
	if err != nil {
		return 0, err
	}

	q := u.Query()
	q.Set("query", query)
	u.RawQuery = q.Encode()

	// Execute the request
	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return 0, err
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("prometheus query failed with status %d", resp.StatusCode)
	}

	// Parse the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var promResp PrometheusResponse
	if err := json.Unmarshal(body, &promResp); err != nil {
		return 0, err
	}

	if promResp.Status != "success" {
		return 0, fmt.Errorf("prometheus query failed: %s", promResp.Status)
	}

	if len(promResp.Data.Result) == 0 {
		return 0, fmt.Errorf("no data returned from prometheus query")
	}

	// Extract the value
	valueInterface := promResp.Data.Result[0].Value[1]
	valueStr, ok := valueInterface.(string)
	if !ok {
		return 0, fmt.Errorf("unexpected value type from prometheus")
	}

	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse prometheus value: %w", err)
	}

	return value, nil
}

// replaceQueryPlaceholders replaces placeholders in Prometheus queries
func (p *PrometheusProvider) replaceQueryPlaceholders(query string, canary *gatewaycdv1alpha1.CanaryDeployment) string {
	replacements := map[string]string{
		"{{.Service}}":         canary.Spec.Service.Name,
		"{{.CanaryService}}":   fmt.Sprintf("%s-canary", canary.Spec.Service.Name),
		"{{.Namespace}}":       canary.Namespace,
		"{{.Name}}":           canary.Name,
	}

	result := query
	for placeholder, value := range replacements {
		result = strings.ReplaceAll(result, placeholder, value)
	}

	return result
}

// compareValues compares two values using the specified operator
func (p *PrometheusProvider) compareValues(value, threshold float64, operator string) bool {
	switch operator {
	case ">":
		return value > threshold
	case ">=":
		return value >= threshold
	case "<":
		return value < threshold
	case "<=":
		return value <= threshold
	case "==":
		return value == threshold
	case "!=":
		return value != threshold
	default:
		return false
	}
}