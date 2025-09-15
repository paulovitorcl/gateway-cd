package api

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	gatewaycdv1alpha1 "gateway-cd/pkg/api/v1alpha1"
)

// Server represents the API server
type Server struct {
	client client.Client
	router *gin.Engine
}

// NewServer creates a new API server
func NewServer(client client.Client) *Server {
	s := &Server{
		client: client,
		router: gin.Default(),
	}

	s.setupRoutes()
	return s
}

// setupRoutes configures the API routes
func (s *Server) setupRoutes() {
	// CORS middleware
	s.router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	api := s.router.Group("/api/v1")
	{
		// Canary deployment routes
		api.GET("/canaries", s.listCanaryDeployments)
		api.GET("/canaries/:namespace/:name", s.getCanaryDeployment)
		api.POST("/canaries", s.createCanaryDeployment)
		api.PUT("/canaries/:namespace/:name", s.updateCanaryDeployment)
		api.DELETE("/canaries/:namespace/:name", s.deleteCanaryDeployment)

		// Canary control routes
		api.POST("/canaries/:namespace/:name/resume", s.resumeCanaryDeployment)
		api.POST("/canaries/:namespace/:name/pause", s.pauseCanaryDeployment)
		api.POST("/canaries/:namespace/:name/abort", s.abortCanaryDeployment)
		api.POST("/canaries/:namespace/:name/promote", s.promoteCanaryDeployment)

		// Status and metrics routes
		api.GET("/canaries/:namespace/:name/status", s.getCanaryStatus)
		api.GET("/canaries/:namespace/:name/metrics", s.getCanaryMetrics)
		api.GET("/canaries/:namespace/:name/history", s.getCanaryHistory)

		// Health check
		api.GET("/health", s.healthCheck)
	}
}

// Run starts the API server
func (s *Server) Run(addr string) error {
	return s.router.Run(addr)
}

// listCanaryDeployments returns all canary deployments
func (s *Server) listCanaryDeployments(c *gin.Context) {
	var canaries gatewaycdv1alpha1.CanaryDeploymentList

	namespace := c.Query("namespace")
	var listOpts []client.ListOption
	if namespace != "" {
		listOpts = append(listOpts, client.InNamespace(namespace))
	}

	if err := s.client.List(context.Background(), &canaries, listOpts...); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, canaries.Items)
}

// getCanaryDeployment returns a specific canary deployment
func (s *Server) getCanaryDeployment(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	var canary gatewaycdv1alpha1.CanaryDeployment
	if err := s.client.Get(context.Background(), types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, &canary); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Canary deployment not found"})
		return
	}

	c.JSON(http.StatusOK, canary)
}

// createCanaryDeployment creates a new canary deployment
func (s *Server) createCanaryDeployment(c *gin.Context) {
	var canary gatewaycdv1alpha1.CanaryDeployment
	if err := c.ShouldBindJSON(&canary); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.client.Create(context.Background(), &canary); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, canary)
}

// updateCanaryDeployment updates an existing canary deployment
func (s *Server) updateCanaryDeployment(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	var existing gatewaycdv1alpha1.CanaryDeployment
	if err := s.client.Get(context.Background(), types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, &existing); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Canary deployment not found"})
		return
	}

	var updated gatewaycdv1alpha1.CanaryDeployment
	if err := c.ShouldBindJSON(&updated); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Preserve metadata
	updated.ObjectMeta = existing.ObjectMeta
	updated.Status = existing.Status

	if err := s.client.Update(context.Background(), &updated); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updated)
}

// deleteCanaryDeployment deletes a canary deployment
func (s *Server) deleteCanaryDeployment(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	var canary gatewaycdv1alpha1.CanaryDeployment
	if err := s.client.Get(context.Background(), types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, &canary); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Canary deployment not found"})
		return
	}

	if err := s.client.Delete(context.Background(), &canary); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Canary deployment deleted"})
}

// resumeCanaryDeployment resumes a paused canary deployment
func (s *Server) resumeCanaryDeployment(c *gin.Context) {
	s.updateCanaryAnnotation(c, "gateway-cd.io/resume", "true")
}

// pauseCanaryDeployment pauses a running canary deployment
func (s *Server) pauseCanaryDeployment(c *gin.Context) {
	s.updateCanaryAnnotation(c, "gateway-cd.io/pause", "true")
}

// abortCanaryDeployment aborts a canary deployment
func (s *Server) abortCanaryDeployment(c *gin.Context) {
	s.updateCanaryAnnotation(c, "gateway-cd.io/abort", "true")
}

// promoteCanaryDeployment promotes canary to stable
func (s *Server) promoteCanaryDeployment(c *gin.Context) {
	s.updateCanaryAnnotation(c, "gateway-cd.io/promote", "true")
}

// updateCanaryAnnotation is a helper to update canary annotations
func (s *Server) updateCanaryAnnotation(c *gin.Context, key, value string) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	var canary gatewaycdv1alpha1.CanaryDeployment
	if err := s.client.Get(context.Background(), types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, &canary); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Canary deployment not found"})
		return
	}

	if canary.Annotations == nil {
		canary.Annotations = make(map[string]string)
	}
	canary.Annotations[key] = value

	if err := s.client.Update(context.Background(), &canary); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Annotation updated"})
}

// getCanaryStatus returns the current status of a canary deployment
func (s *Server) getCanaryStatus(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	var canary gatewaycdv1alpha1.CanaryDeployment
	if err := s.client.Get(context.Background(), types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, &canary); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Canary deployment not found"})
		return
	}

	// Enhanced status response
	status := map[string]interface{}{
		"phase":             canary.Status.Phase,
		"message":           canary.Status.Message,
		"currentStep":       canary.Status.CurrentStep,
		"totalSteps":        len(canary.Spec.TrafficSplit),
		"canaryWeight":      canary.Status.CanaryWeight,
		"stableWeight":      canary.Status.StableWeight,
		"lastTransition":    canary.Status.LastTransitionTime,
		"conditions":        canary.Status.Conditions,
		"analysisRun":       canary.Status.AnalysisRun,
		"canPause":          canary.Status.Phase == gatewaycdv1alpha1.CanaryDeploymentPhaseProgressing,
		"canResume":         canary.Status.Phase == gatewaycdv1alpha1.CanaryDeploymentPhasePaused,
		"canAbort":          canary.Status.Phase == gatewaycdv1alpha1.CanaryDeploymentPhaseProgressing || canary.Status.Phase == gatewaycdv1alpha1.CanaryDeploymentPhasePaused,
		"canPromote":        canary.Status.Phase == gatewaycdv1alpha1.CanaryDeploymentPhasePaused,
	}

	c.JSON(http.StatusOK, status)
}

// getCanaryMetrics returns metrics for a canary deployment
func (s *Server) getCanaryMetrics(c *gin.Context) {
	// This would integrate with your metrics provider
	// For now, return mock data
	metrics := map[string]interface{}{
		"successRate":    0.995,
		"averageLatency": 150,
		"requestCount":   1250,
		"errorRate":      0.005,
		"throughput":     50.2,
		"timestamp":      metav1.Now(),
	}

	c.JSON(http.StatusOK, metrics)
}

// getCanaryHistory returns the deployment history
func (s *Server) getCanaryHistory(c *gin.Context) {
	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "10")
	limit, _ := strconv.Atoi(limitStr)

	// This would query your database for historical data
	// For now, return mock data
	history := []map[string]interface{}{
		{
			"timestamp":   metav1.Now(),
			"phase":       "Progressing",
			"step":        2,
			"weight":      25,
			"message":     "Traffic split updated: 25% canary, 75% stable",
		},
		{
			"timestamp":   metav1.Now(),
			"phase":       "Progressing",
			"step":        1,
			"weight":      10,
			"message":     "Traffic split updated: 10% canary, 90% stable",
		},
	}

	if len(history) > limit {
		history = history[:limit]
	}

	c.JSON(http.StatusOK, history)
}

// healthCheck returns the API health status
func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": metav1.Now(),
	})
}