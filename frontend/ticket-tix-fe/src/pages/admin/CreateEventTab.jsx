import { useState, useRef } from "react";
import api from "../../api/api";
import { useToast } from "../../context/ToastContext";
import { useForm } from "../../hooks/useForm";
import { FieldGroup, Btn, Spinner } from "../../components/ui/index.jsx";

export function CreateEventTab({ onCreated }) {
  const toast   = useToast();
  const fileRef = useRef();
  const [loading, setLoading]   = useState(false);
  const [files, setFiles]       = useState([]);
  const [previews, setPreviews] = useState([]);
  const { values, errors, set, validate } = useForm({
    name: "", description: "", location: "", startTime: "", endTime: "",
  });

  const rules = {
    name:      (v) => !v?.trim() ? "Required" : null,
    location:  (v) => !v?.trim() ? "Required" : null,
    startTime: (v) => !v ? "Required" : null,
    endTime:   (v, a) =>
      !v ? "Required"
      : a.startTime && new Date(v) <= new Date(a.startTime)
      ? "Must be after start"
      : null,
  };

  const onFiles = (e) => {
    const f = Array.from(e.target.files);
    setFiles(f);
    setPreviews(f.map((x) => URL.createObjectURL(x)));
  };

  const rmFile = (i) => {
    setFiles((p) => p.filter((_, j) => j !== i));
    setPreviews((p) => p.filter((_, j) => j !== i));
  };

  const submit = async () => {
    if (!validate(rules)) return;
    try {
      setLoading(true);
      const fd = new FormData();
      fd.append("name", values.name);
      fd.append("description", values.description);
      fd.append("location", values.location);
      fd.append("start_time", new Date(values.startTime).toISOString());
      fd.append("end_time", new Date(values.endTime).toISOString());
      files.forEach((f) => fd.append("images", f));
      const ev = await api.create(fd);
      toast("Event created!", "success");
      onCreated(ev);
    } catch (e) {
      toast(e.message, "error");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="create-event-form anim-fade-in">
      <div className="create-event-form__row-2">
        <FieldGroup label="Event Name *" error={errors.name}>
          <input
            placeholder="e.g. Jazz Night Jakarta"
            value={values.name}
            onChange={(e) => set("name", e.target.value)}
          />
        </FieldGroup>
        <FieldGroup label="Location *" error={errors.location}>
          <input
            placeholder="e.g. Jakarta Convention Center"
            value={values.location}
            onChange={(e) => set("location", e.target.value)}
          />
        </FieldGroup>
      </div>

      <FieldGroup label="Description">
        <textarea
          placeholder="Describe your event..."
          value={values.description}
          onChange={(e) => set("description", e.target.value)}
        />
      </FieldGroup>

      <div className="create-event-form__row-2">
        <FieldGroup label="Start Date & Time *" error={errors.startTime}>
          <input
            type="datetime-local"
            value={values.startTime}
            onChange={(e) => set("startTime", e.target.value)}
          />
        </FieldGroup>
        <FieldGroup label="End Date & Time *" error={errors.endTime}>
          <input
            type="datetime-local"
            value={values.endTime}
            onChange={(e) => set("endTime", e.target.value)}
          />
        </FieldGroup>
      </div>

      <FieldGroup label="Event Images" hint="First image becomes the cover photo">
        <div
          className="upload-zone"
          onClick={() => fileRef.current.click()}
        >
          <div className="upload-zone__icon">⊕</div>
          <div className="upload-zone__label">Click to upload images</div>
          <div className="upload-zone__hint">JPG, PNG, WebP · Multiple allowed</div>
        </div>
        <input
          ref={fileRef}
          type="file"
          multiple
          accept="image/*"
          style={{ display: "none" }}
          onChange={onFiles}
        />
        {previews.length > 0 && (
          <div className="image-preview-list">
            {previews.map((src, i) => (
              <div key={i} className={`image-preview${i === 0 ? " image-preview--cover" : ""}`}>
                <img src={src} alt="" />
                {i === 0 && <span className="image-preview__cover-tag">COVER</span>}
                <button
                  className="image-preview__remove"
                  onClick={(e) => { e.stopPropagation(); rmFile(i); }}
                >
                  ×
                </button>
              </div>
            ))}
          </div>
        )}
      </FieldGroup>

      <div style={{ paddingTop: 4 }}>
        <Btn onClick={submit} disabled={loading} style={{ minWidth: 180 }}>
          {loading && <Spinner size={14} color="#000" />}
          {loading ? "Creating..." : "Create Event →"}
        </Btn>
      </div>
    </div>
  );
}

