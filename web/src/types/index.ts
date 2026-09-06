export type ProviderType = "WhatsApp" | "SMS" | "WebPortal";

export type UrgencyTier = 
  | "TIER_1_CRITICAL" 
  | "TIER_2_HIGH" 
  | "TIER_3_MEDIUM" 
  | "TIER_4_ROUTINE";

export interface UrgencyResult {
  score: number;
  tier: UrgencyTier;
  sla_hours: number;
  factors: string[];
  hazard_boost: number;
  velocity_rate: number;
}

export interface FourDimensionalScore {
  need: number;
  confidence: number;
  equity: number;
  actionability: number;
}

export interface CitizenSignal {
  id: string;
  provider: ProviderType;
  raw_text: string;
  language: string;
  translated_text?: string;
  location_hint?: string;
  sender_phone?: string;
  timestamp: string;
  metadata?: Record<string, any>;
}

export interface Cluster {
  id: string;
  ward_id: string;
  ward_name: string;
  department: string;
  title: string;
  description: string;
  signal_count: number;
  signal_ids: string[];
  channels: string[];
  urgency: UrgencyResult;
  scores: FourDimensionalScore;
  recommendation: string;
  is_blind_spot: boolean;
  status: "PENDING" | "ACCEPTED" | "INVESTIGATING" | "REJECTED";
  centroid_lat: number;
  centroid_lng: number;
  created_at: string;
  updated_at: string;
}

export interface Ward {
  id: string;
  name: string;
  zone: string;
  lat: number;
  lng: number;
  infra_index: number;
  population: number;
  historical_spend_cr: number;
  critical_facilities: string[];
  active_cluster_count: number;
  is_blind_spot: boolean;
}

export interface AuditLog {
  id: string;
  cluster_id: string;
  action: string;
  officer: string;
  notes: string;
  timestamp: string;
}

export interface ApiResponse<T> {
  success: boolean;
  message?: string;
  data?: T;
  error?: string;
}
