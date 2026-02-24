import { useState, useRef } from "react";
import api from "../../api/api";
import { useToast } from "../../context/ToastContext";
import { EventBadge } from "../../components/EventBadge";
import { EmptyState, Btn, Spinner, Label, Divider } from "../../components/ui/index.jsx";

export function ImagesTab({ event }) {
  const toast   = useToast();
  const fileRef = useRef();
  const [loading, setLoading]   = useState(false);
  const [images, setImages]     = useState(event?.images || []);
  const [files, setFiles]       = useState([]);
  const [previews, setPreviews] = useState([]);

  const onFiles = (e) => {
    const f = Array.from(e.target.files);
    setFiles(f);
    setPreviews(f.map((x) => URL.createObjectURL(x)));
  };

  const upload = async () => {
    if (!files.length) { toast("Select at least one image", "error"); return; }
    try {
      setLoading(true);
      const fd = new FormData();
      files.forEach((f) => fd.append("images", f));
      await api.uploadImages(event.id, fd);
      const updated = await api.detail(event.id);
      setImages(updated.images || []);
      setFiles([]);
      setPreviews([]);
      toast("Images uploaded!", "success");
    } catch (e) {
      toast(e.message, "error");
    } finally {
      setLoading(false);
    }
  };

  const del = async (img) => {
    try {
      await api.deleteImage(event.id, img.id);
      setImages((p) => p.filter((x) => x.id !== img.id));
      toast("Deleted", "success");
    } catch (e) {
      toast(e.message, "error");
    }
  };

  return (
    <div className="anim-fade-in">
      <EventBadge event={event} />

      {images.length > 0 ? (
        <div style={{ marginBottom: 32 }}>
          <Label>Current Images ({images.length})</Label>
          <div className="images-grid">
            {images.map((img) => (
              <div
                key={img.id}
                className={`image-thumb${img.is_primary ? " image-thumb--primary" : ""}`}
              >
                <img src={img.image_url} alt="" />
                {img.is_primary && (
                  <span className="image-thumb__primary-tag">PRIMARY</span>
                )}
                <button className="image-thumb__delete" onClick={() => del(img)}>
                  ×
                </button>
              </div>
            ))}
          </div>
          <Divider />
        </div>
      ) : (
        <EmptyState icon="🖼" title="No images yet" subtitle="Upload below" />
      )}

      <Label>Upload New Images</Label>
      <div
        className="upload-zone"
        onClick={() => fileRef.current.click()}
        style={{ margin: "12px 0 16px" }}
      >
        <div className="upload-zone__icon">⊕</div>
        <div className="upload-zone__label">Click to select images</div>
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
        <div className="upload-previews">
          {previews.map((src, i) => (
            <img key={i} src={src} alt="" />
          ))}
        </div>
      )}

      <Btn
        onClick={upload}
        disabled={loading || !files.length}
        style={{ minWidth: 160 }}
      >
        {loading && <Spinner size={14} color="#000" />}
        {loading ? "Uploading..." : `Upload${files.length ? ` (${files.length})` : ""}`}
      </Btn>
    </div>
  );
}

