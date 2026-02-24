import { useState, useCallback, useEffect } from "react";
import api from "../api/api";

export function useBrowse() {
  const [events, setEvents] = useState([]);
  const [loading, setLoading] = useState(false);
  const [loadingMore, setMore] = useState(false);
  const [hasMore, setHasMore] = useState(false);
  const [cursor, setCursor] = useState("");
  const [error, setError] = useState(null);
  const [filters, setFilters] = useState({});

  const doFetch = useCallback(
    async (f, cur) => {
      try {
        cur === "" ? setLoading(true) : setMore(true);
        setError(null);
        const res = await api.browse({
          event_name: f.eventName,
          location: f.location,
          start_date: f.startDate,
          end_date: f.endDate,
          cursor: cur,
          limit: 12,
        });
        cur === ""
          ? setEvents(res.events || [])
          : setEvents((p) => [...p, ...(res.events || [])]);
        setHasMore(res.has_more);
        setCursor(res.next_cursor || "");
      } catch (e) {
        setError(e.message);
      } finally {
        setLoading(false);
        setMore(false);
      }
    },
    []
  );

  useEffect(() => {
    doFetch({}, "");
  }, [doFetch]);

  const search = useCallback(
    (f) => {
      setFilters(f);
      setCursor("");
      doFetch(f, "");
    },
    [doFetch]
  );

  const loadMore = useCallback(() => {
    if (hasMore && !loadingMore) doFetch(filters, cursor);
  }, [hasMore, loadingMore, filters, cursor, doFetch]);

  return { events, loading, loadingMore, hasMore, error, search, loadMore };
}
