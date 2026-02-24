import { useBrowse } from "../hooks/useBrowse";
import { EventCard } from "../components/EventCard";
import { SearchBar } from "../components/SearchBar";
import { CardSkeleton, EmptyState, Btn, Spinner, GoldLine } from "../components/ui/index.jsx";

export function BrowsePage({ onEventClick }) {
  const { events, loading, loadingMore, hasMore, error, search, loadMore } = useBrowse();

  return (
    <div className="browse-page">
      <div className="browse-page__header">
        <GoldLine />
        <h2 className="display browse-page__title">Upcoming Events</h2>
        <p className="browse-page__subtitle">Discover live experiences near you</p>
      </div>

      <SearchBar onSearch={search} loading={loading} />

      {error && <div className="browse-page__error">{error}</div>}

      {loading ? (
        <div className="stagger browse-page__grid">
          {Array.from({ length: 6 }).map((_, i) => (
            <CardSkeleton key={i} />
          ))}
        </div>
      ) : events.length === 0 ? (
        <EmptyState icon="◎" title="No events found" subtitle="Try adjusting your filters" />
      ) : (
        <>
          <div className="stagger browse-page__grid">
            {events.map((e, i) => (
              <EventCard
                key={e.id}
                event={e}
                onClick={() => onEventClick(e.id)}
                style={{ animationDelay: `${(i % 12) * 0.05}s` }}
              />
            ))}
          </div>

          {hasMore && (
            <div className="browse-page__load-more">
              <Btn
                onClick={loadMore}
                disabled={loadingMore}
                variant="outline"
                style={{ minWidth: 160 }}
              >
                {loadingMore && <Spinner size={14} />}
                {loadingMore ? "Loading..." : "Load More"}
              </Btn>
            </div>
          )}

          {!hasMore && events.length > 0 && (
            <div className="browse-page__end">— END OF RESULTS —</div>
          )}
        </>
      )}
    </div>
  );
}

