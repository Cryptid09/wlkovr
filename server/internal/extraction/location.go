package extraction

import (
	"regexp"
	"strings"
	"unicode"

	"walkover/server/internal/models"
)

type wardReference struct {
	ID      string
	Name    string
	Aliases []string
}

type wardResolution struct {
	ID         string
	Name       string
	Source     string
	Confidence float64
	Rationale  string
}

var wardCatalog = []wardReference{
	{"indore-ward-01", "Ward 1 - Banganga", []string{"banganga", "बाणगंगा", "laxmibai nagar", "laxmibai station"}},
	{"indore-ward-02", "Ward 14 - Chandan Nagar", []string{"chandan nagar", "chandan nagar", "चंदन नगर", "dhar road", "community clinic"}},
	{"indore-ward-03", "Ward 22 - Vijay Nagar", []string{"vijay nagar", "vijaynagar", "विजय नगर", "brts hub", "mother child hospital"}},
	{"indore-ward-04", "Ward 28 - Old Palasia", []string{"old palasia", "palasia", "पलासिया", "industry house", "trauma wing"}},
	{"indore-ward-05", "Ward 35 - Rajwada & Sarafa", []string{"rajwada", "राजवाड़ा", "sarafa", "सराफा", "bartan bazar", "heritage market"}},
	{"indore-ward-06", "Ward 42 - Bhawarkua & Vishnupuri", []string{"bhawarkua", "bhanwarkua", "भंवरकुआ", "vishnupuri", "davv", "rajiv gandhi square"}},
	{"indore-ward-07", "Ward 49 - Annapurna", []string{"annapurna", "अन्नपूर्णा", "ranjeet hanuman"}},
	{"indore-ward-08", "Ward 55 - Sudama Nagar", []string{"sudama nagar", "sudama", "सुदामा", "phooti kothi"}},
	{"indore-ward-09", "Ward 60 - Khajrana", []string{"khajrana", "khajrane", "khajarana", "खजराना", "khajrana mandir", "khajrana temple", "kalka mata"}},
	{"indore-ward-10", "Ward 64 - Sukhliya", []string{"sukhliya", "सुखलिया", "mr10", "mr 10", "bapat square"}},
	{"indore-ward-11", "Ward 71 - Malharganj", []string{"malharganj", "मल्हारगंज", "grain mandi"}},
	{"indore-ward-12", "Ward 78 - Rau & Bypass Corridor", []string{"rau", "राऊ", "silicon city", "bypass junction", "bypass corridor"}},
}

var nonWord = regexp.MustCompile(`[^\p{L}\p{N}]+`)

func normalizedLocationText(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || unicode.IsSpace(r) {
			return r
		}
		return ' '
	}, value)
	return strings.Join(strings.Fields(nonWord.ReplaceAllString(value, " ")), " ")
}

func resolveWard(message, hint, modelID, modelName string) wardResolution {
	if ward, alias := matchWardAlias(hint); ward != nil {
		return wardResolution{ward.ID, ward.Name, "explicit", 0.98, "Matched the supplied location hint: " + alias}
	}
	if ward, alias := matchWardAlias(message); ward != nil {
		source := "alias"
		if strings.Contains(alias, "hospital") || strings.Contains(alias, "temple") || strings.Contains(alias, "mandir") || strings.Contains(alias, "square") || strings.Contains(alias, "clinic") || strings.Contains(alias, "davv") {
			source = "landmark"
		}
		return wardResolution{ward.ID, ward.Name, source, 0.9, "Matched message location reference: " + alias}
	}
	for _, ward := range wardCatalog {
		if modelID == ward.ID || strings.EqualFold(strings.TrimSpace(modelName), ward.Name) {
			return wardResolution{ward.ID, ward.Name, "model_inferred", 0.55, "Sahyog-Ai selected the closest ward from the approved catalog"}
		}
	}

	// Product policy requires actionable routing even without a location. Ward
	// 35 is the central civic routing area, not a claim that the incident was
	// observed there; the low confidence and rationale keep that distinction
	// visible to officials.
	return wardResolution{"indore-ward-05", "Ward 35 - Rajwada & Sarafa", "model_inferred", 0.2, "No location was stated; provisionally routed to central Indore for verification"}
}

func matchWardAlias(value string) (*wardReference, string) {
	normalized := normalizedLocationText(value)
	if normalized == "" {
		return nil, ""
	}
	for i := range wardCatalog {
		for _, alias := range wardCatalog[i].Aliases {
			if strings.Contains(normalized, normalizedLocationText(alias)) {
				return &wardCatalog[i], alias
			}
		}
	}
	return nil, ""
}

func canonicalCategory(suggested, department, issue string, hazards []string) models.IssueCategory {
	for _, category := range []models.IssueCategory{
		models.CategoryWater, models.CategoryRoads, models.CategoryTransport,
		models.CategorySanitation, models.CategoryElectricity, models.CategoryPublicHealth,
		models.CategoryFire, models.CategoryOther,
	} {
		if strings.EqualFold(strings.TrimSpace(suggested), string(category)) {
			return category
		}
	}
	text := strings.ToLower(department + " " + issue + " " + strings.Join(hazards, " "))
	switch {
	case strings.Contains(text, "fire") || strings.Contains(text, "आग"):
		return models.CategoryFire
	case strings.Contains(text, "electric") || strings.Contains(text, "wire") || strings.Contains(text, "power"):
		return models.CategoryElectricity
	case strings.Contains(text, "ambulance") || strings.Contains(text, "transport") || strings.Contains(text, "traffic") || strings.Contains(text, "brts"):
		return models.CategoryTransport
	case strings.Contains(text, "road") || strings.Contains(text, "pothole") || strings.Contains(text, "cave"):
		return models.CategoryRoads
	case strings.Contains(text, "manhole") || strings.Contains(text, "open drain") || strings.Contains(text, "sanitation") || strings.Contains(text, "waste") || strings.Contains(text, "garbage"):
		return models.CategorySanitation
	case strings.Contains(text, "water") || strings.Contains(text, "sewer") || strings.Contains(text, "drain"):
		return models.CategoryWater
	case strings.Contains(text, "health") || strings.Contains(text, "hospital"):
		return models.CategoryPublicHealth
	default:
		return models.CategoryOther
	}
}

func normalizedLanguage(candidate, text string) string {
	if value := strings.TrimSpace(candidate); value != "" {
		return value
	}
	for _, r := range text {
		if r >= '\u0900' && r <= '\u097f' {
			return "Hindi"
		}
	}
	commonHinglish := []string{" hai", " ke ", " ka ", "pani", "gaddha", "aag", "yaha", "karo"}
	lower := " " + strings.ToLower(text) + " "
	for _, token := range commonHinglish {
		if strings.Contains(lower, token) {
			return "Hinglish"
		}
	}
	return "English"
}
