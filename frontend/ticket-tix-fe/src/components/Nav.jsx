import { Label } from "./ui/index.jsx";

export function Nav({ page, navigate }) {
  return (
    <nav className="nav">
      <button className="nav__logo-btn" onClick={() => navigate("browse")}>
        <div className="nav__logo-icon">
          <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
            <rect x="1" y="3" width="12" height="8" rx="1" stroke="#080808" strokeWidth="1.5" />
            <path d="M4 3V2M7 3V2M10 3V2" stroke="#080808" strokeWidth="1.5" strokeLinecap="round" />
            <path d="M4 7h6M4 9.5h4" stroke="#080808" strokeWidth="1.2" strokeLinecap="round" />
          </svg>
        </div>
        <span className="nav__logo-text">
          Ticket<span className="nav__logo-accent">Tix</span>
        </span>
      </button>
      <div className="nav__links">
        {["browse", "admin"].map((p) => (
          <button
            key={p}
            onClick={() => navigate(p)}
            className={`nav__link${page === p ? " nav__link--active" : ""}`}
          >
            {p}
          </button>
        ))}
      </div>
    </nav>
  );
}

