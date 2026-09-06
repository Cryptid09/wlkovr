"use client";

import { FourDimensionalScore } from "@/types";
import { PolarAngleAxis, PolarGrid, Radar, RadarChart, ResponsiveContainer } from "recharts";

const dimensions = [
  { key: "need", label: "Need", color: "#ff5600", detail: "Severity × corroborated volume" },
  { key: "confidence", label: "Confidence", color: "#16803c", detail: "Independent corroboration" },
  { key: "equity", label: "Equity", color: "#7c3aed", detail: "Under-service priority" },
  { key: "actionability", label: "Actionability", color: "#0369a1", detail: "Location & ownership resolved" },
] as const;

export function FourDScoreChart({ scores }: { scores: FourDimensionalScore }) {
  const data = dimensions.map((dimension) => ({ subject: dimension.label, value: scores[dimension.key], fullMark: 100 }));

  return (
    <section className="rounded-2xl border border-[#d3cec6] bg-white p-5">
      <div className="flex items-start justify-between gap-4">
        <div>
          <p className="text-sm font-medium text-[#111111]">Transparent priority dimensions</p>
          <p className="mt-1 text-xs leading-5 text-[#626260]">No hidden composite score. Each dimension remains visible for review.</p>
        </div>
        <span className="rounded-full bg-[#fff0ea] px-2.5 py-1 text-[11px] font-medium text-[#c13c00]">Illustrative weights</span>
      </div>
      <div className="mt-3 h-48" aria-label="Radar chart showing the four transparent priority dimensions">
        <ResponsiveContainer width="100%" height="100%">
          <RadarChart data={data} outerRadius="68%">
            <PolarGrid stroke="#d3cec6" />
            <PolarAngleAxis dataKey="subject" tick={{ fill: "#626260", fontSize: 11 }} />
            <Radar dataKey="value" stroke="#ff5600" fill="#ff5600" fillOpacity={0.18} />
          </RadarChart>
        </ResponsiveContainer>
      </div>
      <div className="grid grid-cols-2 gap-3">
        {dimensions.map((dimension) => (
          <div key={dimension.key} className="rounded-xl bg-[#f5f1ec] p-3">
            <div className="flex items-center justify-between text-xs"><span className="font-medium text-[#111111]">{dimension.label}</span><span style={{ color: dimension.color }} className="font-semibold">{scores[dimension.key]}%</span></div>
            <div className="mt-2 h-1.5 overflow-hidden rounded-full bg-[#e4dfd8]"><div className="h-full rounded-full transition-all duration-700" style={{ width: `${scores[dimension.key]}%`, backgroundColor: dimension.color }} /></div>
            <p className="mt-1.5 text-[10px] text-[#7b7b78]">{dimension.detail}</p>
          </div>
        ))}
      </div>
    </section>
  );
}
