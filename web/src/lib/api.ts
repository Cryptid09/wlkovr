import { ApiResponse, Cluster, Ward, CitizenSignal, SignalFeedItem } from "@/types";
import { mockClusters, mockFeed, mockSignals, mockWards } from "@/lib/mock-dashboard-data";

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1";
export const WS_BASE = process.env.NEXT_PUBLIC_WS_URL || "ws://localhost:8080/ws";
// Live backend data is the default. Mock data is opt-in via
// NEXT_PUBLIC_USE_MOCK_DATA=true, so a demo can never silently present
// fabricated clusters and scores as though they came from the platform.
const USE_MOCK_DATA = process.env.NEXT_PUBLIC_USE_MOCK_DATA === "true";
export type DataMode = "live" | "demo";
export const DEFAULT_DATA_MODE: DataMode = USE_MOCK_DATA ? "demo" : "live";

export async function fetchWards(mode: DataMode = DEFAULT_DATA_MODE): Promise<Ward[]> {
  if (mode === "demo") return mockWards;
  try {
    const res = await fetch(`${API_BASE}/wards`, { cache: "no-store" });
    const json: ApiResponse<Ward[]> = await res.json();
    if (json.success && json.data) return json.data;
    throw new Error(json.error || "Failed to fetch wards");
  } catch (err) {
    console.error("[API] Could not load wards from the backend:", err);
    return [];
  }
}

export async function fetchClusters(mode: DataMode = DEFAULT_DATA_MODE): Promise<Cluster[]> {
  if (mode === "demo") return mockClusters;
  try {
    const res = await fetch(`${API_BASE}/clusters`, { cache: "no-store" });
    const json: ApiResponse<Cluster[]> = await res.json();
    if (json.success && json.data) return json.data;
    throw new Error(json.error || "Failed to fetch clusters");
  } catch (err) {
    console.error("[API] Could not load clusters from the backend:", err);
    return [];
  }
}

export async function fetchSignals(mode: DataMode = DEFAULT_DATA_MODE): Promise<CitizenSignal[]> {
  if (mode === "demo") return mockSignals;
  try {
    const res = await fetch(`${API_BASE}/signals`, { cache: "no-store" });
    const json: ApiResponse<CitizenSignal[]> = await res.json();
    if (json.success && json.data) return json.data;
    throw new Error(json.error || "Failed to fetch signals");
  } catch (err) {
    console.error("[API] Could not load signals from the backend:", err);
    return [];
  }
}

export async function fetchFeed(mode: DataMode = DEFAULT_DATA_MODE): Promise<SignalFeedItem[]> {
  if (mode === "demo") return mockFeed;
  const res = await fetch(`${API_BASE}/feed?limit=50`, { cache: "no-store" });
  if (!res.ok) throw new Error(`Feed request failed (${res.status})`);
  const json: ApiResponse<SignalFeedItem[]> = await res.json();
  if (!json.success || !json.data) throw new Error(json.error || "Failed to fetch live feed");
  return json.data;
}

export async function backendIsHealthy(): Promise<boolean> {
  try {
    const response = await fetch(`${API_BASE}/health`, { cache: "no-store" });
    return response.ok;
  } catch {
    return false;
  }
}

export async function recordDecision(
  clusterId: string,
  action: "ACCEPTED" | "REJECTED" | "INVESTIGATING",
  officer: string,
  notes: string
): Promise<boolean> {
  try {
    const res = await fetch(`${API_BASE}/decisions`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        cluster_id: clusterId,
        action,
        officer,
        notes,
      }),
    });
    const json: ApiResponse<unknown> = await res.json();
    return json.success;
  } catch (err) {
    console.error("[API] Decision submission failed:", err);
    return false;
  }
}

export async function triggerSimulatedSignal(): Promise<CitizenSignal | null> {
  try {
    const res = await fetch(`${API_BASE}/demo/simulate`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
    });
    const json: ApiResponse<CitizenSignal> = await res.json();
    return json.data || null;
  } catch (err) {
    console.error("[API] Simulation trigger failed:", err);
    return null;
  }
}

export type AssistantAnswer = { answer: string; grounded_on: string[]; source: "gemini" | "offline" };

/** Asks the backend assistant a question grounded in the evidence on screen. */
export async function askAssistant(
  question: string,
  clusterId?: string,
  wardId?: string
): Promise<AssistantAnswer> {
  const res = await fetch(`${API_BASE}/assistant`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ question, cluster_id: clusterId, ward_id: wardId }),
  });
  const json: ApiResponse<AssistantAnswer> = await res.json();
  if (!json.success || !json.data) throw new Error(json.error || "The assistant could not answer");
  return json.data;
}
