"use client";

import React, { useEffect, useState, useRef } from "react";
import {
  AlertTriangle,
  Activity,
  CheckCircle,
  EyeOff,
  Flame,
  MessageSquare,
  Radio,
  Send,
  ShieldAlert,
  Clock,
  MapPin,
  TrendingUp,
  Sparkles,
  ChevronRight,
  RefreshCw,
  Building2,
  Check,
  X,
  Search,
  Filter,
} from "lucide-react";
import { Cluster, Ward, CitizenSignal, UrgencyTier } from "@/types";
import { fetchClusters, fetchWards, fetchSignals, recordDecision, triggerSimulatedSignal, WS_BASE } from "@/lib/api";

export default function Dashboard() {
  const [clusters, setClusters] = useState<Cluster[]>([]);
  const [wards, setWards] = useState<Ward[]>([]);
  const [signals, setSignals] = useState<CitizenSignal[]>([]);
  const [selectedCluster, setSelectedCluster] = useState<Cluster | null>(null);
  const [selectedWard, setSelectedWard] = useState<Ward | null>(null);
  const [wsConnected, setWsConnected] = useState<boolean>(false);
  const [isSimulating, setIsSimulating] = useState<boolean>(false);
  const [officerName, setOfficerName] = useState<string>("Municipal Commissioner");
  const [decisionNotes, setDecisionNotes] = useState<string>("");
  const [isSubmitting, setIsSubmitting] = useState<boolean>(false);
  const [actionSuccess, setActionSuccess] = useState<string | null>(null);
  const [filterTier, setFilterTier] = useState<string>("ALL");
  const [activeTab, setActiveTab] = useState<"clusters" | "wards" | "signals">("clusters");

  const wsRef = useRef<WebSocket | null>(null);

  // Initial Data Load
  const loadData = async () => {
    const [cData, wData, sData] = await Promise.all([
      fetchClusters(),
      fetchWards(),
      fetchSignals(),
    ]);
    if (cData.length > 0) {
      setClusters(cData);
      setSelectedCluster(cData[0]);
    }
    if (wData.length > 0) setWards(wData);
    if (sData.length > 0) setSignals(sData);
  };

  useEffect(() => {
    loadData();

    // Setup Realtime WebSocket connection to Go Backend
    const connectWS = () => {
      try {
        const ws = new WebSocket(WS_BASE);
        wsRef.current = ws;

        ws.onopen = () => {
          setWsConnected(true);
        };

        ws.onmessage = (event) => {
          try {
            const msg = JSON.parse(event.data);
            if (msg.type === "SIGNAL_RECEIVED") {
              const newSignal: CitizenSignal = msg.data;
              setSignals((prev) => [newSignal, ...prev.slice(0, 49)]);
            } else if (msg.type === "DECISION_RECORDED") {
              const { cluster } = msg.data;
              setClusters((prev) =>
                prev.map((c) => (c.id === cluster.id ? cluster : c))
              );
              if (selectedCluster?.id === cluster.id) {
                setSelectedCluster(cluster);
              }
            }
          } catch (e) {
            console.error("WS message parse error:", e);
          }
        };

        ws.onclose = () => {
          setWsConnected(false);
          // Reconnect after 3 seconds
          setTimeout(connectWS, 3000);
        };

        ws.onerror = () => {
          setWsConnected(false);
        };
      } catch (e) {
        console.warn("WS connection failed, running offline mode");
      }
    };

    connectWS();

    return () => {
      if (wsRef.current) wsRef.current.close();
    };
  }, []);

  // Handle Decision
  const handleDecision = async (action: "ACCEPTED" | "REJECTED" | "INVESTIGATING") => {
    if (!selectedCluster) return;
    setIsSubmitting(true);
    const success = await recordDecision(
      selectedCluster.id,
      action,
      officerName,
      decisionNotes || `Action marked as ${action} by ${officerName}`
    );

    if (success) {
      setActionSuccess(`Decision recorded: ${action}`);
      setTimeout(() => setActionSuccess(null), 4000);
      setDecisionNotes("");
      loadData();
    }
    setIsSubmitting(false);
  };

  // Handle Simulation
  const handleSimulate = async () => {
    setIsSimulating(true);
    await triggerSimulatedSignal();
    await loadData();
    setIsSimulating(false);
  };

  // Filtered clusters
  const filteredClusters = clusters.filter((c) => {
    if (filterTier === "ALL") return true;
    return c.urgency.tier === filterTier;
  });

  // Calculate Metrics
  const totalSignals = signals.length + clusters.reduce((acc, c) => acc + c.signal_count, 0);
  const criticalCount = clusters.filter((c) => c.urgency.tier === "TIER_1_CRITICAL").length;
  const blindSpotsCount = wards.filter((w) => w.is_blind_spot).length;

  const getTierBadge = (tier: UrgencyTier) => {
    switch (tier) {
      case "TIER_1_CRITICAL":
        return (
          <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-semibold bg-red-500/10 text-red-500 border border-red-500/20 animate-pulse">
            <Flame className="w-3.5 h-3.5" /> Tier 1 Critical (&lt;4h SLA)
          </span>
        );
      case "TIER_2_HIGH":
        return (
          <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-semibold bg-orange-500/10 text-orange-500 border border-orange-500/20">
            <AlertTriangle className="w-3.5 h-3.5" /> Tier 2 High (&lt;24h SLA)
          </span>
        );
      case "TIER_3_MEDIUM":
        return (
          <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-semibold bg-yellow-500/10 text-yellow-500 border border-yellow-500/20">
            <Clock className="w-3.5 h-3.5" /> Tier 3 Medium (&lt;72h SLA)
          </span>
        );
      default:
        return (
          <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-semibold bg-emerald-500/10 text-emerald-500 border border-emerald-500/20">
            <CheckCircle className="w-3.5 h-3.5" /> Tier 4 Routine
          </span>
        );
    }
  };

  return (
    <div className="min-h-screen bg-[#090D16] text-slate-100 font-sans">
      {/* Top Navigation Bar */}
      <header className="sticky top-0 z-50 backdrop-blur-xl bg-[#090D16]/80 border-b border-slate-800/80 px-6 py-3.5">
        <div className="max-w-7xl mx-auto flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-xl bg-gradient-to-tr from-blue-600 to-indigo-500 flex items-center justify-center shadow-lg shadow-indigo-500/20 border border-indigo-400/30">
              <ShieldAlert className="w-5 h-5 text-white" />
            </div>
            <div>
              <div className="flex items-center gap-2">
                <h1 className="text-base font-bold tracking-tight text-white">
                  Zen Civic Intelligence Platform
                </h1>
                <span className="text-[10px] uppercase font-bold tracking-wider px-2 py-0.5 rounded-md bg-indigo-500/10 text-indigo-400 border border-indigo-500/20">
                  Indore Command
                </span>
              </div>
              <p className="text-xs text-slate-400">
                AI Omnichannel Governance &amp; Decision Support System
              </p>
            </div>
          </div>

          <div className="flex items-center gap-3">
            {/* Live WebSocket Status */}
            <div className="flex items-center gap-2 px-3 py-1.5 rounded-full bg-slate-900 border border-slate-800 text-xs">
              <span
                className={`w-2 h-2 rounded-full ${
                  wsConnected ? "bg-emerald-400 animate-ping" : "bg-amber-400"
                }`}
              />
              <span className="text-slate-300 font-medium">
                {wsConnected ? "Live Engine Connected" : "Connecting..."}
              </span>
            </div>

            {/* Simulated WhatsApp Ingestion Trigger */}
            <button
              onClick={handleSimulate}
              disabled={isSimulating}
              className="flex items-center gap-2 px-4 py-1.5 rounded-xl bg-gradient-to-r from-emerald-600 to-teal-600 hover:from-emerald-500 hover:to-teal-500 text-white text-xs font-semibold shadow-lg shadow-emerald-500/20 transition-all active:scale-95 disabled:opacity-50 cursor-pointer"
            >
              <Send className="w-3.5 h-3.5" />
              {isSimulating ? "Ingesting..." : "Simulate WhatsApp Signal"}
            </button>
          </div>
        </div>
      </header>

      {/* Main Container */}
      <main className="max-w-7xl mx-auto px-6 py-6 space-y-6">
        {/* Metric Cards Banner */}
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <div className="p-4 rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-sm relative overflow-hidden">
            <div className="flex items-center justify-between">
              <span className="text-xs font-medium text-slate-400 uppercase tracking-wider">
                Corroborated Signals
              </span>
              <MessageSquare className="w-4 h-4 text-blue-400" />
            </div>
            <div className="mt-2 flex items-baseline gap-2">
              <span className="text-2xl font-bold text-white">{totalSignals}</span>
              <span className="text-xs text-emerald-400 font-medium">+3 live</span>
            </div>
            <p className="text-[11px] text-slate-500 mt-1">WhatsApp, SMS &amp; Web normalized</p>
          </div>

          <div className="p-4 rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-sm relative overflow-hidden">
            <div className="flex items-center justify-between">
              <span className="text-xs font-medium text-slate-400 uppercase tracking-wider">
                Active Clusters
              </span>
              <Activity className="w-4 h-4 text-indigo-400" />
            </div>
            <div className="mt-2 flex items-baseline gap-2">
              <span className="text-2xl font-bold text-white">{clusters.length}</span>
              <span className="text-xs text-slate-400">across 12 wards</span>
            </div>
            <p className="text-[11px] text-slate-500 mt-1">Grouped by Gemini embeddings</p>
          </div>

          <div className="p-4 rounded-2xl bg-red-950/20 border border-red-900/30 backdrop-blur-sm relative overflow-hidden">
            <div className="flex items-center justify-between">
              <span className="text-xs font-medium text-red-400 uppercase tracking-wider">
                Tier-1 Emergencies
              </span>
              <Flame className="w-4 h-4 text-red-500 animate-pulse" />
            </div>
            <div className="mt-2 flex items-baseline gap-2">
              <span className="text-2xl font-bold text-red-400">{criticalCount}</span>
              <span className="text-xs text-red-500/80 font-medium">&lt;4h SLA</span>
            </div>
            <p className="text-[11px] text-red-400/60 mt-1">Urgency Decision Engine triggers</p>
          </div>

          <div className="p-4 rounded-2xl bg-purple-950/20 border border-purple-900/30 backdrop-blur-sm relative overflow-hidden">
            <div className="flex items-center justify-between">
              <span className="text-xs font-medium text-purple-400 uppercase tracking-wider">
                Civic Blind Spots
              </span>
              <EyeOff className="w-4 h-4 text-purple-400" />
            </div>
            <div className="mt-2 flex items-baseline gap-2">
              <span className="text-2xl font-bold text-purple-400">{blindSpotsCount}</span>
              <span className="text-xs text-purple-400/80 font-medium">Silent Wards</span>
            </div>
            <p className="text-[11px] text-purple-400/60 mt-1">High need, zero digital voice</p>
          </div>
        </div>

        {/* Action Success Alert */}
        {actionSuccess && (
          <div className="p-3 rounded-xl bg-emerald-500/10 border border-emerald-500/30 text-emerald-400 text-xs flex items-center gap-2 animate-fadeIn">
            <CheckCircle className="w-4 h-4 shrink-0" />
            <span className="font-semibold">{actionSuccess}</span>
          </div>
        )}

        {/* Split View: Left Queue + Right Detail Panel */}
        <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
          {/* Left Column: Priority Clusters & Wards (5 Cols) */}
          <div className="lg:col-span-5 space-y-4">
            {/* Filter Tabs */}
            <div className="flex items-center justify-between border-b border-slate-800 pb-3">
              <div className="flex gap-2">
                <button
                  onClick={() => setActiveTab("clusters")}
                  className={`px-3 py-1.5 rounded-lg text-xs font-semibold transition-all cursor-pointer ${
                    activeTab === "clusters"
                      ? "bg-indigo-600 text-white shadow-md shadow-indigo-600/20"
                      : "text-slate-400 hover:text-white"
                  }`}
                >
                  Hotspot Queue ({filteredClusters.length})
                </button>
                <button
                  onClick={() => setActiveTab("wards")}
                  className={`px-3 py-1.5 rounded-lg text-xs font-semibold transition-all cursor-pointer ${
                    activeTab === "wards"
                      ? "bg-indigo-600 text-white shadow-md shadow-indigo-600/20"
                      : "text-slate-400 hover:text-white"
                  }`}
                >
                  Ward Intelligence ({wards.length})
                </button>
                <button
                  onClick={() => setActiveTab("signals")}
                  className={`px-3 py-1.5 rounded-lg text-xs font-semibold transition-all cursor-pointer ${
                    activeTab === "signals"
                      ? "bg-indigo-600 text-white shadow-md shadow-indigo-600/20"
                      : "text-slate-400 hover:text-white"
                  }`}
                >
                  Live Stream ({signals.length})
                </button>
              </div>

              {activeTab === "clusters" && (
                <select
                  value={filterTier}
                  onChange={(e) => setFilterTier(e.target.value)}
                  className="bg-slate-900 border border-slate-800 text-slate-300 text-xs rounded-lg px-2 py-1 outline-none cursor-pointer"
                >
                  <option value="ALL">All Tiers</option>
                  <option value="TIER_1_CRITICAL">Tier 1 Critical</option>
                  <option value="TIER_2_HIGH">Tier 2 High</option>
                  <option value="TIER_3_MEDIUM">Tier 3 Medium</option>
                </select>
              )}
            </div>

            {/* List Content */}
            <div className="space-y-3 max-h-[620px] overflow-y-auto pr-1">
              {activeTab === "clusters" &&
                filteredClusters.map((cluster) => {
                  const isSelected = selectedCluster?.id === cluster.id;
                  return (
                    <div
                      key={cluster.id}
                      onClick={() => setSelectedCluster(cluster)}
                      className={`p-4 rounded-2xl border transition-all cursor-pointer ${
                        isSelected
                          ? "bg-indigo-950/20 border-indigo-500/50 shadow-lg shadow-indigo-500/10"
                          : "bg-slate-900/40 border-slate-800/60 hover:bg-slate-900/80 hover:border-slate-700"
                      }`}
                    >
                      <div className="flex items-start justify-between gap-2">
                        <div>
                          <span className="text-[11px] font-semibold text-slate-400 uppercase tracking-wider">
                            {cluster.department}
                          </span>
                          <h3 className="text-sm font-bold text-white mt-0.5">
                            {cluster.title}
                          </h3>
                          <div className="flex items-center gap-2 mt-1 text-xs text-slate-400">
                            <MapPin className="w-3.5 h-3.5 text-indigo-400" />
                            <span>{cluster.ward_name}</span>
                          </div>
                        </div>
                        {getTierBadge(cluster.urgency.tier)}
                      </div>

                      {/* Score Bar Mini Strip */}
                      <div className="mt-3 pt-3 border-t border-slate-800/60 grid grid-cols-4 gap-2 text-center text-[10px]">
                        <div className="bg-slate-950/40 p-1.5 rounded-lg">
                          <span className="text-slate-400 block font-medium">Need</span>
                          <span className="text-blue-400 font-bold">{cluster.scores.need}%</span>
                        </div>
                        <div className="bg-slate-950/40 p-1.5 rounded-lg">
                          <span className="text-slate-400 block font-medium">Confidence</span>
                          <span className="text-emerald-400 font-bold">{cluster.scores.confidence}%</span>
                        </div>
                        <div className="bg-slate-950/40 p-1.5 rounded-lg">
                          <span className="text-slate-400 block font-medium">Equity</span>
                          <span className="text-purple-400 font-bold">{cluster.scores.equity}%</span>
                        </div>
                        <div className="bg-slate-950/40 p-1.5 rounded-lg">
                          <span className="text-slate-400 block font-medium">Action</span>
                          <span className="text-amber-400 font-bold">{cluster.scores.actionability}%</span>
                        </div>
                      </div>
                    </div>
                  );
                })}

              {activeTab === "wards" &&
                wards.map((ward) => (
                  <div
                    key={ward.id}
                    onClick={() => setSelectedWard(ward)}
                    className={`p-4 rounded-2xl border transition-all cursor-pointer ${
                      ward.is_blind_spot
                        ? "bg-purple-950/20 border-purple-800/40"
                        : "bg-slate-900/40 border-slate-800/60 hover:bg-slate-900"
                    }`}
                  >
                    <div className="flex items-start justify-between">
                      <div>
                        <h4 className="text-sm font-bold text-white">{ward.name}</h4>
                        <span className="text-xs text-slate-400">{ward.zone}</span>
                      </div>
                      {ward.is_blind_spot && (
                        <span className="px-2 py-0.5 rounded-md text-[10px] font-bold bg-purple-500/20 text-purple-300 border border-purple-500/30">
                          Civic Blind Spot
                        </span>
                      )}
                    </div>
                    <div className="mt-2 text-xs text-slate-400 flex items-center justify-between">
                      <span>Infra Index: {(ward.infra_index * 100).toFixed(0)}%</span>
                      <span>Population: {ward.population.toLocaleString()}</span>
                      <span>Hist. Spend: ₹{ward.historical_spend_cr} Cr</span>
                    </div>
                  </div>
                ))}

              {activeTab === "signals" &&
                signals.map((sig) => (
                  <div
                    key={sig.id}
                    className="p-3.5 rounded-xl bg-slate-900/50 border border-slate-800/60 space-y-1.5"
                  >
                    <div className="flex items-center justify-between text-xs">
                      <span className="font-semibold text-emerald-400 flex items-center gap-1">
                        <MessageSquare className="w-3.5 h-3.5" />
                        {sig.provider} ({sig.language})
                      </span>
                      <span className="text-slate-500 text-[11px]">
                        {new Date(sig.timestamp).toLocaleTimeString()}
                      </span>
                    </div>
                    <p className="text-xs text-slate-200">{sig.raw_text}</p>
                    {sig.location_hint && (
                      <span className="text-[10px] text-slate-400 block">
                        📍 Hint: {sig.location_hint}
                      </span>
                    )}
                  </div>
                ))}
            </div>
          </div>

          {/* Right Column: Cluster Detail, 4D Explainability & Human Action Panel (7 Cols) */}
          <div className="lg:col-span-7">
            {selectedCluster ? (
              <div className="p-6 rounded-3xl bg-slate-900/70 border border-slate-800 backdrop-blur-md space-y-6">
                {/* Header */}
                <div className="flex items-start justify-between gap-4">
                  <div>
                    <div className="flex items-center gap-2">
                      <span className="text-xs font-bold uppercase tracking-wider text-indigo-400">
                        {selectedCluster.department}
                      </span>
                      <span className="text-xs text-slate-500">•</span>
                      <span className="text-xs text-slate-400">
                        {selectedCluster.signal_count} Corroborating Signals ({selectedCluster.channels.join(", ")})
                      </span>
                    </div>
                    <h2 className="text-xl font-bold text-white mt-1">
                      {selectedCluster.title}
                    </h2>
                    <p className="text-xs text-slate-400 mt-0.5">
                      📍 {selectedCluster.ward_name} (Centroid: {selectedCluster.centroid_lat}, {selectedCluster.centroid_lng})
                    </p>
                  </div>
                  {getTierBadge(selectedCluster.urgency.tier)}
                </div>

                {/* Description */}
                <div className="p-3.5 rounded-xl bg-slate-950/50 border border-slate-800/70 text-xs text-slate-300">
                  {selectedCluster.description}
                </div>

                {/* Grounded Recommendation Banner */}
                <div className="p-4 rounded-2xl bg-gradient-to-r from-indigo-950/30 to-blue-950/20 border border-indigo-500/30 space-y-1.5">
                  <div className="flex items-center gap-2 text-indigo-300 text-xs font-bold">
                    <Sparkles className="w-4 h-4 text-indigo-400" />
                    <span>Grounded AI Policy Recommendation</span>
                  </div>
                  <p className="text-xs text-slate-200 leading-relaxed">
                    {selectedCluster.recommendation}
                  </p>
                </div>

                {/* 4-Dimensional Priority Visualizer */}
                <div className="space-y-3">
                  <h4 className="text-xs font-bold uppercase tracking-wider text-slate-400">
                    Transparent 4D Priority Scores (Explainable Governance)
                  </h4>
                  <div className="grid grid-cols-2 gap-3">
                    {/* Need */}
                    <div className="p-3 rounded-xl bg-slate-950/40 border border-slate-800/80 space-y-1.5">
                      <div className="flex justify-between text-xs">
                        <span className="text-slate-400 font-medium">Need Score</span>
                        <span className="text-blue-400 font-bold">{selectedCluster.scores.need}%</span>
                      </div>
                      <div className="w-full bg-slate-800 h-2 rounded-full overflow-hidden">
                        <div
                          className="bg-blue-500 h-full rounded-full transition-all duration-500"
                          style={{ width: `${selectedCluster.scores.need}%` }}
                        />
                      </div>
                      <span className="text-[10px] text-slate-500 block">Volume &times; Urgency Decision Engine</span>
                    </div>

                    {/* Confidence */}
                    <div className="p-3 rounded-xl bg-slate-950/40 border border-slate-800/80 space-y-1.5">
                      <div className="flex justify-between text-xs">
                        <span className="text-slate-400 font-medium">Confidence Score</span>
                        <span className="text-emerald-400 font-bold">{selectedCluster.scores.confidence}%</span>
                      </div>
                      <div className="w-full bg-slate-800 h-2 rounded-full overflow-hidden">
                        <div
                          className="bg-emerald-500 h-full rounded-full transition-all duration-500"
                          style={{ width: `${selectedCluster.scores.confidence}%` }}
                        />
                      </div>
                      <span className="text-[10px] text-slate-500 block">Multi-channel corroboration</span>
                    </div>

                    {/* Equity */}
                    <div className="p-3 rounded-xl bg-slate-950/40 border border-slate-800/80 space-y-1.5">
                      <div className="flex justify-between text-xs">
                        <span className="text-slate-400 font-medium">Equity Priority</span>
                        <span className="text-purple-400 font-bold">{selectedCluster.scores.equity}%</span>
                      </div>
                      <div className="w-full bg-slate-800 h-2 rounded-full overflow-hidden">
                        <div
                          className="bg-purple-500 h-full rounded-full transition-all duration-500"
                          style={{ width: `${selectedCluster.scores.equity}%` }}
                        />
                      </div>
                      <span className="text-[10px] text-slate-500 block">Inverse of ward infrastructure index</span>
                    </div>

                    {/* Actionability */}
                    <div className="p-3 rounded-xl bg-slate-950/40 border border-slate-800/80 space-y-1.5">
                      <div className="flex justify-between text-xs">
                        <span className="text-slate-400 font-medium">Actionability</span>
                        <span className="text-amber-400 font-bold">{selectedCluster.scores.actionability}%</span>
                      </div>
                      <div className="w-full bg-slate-800 h-2 rounded-full overflow-hidden">
                        <div
                          className="bg-amber-500 h-full rounded-full transition-all duration-500"
                          style={{ width: `${selectedCluster.scores.actionability}%` }}
                        />
                      </div>
                      <span className="text-[10px] text-slate-500 block">Location &amp; Department verified</span>
                    </div>
                  </div>
                </div>

                {/* Urgency Decision Engine Factor Breakdown */}
                <div className="p-4 rounded-2xl bg-slate-950/40 border border-slate-800 space-y-2">
                  <span className="text-xs font-bold uppercase tracking-wider text-slate-400 block">
                    ⚡ Urgency Decision Engine Drivers
                  </span>
                  <div className="flex flex-wrap gap-2">
                    {selectedCluster.urgency.factors.map((factor, idx) => (
                      <span
                        key={idx}
                        className="px-2.5 py-1 rounded-lg text-[11px] font-medium bg-slate-800 text-slate-200 border border-slate-700/80 flex items-center gap-1.5"
                      >
                        <AlertTriangle className="w-3 h-3 text-amber-400" />
                        {factor}
                      </span>
                    ))}
                  </div>
                </div>

                {/* Human-in-the-Loop Policymaker Action Panel */}
                <div className="p-5 rounded-2xl bg-slate-950/80 border border-indigo-900/30 space-y-4">
                  <div className="flex items-center justify-between">
                    <div>
                      <h4 className="text-sm font-bold text-white flex items-center gap-2">
                        <CheckCircle className="w-4 h-4 text-indigo-400" />
                        Policymaker Action &amp; Audit Trail
                      </h4>
                      <p className="text-[11px] text-slate-400 mt-0.5">
                        System recommends; municipal officials record the binding decision.
                      </p>
                    </div>
                    <span className="text-xs px-2.5 py-1 rounded-md font-bold bg-slate-900 text-indigo-300 border border-indigo-500/30">
                      Status: {selectedCluster.status}
                    </span>
                  </div>

                  <div className="space-y-2">
                    <input
                      type="text"
                      value={decisionNotes}
                      onChange={(e) => setDecisionNotes(e.target.value)}
                      placeholder="Add dispatch notes or instructions for field engineer..."
                      className="w-full bg-slate-900 border border-slate-800 rounded-xl px-3.5 py-2 text-xs text-white placeholder-slate-500 focus:outline-none focus:border-indigo-500"
                    />
                  </div>

                  <div className="flex items-center gap-3">
                    <button
                      onClick={() => handleDecision("ACCEPTED")}
                      disabled={isSubmitting}
                      className="flex-1 flex items-center justify-center gap-2 py-2.5 rounded-xl bg-emerald-600 hover:bg-emerald-500 text-white text-xs font-bold transition-all cursor-pointer active:scale-95 disabled:opacity-50"
                    >
                      <Check className="w-4 h-4" /> Accept &amp; Dispatch Field Team
                    </button>
                    <button
                      onClick={() => handleDecision("INVESTIGATING")}
                      disabled={isSubmitting}
                      className="flex-1 flex items-center justify-center gap-2 py-2.5 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-bold transition-all cursor-pointer active:scale-95 disabled:opacity-50"
                    >
                      <RefreshCw className="w-4 h-4" /> Request Site Inspection
                    </button>
                    <button
                      onClick={() => handleDecision("REJECTED")}
                      disabled={isSubmitting}
                      className="px-4 py-2.5 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-300 text-xs font-semibold transition-all cursor-pointer active:scale-95 disabled:opacity-50"
                    >
                      <X className="w-4 h-4" /> Reject
                    </button>
                  </div>
                </div>
              </div>
            ) : (
              <div className="h-full flex items-center justify-center p-12 rounded-3xl bg-slate-900/40 border border-slate-800 text-slate-500 text-sm">
                Select a hotspot cluster from the queue to view full 4D intelligence and take action.
              </div>
            )}
          </div>
        </div>
      </main>
    </div>
  );
}
