import { ApiResponse, Cluster, Ward, CitizenSignal, AuditLog } from "@/types";

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1";
export const WS_BASE = process.env.NEXT_PUBLIC_WS_URL || "ws://localhost:8080/ws";

export async function fetchWards(): Promise<Ward[]> {
  try {
    const res = await fetch(`${API_BASE}/wards`, { cache: "no-store" });
    const json: ApiResponse<Ward[]> = await res.json();
    if (json.success && json.data) return json.data;
    throw new Error(json.error || "Failed to fetch wards");
  } catch (err) {
    console.warn("[API] Falling back to default wards:", err);
    return [];
  }
}

export async function fetchClusters(): Promise<Cluster[]> {
  try {
    const res = await fetch(`${API_BASE}/clusters`, { cache: "no-store" });
    const json: ApiResponse<Cluster[]> = await res.json();
    if (json.success && json.data) return json.data;
    throw new Error(json.error || "Failed to fetch clusters");
  } catch (err) {
    console.warn("[API] Falling back to default clusters:", err);
    return [];
  }
}

export async function fetchSignals(): Promise<CitizenSignal[]> {
  try {
    const res = await fetch(`${API_BASE}/signals`, { cache: "no-store" });
    const json: ApiResponse<CitizenSignal[]> = await res.json();
    if (json.success && json.data) return json.data;
    throw new Error(json.error || "Failed to fetch signals");
  } catch (err) {
    console.warn("[API] Falling back to default signals:", err);
    return [];
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
    const json: ApiResponse<any> = await res.json();
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
