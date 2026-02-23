import { useState, useEffect, useCallback, useRef, createContext, useContext } from "react";

// ─── GLOBAL STYLES ────────────────────────────────────────────────────────────
const STYLES = `
@import url('https://fonts.googleapis.com/css2?family=Cormorant+Garamond:ital,wght@0,300;0,400;0,600;0,700;1,300;1,400&family=Outfit:wght@300;400;500;600&display=swap');
*,*::before,*::after{box-sizing:border-box;margin:0;padding:0;}
:root{
  --black:#080808;--ink:#0f0f0f;--surface:#141414;--raised:#1c1c1c;
  --card:#181818;--border:#272727;--border2:#333;--muted:#4a4a4a;
  --soft:#888880;--text:#e2ddd5;--light:#f0ebe0;
  --gold:#c8a84b;--gold2:#e2c06a;--amber:#d4751a;--red:#c0392b;--green:#27ae60;
}
body{background:var(--black);color:var(--text);font-family:'Outfit',sans-serif;font-weight:300;-webkit-font-smoothing:antialiased;}
::-webkit-scrollbar{width:3px;}::-webkit-scrollbar-track{background:var(--black);}::-webkit-scrollbar-thumb{background:var(--border2);border-radius:2px;}
.display{font-family:'Cormorant Garamond',serif;font-weight:600;letter-spacing:-0.02em;line-height:1.1;}
@keyframes fadeUp{from{opacity:0;transform:translateY(20px)}to{opacity:1;transform:translateY(0)}}
@keyframes fadeIn{from{opacity:0}to{opacity:1}}
@keyframes scaleIn{from{opacity:0;transform:scale(0.97)}to{opacity:1;transform:scale(1)}}
@keyframes spin{to{transform:rotate(360deg)}}
@keyframes shimmer{0%{background-position:-400px 0}100%{background-position:400px 0}}
@keyframes toastIn{from{opacity:0;transform:translateX(12px)}to{opacity:1;transform:translateX(0)}}
@keyframes slideDown{from{opacity:0;transform:translateY(-6px)}to{opacity:1;transform:translateY(0)}}
.anim-fade-up{animation:fadeUp 0.5s cubic-bezier(0.22,1,0.36,1) both;}
.anim-fade-in{animation:fadeIn 0.3s ease both;}
.stagger>*:nth-child(1){animation-delay:0.04s}.stagger>*:nth-child(2){animation-delay:0.08s}
.stagger>*:nth-child(3){animation-delay:0.12s}.stagger>*:nth-child(4){animation-delay:0.16s}
.stagger>*:nth-child(5){animation-delay:0.20s}.stagger>*:nth-child(6){animation-delay:0.24s}
.stagger>*:nth-child(7){animation-delay:0.28s}.stagger>*:nth-child(8){animation-delay:0.32s}
.stagger>*:nth-child(9){animation-delay:0.36s}.stagger>*:nth-child(10){animation-delay:0.40s}
.stagger>*:nth-child(11){animation-delay:0.44s}.stagger>*:nth-child(12){animation-delay:0.48s}
.skeleton{background:linear-gradient(90deg,var(--surface) 25%,var(--raised) 50%,var(--surface) 75%);background-size:400px 100%;animation:shimmer 1.4s ease infinite;border-radius:4px;}
input,select,textarea{background:var(--surface);border:1px solid var(--border);color:var(--text);font-family:'Outfit',sans-serif;font-size:14px;font-weight:300;padding:11px 14px;border-radius:6px;outline:none;width:100%;transition:border-color 0.2s,box-shadow 0.2s;letter-spacing:0.01em;}
input:focus,select:focus,textarea:focus{border-color:var(--gold);box-shadow:0 0 0 3px rgba(200,168,75,0.08);}
input::placeholder,textarea::placeholder{color:var(--muted);}
textarea{resize:vertical;min-height:90px;line-height:1.5;}
button{cursor:pointer;font-family:'Outfit',sans-serif;}
body::before{content:'';position:fixed;inset:0;pointer-events:none;opacity:0.02;background-image:url("data:image/svg+xml,%3Csvg viewBox='0 0 256 256' xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='n'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.9' numOctaves='4' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23n)'/%3E%3C/svg%3E");background-size:128px;z-index:999;}
`;

const API_BASE = "http://localhost:50061";

// ─── API ──────────────────────────────────────────────────────────────────────
async function apiFetch(path, opts = {}) {
  const res = await fetch(`${API_BASE}${path}`, opts);
  if (!res.ok) { const b = await res.json().catch(() => ({})); throw new Error(b.error || `HTTP ${res.status}`); }
  if (res.status === 204) return null;
  return res.json();
}
const api = {
  browse: (p) => { const q = new URLSearchParams(); Object.entries(p).forEach(([k,v]) => v && q.set(k,v)); return apiFetch(`/events?${q}`); },
  detail: (id) => apiFetch(`/event/${id}`),
  create: (fd) => apiFetch("/events", { method: "POST", body: fd }),
  uploadImages: (id, fd) => apiFetch(`/events/${id}/images`, { method: "POST", body: fd }),
  deleteImage: (eid, iid) => apiFetch(`/events/${eid}/images/${iid}`, { method: "DELETE" }),
  createCategory: (eid, data) => apiFetch(`/events/${eid}/categories`, { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify(data) }),
  deleteCategory: (eid, cid) => apiFetch(`/events/${eid}/categories/${cid}`, { method: "DELETE" }),
};

// ─── TOAST ────────────────────────────────────────────────────────────────────
const ToastCtx = createContext(null);
let _toastId = 0;
function ToastProvider({ children }) {
  const [toasts, setToasts] = useState([]);
  const show = useCallback((message, type = "success") => {
    const id = ++_toastId;
    setToasts(t => [...t, { id, message, type }]);
    setTimeout(() => setToasts(t => t.filter(x => x.id !== id)), 3500);
  }, []);
  return (
    <ToastCtx.Provider value={show}>
      {children}
      <div style={{ position: "fixed", bottom: 24, right: 24, zIndex: 9999, display: "flex", flexDirection: "column", gap: 10 }}>
        {toasts.map(t => (
          <div key={t.id} onClick={() => setToasts(x => x.filter(i => i.id !== t.id))} style={{ background: t.type === "error" ? "#1e0808" : "#081e0e", border: `1px solid ${t.type === "error" ? "rgba(192,57,43,0.5)" : "rgba(39,174,96,0.4)"}`, color: t.type === "error" ? "#e57373" : "#81c995", padding: "12px 18px", borderRadius: 8, fontSize: 13, cursor: "pointer", animation: "toastIn 0.3s ease both", boxShadow: "0 8px 32px rgba(0,0,0,0.5)", minWidth: 220, maxWidth: 320 }}>
            <span style={{ marginRight: 8 }}>{t.type === "error" ? "✕" : "✓"}</span>{t.message}
          </div>
        ))}
      </div>
    </ToastCtx.Provider>
  );
}
const useToast = () => useContext(ToastCtx);

// ─── UTILS ────────────────────────────────────────────────────────────────────
const fmt = {
  date: d => new Date(d).toLocaleDateString("en-GB", { day: "numeric", month: "short", year: "numeric" }),
  time: d => new Date(d).toLocaleTimeString("en-GB", { hour: "2-digit", minute: "2-digit" }),
  currency: v => Number(v).toLocaleString("id-ID", { style: "currency", currency: "IDR", maximumFractionDigits: 0 }),
};

// ─── UI PRIMITIVES ────────────────────────────────────────────────────────────
function Spinner({ size = 18, color = "var(--gold)" }) {
  return <div style={{ width: size, height: size, border: "2px solid var(--border)", borderTopColor: color, borderRadius: "50%", animation: "spin 0.7s linear infinite", flexShrink: 0 }} />;
}
function Skeleton({ w = "100%", h = 16, radius = 4, style: s = {} }) {
  return <div className="skeleton" style={{ width: w, height: h, borderRadius: radius, flexShrink: 0, ...s }} />;
}
function Label({ children }) {
  return <label style={{ display: "block", fontSize: 11, fontWeight: 500, letterSpacing: "0.1em", textTransform: "uppercase", color: "var(--soft)", marginBottom: 6 }}>{children}</label>;
}
function FieldGroup({ label, children, error, hint }) {
  return (
    <div style={{ display: "flex", flexDirection: "column", gap: 6 }}>
      {label && <Label>{label}</Label>}
      {children}
      {hint && !error && <span style={{ fontSize: 12, color: "var(--muted)" }}>{hint}</span>}
      {error && <span style={{ fontSize: 12, color: "var(--red)" }}>{error}</span>}
    </div>
  );
}
function GoldLine() { return <div style={{ width: 32, height: 2, background: "var(--gold)", marginBottom: 12 }} />; }
function Divider({ m = "24px 0" }) { return <hr style={{ border: "none", borderTop: "1px solid var(--border)", margin: m }} />; }
function BackBtn({ onClick }) {
  return (
    <button onClick={onClick} style={{ background: "none", border: "none", color: "var(--soft)", fontSize: 13, display: "inline-flex", alignItems: "center", gap: 8, padding: "4px 0", transition: "color 0.2s" }}
      onMouseEnter={e => e.currentTarget.style.color = "var(--text)"} onMouseLeave={e => e.currentTarget.style.color = "var(--soft)"}>
      <span style={{ fontSize: 18 }}>←</span> Back to Events
    </button>
  );
}
function Btn({ children, onClick, disabled, variant = "gold", size = "md", style: s = {} }) {
  const base = { display: "inline-flex", alignItems: "center", justifyContent: "center", gap: 8, fontFamily: "'Outfit',sans-serif", fontWeight: 500, fontSize: size === "sm" ? 12 : 13, letterSpacing: "0.06em", textTransform: "uppercase", padding: size === "sm" ? "7px 14px" : "10px 22px", borderRadius: 6, border: "none", cursor: "pointer", transition: "all 0.2s", whiteSpace: "nowrap", opacity: disabled ? 0.45 : 1, ...s };
  const variants = {
    gold:    { background: "var(--gold)", color: "var(--black)" },
    outline: { background: "transparent", color: "var(--text)", border: "1px solid var(--border2)" },
    danger:  { background: "transparent", color: "var(--red)", border: "1px solid rgba(192,57,43,0.3)" },
    ghost:   { background: "transparent", color: "var(--soft)", border: "none" },
  };
  return <button onClick={onClick} disabled={disabled} style={{ ...base, ...variants[variant] }}>{children}</button>;
}
function EmptyState({ icon = "◎", title, subtitle }) {
  return (
    <div style={{ textAlign: "center", padding: "60px 24px" }}>
      <div style={{ fontSize: 36, marginBottom: 14, opacity: 0.25 }}>{icon}</div>
      <div style={{ fontSize: 18, fontFamily: "'Cormorant Garamond',serif", marginBottom: 8 }}>{title}</div>
      {subtitle && <div style={{ fontSize: 13, color: "var(--muted)" }}>{subtitle}</div>}
    </div>
  );
}
function CardSkeleton() {
  return <div style={{ background: "var(--card)", border: "1px solid var(--border)", borderRadius: 12, overflow: "hidden" }}><Skeleton h={200} radius={0} /><div style={{ padding: 20 }}><Skeleton h={11} w="40%" s={{ marginBottom: 12 }} /><Skeleton h={22} w="80%" s={{ marginBottom: 8 }} /><Skeleton h={14} w="55%" /></div></div>;
}

// ─── HOOKS ────────────────────────────────────────────────────────────────────
function useForm(init) {
  const [values, setValues] = useState(init);
  const [errors, setErrors] = useState({});
  const set = (f, v) => { setValues(p => ({ ...p, [f]: v })); setErrors(p => ({ ...p, [f]: undefined })); };
  const reset = () => { setValues(init); setErrors({}); };
  const validate = (rules) => {
    const e = {};
    for (const [f, r] of Object.entries(rules)) { const m = r(values[f], values); if (m) e[f] = m; }
    setErrors(e); return Object.keys(e).length === 0;
  };
  return { values, errors, set, reset, validate };
}

function useBrowse() {
  const [events, setEvents]       = useState([]);
  const [loading, setLoading]     = useState(false);
  const [loadingMore, setMore]    = useState(false);
  const [hasMore, setHasMore]     = useState(false);
  const [cursor, setCursor]       = useState("");
  const [error, setError]         = useState(null);
  const [filters, setFilters]     = useState({});

  const doFetch = useCallback(async (f, cur) => {
    try {
      cur === "" ? setLoading(true) : setMore(true);
      setError(null);
      const res = await api.browse({ event_name: f.eventName, location: f.location, start_date: f.startDate, end_date: f.endDate, cursor: cur, limit: 12 });
      cur === "" ? setEvents(res.events || []) : setEvents(p => [...p, ...(res.events || [])]);
      setHasMore(res.has_more); setCursor(res.next_cursor || "");
    } catch (e) { setError(e.message); }
    finally { setLoading(false); setMore(false); }
  }, []);

  useEffect(() => { doFetch({}, ""); }, []);

  const search = useCallback((f) => { setFilters(f); setCursor(""); doFetch(f, ""); }, [doFetch]);
  const loadMore = useCallback(() => { if (hasMore && !loadingMore) doFetch(filters, cursor); }, [hasMore, loadingMore, filters, cursor, doFetch]);
  return { events, loading, loadingMore, hasMore, error, search, loadMore };
}

// ─── NAV ─────────────────────────────────────────────────────────────────────
function Nav({ page, navigate }) {
  return (
    <nav style={{ position: "fixed", top: 0, left: 0, right: 0, zIndex: 50, height: 64, background: "rgba(8,8,8,0.92)", backdropFilter: "blur(12px)", borderBottom: "1px solid var(--border)", display: "flex", alignItems: "center", justifyContent: "space-between", padding: "0 32px" }}>
      <button onClick={() => navigate("browse")} style={{ background: "none", border: "none", cursor: "pointer", display: "flex", alignItems: "center", gap: 10 }}>
        <div style={{ width: 28, height: 28, background: "var(--gold)", borderRadius: 4, display: "flex", alignItems: "center", justifyContent: "center" }}>
          <svg width="14" height="14" viewBox="0 0 14 14" fill="none"><rect x="1" y="3" width="12" height="8" rx="1" stroke="#080808" strokeWidth="1.5"/><path d="M4 3V2M7 3V2M10 3V2" stroke="#080808" strokeWidth="1.5" strokeLinecap="round"/><path d="M4 7h6M4 9.5h4" stroke="#080808" strokeWidth="1.2" strokeLinecap="round"/></svg>
        </div>
        <span style={{ fontFamily: "'Cormorant Garamond',serif", fontSize: 20, fontWeight: 600, color: "var(--light)", letterSpacing: "-0.01em" }}>Ticket<span style={{ color: "var(--gold)" }}>Tix</span></span>
      </button>
      <div style={{ display: "flex", gap: 4 }}>
        {["browse", "admin"].map(p => (
          <button key={p} onClick={() => navigate(p)} style={{ background: page === p ? "rgba(200,168,75,0.1)" : "transparent", border: `1px solid ${page === p ? "rgba(200,168,75,0.25)" : "transparent"}`, color: page === p ? "var(--gold)" : "var(--soft)", fontFamily: "'Outfit',sans-serif", fontSize: 13, fontWeight: 500, letterSpacing: "0.06em", textTransform: "uppercase", padding: "6px 16px", borderRadius: 6, cursor: "pointer", transition: "all 0.2s" }}>
            {p}
          </button>
        ))}
      </div>
    </nav>
  );
}

// ─── EVENT CARD ───────────────────────────────────────────────────────────────
function EventCard({ event, onClick, style: s = {} }) {
  const img = event.images?.[0]?.image_url || null;
  const [hovered, setHovered] = useState(false);
  return (
    <article onClick={onClick} className="anim-fade-up"
      onMouseEnter={() => setHovered(true)} onMouseLeave={() => setHovered(false)}
      style={{ background: "var(--card)", border: `1px solid ${hovered ? "var(--border2)" : "var(--border)"}`, borderRadius: 12, overflow: "hidden", cursor: "pointer", transform: hovered ? "translateY(-4px)" : "translateY(0)", boxShadow: hovered ? "0 16px 48px rgba(0,0,0,0.4)" : "none", transition: "all 0.25s cubic-bezier(0.22,1,0.36,1)", ...s }}>
      <div style={{ position: "relative", aspectRatio: "16/9", background: "var(--surface)", overflow: "hidden" }}>
        {img
          ? <img src={img} alt={event.name} style={{ width: "100%", height: "100%", objectFit: "cover", transform: hovered ? "scale(1.04)" : "scale(1)", transition: "transform 0.4s" }} />
          : <div style={{ width: "100%", height: "100%", background: "linear-gradient(135deg,var(--surface),var(--raised))", display: "flex", alignItems: "center", justifyContent: "center" }}><span style={{ opacity: 0.2, fontSize: 32 }}>🎭</span></div>
        }
        <div style={{ position: "absolute", top: 12, left: 12, background: "rgba(8,8,8,0.88)", backdropFilter: "blur(8px)", border: "1px solid rgba(200,168,75,0.2)", borderRadius: 6, padding: "5px 10px", minWidth: 44, textAlign: "center" }}>
          <div style={{ fontSize: 16, fontFamily: "'Cormorant Garamond',serif", fontWeight: 700, color: "var(--gold)", lineHeight: 1 }}>{new Date(event.start_time).getDate()}</div>
          <div style={{ fontSize: 9, fontWeight: 500, letterSpacing: "0.1em", textTransform: "uppercase", color: "var(--soft)", marginTop: 1 }}>{new Date(event.start_time).toLocaleString("en", { month: "short" })}</div>
        </div>
      </div>
      <div style={{ padding: "18px 20px 20px" }}>
        <div style={{ fontSize: 10, fontWeight: 500, letterSpacing: "0.1em", textTransform: "uppercase", color: "var(--gold)", marginBottom: 8 }}>{fmt.time(event.start_time)} · {new Date(event.start_time).getFullYear()}</div>
        <h3 style={{ fontFamily: "'Cormorant Garamond',serif", fontSize: 20, fontWeight: 600, color: "var(--light)", lineHeight: 1.2, marginBottom: 10, letterSpacing: "-0.01em" }}>{event.name}</h3>
        <div style={{ display: "flex", alignItems: "center", gap: 6, color: "var(--soft)", fontSize: 12 }}>
          <svg width="11" height="13" viewBox="0 0 11 13" fill="none"><path d="M5.5 1C3.015 1 1 3.015 1 5.5c0 3.375 4.5 7.5 4.5 7.5s4.5-4.125 4.5-7.5C10 3.015 7.985 1 5.5 1z" stroke="currentColor" strokeWidth="1.2"/><circle cx="5.5" cy="5.5" r="1.5" stroke="currentColor" strokeWidth="1.2"/></svg>
          {event.location}
        </div>
      </div>
    </article>
  );
}

// ─── SEARCH BAR ───────────────────────────────────────────────────────────────
function SearchBar({ onSearch, loading }) {
  const [name, setName]       = useState("");
  const [loc, setLoc]         = useState("");
  const [from, setFrom]       = useState("");
  const [to, setTo]           = useState("");
  const [expanded, setExp]    = useState(false);
  const hasFilters = name || loc || from || to;

  const doSearch = () => onSearch({ eventName: name, location: loc, startDate: from ? new Date(from).toISOString() : "", endDate: to ? new Date(to).toISOString() : "" });
  const clear = () => { setName(""); setLoc(""); setFrom(""); setTo(""); onSearch({}); };

  return (
    <div style={{ background: "var(--surface)", border: "1px solid var(--border)", borderRadius: 12, padding: "20px 24px", marginBottom: 40 }}>
      <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr auto auto", gap: 12, alignItems: "end" }}>
        <div><Label>Event Name</Label><input placeholder="Search events..." value={name} onChange={e => setName(e.target.value)} onKeyDown={e => e.key === "Enter" && doSearch()} /></div>
        <div><Label>Location</Label><input placeholder="City or venue..." value={loc} onChange={e => setLoc(e.target.value)} onKeyDown={e => e.key === "Enter" && doSearch()} /></div>
        <Btn onClick={() => setExp(x => !x)} variant="outline" style={{ height: 42, gap: 6 }}>
          <svg width="13" height="13" viewBox="0 0 14 14" fill="none"><rect x="1" y="2" width="12" height="11" rx="1.5" stroke="currentColor" strokeWidth="1.2"/><path d="M4 1v2M10 1v2M1 6h12" stroke="currentColor" strokeWidth="1.2" strokeLinecap="round"/></svg>
          Dates {expanded ? "▲" : "▼"}
        </Btn>
        <Btn onClick={doSearch} disabled={loading} style={{ height: 42 }}>
          {loading ? <Spinner size={14} color="#000" /> : <svg width="14" height="14" viewBox="0 0 14 14" fill="none"><circle cx="6" cy="6" r="4.5" stroke="currentColor" strokeWidth="1.5"/><path d="M9.5 9.5L13 13" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round"/></svg>}
          Search
        </Btn>
      </div>
      {expanded && (
        <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 12, marginTop: 16, animation: "slideDown 0.2s ease" }}>
          <div><Label>From Date</Label><input type="date" value={from} onChange={e => setFrom(e.target.value)} /></div>
          <div><Label>To Date</Label><input type="date" value={to} onChange={e => setTo(e.target.value)} /></div>
        </div>
      )}
      {hasFilters && (
        <div style={{ marginTop: 14, display: "flex", alignItems: "center", gap: 10, flexWrap: "wrap" }}>
          <span style={{ fontSize: 11, color: "var(--soft)", letterSpacing: "0.05em" }}>ACTIVE:</span>
          {[name && `"${name}"`, loc, from && `From ${from}`, to && `To ${to}`].filter(Boolean).map((t, i) => (
            <span key={i} style={{ fontSize: 11, background: "rgba(200,168,75,0.1)", color: "var(--gold)", border: "1px solid rgba(200,168,75,0.2)", borderRadius: 100, padding: "2px 10px" }}>{t}</span>
          ))}
          <button onClick={clear} style={{ background: "none", border: "none", color: "var(--muted)", fontSize: 11, cursor: "pointer" }}>Clear ×</button>
        </div>
      )}
    </div>
  );
}

// ─── BROWSE PAGE ──────────────────────────────────────────────────────────────
function BrowsePage({ onEventClick }) {
  const { events, loading, loadingMore, hasMore, error, search, loadMore } = useBrowse();
  return (
    <div style={{ maxWidth: 1200, margin: "0 auto", padding: "48px 32px" }}>
      <div style={{ marginBottom: 32 }}>
        <GoldLine />
        <h2 className="display" style={{ fontSize: 38, color: "var(--light)" }}>Upcoming Events</h2>
        <p style={{ marginTop: 6, color: "var(--soft)", fontSize: 14 }}>Discover live experiences near you</p>
      </div>
      <SearchBar onSearch={search} loading={loading} />
      {error && <div style={{ background: "rgba(192,57,43,0.1)", border: "1px solid rgba(192,57,43,0.3)", borderRadius: 8, padding: "14px 18px", marginBottom: 24, color: "#e57373", fontSize: 13 }}>{error}</div>}
      {loading ? (
        <div className="stagger" style={{ display: "grid", gridTemplateColumns: "repeat(auto-fill, minmax(300px,1fr))", gap: 24 }}>
          {Array.from({ length: 6 }).map((_, i) => <CardSkeleton key={i} />)}
        </div>
      ) : events.length === 0 ? (
        <EmptyState icon="◎" title="No events found" subtitle="Try adjusting your filters" />
      ) : (
        <>
          <div className="stagger" style={{ display: "grid", gridTemplateColumns: "repeat(auto-fill, minmax(300px,1fr))", gap: 24 }}>
            {events.map((e, i) => <EventCard key={e.id} event={e} onClick={() => onEventClick(e.id)} style={{ animationDelay: `${(i % 12) * 0.05}s` }} />)}
          </div>
          {hasMore && (
            <div style={{ textAlign: "center", marginTop: 48 }}>
              <Btn onClick={loadMore} disabled={loadingMore} variant="outline" style={{ minWidth: 160 }}>
                {loadingMore && <Spinner size={14} />} {loadingMore ? "Loading..." : "Load More"}
              </Btn>
            </div>
          )}
          {!hasMore && events.length > 0 && <div style={{ textAlign: "center", marginTop: 48, color: "var(--muted)", fontSize: 11, letterSpacing: "0.12em" }}>— END OF RESULTS —</div>}
        </>
      )}
    </div>
  );
}

// ─── EVENT DETAIL PAGE ────────────────────────────────────────────────────────
function EventDetailPage({ eventId, onBack }) {
  const [event, setEvent]     = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError]     = useState(null);
  const [activeImg, setActive]= useState(0);

  useEffect(() => {
    setLoading(true); setError(null);
    api.detail(eventId)
      .then(d => { setEvent(d); setLoading(false); })
      .catch(e => { setError(e.message); setLoading(false); });
  }, [eventId]);

  if (loading) return (
    <div style={{ maxWidth: 1100, margin: "0 auto", padding: "48px 32px" }}>
      <BackBtn onClick={onBack} />
      <div style={{ marginTop: 32, display: "grid", gridTemplateColumns: "1fr 360px", gap: 48 }}>
        <div><Skeleton h={440} radius={12} s={{ marginBottom: 16 }} /><Skeleton h={14} w="30%" s={{ marginBottom: 16 }} /><Skeleton h={48} w="70%" s={{ marginBottom: 12 }} /><Skeleton h={14} w="50%" /></div>
        <Skeleton h={300} radius={12} />
      </div>
    </div>
  );

  if (error || !event) return (
    <div style={{ maxWidth: 800, margin: "0 auto", padding: "48px 32px" }}>
      <BackBtn onClick={onBack} />
      <div style={{ marginTop: 24, color: "#e57373" }}>{error || "Event not found"}</div>
    </div>
  );

  const images = event.images || [];
  const cats   = event.categories || [];
  const img    = images[activeImg]?.image_url || null;

  return (
    <div style={{ maxWidth: 1100, margin: "0 auto", padding: "48px 32px" }} className="anim-fade-in">
      <BackBtn onClick={onBack} />
      <div style={{ marginTop: 32, display: "grid", gridTemplateColumns: "1fr 360px", gap: 48, alignItems: "start" }}>
        {/* Left */}
        <div>
          <div style={{ borderRadius: 12, overflow: "hidden", aspectRatio: "16/9", background: "var(--surface)", marginBottom: 14 }}>
            {img
              ? <img src={img} alt={event.name} style={{ width: "100%", height: "100%", objectFit: "cover" }} />
              : <div style={{ width: "100%", height: "100%", display: "flex", alignItems: "center", justifyContent: "center", opacity: 0.15, fontSize: 60 }}>🎭</div>
            }
          </div>
          {images.length > 1 && (
            <div style={{ display: "flex", gap: 8, marginBottom: 28, flexWrap: "wrap" }}>
              {images.map((im, i) => (
                <button key={i} onClick={() => setActive(i)} style={{ width: 72, height: 48, borderRadius: 6, overflow: "hidden", border: `2px solid ${i === activeImg ? "var(--gold)" : "var(--border)"}`, padding: 0, cursor: "pointer", transition: "border-color 0.2s" }}>
                  <img src={im.image_url} alt="" style={{ width: "100%", height: "100%", objectFit: "cover" }} />
                </button>
              ))}
            </div>
          )}
          <GoldLine />
          <h1 className="display" style={{ fontSize: 40, color: "var(--light)", marginBottom: 20 }}>{event.name}</h1>
          <div style={{ display: "flex", flexDirection: "column", gap: 10, marginBottom: 24 }}>
            {[
              { icon: "📅", text: `${fmt.date(event.start_time)} · ${fmt.time(event.start_time)}` },
              { icon: "🕐", text: `Ends at ${fmt.time(event.end_time)}` },
              { icon: "📍", text: event.location },
            ].map(({ icon, text }) => (
              <div key={text} style={{ display: "flex", alignItems: "center", gap: 10, color: "var(--soft)", fontSize: 14 }}>
                <span style={{ fontSize: 14 }}>{icon}</span><span>{text}</span>
              </div>
            ))}
          </div>
          {event.description && (
            <>
              <Divider />
              <Label>About this event</Label>
              <p style={{ color: "var(--text)", lineHeight: 1.75, fontSize: 15, fontWeight: 300, marginTop: 8 }}>{event.description}</p>
            </>
          )}
        </div>

        {/* Right — sticky ticket panel */}
        <div style={{ position: "sticky", top: 88 }}>
          <div style={{ background: "var(--surface)", border: "1px solid var(--border)", borderRadius: 12, padding: 24 }}>
            <h3 style={{ fontFamily: "'Cormorant Garamond',serif", fontSize: 22, color: "var(--light)", marginBottom: 20 }}>Ticket Categories</h3>
            {cats.length === 0
              ? <p style={{ color: "var(--muted)", fontSize: 13 }}>No categories available yet</p>
              : <div style={{ display: "flex", flexDirection: "column", gap: 12 }}>
                {cats.map(cat => {
                  const avail = cat.available_capacity ?? cat.available_stock ?? 0;
                  const total = cat.total_capacity ?? 0;
                  const pct   = total > 0 ? (avail / total) * 100 : 0;
                  const clr   = pct > 50 ? "var(--green)" : pct > 20 ? "var(--gold)" : "var(--amber)";
                  return (
                    <div key={cat.category_id ?? cat.id} style={{ border: "1px solid var(--border)", borderRadius: 8, padding: 16 }}>
                      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start", marginBottom: 12 }}>
                        <div>
                          <div style={{ fontWeight: 500, fontSize: 15, color: "var(--light)", marginBottom: 3 }}>{cat.name}</div>
                          <div style={{ fontSize: 11, color: "var(--soft)", letterSpacing: "0.06em", textTransform: "uppercase" }}>{cat.category_type}</div>
                        </div>
                        <div style={{ fontFamily: "'Cormorant Garamond',serif", fontSize: 20, fontWeight: 600, color: "var(--gold)", textAlign: "right", lineHeight: 1 }}>{fmt.currency(cat.price)}</div>
                      </div>
                      <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
                        <div style={{ flex: 1, height: 3, background: "var(--border)", borderRadius: 2 }}>
                          <div style={{ height: "100%", width: `${pct}%`, background: clr, borderRadius: 2, transition: "width 0.5s" }} />
                        </div>
                        <span style={{ fontSize: 11, color: clr, fontWeight: 500, whiteSpace: "nowrap" }}>{avail}/{total}</span>
                      </div>
                    </div>
                  );
                })}
              </div>
            }
          </div>
        </div>
      </div>
    </div>
  );
}

// ─── ADMIN PAGE ───────────────────────────────────────────────────────────────
function AdminPage() {
  const [tab, setTab]         = useState(0);
  const [event, setEvent]     = useState(null);
  const TABS = ["1. Create Event", "2. Categories", "3. Images"];

  return (
    <div style={{ maxWidth: 860, margin: "0 auto", padding: "48px 32px" }}>
      <div style={{ marginBottom: 36 }}>
        <GoldLine />
        <h2 className="display" style={{ fontSize: 38, color: "var(--light)" }}>Admin Panel</h2>
        <p style={{ marginTop: 6, color: "var(--soft)", fontSize: 14 }}>Create and manage events</p>
      </div>

      {/* Tabs */}
      <div style={{ display: "flex", gap: 0, marginBottom: 40, borderBottom: "1px solid var(--border)" }}>
        {TABS.map((t, i) => (
          <button key={t} onClick={() => setTab(i)} disabled={i > 0 && !event}
            style={{ background: "none", border: "none", borderBottom: `2px solid ${tab === i ? "var(--gold)" : "transparent"}`, color: tab === i ? "var(--gold)" : (i > 0 && !event) ? "var(--muted)" : "var(--soft)", fontFamily: "'Outfit',sans-serif", fontWeight: 500, fontSize: 13, letterSpacing: "0.05em", textTransform: "uppercase", padding: "10px 20px 12px", cursor: (i > 0 && !event) ? "not-allowed" : "pointer", transition: "all 0.2s", marginBottom: -1 }}>
            {t}
          </button>
        ))}
      </div>

      {tab === 0 && <CreateEventTab onCreated={e => { setEvent(e); setTab(1); }} />}
      {tab === 1 && <CategoriesTab event={event} />}
      {tab === 2 && <ImagesTab event={event} />}
    </div>
  );
}

// ─── CREATE EVENT TAB ─────────────────────────────────────────────────────────
function CreateEventTab({ onCreated }) {
  const toast   = useToast();
  const fileRef = useRef();
  const [loading, setLoading]   = useState(false);
  const [files, setFiles]       = useState([]);
  const [previews, setPreviews] = useState([]);
  const { values, errors, set, validate } = useForm({ name: "", description: "", location: "", startTime: "", endTime: "" });

  const rules = {
    name:      v => !v?.trim() ? "Required" : null,
    location:  v => !v?.trim() ? "Required" : null,
    startTime: v => !v ? "Required" : null,
    endTime:   (v, a) => !v ? "Required" : (a.startTime && new Date(v) <= new Date(a.startTime)) ? "Must be after start" : null,
  };

  const onFiles = e => { const f = Array.from(e.target.files); setFiles(f); setPreviews(f.map(x => URL.createObjectURL(x))); };
  const rmFile  = i => { setFiles(p => p.filter((_,j) => j !== i)); setPreviews(p => p.filter((_,j) => j !== i)); };

  const submit = async () => {
    if (!validate(rules)) return;
    try {
      setLoading(true);
      const fd = new FormData();
      fd.append("name", values.name); fd.append("description", values.description);
      fd.append("location", values.location);
      fd.append("start_time", new Date(values.startTime).toISOString());
      fd.append("end_time",   new Date(values.endTime).toISOString());
      files.forEach(f => fd.append("images", f));
      const ev = await api.create(fd);
      toast("Event created!", "success");
      onCreated(ev);
    } catch (e) { toast(e.message, "error"); }
    finally { setLoading(false); }
  };

  return (
    <div className="anim-fade-in" style={{ display: "flex", flexDirection: "column", gap: 22 }}>
      <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 18 }}>
        <FieldGroup label="Event Name *" error={errors.name}><input placeholder="e.g. Jazz Night Jakarta" value={values.name} onChange={e => set("name", e.target.value)} /></FieldGroup>
        <FieldGroup label="Location *" error={errors.location}><input placeholder="e.g. Jakarta Convention Center" value={values.location} onChange={e => set("location", e.target.value)} /></FieldGroup>
      </div>
      <FieldGroup label="Description"><textarea placeholder="Describe your event..." value={values.description} onChange={e => set("description", e.target.value)} /></FieldGroup>
      <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 18 }}>
        <FieldGroup label="Start Date & Time *" error={errors.startTime}><input type="datetime-local" value={values.startTime} onChange={e => set("startTime", e.target.value)} /></FieldGroup>
        <FieldGroup label="End Date & Time *" error={errors.endTime}><input type="datetime-local" value={values.endTime} onChange={e => set("endTime", e.target.value)} /></FieldGroup>
      </div>

      {/* Image upload */}
      <FieldGroup label="Event Images" hint="First image becomes the cover photo">
        <div onClick={() => fileRef.current.click()} style={{ border: "2px dashed var(--border2)", borderRadius: 10, padding: "28px 24px", textAlign: "center", cursor: "pointer", transition: "border-color 0.2s" }}
          onMouseEnter={e => e.currentTarget.style.borderColor = "var(--gold)"} onMouseLeave={e => e.currentTarget.style.borderColor = "var(--border2)"}>
          <div style={{ fontSize: 28, opacity: 0.35, marginBottom: 8 }}>⊕</div>
          <div style={{ fontSize: 13, color: "var(--soft)" }}>Click to upload images</div>
          <div style={{ fontSize: 11, color: "var(--muted)", marginTop: 3 }}>JPG, PNG, WebP · Multiple allowed</div>
        </div>
        <input ref={fileRef} type="file" multiple accept="image/*" style={{ display: "none" }} onChange={onFiles} />
        {previews.length > 0 && (
          <div style={{ display: "flex", gap: 10, flexWrap: "wrap", marginTop: 10 }}>
            {previews.map((src, i) => (
              <div key={i} style={{ position: "relative", width: 88, height: 62 }}>
                <img src={src} alt="" style={{ width: "100%", height: "100%", objectFit: "cover", borderRadius: 6, border: i === 0 ? "2px solid var(--gold)" : "1px solid var(--border)" }} />
                {i === 0 && <span style={{ position: "absolute", top: 4, left: 4, background: "var(--gold)", color: "#000", fontSize: 7, fontWeight: 700, padding: "1px 4px", borderRadius: 2, letterSpacing: "0.05em" }}>COVER</span>}
                <button onClick={e => { e.stopPropagation(); rmFile(i); }} style={{ position: "absolute", top: 4, right: 4, background: "rgba(0,0,0,0.75)", border: "none", color: "#fff", width: 18, height: 18, borderRadius: "50%", fontSize: 10, cursor: "pointer" }}>×</button>
              </div>
            ))}
          </div>
        )}
      </FieldGroup>

      <div style={{ paddingTop: 4 }}>
        <Btn onClick={submit} disabled={loading} style={{ minWidth: 180 }}>
          {loading && <Spinner size={14} color="#000" />} {loading ? "Creating..." : "Create Event →"}
        </Btn>
      </div>
    </div>
  );
}

// ─── CATEGORIES TAB ───────────────────────────────────────────────────────────
function CategoriesTab({ event }) {
  const toast = useToast();
  const [loading, setLoading] = useState(false);
  const [cats, setCats]       = useState(event?.categories || []);
  const { values, errors, set, reset, validate } = useForm({ name: "", categoryType: "STANDING", price: "", bookType: "FIXED", totalCapacity: "" });

  const rules = {
    name:          v => !v?.trim() ? "Required" : null,
    price:         v => !v || isNaN(v) || Number(v) <= 0 ? "Enter a valid price" : null,
    totalCapacity: v => !v || isNaN(v) || Number(v) <= 0 ? "Enter a valid capacity" : null,
  };

  const add = async () => {
    if (!validate(rules)) return;
    try {
      setLoading(true);
      const cat = await api.createCategory(event.id, { name: values.name, category_type: values.categoryType, price: String(values.price), book_type: values.bookType, total_capacity: Number(values.totalCapacity), available_stock: Number(values.totalCapacity) });
      setCats(p => [...p, cat]); reset(); toast("Category added!", "success");
    } catch (e) { toast(e.message, "error"); }
    finally { setLoading(false); }
  };

  const del = async (id) => {
    try { await api.deleteCategory(event.id, id); setCats(p => p.filter(c => c.id !== id)); toast("Removed", "success"); }
    catch (e) { toast(e.message, "error"); }
  };

  return (
    <div className="anim-fade-in">
      <EventBadge event={event} />
      {cats.length > 0 && (
        <div style={{ marginBottom: 32 }}>
          <Label>Existing Categories</Label>
          <div style={{ display: "flex", flexDirection: "column", gap: 8, marginTop: 10 }}>
            {cats.map(cat => (
              <div key={cat.id} style={{ display: "flex", alignItems: "center", justifyContent: "space-between", background: "var(--surface)", border: "1px solid var(--border)", borderRadius: 8, padding: "12px 16px" }}>
                <div style={{ display: "flex", alignItems: "center", gap: 16, flexWrap: "wrap" }}>
                  <span style={{ fontWeight: 500, color: "var(--text)" }}>{cat.name}</span>
                  <span style={{ fontSize: 11, color: "var(--soft)", letterSpacing: "0.06em" }}>{cat.category_type}</span>
                  <span style={{ fontSize: 13, color: "var(--gold)", fontFamily: "'Cormorant Garamond',serif" }}>{fmt.currency(cat.price)}</span>
                  <span style={{ fontSize: 12, color: "var(--muted)" }}>{cat.total_capacity} seats</span>
                </div>
                <Btn variant="danger" size="sm" onClick={() => del(cat.id)}>Remove</Btn>
              </div>
            ))}
          </div>
          <Divider />
        </div>
      )}
      <Label style={{ marginBottom: 16 }}>Add New Category</Label>
      <div style={{ display: "flex", flexDirection: "column", gap: 18 }}>
        <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 16 }}>
          <FieldGroup label="Name *" error={errors.name}><input placeholder="e.g. VIP, STANDING A" value={values.name} onChange={e => set("name", e.target.value)} /></FieldGroup>
          <FieldGroup label="Type"><select value={values.categoryType} onChange={e => set("categoryType", e.target.value)}><option value="STANDING">Standing</option><option value="SEATED">Seated</option></select></FieldGroup>
        </div>
        <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr 1fr", gap: 16 }}>
          <FieldGroup label="Price (IDR) *" error={errors.price}><input type="number" placeholder="e.g. 250000" value={values.price} onChange={e => set("price", e.target.value)} /></FieldGroup>
          <FieldGroup label="Book Type"><select value={values.bookType} onChange={e => set("bookType", e.target.value)}><option value="FIXED">Fixed</option><option value="FLEXIBLE">Flexible</option></select></FieldGroup>
          <FieldGroup label="Total Capacity *" error={errors.totalCapacity}><input type="number" placeholder="e.g. 500" value={values.totalCapacity} onChange={e => set("totalCapacity", e.target.value)} /></FieldGroup>
        </div>
        <div><Btn onClick={add} disabled={loading} style={{ minWidth: 160 }}>{loading && <Spinner size={14} color="#000" />} {loading ? "Adding..." : "+ Add Category"}</Btn></div>
      </div>
    </div>
  );
}

// ─── IMAGES TAB ───────────────────────────────────────────────────────────────
function ImagesTab({ event }) {
  const toast   = useToast();
  const fileRef = useRef();
  const [loading, setLoading]   = useState(false);
  const [images, setImages]     = useState(event?.images || []);
  const [files, setFiles]       = useState([]);
  const [previews, setPreviews] = useState([]);

  const onFiles = e => { const f = Array.from(e.target.files); setFiles(f); setPreviews(f.map(x => URL.createObjectURL(x))); };

  const upload = async () => {
    if (!files.length) { toast("Select at least one image", "error"); return; }
    try {
      setLoading(true);
      const fd = new FormData();
      files.forEach(f => fd.append("images", f));
      await api.uploadImages(event.id, fd);
      const updated = await api.detail(event.id);
      setImages(updated.images || []); setFiles([]); setPreviews([]);
      toast("Images uploaded!", "success");
    } catch (e) { toast(e.message, "error"); }
    finally { setLoading(false); }
  };

  const del = async (img) => {
    try { await api.deleteImage(event.id, img.id); setImages(p => p.filter(x => x.id !== img.id)); toast("Deleted", "success"); }
    catch (e) { toast(e.message, "error"); }
  };

  return (
    <div className="anim-fade-in">
      <EventBadge event={event} />
      {images.length > 0 ? (
        <div style={{ marginBottom: 32 }}>
          <Label>Current Images ({images.length})</Label>
          <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fill, minmax(150px,1fr))", gap: 12, marginTop: 12 }}>
            {images.map(img => (
              <div key={img.id} style={{ position: "relative", borderRadius: 8, overflow: "hidden", aspectRatio: "4/3", border: `1px solid ${img.is_primary ? "var(--gold)" : "var(--border)"}` }}>
                <img src={img.image_url} alt="" style={{ width: "100%", height: "100%", objectFit: "cover" }} />
                {img.is_primary && <span style={{ position: "absolute", top: 8, left: 8, background: "var(--gold)", color: "#000", fontSize: 7, fontWeight: 700, padding: "2px 6px", borderRadius: 3, letterSpacing: "0.06em" }}>PRIMARY</span>}
                <button onClick={() => del(img)} style={{ position: "absolute", top: 8, right: 8, background: "rgba(192,57,43,0.85)", border: "none", color: "#fff", width: 26, height: 26, borderRadius: "50%", fontSize: 13, cursor: "pointer", display: "flex", alignItems: "center", justifyContent: "center" }}>×</button>
              </div>
            ))}
          </div>
          <Divider />
        </div>
      ) : <EmptyState icon="🖼" title="No images yet" subtitle="Upload below" />}

      <Label>Upload New Images</Label>
      <div onClick={() => fileRef.current.click()} style={{ border: "2px dashed var(--border2)", borderRadius: 10, padding: "28px 24px", textAlign: "center", cursor: "pointer", transition: "border-color 0.2s", margin: "12px 0 16px" }}
        onMouseEnter={e => e.currentTarget.style.borderColor = "var(--gold)"} onMouseLeave={e => e.currentTarget.style.borderColor = "var(--border2)"}>
        <div style={{ fontSize: 28, opacity: 0.35, marginBottom: 8 }}>⊕</div>
        <div style={{ fontSize: 13, color: "var(--soft)" }}>Click to select images</div>
      </div>
      <input ref={fileRef} type="file" multiple accept="image/*" style={{ display: "none" }} onChange={onFiles} />
      {previews.length > 0 && (
        <div style={{ display: "flex", gap: 10, flexWrap: "wrap", marginBottom: 16 }}>
          {previews.map((src, i) => <img key={i} src={src} alt="" style={{ width: 80, height: 56, objectFit: "cover", borderRadius: 6, border: "1px solid var(--border)" }} />)}
        </div>
      )}
      <Btn onClick={upload} disabled={loading || !files.length} style={{ minWidth: 160 }}>
        {loading && <Spinner size={14} color="#000" />} {loading ? "Uploading..." : `Upload${files.length ? ` (${files.length})` : ""}`}
      </Btn>
    </div>
  );
}

// ─── SHARED ───────────────────────────────────────────────────────────────────
function EventBadge({ event }) {
  if (!event) return null;
  return (
    <div style={{ background: "var(--surface)", border: "1px solid var(--border)", borderRadius: 8, padding: "12px 18px", marginBottom: 28, display: "flex", alignItems: "center", gap: 10 }}>
      <div style={{ width: 6, height: 6, background: "var(--gold)", borderRadius: "50%" }} />
      <span style={{ fontSize: 13, color: "var(--soft)" }}>Event:</span>
      <span style={{ fontSize: 14, color: "var(--text)", fontWeight: 500 }}>{event.name}</span>
      <span style={{ marginLeft: "auto", fontSize: 11, color: "var(--muted)" }}>ID #{event.id}</span>
    </div>
  );
}

// ─── APP ROOT ─────────────────────────────────────────────────────────────────
export default function App() {
  const [page, setPage]       = useState("browse");
  const [eventId, setEventId] = useState(null);

  const navigate = (to, id = null) => {
    setEventId(id); setPage(to);
    window.scrollTo({ top: 0, behavior: "smooth" });
  };

  return (
    <ToastProvider>
      <style>{STYLES}</style>
      <div style={{ minHeight: "100vh" }}>
        <Nav page={page} navigate={navigate} />
        <div style={{ paddingTop: 64 }}>
          {page === "browse" && <BrowsePage onEventClick={id => navigate("detail", id)} />}
          {page === "detail" && <EventDetailPage eventId={eventId} onBack={() => navigate("browse")} />}
          {page === "admin"  && <AdminPage />}
        </div>
      </div>
    </ToastProvider>
  );
}
