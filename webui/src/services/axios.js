// webui/src/services/axios.js
import axios from "axios";

export const TOKEN_KEY = "wasa_token";
export const USERNAME_KEY = "wasa_username";
export const UNAUTHORIZED_EVENT = "wasa:unauthorized";

// Temporarily suppress auto-logout 401 handling (e.g., during rename)
let SUPPRESS_401 = false;
export function suppressUnauthorized(on) {
  SUPPRESS_401 = !!on;
}

// webui/src/services/axios.js
const explicit = (import.meta.env.VITE_API_BASE || "").trim();
const baseURL = import.meta.env.DEV ? "/api" : explicit;

console.log("AXIOS baseURL =", baseURL); // for testing


const instance = axios.create({
  baseURL,
  timeout: 1000 * 5,
});

// Attach Authorization on every request except /session
instance.interceptors.request.use((cfg) => {
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

instance.interceptors.response.use(
  (res) => res,
  (err) => {
    if (err?.response?.status === 401) {
      if (SUPPRESS_401) return Promise.reject(err);
      try { window.dispatchEvent(new Event(UNAUTHORIZED_EVENT)); } catch {}
      localStorage.removeItem(TOKEN_KEY);
      localStorage.removeItem(USERNAME_KEY);
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

export default instance;
