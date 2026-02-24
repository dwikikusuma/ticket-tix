const API_BASE = "http://localhost:50061";

async function apiFetch(path, opts = {}) {
  const res = await fetch(`${API_BASE}${path}`, opts);
  if (!res.ok) {
    const b = await res.json().catch(() => ({}));
    throw new Error(b.error || `HTTP ${res.status}`);
  }
  if (res.status === 204) return null;
  return res.json();
}

const api = {
  browse: (p) => {
    const q = new URLSearchParams();
    Object.entries(p).forEach(([k, v]) => v && q.set(k, v));
    return apiFetch(`/events?${q}`);
  },
  detail: (id) => apiFetch(`/event/${id}`),
  create: (fd) => apiFetch("/events", { method: "POST", body: fd }),
  uploadImages: (id, fd) =>
    apiFetch(`/events/${id}/images`, { method: "POST", body: fd }),
  deleteImage: (eid, iid) =>
    apiFetch(`/events/${eid}/images/${iid}`, { method: "DELETE" }),
  createCategory: (eid, data) =>
    apiFetch(`/events/${eid}/categories`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(data),
    }),
  deleteCategory: (eid, cid) =>
    apiFetch(`/events/${eid}/categories/${cid}`, { method: "DELETE" }),
};

export default api;
