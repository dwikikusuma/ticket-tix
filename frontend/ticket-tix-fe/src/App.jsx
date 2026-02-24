import { ToastProvider } from "./context/ToastContext";
import { AppRoutes } from "./routes/AppRoutes";

export default function App() {
  return (
    <ToastProvider>
      <AppRoutes />
    </ToastProvider>
  );
}
