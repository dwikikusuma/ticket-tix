export function EventBadge({ event }) {
  if (!event) return null;
  return (
    <div className="event-badge">
      <div className="event-badge__dot" />
      <span className="event-badge__label">Event:</span>
      <span className="event-badge__name">{event.name}</span>
      <span className="event-badge__id">ID #{event.id}</span>
    </div>
  );
}

