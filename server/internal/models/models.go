package models

import (
	"time"
)

// ProviderType represents the incoming channel
type ProviderType string

const (
	ProviderWhatsApp ProviderType = "WhatsApp"
	ProviderSMS      ProviderType = "SMS"
	ProviderWeb      ProviderType = "WebPortal"
	ProviderTelegram ProviderType = "Telegram"
)

// IssueCategory is the stable, cross-channel taxonomy used by clustering and
// the dashboard. Department names remain useful for routing, but they are too
// broad to decide whether two reports describe the same problem.
type IssueCategory string

const (
	CategoryWater        IssueCategory = "Water"
	CategoryRoads        IssueCategory = "Roads"
	CategoryTransport    IssueCategory = "Transport"
	CategorySanitation   IssueCategory = "Sanitation"
	CategoryElectricity  IssueCategory = "Electricity"
	CategoryPublicHealth IssueCategory = "Public Health"
	CategoryFire         IssueCategory = "Fire/Emergency"
	CategoryOther        IssueCategory = "Other Civic"
)

// UrgencyTier represents the classified urgency tier
type UrgencyTier string

const (
	Tier1Critical UrgencyTier = "TIER_1_CRITICAL" // < 4 hours SLA
	Tier2High     UrgencyTier = "TIER_2_HIGH"     // < 24 hours SLA
	Tier3Medium   UrgencyTier = "TIER_3_MEDIUM"   // < 72 hours SLA
	Tier4Routine  UrgencyTier = "TIER_4_ROUTINE"  // < 7 days SLA
)

// ViasocketPayload represents incoming normalized webhook from viasocket
type ViasocketPayload struct {
	EventID   string                 `json:"event_id,omitempty"`
	Provider  string                 `json:"provider"`
	Sender    string                 `json:"sender,omitempty"`
	Body      string                 `json:"body"`
	MediaURL  string                 `json:"media_url,omitempty"`
	Timestamp string                 `json:"timestamp,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`

	// Twilio names its fields in PascalCase. Go's JSON decoder matches field
	// names case-insensitively, so "Body" already lands in Body — but "From"
	// and "WaId" have no lowercase counterpart above and would be dropped,
	// leaving the sender empty and the citizen unreachable for a reply.
	From string `json:"From,omitempty"`
	WaId string `json:"WaId,omitempty"`

	// Some viasocket flows forward the original request wrapped as
	// {"data": {...}} rather than at the top level.
	Data *ViasocketPayload `json:"data,omitempty"`
}

// CitizenSignal represents the canonical signal standard across all channels
type CitizenSignal struct {
	ID             string                 `json:"id" firestore:"id"`
	Provider       ProviderType           `json:"provider" firestore:"provider"`
	RawText        string                 `json:"raw_text" firestore:"raw_text"`
	Language       string                 `json:"language" firestore:"language"`
	TranslatedText string                 `json:"translated_text,omitempty" firestore:"translated_text,omitempty"`
	LocationHint   string                 `json:"location_hint,omitempty" firestore:"location_hint,omitempty"`
	SenderPhone    string                 `json:"sender_phone,omitempty" firestore:"sender_phone,omitempty"`
	Timestamp      time.Time              `json:"timestamp" firestore:"timestamp"`
	Metadata       map[string]interface{} `json:"metadata,omitempty" firestore:"metadata,omitempty"`
}

// AIExtraction represents structured output extracted by Gemini
type AIExtraction struct {
	SignalID           string        `json:"signal_id" firestore:"signal_id"`
	Issue              string        `json:"issue" firestore:"issue"`
	IssueCategory      IssueCategory `json:"issue_category" firestore:"issue_category"`
	WardID             string        `json:"ward_id" firestore:"ward_id"`
	WardName           string        `json:"ward_name" firestore:"ward_name"`
	LocationSource     string        `json:"location_source" firestore:"location_source"`
	LocationConfidence float64       `json:"location_confidence" firestore:"location_confidence"`
	LocationRationale  string        `json:"location_rationale" firestore:"location_rationale"`
	Department         string        `json:"department" firestore:"department"`
	BaseUrgency        int           `json:"base_urgency" firestore:"base_urgency"` // 1 - 5
	HazardTags         []string      `json:"hazard_tags" firestore:"hazard_tags"`
	Intent             string        `json:"intent" firestore:"intent"`
	Summary            string        `json:"summary" firestore:"summary"`
	DetectedLanguage   string        `json:"detected_language" firestore:"detected_language"`
	TranslatedText     string        `json:"translated_text,omitempty" firestore:"translated_text,omitempty"`
	AnalysisSource     string        `json:"analysis_source" firestore:"analysis_source"`
	Embedding          []float32     `json:"embedding,omitempty" firestore:"embedding,omitempty"`
	ConfidenceScore    float64       `json:"confidence_score" firestore:"confidence_score"`
	CreatedAt          time.Time     `json:"created_at" firestore:"created_at"`
}

// SignalFeedItem is the public, privacy-safe view of a verified report. It
// intentionally excludes sender identifiers and provider metadata.
type SignalFeedItem struct {
	ID                 string        `json:"id"`
	Provider           ProviderType  `json:"provider"`
	RawText            string        `json:"raw_text"`
	Language           string        `json:"language"`
	TranslatedText     string        `json:"translated_text,omitempty"`
	Timestamp          time.Time     `json:"timestamp"`
	Status             string        `json:"status"`
	Issue              string        `json:"issue"`
	IssueCategory      IssueCategory `json:"issue_category"`
	Department         string        `json:"department"`
	WardID             string        `json:"ward_id"`
	WardName           string        `json:"ward_name"`
	LocationSource     string        `json:"location_source"`
	LocationConfidence float64       `json:"location_confidence"`
	LocationRationale  string        `json:"location_rationale"`
	BaseUrgency        int           `json:"base_urgency"`
	Severity           string        `json:"severity"`
	HazardTags         []string      `json:"hazard_tags"`
	Summary            string        `json:"summary"`
	AIConfidence       float64       `json:"ai_confidence"`
	AnalysisSource     string        `json:"analysis_source"`
	ClusterID          string        `json:"cluster_id"`
	ClusterTitle       string        `json:"cluster_title"`
	ClusterStatus      string        `json:"cluster_status"`
	ClusterUrgency     UrgencyResult `json:"cluster_urgency"`
	ClusterSignalCount int           `json:"cluster_signal_count"`
}

// UrgencyResult contains multi-factor scoring and SLA breakdown
type UrgencyResult struct {
	Score        float64     `json:"score"` // 0 - 100
	Tier         UrgencyTier `json:"tier"`
	SLAHours     int         `json:"sla_hours"`
	Factors      []string    `json:"factors"`
	HazardBoost  float64     `json:"hazard_boost"`
	VelocityRate float64     `json:"velocity_rate"`
}

// FourDimensionalScore holds the transparent 4 priority scores (0 - 100)
type FourDimensionalScore struct {
	Need          float64 `json:"need"`
	Confidence    float64 `json:"confidence"`
	Equity        float64 `json:"equity"`
	Actionability float64 `json:"actionability"`
}

// Cluster represents a grouped set of related citizen reports in a ward
type Cluster struct {
	ID             string               `json:"id" firestore:"id"`
	WardID         string               `json:"ward_id" firestore:"ward_id"`
	WardName       string               `json:"ward_name" firestore:"ward_name"`
	Department     string               `json:"department" firestore:"department"`
	Title          string               `json:"title" firestore:"title"`
	Description    string               `json:"description" firestore:"description"`
	SignalCount    int                  `json:"signal_count" firestore:"signal_count"`
	SignalIDs      []string             `json:"signal_ids" firestore:"signal_ids"`
	Channels       []string             `json:"channels" firestore:"channels"`
	Urgency        UrgencyResult        `json:"urgency" firestore:"urgency"`
	Scores         FourDimensionalScore `json:"scores" firestore:"scores"`
	Recommendation string               `json:"recommendation" firestore:"recommendation"`
	IsBlindSpot    bool                 `json:"is_blind_spot" firestore:"is_blind_spot"`
	Status         string               `json:"status" firestore:"status"` // PENDING, ACCEPTED, INVESTIGATING, REJECTED
	CentroidLat    float64              `json:"centroid_lat" firestore:"centroid_lat"`
	CentroidLng    float64              `json:"centroid_lng" firestore:"centroid_lng"`
	CreatedAt      time.Time            `json:"created_at" firestore:"created_at"`
	UpdatedAt      time.Time            `json:"updated_at" firestore:"updated_at"`
}

// Ward represents an Indore ward reference profile
type Ward struct {
	ID                 string   `json:"id" firestore:"id"`
	Name               string   `json:"name" firestore:"name"`
	Zone               string   `json:"zone" firestore:"zone"`
	Lat                float64  `json:"lat" firestore:"lat"`
	Lng                float64  `json:"lng" firestore:"lng"`
	InfraIndex         float64  `json:"infra_index" firestore:"infra_index"` // 0.0 - 1.0 (Lower = worse infra/more needy)
	Population         int      `json:"population" firestore:"population"`
	HistoricalSpend    float64  `json:"historical_spend_cr" firestore:"historical_spend_cr"`
	CriticalFacilities []string `json:"critical_facilities" firestore:"critical_facilities"`
	ActiveClusterCount int      `json:"active_cluster_count" firestore:"active_cluster_count"`
	IsBlindSpot        bool     `json:"is_blind_spot" firestore:"is_blind_spot"`
}

// DecisionRequest is the body sent by policymakers when taking action
type DecisionRequest struct {
	ClusterID string `json:"cluster_id" binding:"required"`
	Action    string `json:"action" binding:"required"` // ACCEPT, REJECT, INVESTIGATE
	Officer   string `json:"officer" binding:"required"`
	Notes     string `json:"notes"`
}

// AuditLog records policymaker actions immutably
type AuditLog struct {
	ID        string    `json:"id" firestore:"id"`
	ClusterID string    `json:"cluster_id" firestore:"cluster_id"`
	Action    string    `json:"action" firestore:"action"`
	Officer   string    `json:"officer" firestore:"officer"`
	Notes     string    `json:"notes" firestore:"notes"`
	Timestamp time.Time `json:"timestamp" firestore:"timestamp"`
}

// WebSocketMessage represents real-time events sent to frontend
type WebSocketMessage struct {
	Type      string      `json:"type"` // SIGNAL_RECEIVED, CLUSTER_UPDATED, DECISION_RECORDED
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// ApiResponse standard wrapper for REST endpoints
type ApiResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}
