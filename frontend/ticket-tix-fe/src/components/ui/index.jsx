export function Spinner({ size = 18, color = "var(--gold)" }) {
  return (
    <div
      className="spinner"
      style={{
        width: size,
        height: size,
        borderTopColor: color,
      }}
    />
  );
}

export function Skeleton({ w = "100%", h = 16, radius = 4, style: s = {} }) {
  return (
    <div
      className="skeleton"
      style={{ width: w, height: h, borderRadius: radius, flexShrink: 0, ...s }}
    />
  );
}

export function Label({ children }) {
  return <label className="field-group__label">{children}</label>;
}

export function FieldGroup({ label, children, error, hint }) {
  return (
    <div className="field-group">
      {label && <Label>{label}</Label>}
      {children}
      {hint && !error && <span className="field-group__hint">{hint}</span>}
      {error && <span className="field-group__error">{error}</span>}
    </div>
  );
}

export function GoldLine() {
  return <div className="gold-line" />;
}

export function Divider({ m = "24px 0" }) {
  return <hr className="divider" style={{ margin: m }} />;
}

export function BackBtn({ onClick }) {
  return (
    <button className="back-btn" onClick={onClick}>
      <span className="back-btn__arrow">←</span> Back to Events
    </button>
  );
}

export function Btn({
  children,
  onClick,
  disabled,
  variant = "gold",
  size = "md",
  style: s = {},
}) {
  return (
    <button
      onClick={onClick}
      disabled={disabled}
      className={`btn btn--${size} btn--${variant}`}
      style={s}
    >
      {children}
    </button>
  );
}

export function EmptyState({ icon = "◎", title, subtitle }) {
  return (
    <div className="empty-state">
      <div className="empty-state__icon">{icon}</div>
      <div className="empty-state__title">{title}</div>
      {subtitle && <div className="empty-state__subtitle">{subtitle}</div>}
    </div>
  );
}

export function CardSkeleton() {
  return (
    <div className="card-skeleton">
      <Skeleton h={200} radius={0} />
      <div className="card-skeleton__body">
        <Skeleton h={11} w="40%" />
        <Skeleton h={22} w="80%" />
        <Skeleton h={14} w="55%" />
      </div>
    </div>
  );
}

