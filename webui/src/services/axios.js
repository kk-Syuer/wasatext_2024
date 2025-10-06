// webui/src/services/axios.js
import axios from "axios";

export const TOKEN_KEY = "wasa_token";
export const USERNAME_KEY = "wasa_username";
export const UNAUTHORIZED_EVENT = "wasa:unauthorized";

// 在改名等操作期间，临时关闭 401 自动登出
let SUPPRESS_401 = false;
export function suppressUnauthorized(on) {
  SUPPRESS_401 = !!on;
}

// In dev we use Vite proxy (/api -> http://localhost:3000).
// In prod, set VITE_API_BASE to your backend origin or path prefix.
// If VITE_API_BASE is unset, we'll still default to "/api" (works if FE is reverse-proxied by BE).
const http = axios.create({
  baseURL: __API_URL__,
  timeout: 1000 * 5,
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
      cfg.headers = cfg.headers || {};
      cfg.headers.Authorization = `Bearer ${token}`;
    }
  }
  return cfg;
});

http.interceptors.response.use(
  (res) => res,
  (err) => {
    if (err?.response?.status === 401) {
      // 改名中的瞬时 401，不触发登出
      if (SUPPRESS_401) {
        return Promise.reject(err);
      }
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
