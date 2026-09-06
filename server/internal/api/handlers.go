package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"walkover/server/config"
	"walkover/server/internal/clustering"
	"walkover/server/internal/models"
	"walkover/server/internal/urgency"
)

// Handler manages all REST API request processing
type Handler struct {
	cfg             *config.Config
	hub             *Hub
	urgencyEngine   *urgency.Engine
	clusterEngine   *clustering.Engine
	wards           []models.Ward
	signals         []models.CitizenSignal
	clusters        []models.Cluster
	auditLogs       []models.AuditLog
	mutex           sync.RWMutex
}

// NewHandler initializes API handlers and seeds initial memory store from wards dataset
func NewHandler(cfg *config.Config, hub *Hub) *Handler {
	urgencyEngine := urgency.NewEngine()
	clusterEngine := clustering.NewEngine(urgencyEngine)

	h := &Handler{
		cfg:           cfg,
		hub:           hub,
		urgencyEngine: urgencyEngine,
		clusterEngine: clusterEngine,
		signals:       make([]models.CitizenSignal, 0),
		clusters:      make([]models.Cluster, 0),
		auditLogs:     make([]models.AuditLog, 0),
	}

	h.loadInitialWardsAndSeedData()
	return h
}

// loadInitialWardsAndSeedData reads indore_wards.json and creates initial seed clusters
func (h *Handler) loadInitialWardsAndSeedData() {
	wardsFile := "data/indore_wards.json"
	data, err := os.ReadFile(wardsFile)
	if err != nil {
		data, err = os.ReadFile("../data/indore_wards.json")
	}

	if err == nil {
		var wards []models.Ward
		if err := json.Unmarshal(data, &wards); err == nil {
			h.wards = wards
		}
	}

	// Create initial seed clusters demonstrating Hotspots and Blind Spots
	now := time.Now()

	// Cluster 1: Critical Water Contamination in Ward 14 (Chandan Nagar)
	c1Urgency := h.urgencyEngine.CalculateUrgency(
		5,
		[]string{"CONTAMINATED_WATER", "SEWAGE_MIXING"},
		8,
		6,
		[]string{"Community Clinic"},
		now.Add(-4*time.Hour),
	)
	c1Scores := h.clusterEngine.ComputeFourDimensionalScores(8, c1Urgency.Score, 2, 0.32, true, true)
	c1Rec := h.clusterEngine.GenerateGroundedRecommendation("Water Supply & Sewerage", "Ward 14 (Chandan Nagar)", "Sewage Contamination in Main Drinking Line", 8, c1Urgency, []string{"WhatsApp", "SMS"})

	cluster1 := models.Cluster{
		ID:             "cluster-indore-001",
		WardID:         "indore-ward-02",
		WardName:       "Ward 14 - Chandan Nagar",
		Department:     "Water Supply & Sewerage",
		Title:          "Sewage Contamination in Main Drinking Line",
		Description:    "Multiple citizen reports describing brownish, foul-smelling tap water and sewage backflow near street 4 community clinic.",
		SignalCount:    8,
		SignalIDs:      []string{"sig-01", "sig-02", "sig-03", "sig-04", "sig-05", "sig-06", "sig-07", "sig-08"},
		Channels:       []string{"WhatsApp", "SMS"},
		Urgency:        c1Urgency,
		Scores:         c1Scores,
		Recommendation: c1Rec,
		IsBlindSpot:    false,
		Status:         "PENDING",
		CentroidLat:    22.7092,
		CentroidLng:    75.8236,
		CreatedAt:      now.Add(-3 * time.Hour),
		UpdatedAt:      now,
	}

	// Cluster 2: Arterial Road Cave-in in Ward 22 (Vijay Nagar)
	c2Urgency := h.urgencyEngine.CalculateUrgency(
		4,
		[]string{"CAVE_IN", "HOSPITAL_ROUTE"},
		12,
		4,
		[]string{"Mother & Child Care Hospital", "BRTS Main Hub"},
		now.Add(-12*time.Hour),
	)
	c2Scores := h.clusterEngine.ComputeFourDimensionalScores(12, c2Urgency.Score, 3, 0.84, true, true)
	c2Rec := h.clusterEngine.GenerateGroundedRecommendation("Public Works Department (PWD)", "Ward 22 (Vijay Nagar)", "Major Road Caved-in on Hospital Approach Road", 12, c2Urgency, []string{"WhatsApp", "SMS", "WebPortal"})

	cluster2 := models.Cluster{
		ID:             "cluster-indore-002",
		WardID:         "indore-ward-03",
		WardName:       "Ward 22 - Vijay Nagar",
		Department:     "Public Works Department (PWD)",
		Title:          "Major Road Caved-in on Hospital Approach Road",
		Description:    "Deep crater and road surface collapse near BRTS intersection obstructing emergency ambulance ingress.",
		SignalCount:    12,
		SignalIDs:      []string{"sig-11", "sig-12", "sig-13", "sig-14"},
		Channels:       []string{"WhatsApp", "SMS", "WebPortal"},
		Urgency:        c2Urgency,
		Scores:         c2Scores,
		Recommendation: c2Rec,
		IsBlindSpot:    false,
		Status:         "INVESTIGATING",
		CentroidLat:    22.7533,
		CentroidLng:    75.8937,
		CreatedAt:      now.Add(-10 * time.Hour),
		UpdatedAt:      now.Add(-1 * time.Hour),
	}

	// Cluster 3: Open Manhole in Ward 60 (Khajrana)
	c3Urgency := h.urgencyEngine.CalculateUrgency(
		5,
		[]string{"MANHOLE"},
		6,
		3,
		[]string{"High-density Slum Cluster"},
		now.Add(-24*time.Hour),
	)
	c3Scores := h.clusterEngine.ComputeFourDimensionalScores(6, c3Urgency.Score, 1, 0.44, true, true)
	c3Rec := h.clusterEngine.GenerateGroundedRecommendation("Sanitation & Drainage", "Ward 60 (Khajrana)", "Uncovered Deep Drainage Manhole on School Path", 6, c3Urgency, []string{"WhatsApp"})

	cluster3 := models.Cluster{
		ID:             "cluster-indore-003",
		WardID:         "indore-ward-09",
		WardName:       "Ward 60 - Khajrana",
		Department:     "Sanitation & Drainage",
		Title:          "Uncovered Deep Drainage Manhole on School Path",
		Description:    "Drain cover broken during monsoon runoff; high hazard for pedestrians and school children.",
		SignalCount:    6,
		SignalIDs:      []string{"sig-21", "sig-22", "sig-23"},
		Channels:       []string{"WhatsApp"},
		Urgency:        c3Urgency,
		Scores:         c3Scores,
		Recommendation: c3Rec,
		IsBlindSpot:    false,
		Status:         "PENDING",
		CentroidLat:    22.7314,
		CentroidLng:    75.9083,
		CreatedAt:      now.Add(-18 * time.Hour),
		UpdatedAt:      now.Add(-2 * time.Hour),
	}

	h.clusters = []models.Cluster{cluster1, cluster2, cluster3}

	// Update blind spot markers on wards
	h.wards = h.clusterEngine.DetectBlindSpots(h.wards, h.clusters)

	// Seed recent signals
	h.signals = []models.CitizenSignal{
		{
			ID:           "sig-01",
			Provider:     models.ProviderWhatsApp,
			RawText:      "हमारे यहाँ नल से गंदा पानी आ रहा है बहुत बदबू है चंदन नगर गली 4",
			Language:     "Hindi",
			LocationHint: "Chandan Nagar Street 4",
			SenderPhone:  "+91-98260XXXXX",
			Timestamp:    now.Add(-25 * time.Minute),
		},
		{
			ID:           "sig-02",
			Provider:     models.ProviderWhatsApp,
			RawText:      "Urgent: Sewage mixing with drinking water line near clinic in Chandan nagar.",
			Language:     "English",
			LocationHint: "Chandan Nagar Near Clinic",
			SenderPhone:  "+91-97550XXXXX",
			Timestamp:    now.Add(-15 * time.Minute),
		},
		{
			ID:           "sig-11",
			Provider:     models.ProviderSMS,
			RawText:      "Vijay Nagar square ke paas ambulance route par road dhas gayi hai",
			Language:     "Hinglish",
			LocationHint: "Vijay Nagar Square",
			SenderPhone:  "+91-94250XXXXX",
			Timestamp:    now.Add(-40 * time.Minute),
		},
	}
}

// HealthCheck responds with server status
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Civic Development Intelligence Engine is running",
		Data: gin.H{
			"version":   "1.0.0-gdg-indore",
			"timestamp": time.Now(),
			"status":    "operational",
		},
	})
}

// GetWards returns all Indore wards with blind spot status
func (h *Handler) GetWards(c *gin.Context) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	c.JSON(http.StatusOK, models.ApiResponse{
		Success: true,
		Data:    h.wards,
	})
}

// GetClusters returns active issue clusters
func (h *Handler) GetClusters(c *gin.Context) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	c.JSON(http.StatusOK, models.ApiResponse{
		Success: true,
		Data:    h.clusters,
	})
}

// GetClusterByID returns single cluster detail
func (h *Handler) GetClusterByID(c *gin.Context) {
	id := c.Param("id")
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for _, cl := range h.clusters {
		if cl.ID == id {
			c.JSON(http.StatusOK, models.ApiResponse{
				Success: true,
				Data:    cl,
			})
			return
		}
	}

	c.JSON(http.StatusNotFound, models.ApiResponse{
		Success: false,
		Error:   "Cluster not found",
	})
}

// GetSignals returns recent citizen signals feed
func (h *Handler) GetSignals(c *gin.Context) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	c.JSON(http.StatusOK, models.ApiResponse{
		Success: true,
		Data:    h.signals,
	})
}

// GetAuditLogs returns policymaker decision audit history
func (h *Handler) GetAuditLogs(c *gin.Context) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	c.JSON(http.StatusOK, models.ApiResponse{
		Success: true,
		Data:    h.auditLogs,
	})
}

// HandleViasocketWebhook processes incoming WhatsApp/SMS webhook from viasocket
func (h *Handler) HandleViasocketWebhook(c *gin.Context) {
	var payload models.ViasocketPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, models.ApiResponse{
			Success: false,
			Error:   "Invalid webhook payload: " + err.Error(),
		})
		return
	}

	provider := models.ProviderWhatsApp
	if payload.Provider == "SMS" {
		provider = models.ProviderSMS
	} else if payload.Provider == "WebPortal" {
		provider = models.ProviderWeb
	}

	signal := models.CitizenSignal{
		ID:          "sig-" + uuid.New().String()[:8],
		Provider:    provider,
		RawText:     payload.Body,
		Language:    "auto-detected",
		SenderPhone: payload.Sender,
		Timestamp:   time.Now(),
		Metadata:    payload.Metadata,
	}

	h.mutex.Lock()
	h.signals = append([]models.CitizenSignal{signal}, h.signals...)
	if len(h.signals) > 50 {
		h.signals = h.signals[:50]
	}
	h.mutex.Unlock()

	// Broadcast live signal to connected Next.js dashboard
	h.hub.Broadcast("SIGNAL_RECEIVED", signal)

	c.JSON(http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Signal ingested and broadcasted successfully",
		Data: gin.H{
			"signal_id": signal.ID,
			"timestamp": signal.Timestamp,
		},
	})
}

// RecordDecision handles policymaker action (Accept, Reject, Investigate)
func (h *Handler) RecordDecision(c *gin.Context) {
	var req models.DecisionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ApiResponse{
			Success: false,
			Error:   "Invalid decision request: " + err.Error(),
		})
		return
	}

	h.mutex.Lock()
	defer h.mutex.Unlock()

	var updatedCluster *models.Cluster
	for i := range h.clusters {
		if h.clusters[i].ID == req.ClusterID {
			h.clusters[i].Status = req.Action
			h.clusters[i].UpdatedAt = time.Now()
			updatedCluster = &h.clusters[i]
			break
		}
	}

	if updatedCluster == nil {
		c.JSON(http.StatusNotFound, models.ApiResponse{
			Success: false,
			Error:   "Cluster not found",
		})
		return
	}

	audit := models.AuditLog{
		ID:        "audit-" + uuid.New().String()[:8],
		ClusterID: req.ClusterID,
		Action:    req.Action,
		Officer:   req.Officer,
		Notes:     req.Notes,
		Timestamp: time.Now(),
	}
	h.auditLogs = append([]models.AuditLog{audit}, h.auditLogs...)

	// Broadcast decision event
	h.hub.Broadcast("DECISION_RECORDED", gin.H{
		"cluster": updatedCluster,
		"audit":   audit,
	})

	c.JSON(http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Policymaker decision recorded in audit log",
		Data: gin.H{
			"cluster_id": req.ClusterID,
			"status":     req.Action,
			"audit_id":   audit.ID,
		},
	})
}

// SimulateLiveMessage triggers a live simulated incoming citizen signal for testing
func (h *Handler) SimulateLiveMessage(c *gin.Context) {
	sampleMessages := []struct {
		Text     string
		Location string
		Phone    string
		Provider models.ProviderType
	}{
		{
			Text:     "Emergency: Live electric cable snapped near Rajwada market bus stop!",
			Location: "Rajwada & Sarafa",
			Phone:    "+91-9893012345",
			Provider: models.ProviderWhatsApp,
		},
		{
			Text:     "पानी की मेन पाइपलाइन फूट गई है बाणगंगा मेन रोड पर पूरा पानी बह रहा है",
			Location: "Banganga",
			Phone:    "+91-9827054321",
			Provider: models.ProviderWhatsApp,
		},
		{
			Text:     "Broken streetlight pole fallen across street 12 near Annapurna temple",
			Location: "Annapurna",
			Phone:    "+91-9425098765",
			Provider: models.ProviderSMS,
		},
	}

	idx := int(time.Now().UnixNano()) % len(sampleMessages)
	sample := sampleMessages[idx]

	signal := models.CitizenSignal{
		ID:           "sim-" + uuid.New().String()[:8],
		Provider:     sample.Provider,
		RawText:      sample.Text,
		Language:     "Hindi/English",
		LocationHint: sample.Location,
		SenderPhone:  sample.Phone,
		Timestamp:    time.Now(),
	}

	h.mutex.Lock()
	h.signals = append([]models.CitizenSignal{signal}, h.signals...)
	h.mutex.Unlock()

	// Broadcast live message event to frontend
	h.hub.Broadcast("SIGNAL_RECEIVED", signal)

	c.JSON(http.StatusOK, models.ApiResponse{
		Success: true,
		Message: fmt.Sprintf("Simulated live %s message generated", sample.Provider),
		Data:    signal,
	})
}
