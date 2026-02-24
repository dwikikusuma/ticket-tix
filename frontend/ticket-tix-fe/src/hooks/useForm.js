import { useState } from "react";

export function useForm(init) {
  const [values, setValues] = useState(init);
  const [errors, setErrors] = useState({});

  const set = (f, v) => {
    setValues((p) => ({ ...p, [f]: v }));
    setErrors((p) => ({ ...p, [f]: undefined }));
  };

  const reset = () => {
    setValues(init);
    setErrors({});
  };

  const validate = (rules) => {
    const e = {};
    for (const [f, r] of Object.entries(rules)) {
      const m = r(values[f], values);
      if (m) e[f] = m;
    }
    setErrors(e);
    return Object.keys(e).length === 0;
  };

  return { values, errors, set, reset, validate };
}

