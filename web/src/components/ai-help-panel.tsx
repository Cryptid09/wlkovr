"use client";

import { useEffect, useState } from "react";
import { ArrowUp, Bot, CheckCircle2, Sparkles, X } from "lucide-react";
import { Cluster, Ward } from "@/types";
import { askAssistant } from "@/lib/api";

type Message = { role: "assistant" | "user"; text: string; source?: "gemini" | "offline" };

const suggestions = ["Why is this a priority?", "What should be verified first?", "Explain the four scores", "Which wards might be hiding unmet need?"];

function guidance(question: string, cluster: Cluster | null, ward: Ward | null) {
  const subject = cluster?.title || ward?.name || "this dashboard";
  if (question.toLowerCase().includes("score")) return "Need describes seriousness and spread, Confidence describes evidence strength, Equity highlights under-served communities, and Actionability describes whether the issue is ready for responsible investigation. They are intentionally not combined.";
  if (question.toLowerCase().includes("verified") || question.toLowerCase().includes("first")) return cluster ? `Start with the location and hazard: verify the ${cluster.title.toLowerCase()} on site, confirm the responsible department, and check whether existing plans already cover it. This is a recommendation for investigation, not an automatic approval.` : "Start by confirming the ward, the affected service, and the quality of available evidence. Low participation should lead to outreach or field verification, not a conclusion that there is no need.";
  return cluster ? `${subject} is visible because ${cluster.signal_count} citizen signals were grouped into one issue. Its current urgency is ${cluster.urgency.score}/100, with evidence arriving through ${cluster.channels.join(", ")}. Review the original signals and representation limits before recording a decision.` : `The dashboard is showing ${subject} as a context area. Compare its infrastructure index, population, historic spend, and signal volume before deciding what deserves field verification.`;
}

export function AIHelpPanel({ open, onClose, cluster, ward }: { open: boolean; onClose: () => void; cluster: Cluster | null; ward: Ward | null }) {
  const [question, setQuestion] = useState("");
  const [pending, setPending] = useState(false);
  const [messages, setMessages] = useState<Message[]>([{ role: "assistant", text: "I can help you read the evidence, scores, and recommended next steps on this page." }]);

  useEffect(() => { if (!open) return; const closeOnEscape = (event: KeyboardEvent) => { if (event.key === "Escape") onClose(); }; window.addEventListener("keydown", closeOnEscape); return () => window.removeEventListener("keydown", closeOnEscape); }, [open, onClose]);
  const ask = async (value = question) => {
    const trimmed = value.trim();
    if (!trimmed || pending) return;
    setMessages((old) => [...old, { role: "user", text: trimmed }]);
    setQuestion("");
    setPending(true);
    try {
      const reply = await askAssistant(trimmed, cluster?.id, ward?.id);
      setMessages((old) => [...old, { role: "assistant", text: reply.answer, source: reply.source }]);
    } catch {
      // The backend is unreachable; fall back to the local explanation rather
      // than leaving the official with an empty panel.
      setMessages((old) => [...old, { role: "assistant", text: guidance(trimmed, cluster, ward) }]);
    } finally {
      setPending(false);
    }
  };
  const context = cluster?.title || ward?.name || "Dashboard overview";

  return <><div aria-hidden={!open} onClick={onClose} className={`fixed inset-0 z-40 bg-[#111111]/20 transition-opacity duration-300 ${open ? "opacity-100" : "pointer-events-none opacity-0"}`} /><aside aria-label="Zen AI help panel" aria-hidden={!open} className={`fixed right-0 top-0 z-50 flex h-full w-full max-w-[390px] flex-col border-l border-[#d3cec6] bg-white shadow-2xl transition-transform duration-300 ease-out ${open ? "translate-x-0" : "pointer-events-none translate-x-full"}`}><div className="flex items-center justify-between border-b border-[#ebe7e1] px-5 py-4"><div className="flex items-center gap-3"><div className="grid h-9 w-9 place-items-center rounded-xl bg-[#111111] text-white"><Bot className="h-4 w-4" /></div><div><p className="text-sm font-medium">Zen AI Guide</p><p className="text-[11px] text-[#626260]">Contextual help · human review stays central</p></div></div><button onClick={onClose} className="grid h-9 w-9 place-items-center rounded-lg text-[#626260] hover:bg-[#f5f1ec]" aria-label="Close AI help"><X className="h-4 w-4" /></button></div><div className="border-b border-[#ebe7e1] bg-[#fff8f4] px-5 py-3"><p className="text-[10px] font-medium uppercase tracking-[0.1em] text-[#c13c00]">Current context</p><p className="mt-1 truncate text-xs font-medium text-[#4a3028]">{context}</p></div><div className="flex-1 space-y-3 overflow-y-auto px-5 py-5">{messages.map((message, index) => <div key={`${message.role}-${index}`} className={message.role === "user" ? "ml-8 rounded-xl bg-[#111111] px-3 py-2.5 text-xs leading-5 text-white" : "mr-5 rounded-xl bg-[#f5f1ec] px-3 py-2.5 text-xs leading-5 text-[#4a4a47]"}>{message.role === "assistant" && <Sparkles className="mr-1 inline h-3 w-3 text-[#ff5600]" />}{message.text}{message.source === "offline" && <span className="mt-1 block text-[10px] text-[#9c9fa5]">Answered from stored evidence — the AI service was unreachable.</span>}</div>)}{pending && <div className="mr-5 rounded-xl bg-[#f5f1ec] px-3 py-2.5 text-xs italic leading-5 text-[#7b7b78]">Reading the evidence…</div>}<div className="pt-2"><p className="mb-2 text-[11px] font-medium text-[#7b7b78]">Suggested questions</p><div className="space-y-2">{suggestions.map((suggestion) => <button key={suggestion} onClick={() => void ask(suggestion)} className="flex w-full items-center justify-between rounded-lg border border-[#d3cec6] px-3 py-2.5 text-left text-xs text-[#626260] hover:border-[#111111] hover:text-[#111111]">{suggestion}<ArrowUp className="h-3.5 w-3.5 rotate-45" /></button>)}</div></div></div><div className="border-t border-[#ebe7e1] p-4"><form onSubmit={(event) => { event.preventDefault(); void ask(); }} className="flex items-center gap-2 rounded-xl border border-[#d3cec6] bg-white p-1.5 focus-within:border-[#111111]"><input value={question} onChange={(event) => setQuestion(event.target.value)} className="min-w-0 flex-1 px-2 text-xs outline-none" placeholder="Ask about this evidence…" aria-label="Ask Zen AI" /><button type="submit" className="grid h-8 w-8 place-items-center rounded-lg bg-[#111111] text-white disabled:opacity-40" disabled={!question.trim() || pending} aria-label="Send question"><ArrowUp className="h-4 w-4" /></button></form><p className="mt-2 flex gap-1 text-[10px] leading-4 text-[#9c9fa5]"><CheckCircle2 className="mt-0.5 h-3 w-3 shrink-0" />AI guidance is illustrative. Verify evidence before taking official action.</p></div></aside></>;
}
