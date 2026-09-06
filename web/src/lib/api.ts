import { ApiResponse, Cluster, Ward, CitizenSignal } from "@/types";
import { mockClusters, mockSignals, mockWards } from "@/lib/mock-dashboard-data";

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1";
export const WS_BASE = process.env.NEXT_PUBLIC_WS_URL || "ws://localhost:8080/ws";
// Live backend data is the default. Mock data is opt-in via
// NEXT_PUBLIC_USE_MOCK_DATA=true, so a demo can never silently present
// fabricated clusters and scores as though they came from the platform.
const USE_MOCK_DATA = process.env.NEXT_PUBLIC_USE_MOCK_DATA === "true";

export async function fetchWards(): Promise<Ward[]> {
  if (USE_MOCK_DATA) return mockWards;
  try {
    const res = await fetch(`${API_BASE}/wards`, { cache: "no-store" });
    const json: ApiResponse<Ward[]> = await res.json();
    if (json.success && json.data) return json.data;
    throw new Error(json.error || "Failed to fetch wards");
  } catch (err) {
    console.error("[API] Could not load wards from the backend:", err);
    return USE_MOCK_DATA ? mockWards : [];
  }
}

export async function fetchClusters(): Promise<Cluster[]> {
  if (USE_MOCK_DATA) return mockClusters;
  try {
    const res = await fetch(`${API_BASE}/clusters`, { cache: "no-store" });
    const json: ApiResponse<Cluster[]> = await res.json();
    if (json.success && json.data) return json.data;
    throw new Error(json.error || "Failed to fetch clusters");
  } catch (err) {
    console.error("[API] Could not load clusters from the backend:", err);
    return USE_MOCK_DATA ? mockClusters : [];
  }
}

export async function fetchSignals(): Promise<CitizenSignal[]> {
  if (USE_MOCK_DATA) return mockSignals;
  try {
    const res = await fetch(`${API_BASE}/signals`, { cache: "no-store" });
    const json: ApiResponse<CitizenSignal[]> = await res.json();
    if (json.success && json.data) return json.data;
    throw new Error(json.error || "Failed to fetch signals");
  } catch (err) {
    console.error("[API] Could not load signals from the backend:", err);
    return USE_MOCK_DATA ? mockSignals : [];
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
