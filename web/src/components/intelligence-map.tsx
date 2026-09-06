"use client";

import { useEffect, useState } from "react";
import { CircleMarker, MapContainer, Popup, TileLayer } from "react-leaflet";
import "leaflet/dist/leaflet.css";
import { Cluster, Ward } from "@/types";

const center: [number, number] = [22.7196, 75.8577];

export function IntelligenceMap({
  clusters,
  wards,
  selectedCluster,
  onSelectCluster,
  onSelectWard,
}: {
  clusters: Cluster[];
  wards: Ward[];
  selectedCluster: Cluster | null;
  onSelectCluster: (cluster: Cluster) => void;
  onSelectWard: (ward: Ward) => void;
}) {
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  if (!mounted) {
    return (
      <div className="grid h-[335px] place-items-center rounded-2xl bg-[#e8e2d8] text-xs text-[#626260]">
        Loading Indore intelligence map…
      </div>
    );
  }

  return (
    <div
      className="relative z-0 h-[335px] overflow-hidden rounded-2xl border border-[#d3cec6]"
      aria-label="Indore civic intelligence map"
    >
      <MapContainer
        key="indore-map-container"
        center={center}
        zoom={12}
        scrollWheelZoom={false}
        className="relative z-0 h-full w-full"
      >
        <TileLayer
          attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
          url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
        />
        {clusters.map((cluster) => (
          <CircleMarker
            key={`cluster-marker-${cluster.id}`}
            center={[cluster.centroid_lat, cluster.centroid_lng]}
            pathOptions={{
              color: "#ffffff",
              weight: selectedCluster?.id === cluster.id ? 5 : 2,
              fillColor: cluster.urgency.tier === "TIER_1_CRITICAL" ? "#c41c1c" : "#ff5600",
              fillOpacity: 0.95,
            }}
            radius={selectedCluster?.id === cluster.id ? 13 : 10}
            eventHandlers={{ click: () => onSelectCluster(cluster) }}
          >
            <Popup>
              <strong>{cluster.title}</strong>
              <br />
              {cluster.ward_name}
              <br />
              {cluster.signal_count} corroborating signals
            </Popup>
          </CircleMarker>
        ))}
        {wards
          .filter((ward) => ward.is_blind_spot)
          .map((ward) => (
            <CircleMarker
              key={`blindspot-marker-${ward.id}`}
              center={[ward.lat, ward.lng]}
              radius={9}
              pathOptions={{
                color: "#ffffff",
                weight: 2,
                fillColor: "#7c3aed",
                fillOpacity: 0.95,
              }}
              eventHandlers={{ click: () => onSelectWard(ward) }}
            >
              <Popup>
                <strong>{ward.name}</strong>
                <br />
                Civic data blind spot
              </Popup>
            </CircleMarker>
          ))}
      </MapContainer>
      <div className="pointer-events-none relative z-10 -mt-[54px] ml-4 inline-flex gap-3 rounded-lg bg-white/95 px-3 py-2 text-[11px] text-[#626260]">
        <span>
          <i className="mr-1 inline-block h-2 w-2 rounded-full bg-[#c41c1c]" />
          Hotspot
        </span>
        <span>
          <i className="mr-1 inline-block h-2 w-2 rounded-full bg-violet-600" />
          Blind spot
        </span>
      </div>
    </div>
  );
}
