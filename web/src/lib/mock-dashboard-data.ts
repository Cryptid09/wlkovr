import { CitizenSignal, Cluster, SignalFeedItem, Ward } from "@/types";

const now = new Date();
const ago = (minutes: number) => new Date(now.getTime() - minutes * 60_000).toISOString();

export const mockWards: Ward[] = [
  { id: "indore-ward-01", name: "Ward 1 - Banganga", zone: "Zone 1 (North)", lat: 22.7533, lng: 75.8577, infra_index: 0.38, population: 42000, historical_spend_cr: 2.4, critical_facilities: ["Primary Health Center", "Govt Girls High School"], active_cluster_count: 0, is_blind_spot: true },
  { id: "indore-ward-02", name: "Ward 14 - Chandan Nagar", zone: "Zone 14 (West)", lat: 22.7092, lng: 75.8236, infra_index: 0.32, population: 58000, historical_spend_cr: 1.8, critical_facilities: ["Community Clinic", "District Primary School"], active_cluster_count: 1, is_blind_spot: false },
  { id: "indore-ward-03", name: "Ward 22 - Vijay Nagar", zone: "Zone 7 (East)", lat: 22.7533, lng: 75.8937, infra_index: 0.84, population: 65000, historical_spend_cr: 12.5, critical_facilities: ["Mother & Child Care Hospital", "BRTS Main Hub"], active_cluster_count: 1, is_blind_spot: false },
  { id: "indore-ward-09", name: "Ward 60 - Khajrana", zone: "Zone 6 (East)", lat: 22.7314, lng: 75.9083, infra_index: 0.44, population: 72000, historical_spend_cr: 4.1, critical_facilities: ["Maternity Clinic", "High-density Slum Cluster"], active_cluster_count: 1, is_blind_spot: false },
  { id: "indore-ward-12", name: "Ward 78 - Rau & Bypass Corridor", zone: "Zone 13 (Outer South)", lat: 22.6394, lng: 75.8038, infra_index: 0.35, population: 39000, historical_spend_cr: 2.1, critical_facilities: ["Rural Influx Transit Hub", "Govt Primary School"], active_cluster_count: 0, is_blind_spot: true },
  { id: "indore-ward-04", name: "Ward 28 - Old Palasia", zone: "Zone 8 (Central-East)", lat: 22.7244, lng: 75.8839, infra_index: 0.88, population: 38000, historical_spend_cr: 9.2, critical_facilities: ["District Hospital Trauma Wing", "Major Intersection"], active_cluster_count: 0, is_blind_spot: false },
  { id: "indore-ward-05", name: "Ward 35 - Rajwada & Sarafa", zone: "Zone 2 (Central Heritage)", lat: 22.7187, lng: 75.8558, infra_index: 0.62, population: 49000, historical_spend_cr: 7.1, critical_facilities: ["Heritage Market", "Fire Station Substation"], active_cluster_count: 0, is_blind_spot: false },
  { id: "indore-ward-06", name: "Ward 42 - Bhawarkua & Vishnupuri", zone: "Zone 9 (South)", lat: 22.6926, lng: 75.8676, infra_index: 0.72, population: 52000, historical_spend_cr: 8.0, critical_facilities: ["DAVV University Campus", "City Bus Terminal"], active_cluster_count: 0, is_blind_spot: false },
  { id: "indore-ward-07", name: "Ward 49 - Annapurna", zone: "Zone 11 (South-West)", lat: 22.6989, lng: 75.8384, infra_index: 0.76, population: 46000, historical_spend_cr: 6.8, critical_facilities: ["Annapurna Temple Zone", "Senior Secondary School"], active_cluster_count: 0, is_blind_spot: false },
  { id: "indore-ward-08", name: "Ward 55 - Sudama Nagar", zone: "Zone 12 (West)", lat: 22.7056, lng: 75.8309, infra_index: 0.58, population: 61000, historical_spend_cr: 5.4, critical_facilities: ["Secondary Drainage Canal", "Community Health Center"], active_cluster_count: 0, is_blind_spot: false },
  { id: "indore-ward-10", name: "Ward 64 - Sukhliya", zone: "Zone 4 (North-East)", lat: 22.7601, lng: 75.8752, infra_index: 0.65, population: 54000, historical_spend_cr: 6.0, critical_facilities: ["MR-10 Metro Pillar Area", "Sub-health center"], active_cluster_count: 0, is_blind_spot: false },
  { id: "indore-ward-11", name: "Ward 71 - Malharganj", zone: "Zone 3 (North-West)", lat: 22.7299, lng: 75.8451, infra_index: 0.54, population: 43000, historical_spend_cr: 4.8, critical_facilities: ["Old Wholesale Grain Mandi", "District Dispensary"], active_cluster_count: 0, is_blind_spot: false },
];

export const mockClusters: Cluster[] = [
  { id: "cluster-indore-001", ward_id: "indore-ward-02", ward_name: "Ward 14 - Chandan Nagar", department: "Water Supply & Sewerage", title: "Sewage contamination in drinking line", description: "Multiple reports describe brown, foul-smelling tap water and sewage backflow near the Street 4 community clinic.", signal_count: 8, signal_ids: ["sig-01", "sig-02"], channels: ["WhatsApp", "SMS"], urgency: { score: 96, tier: "TIER_1_CRITICAL", sla_hours: 4, factors: ["Contaminated drinking water", "Sewage mixing", "Community clinic nearby"], hazard_boost: 2, velocity_rate: 8 }, scores: { need: 96, confidence: 72, equity: 68, actionability: 100 }, recommendation: "Verify the water line and sewage crossover immediately; dispatch water-quality testing and an emergency repair inspection before recommending a capital intervention.", is_blind_spot: false, status: "PENDING", centroid_lat: 22.7092, centroid_lng: 75.8236, created_at: ago(240), updated_at: ago(15) },
  { id: "cluster-indore-002", ward_id: "indore-ward-03", ward_name: "Ward 22 - Vijay Nagar", department: "Public Works Department", title: "Hospital approach road cave-in", description: "A road-surface collapse near the BRTS intersection is obstructing ambulance access to the hospital corridor.", signal_count: 12, signal_ids: ["sig-11", "sig-12"], channels: ["WhatsApp", "SMS", "WebPortal"], urgency: { score: 82, tier: "TIER_1_CRITICAL", sla_hours: 4, factors: ["Ambulance route blocked", "Deep road cave-in", "High signal velocity"], hazard_boost: 1.9, velocity_rate: 12 }, scores: { need: 92, confidence: 91, equity: 16, actionability: 100 }, recommendation: "Secure the affected road section and verify the temporary ambulance diversion with the hospital and PWD field team.", is_blind_spot: false, status: "INVESTIGATING", centroid_lat: 22.7533, centroid_lng: 75.8937, created_at: ago(600), updated_at: ago(65) },
  { id: "cluster-indore-003", ward_id: "indore-ward-09", ward_name: "Ward 60 - Khajrana", department: "Sanitation & Drainage", title: "Open manhole on school path", description: "A drainage cover is broken after monsoon runoff, creating a serious pedestrian hazard for children and nearby residents.", signal_count: 6, signal_ids: ["sig-21", "sig-22"], channels: ["WhatsApp"], urgency: { score: 74, tier: "TIER_2_HIGH", sla_hours: 24, factors: ["Open drainage manhole", "School walking route", "High-density settlement"], hazard_boost: 1.8, velocity_rate: 6 }, scores: { need: 71, confidence: 52, equity: 56, actionability: 100 }, recommendation: "Install a temporary barricade and cover, then verify whether the damaged drain needs a permanent replacement.", is_blind_spot: false, status: "PENDING", centroid_lat: 22.7314, centroid_lng: 75.9083, created_at: ago(1440), updated_at: ago(120) },
];

export const mockSignals: CitizenSignal[] = [
  { id: "sig-01", provider: "WhatsApp", raw_text: "हमारे यहाँ नल से गंदा पानी आ रहा है, बहुत बदबू है।", language: "Hindi", location_hint: "Chandan Nagar, Street 4", timestamp: ago(15) },
  { id: "sig-02", provider: "WhatsApp", raw_text: "Urgent: sewage is mixing with the drinking water line near the clinic.", language: "English", location_hint: "Chandan Nagar clinic", timestamp: ago(24) },
  { id: "sig-11", provider: "SMS", raw_text: "Vijay Nagar square ke paas ambulance route par road dhas gayi hai.", language: "Hinglish", location_hint: "Vijay Nagar Square", timestamp: ago(40) },
];

export const mockFeed: SignalFeedItem[] = mockSignals.map((signal, index) => {
  const cluster = mockClusters[index];
  const categories = ["Water", "Water", "Transport"] as const;
  return {
    id: signal.id,
    provider: signal.provider,
    raw_text: signal.raw_text,
    language: signal.language,
    timestamp: signal.timestamp,
    status: "VERIFIED",
    issue: cluster.title,
    issue_category: categories[index],
    department: cluster.department,
    ward_id: cluster.ward_id,
    ward_name: cluster.ward_name,
    location_source: "explicit",
    location_confidence: 0.96,
    location_rationale: `Matched ${signal.location_hint}`,
    base_urgency: index < 2 ? 5 : 4,
    severity: index < 2 ? "CRITICAL" : "HIGH",
    hazard_tags: index < 2 ? ["CONTAMINATED_WATER"] : ["HOSPITAL_ROUTE_BLOCKED"],
    summary: cluster.description,
    ai_confidence: 0.94,
    analysis_source: "gemini",
    cluster_id: cluster.id,
    cluster_title: cluster.title,
    cluster_status: cluster.status,
    cluster_urgency: cluster.urgency,
    cluster_signal_count: cluster.signal_count,
  };
});
