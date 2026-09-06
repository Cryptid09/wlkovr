import { ApiResponse, Cluster, Ward, CitizenSignal } from "@/types";
import { mockClusters, mockSignals, mockWards } from "@/lib/mock-dashboard-data";

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1";
export const WS_BASE = process.env.NEXT_PUBLIC_WS_URL || "ws://localhost:8080/ws";
// Keep the hackathon UI explorable with deterministic local data until live data is explicitly enabled.
const USE_MOCK_DATA = process.env.NEXT_PUBLIC_USE_MOCK_DATA !== "false";

export async function fetchWards(): Promise<Ward[]> {
  if (USE_MOCK_DATA) return mockWards;
  try {
    const res = await fetch(`${API_BASE}/wards`, { cache: "no-store" });
    const json: ApiResponse<Ward[]> = await res.json();
    if (json.success && json.data) return json.data;
    throw new Error(json.error || "Failed to fetch wards");
  } catch (err) {
    console.warn("[API] Falling back to default wards:", err);
    return mockWards;
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
    console.warn("[API] Falling back to default clusters:", err);
    return mockClusters;
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
    console.warn("[API] Falling back to default signals:", err);
    return mockSignals;
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
