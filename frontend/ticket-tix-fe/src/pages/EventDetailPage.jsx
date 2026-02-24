import { useState, useEffect } from "react";
import api from "../api/api";
import fmt from "../utils/fmt";
import { BackBtn, GoldLine, Divider, Label, Skeleton } from "../components/ui/index.jsx";

export function EventDetailPage({ eventId, onBack }) {
  const [event, setEvent]     = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError]     = useState(null);
  const [activeImg, setActive]= useState(0);

  useEffect(() => {
    setLoading(true);
    setError(null);
    api
      .detail(eventId)
      .then((d) => { setEvent(d); setLoading(false); })
      .catch((e) => { setError(e.message); setLoading(false); });
  }, [eventId]);

  if (loading) {
    return (
      <div className="detail-page">
        <BackBtn onClick={onBack} />
        <div className="detail-page__layout">
          <div>
            <Skeleton h={440} radius={12} style={{ marginBottom: 16 }} />
            <Skeleton h={14} w="30%" style={{ marginBottom: 16 }} />
            <Skeleton h={48} w="70%" style={{ marginBottom: 12 }} />
            <Skeleton h={14} w="50%" />
          </div>
          <Skeleton h={300} radius={12} />
        </div>
      </div>
    );
  }

  if (error || !event) {
    return (
      <div className="detail-page">
        <BackBtn onClick={onBack} />
        <div style={{ marginTop: 24, color: "#e57373" }}>{error || "Event not found"}</div>
      </div>
    );
  }

  const images = event.images || [];
  const cats   = event.categories || [];
  const img    = images[activeImg]?.image_url || null;

  return (
    <div className="detail-page anim-fade-in">
      <BackBtn onClick={onBack} />
      <div className="detail-page__layout">
        {/* Left column */}
        <div>
          <div className="detail-page__main-image">
            {img ? (
              <img src={img} alt={event.name} />
            ) : (
              <div className="detail-page__no-image">🎭</div>
            )}
          </div>

          {images.length > 1 && (
            <div className="detail-page__thumbs">
              {images.map((im, i) => (
                <button
                  key={i}
                  onClick={() => setActive(i)}
                  className={`detail-page__thumb${i === activeImg ? " detail-page__thumb--active" : ""}`}
                >
                  <img src={im.image_url} alt="" />
                </button>
              ))}
            </div>
          )}

          <GoldLine />
          <h1 className="display detail-page__title">{event.name}</h1>

          <div className="detail-page__meta">
            {[
              { icon: "📅", text: `${fmt.date(event.start_time)} · ${fmt.time(event.start_time)}` },
              { icon: "🕐", text: `Ends at ${fmt.time(event.end_time)}` },
              { icon: "📍", text: event.location },
            ].map(({ icon, text }) => (
              <div key={text} className="detail-page__meta-item">
                <span>{icon}</span>
                <span>{text}</span>
              </div>
            ))}
          </div>

          {event.description && (
            <>
              <Divider />
              <Label>About this event</Label>
              <p className="detail-page__description">{event.description}</p>
            </>
          )}
        </div>

        {/* Right — sticky ticket panel */}
        <div className="ticket-panel">
          <div className="ticket-panel__card">
            <h3 className="ticket-panel__title">Ticket Categories</h3>
            {cats.length === 0 ? (
              <p className="ticket-panel__empty">No categories available yet</p>
            ) : (
              <div className="ticket-panel__list">
                {cats.map((cat) => {
                  const avail = cat.available_capacity ?? cat.available_stock ?? 0;
                  const total = cat.total_capacity ?? 0;
                  const pct   = total > 0 ? (avail / total) * 100 : 0;
                  const clr   =
                    pct > 50 ? "var(--green)" : pct > 20 ? "var(--gold)" : "var(--amber)";
                  return (
                    <div key={cat.category_id ?? cat.id} className="ticket-category">
                      <div className="ticket-category__header">
                        <div>
                          <div className="ticket-category__name">{cat.name}</div>
                          <div className="ticket-category__type">{cat.category_type}</div>
                        </div>
                        <div className="ticket-category__price">
                          {fmt.currency(cat.price)}
                        </div>
                      </div>
                      <div className="ticket-category__bar-row">
                        <div className="ticket-category__bar-track">
                          <div
                            className="ticket-category__bar-fill"
                            style={{ width: `${pct}%`, background: clr }}
                          />
                        </div>
                        <span
                          className="ticket-category__bar-label"
                          style={{ color: clr }}
                        >
                          {avail}/{total}
                        </span>
                      </div>
                    </div>
                  );
                })}
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}

