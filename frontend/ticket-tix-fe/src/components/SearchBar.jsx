import { useState } from "react";
import { Label, Btn, Spinner } from "./ui/index.jsx";

export function SearchBar({ onSearch, loading }) {
  const [name, setName]     = useState("");
  const [loc, setLoc]       = useState("");
  const [from, setFrom]     = useState("");
  const [to, setTo]         = useState("");
  const [expanded, setExp]  = useState(false);

  const hasFilters = name || loc || from || to;

  const doSearch = () =>
    onSearch({
      eventName: name,
      location: loc,
      startDate: from ? new Date(from).toISOString() : "",
      endDate: to ? new Date(to).toISOString() : "",
    });

  const clear = () => {
    setName(""); setLoc(""); setFrom(""); setTo("");
    onSearch({});
  };

  return (
    <div className="search-bar">
      <div className="search-bar__row">
        <div>
          <Label>Event Name</Label>
          <input
            placeholder="Search events..."
            value={name}
            onChange={(e) => setName(e.target.value)}
            onKeyDown={(e) => e.key === "Enter" && doSearch()}
          />
        </div>
        <div>
          <Label>Location</Label>
          <input
            placeholder="City or venue..."
            value={loc}
            onChange={(e) => setLoc(e.target.value)}
            onKeyDown={(e) => e.key === "Enter" && doSearch()}
          />
        </div>
        <Btn onClick={() => setExp((x) => !x)} variant="outline" style={{ height: 42, gap: 6 }}>
          <svg width="13" height="13" viewBox="0 0 14 14" fill="none">
            <rect x="1" y="2" width="12" height="11" rx="1.5" stroke="currentColor" strokeWidth="1.2" />
            <path d="M4 1v2M10 1v2M1 6h12" stroke="currentColor" strokeWidth="1.2" strokeLinecap="round" />
          </svg>
          Dates {expanded ? "▲" : "▼"}
        </Btn>
        <Btn onClick={doSearch} disabled={loading} style={{ height: 42 }}>
          {loading ? (
            <Spinner size={14} color="#000" />
          ) : (
            <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
              <circle cx="6" cy="6" r="4.5" stroke="currentColor" strokeWidth="1.5" />
              <path d="M9.5 9.5L13 13" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
            </svg>
          )}
          Search
        </Btn>
      </div>

      {expanded && (
        <div className="search-bar__dates">
          <div>
            <Label>From Date</Label>
            <input type="date" value={from} onChange={(e) => setFrom(e.target.value)} />
          </div>
          <div>
            <Label>To Date</Label>
            <input type="date" value={to} onChange={(e) => setTo(e.target.value)} />
          </div>
        </div>
      )}

      {hasFilters && (
        <div className="search-bar__active-filters">
          <span className="search-bar__filter-label">ACTIVE:</span>
          {[
            name && `"${name}"`,
            loc,
            from && `From ${from}`,
            to && `To ${to}`,
          ]
            .filter(Boolean)
            .map((t, i) => (
              <span key={i} className="search-bar__filter-tag">{t}</span>
            ))}
          <button className="search-bar__clear-btn" onClick={clear}>
            Clear ×
          </button>
        </div>
      )}
    </div>
  );
}

