import { useState } from "react";
import { GoldLine } from "../../components/ui/index.jsx";
import { CreateEventTab } from "./CreateEventTab";
import { CategoriesTab } from "./CategoriesTab";
import { ImagesTab } from "./ImagesTab";

const TABS = ["1. Create Event", "2. Categories", "3. Images"];

export function AdminPage() {
  const [tab, setTab]     = useState(0);
  const [event, setEvent] = useState(null);

  return (
    <div className="admin-page">
      <div className="admin-page__header">
        <GoldLine />
        <h2 className="display admin-page__title">Admin Panel</h2>
        <p className="admin-page__subtitle">Create and manage events</p>
      </div>

      <div className="admin-tabs">
        {TABS.map((t, i) => (
          <button
            key={t}
            onClick={() => setTab(i)}
            disabled={i > 0 && !event}
            className={`admin-tab${tab === i ? " admin-tab--active" : ""}${i > 0 && !event ? " admin-tab--disabled" : ""}`}
          >
            {t}
          </button>
        ))}
      </div>

      {tab === 0 && <CreateEventTab onCreated={(e) => { setEvent(e); setTab(1); }} />}
      {tab === 1 && <CategoriesTab event={event} />}
      {tab === 2 && <ImagesTab event={event} />}
    </div>
  );
}

