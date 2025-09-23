// webui/src/services/axios.js
import axios from "axios";

export const TOKEN_KEY = "wasa_token";
export const USERNAME_KEY = "wasa_username";
export const UNAUTHORIZED_EVENT = "wasa:unauthorized";

// In dev we use Vite proxy (/api -> http://localhost:3000), in prod hit the API URL
const isDev = typeof window !== "undefined" && window.location?.port === "5173";

const http = axios.create({
  baseURL: isDev ? "/api" : __API_URL__,
  timeout: 15000,
});

// Attach Authorization on every request except /session
http.interceptors.request.use((cfg) => {
  const url = String(cfg.url || "");
  const isSession = /(^|\/)session(?:[/?].*)?$/i.test(url);
  if (!isSession) {
    // token == username in this project
    const token =
      localStorage.getItem(TOKEN_KEY) ||
      localStorage.getItem(USERNAME_KEY) ||
      "";
    if (token) {
      cfg.headers.Authorization = `Bearer ${token}`;
    }
  }
  return cfg;
});

http.interceptors.response.use(
  (res) => res,
  (err) => {
    if (err?.response?.status === 401) {
      // Tell views to stop timers NOW
      try { window.dispatchEvent(new Event(UNAUTHORIZED_EVENT)); } catch {}

      // Clear creds
      localStorage.removeItem(TOKEN_KEY);
      localStorage.removeItem(USERNAME_KEY);

      // Defer navigation to avoid unmounting mid-render
      setTimeout(() => {
        if (location.hash !== "#/login") location.hash = "#/login";
      }, 0);
    }
    return Promise.reject(err);
  }
);


// Helper to set both keys right after login or rename
export function setAuthUser(username) {
  localStorage.setItem(TOKEN_KEY, username);
  localStorage.setItem(USERNAME_KEY, username);
}

export default http;
