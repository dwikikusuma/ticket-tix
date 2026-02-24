import { useState } from "react";
import api from "../../api/api";
import fmt from "../../utils/fmt";
import { useToast } from "../../context/ToastContext";
import { useForm } from "../../hooks/useForm";
import { EventBadge } from "../../components/EventBadge";
import { FieldGroup, Btn, Spinner, Label, Divider } from "../../components/ui/index.jsx";

export function CategoriesTab({ event }) {
  const toast = useToast();
  const [loading, setLoading] = useState(false);
  const [cats, setCats]       = useState(event?.categories || []);
  const { values, errors, set, reset, validate } = useForm({
    name: "", categoryType: "STANDING", price: "", bookType: "FIXED", totalCapacity: "",
  });

  const rules = {
    name:          (v) => !v?.trim() ? "Required" : null,
    price:         (v) => !v || isNaN(v) || Number(v) <= 0 ? "Enter a valid price" : null,
    totalCapacity: (v) => !v || isNaN(v) || Number(v) <= 0 ? "Enter a valid capacity" : null,
  };

  const add = async () => {
    if (!validate(rules)) return;
    try {
      setLoading(true);
      const cat = await api.createCategory(event.id, {
        name: values.name,
        category_type: values.categoryType,
        price: String(values.price),
        book_type: values.bookType,
        total_capacity: Number(values.totalCapacity),
        available_stock: Number(values.totalCapacity),
      });
      setCats((p) => [...p, cat]);
      reset();
      toast("Category added!", "success");
    } catch (e) {
      toast(e.message, "error");
    } finally {
      setLoading(false);
    }
  };

  const del = async (id) => {
    try {
      await api.deleteCategory(event.id, id);
      setCats((p) => p.filter((c) => c.id !== id));
      toast("Removed", "success");
    } catch (e) {
      toast(e.message, "error");
    }
  };

  return (
    <div className="anim-fade-in">
      <EventBadge event={event} />

      {cats.length > 0 && (
        <div style={{ marginBottom: 32 }}>
          <Label>Existing Categories</Label>
          <div className="category-list">
            {cats.map((cat) => (
              <div key={cat.id} className="category-item">
                <div className="category-item__info">
                  <span className="category-item__name">{cat.name}</span>
                  <span className="category-item__type">{cat.category_type}</span>
                  <span className="category-item__price">{fmt.currency(cat.price)}</span>
                  <span className="category-item__capacity">{cat.total_capacity} seats</span>
                </div>
                <Btn variant="danger" size="sm" onClick={() => del(cat.id)}>
                  Remove
                </Btn>
              </div>
            ))}
          </div>
          <Divider />
        </div>
      )}

      <Label style={{ marginBottom: 16 }}>Add New Category</Label>
      <div style={{ display: "flex", flexDirection: "column", gap: 18 }}>
        <div className="categories-form__row-2">
          <FieldGroup label="Name *" error={errors.name}>
            <input
              placeholder="e.g. VIP, STANDING A"
              value={values.name}
              onChange={(e) => set("name", e.target.value)}
            />
          </FieldGroup>
          <FieldGroup label="Type">
            <select value={values.categoryType} onChange={(e) => set("categoryType", e.target.value)}>
              <option value="STANDING">Standing</option>
              <option value="SEATED">Seated</option>
            </select>
          </FieldGroup>
        </div>

        <div className="categories-form__row-3">
          <FieldGroup label="Price (IDR) *" error={errors.price}>
            <input
              type="number"
              placeholder="e.g. 250000"
              value={values.price}
              onChange={(e) => set("price", e.target.value)}
            />
          </FieldGroup>
          <FieldGroup label="Book Type">
            <select value={values.bookType} onChange={(e) => set("bookType", e.target.value)}>
              <option value="FIXED">Fixed</option>
              <option value="FLEXIBLE">Flexible</option>
            </select>
          </FieldGroup>
          <FieldGroup label="Total Capacity *" error={errors.totalCapacity}>
            <input
              type="number"
              placeholder="e.g. 500"
              value={values.totalCapacity}
              onChange={(e) => set("totalCapacity", e.target.value)}
            />
          </FieldGroup>
        </div>

        <div>
          <Btn onClick={add} disabled={loading} style={{ minWidth: 160 }}>
            {loading && <Spinner size={14} color="#000" />}
            {loading ? "Adding..." : "+ Add Category"}
          </Btn>
        </div>
      </div>
    </div>
  );
}

