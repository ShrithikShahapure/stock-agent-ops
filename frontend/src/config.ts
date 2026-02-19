declare global {
  interface Window {
    __API_URL__?: string;
  }
}

export const API_URL: string =
  window.__API_URL__ ||
  import.meta.env.VITE_API_URL ||
  "http://localhost:8000";
