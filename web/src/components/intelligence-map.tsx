"use client";

import { CircleMarker, MapContainer, Popup, TileLayer, useMap } from "react-leaflet";
import { useEffect } from "react";
import "leaflet/dist/leaflet.css";
import { Cluster, Ward } from "@/types";

const center: [number, number] = [22.7196, 75.8577];

function MapFocus({ cluster }: { cluster: Cluster | null }) {
  const map = useMap();
  useEffect(() => {
    if (cluster) map.flyTo([cluster.centroid_lat, cluster.centroid_lng], Math.max(map.getZoom(), 13), { duration: 0.7 });
  }, [cluster, map]);
  return null;
}

function markerPosition(cluster: Cluster, index: number, clusters: Cluster[]): [number, number] {
  const siblings = clusters.filter((item) => item.ward_id === cluster.ward_id);
  if (siblings.length < 2) return [cluster.centroid_lat, cluster.centroid_lng];
  const siblingIndex = siblings.findIndex((item) => item.id === cluster.id);
  const angle = (2 * Math.PI * siblingIndex) / siblings.length + index * 0.01;
  return [cluster.centroid_lat + Math.cos(angle) * 0.0018, cluster.centroid_lng + Math.sin(angle) * 0.0018];
}

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
        <MapFocus cluster={selectedCluster} />
        <TileLayer
          attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
          url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
        />
        {clusters.map((cluster, index) => (
          <CircleMarker
            key={`cluster-marker-${cluster.id}`}
            center={markerPosition(cluster, index, clusters)}
            pathOptions={{
              className: selectedCluster?.id === cluster.id ? "live-marker-pulse" : "",
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
