import fmt from "../utils/fmt";

export function EventCard({ event, onClick, style: s = {} }) {
  const img = event.image_url || null;

  return (
    <article
      onClick={onClick}
      className="event-card anim-fade-up"
      style={s}
    >
      <div className="event-card__image-wrap">
        {img ? (
          <img src={img} alt={event.name} className="event-card__image" />
        ) : (
          <div className="event-card__no-image">
            <span>🎭</span>
          </div>
        )}
        <div className="event-card__date-badge">
          <div className="event-card__date-day">
            {new Date(event.start_time).getDate()}
          </div>
          <div className="event-card__date-month">
            {new Date(event.start_time).toLocaleString("en", { month: "short" })}
          </div>
        </div>
      </div>
      <div className="event-card__body">
        <div className="event-card__time">
          {fmt.time(event.start_time)} · {new Date(event.start_time).getFullYear()}
        </div>
        <h3 className="event-card__title">{event.name}</h3>
        <div className="event-card__location">
          <svg width="11" height="13" viewBox="0 0 11 13" fill="none">
            <path
              d="M5.5 1C3.015 1 1 3.015 1 5.5c0 3.375 4.5 7.5 4.5 7.5s4.5-4.125 4.5-7.5C10 3.015 7.985 1 5.5 1z"
              stroke="currentColor"
              strokeWidth="1.2"
            />
            <circle cx="5.5" cy="5.5" r="1.5" stroke="currentColor" strokeWidth="1.2" />
          </svg>
          {event.location}
        </div>
      </div>
    </article>
  );
}

