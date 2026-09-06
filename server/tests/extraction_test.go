package tests

import (
	"context"
	"testing"

	"walkover/server/internal/extraction"
	"walkover/server/internal/models"
)

func TestExtractSignal_PureHindiScenarios(t *testing.T) {
	ctx := context.Background()
	ext, err := extraction.NewExtractor(ctx, "", "gemini-3.6-flash", "gemini-embedding-001")
	if err != nil {
		t.Fatalf("Failed to initialize extractor: %v", err)
	}
	defer ext.Close()

	testCases := []struct {
		name               string
		rawText            string
		expectedWardID     string
		expectedDepartment string
		expectedHazard     string
		expectedUrgency    int
	}{
		{
			name:               "Chandan Nagar Tap Water Contamination",
			rawText:            "चंदन नगर में पीने के पानी में सीवेज का बदबूदार गंदा पानी मिल कर आ रहा है",
			expectedWardID:     "indore-ward-02",
			expectedDepartment: "Water Supply & Sewerage",
			expectedHazard:     "CONTAMINATED_WATER",
			expectedUrgency:    5,
		},
		{
			name:               "Rajwada Live Wire Sparking",
			rawText:            "राजवाड़ा सराफा बाजार में बिजली का नंगा तार गिर गया है स्पार्क हो रहा है",
			expectedWardID:     "indore-ward-05",
			expectedDepartment: "Electricity & Power",
			expectedHazard:     "LIVE_WIRE",
			expectedUrgency:    5,
		},
		{
			name:               "Annapurna Broken Sewer Cover",
			rawText:            "अन्नपूर्णा मंदिर रोड पर सीवर का ढक्कन टूटा हुआ है और गंदा पानी बह रहा है",
			expectedWardID:     "indore-ward-07",
			expectedDepartment: "Water Supply & Sewerage",
			expectedHazard:     "OPEN_MANHOLE",
			expectedUrgency:    4,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			signal := models.CitizenSignal{
				ID:       "sig-hindi",
				Provider: models.ProviderWhatsApp,
				RawText:  tc.rawText,
			}

			result, err := ext.ExtractSignal(ctx, signal)
			if err != nil {
				t.Fatalf("ExtractSignal failed: %v", err)
			}

			if result.WardID != tc.expectedWardID {
				t.Errorf("Expected ward ID %s, got %s", tc.expectedWardID, result.WardID)
			}
			if result.Department != tc.expectedDepartment {
				t.Errorf("Expected department %s, got %s", tc.expectedDepartment, result.Department)
			}
			if result.BaseUrgency < tc.expectedUrgency {
				t.Errorf("Expected base urgency >= %d, got %d", tc.expectedUrgency, result.BaseUrgency)
			}

			hasHazard := false
			for _, h := range result.HazardTags {
				if h == tc.expectedHazard {
					hasHazard = true
					break
				}
			}
			if !hasHazard {
				t.Errorf("Missing expected hazard %s. Got: %v", tc.expectedHazard, result.HazardTags)
			}
		})
	}
}

func TestExtractSignal_HinglishScenarios(t *testing.T) {
	ctx := context.Background()
	ext, err := extraction.NewExtractor(ctx, "", "gemini-3.6-flash", "gemini-embedding-001")
	if err != nil {
		t.Fatalf("Failed to initialize extractor: %v", err)
	}
	defer ext.Close()

	testCases := []struct {
		name               string
		rawText            string
		expectedWardID     string
		expectedDepartment string
		expectedHazard     string
		expectedUrgency    int
	}{
		{
			name:               "Vijay Nagar Hospital Route Cave-In",
			rawText:            "Vijay Nagar hospital route par road cave-in sinkhole ho gaya hai ambulance nahi nikal rahi",
			expectedWardID:     "indore-ward-03",
			expectedDepartment: "Public Works / Roads",
			expectedHazard:     "HOSPITAL_ROUTE_BLOCKED",
			expectedUrgency:    5,
		},
		{
			name:               "Sukhliya Pipeline Burst",
			rawText:            "Sukhliya MR-10 metro pillar area me drinking water pipeline burst ho gayi hai",
			expectedWardID:     "indore-ward-10",
			expectedDepartment: "Water Supply & Sewerage",
			expectedHazard:     "CONTAMINATED_WATER",
			expectedUrgency:    3,
		},
		{
			name:               "Rau Bypass Colony Drainage Choked",
			rawText:            "Rau bypass colony entry par main drainage chocked ho gaya hai gharo ke andar pani ghus raha hai",
			expectedWardID:     "indore-ward-12",
			expectedDepartment: "Water Supply & Sewerage",
			expectedHazard:     "DRAINAGE_OVERFLOW",
			expectedUrgency:    4,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			signal := models.CitizenSignal{
				ID:       "sig-hinglish",
				Provider: models.ProviderWhatsApp,
				RawText:  tc.rawText,
			}

			result, err := ext.ExtractSignal(ctx, signal)
			if err != nil {
				t.Fatalf("ExtractSignal failed: %v", err)
			}

			if result.WardID != tc.expectedWardID {
				t.Errorf("Expected ward ID %s, got %s", tc.expectedWardID, result.WardID)
			}
			if result.Department != tc.expectedDepartment {
				t.Errorf("Expected department %s, got %s", tc.expectedDepartment, result.Department)
			}
			if result.BaseUrgency < tc.expectedUrgency {
				t.Errorf("Expected base urgency >= %d, got %d", tc.expectedUrgency, result.BaseUrgency)
			}

			hasHazard := false
			for _, h := range result.HazardTags {
				if h == tc.expectedHazard {
					hasHazard = true
					break
				}
			}
			if !hasHazard {
				t.Errorf("Missing expected hazard %s. Got: %v", tc.expectedHazard, result.HazardTags)
			}
		})
	}
}

func TestExtractSignal_UrgencyAndIntentCorrelation(t *testing.T) {
	ctx := context.Background()
	ext, err := extraction.NewExtractor(ctx, "", "gemini-3.6-flash", "gemini-embedding-001")
	if err != nil {
		t.Fatalf("Failed to initialize extractor: %v", err)
	}
	defer ext.Close()

	// High urgency emergency
	emergencySignal := models.CitizenSignal{
		ID:       "sig-emerg",
		Provider: models.ProviderWhatsApp,
		RawText:  "Live sparking wire fell on crowded street in Rajwada emergency help needed",
	}
	emergRes, _ := ext.ExtractSignal(ctx, emergencySignal)
	if emergRes.BaseUrgency < 4 {
		t.Errorf("Expected emergency urgency >= 4, got %d", emergRes.BaseUrgency)
	}
	if emergRes.Intent != "emergency_report" {
		t.Errorf("Expected emergency intent 'emergency_report', got %s", emergRes.Intent)
	}

	// Routine maintenance complaint
	routineSignal := models.CitizenSignal{
		ID:       "sig-routine",
		Provider: models.ProviderSMS,
		RawText:  "Streetlight in Old Palasia not working for 2 nights",
	}
	routineRes, _ := ext.ExtractSignal(ctx, routineSignal)
	if routineRes.BaseUrgency > 3 {
		t.Errorf("Expected routine urgency <= 3, got %d", routineRes.BaseUrgency)
	}
	if routineRes.Intent != "complaint" {
		t.Errorf("Expected routine intent 'complaint', got %s", routineRes.Intent)
	}
}

func TestExtractSignal_AllIndoreWardsResolution(t *testing.T) {
	ctx := context.Background()
	ext, err := extraction.NewExtractor(ctx, "", "gemini-3.6-flash", "gemini-embedding-001")
	if err != nil {
		t.Fatalf("Failed to initialize extractor: %v", err)
	}
	defer ext.Close()

	wardsCatalog := []struct {
		landmark string
		expected string
	}{
		{"Laxmibai nagar station in Banganga", "indore-ward-01"},
		{"Dhar road near Chandan Nagar clinic", "indore-ward-02"},
		{"BRTS square in Vijay Nagar", "indore-ward-03"},
		{"Industry house in Old Palasia", "indore-ward-04"},
		{"Heritage market in Rajwada and Sarafa", "indore-ward-05"},
		{"DAVV campus in Bhawarkua", "indore-ward-06"},
		{"Annapurna temple road", "indore-ward-07"},
		{"Phooti kothi near Sudama Nagar", "indore-ward-08"},
		{"Khajrana shrine transit gate", "indore-ward-09"},
		{"MR-10 metro pillar in Sukhliya", "indore-ward-10"},
		{"Grain mandi in Malharganj", "indore-ward-11"},
		{"Silicon city at Rau bypass corridor", "indore-ward-12"},
	}

	for _, w := range wardsCatalog {
		sig := models.CitizenSignal{
			ID:       "sig-ward",
			Provider: models.ProviderWhatsApp,
			RawText:  "General garbage cleaning required at " + w.landmark,
		}
		res, err := ext.ExtractSignal(ctx, sig)
		if err != nil {
			t.Fatalf("Failed to resolve ward for %s: %v", w.landmark, err)
		}
		if res.WardID != w.expected {
			t.Errorf("Failed to resolve landmark %q: expected %s, got %s", w.landmark, w.expected, res.WardID)
		}
	}
}
