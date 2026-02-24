import { useState } from "react";
import { Nav } from "../components/Nav";
import { BrowsePage } from "../pages/BrowsePage";
import { EventDetailPage } from "../pages/EventDetailPage";
import { AdminPage } from "../pages/admin/AdminPage";

export function AppRoutes() {
  const [page, setPage]       = useState("browse");
  const [eventId, setEventId] = useState(null);

  const navigate = (to, id = null) => {
    setEventId(id);
    setPage(to);
    window.scrollTo({ top: 0, behavior: "smooth" });
  };

  return (
    <div className="page-wrapper">
      <Nav page={page} navigate={navigate} />
      <div className="page-content">
        {page === "browse" && (
          <BrowsePage onEventClick={(id) => navigate("detail", id)} />
        )}
        {page === "detail" && (
          <EventDetailPage eventId={eventId} onBack={() => navigate("browse")} />
        )}
        {page === "admin" && <AdminPage />}
      </div>
    </div>
  );
}

