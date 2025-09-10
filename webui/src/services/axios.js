// webui/src/services/axios.js
import axios from "axios";

export const TOKEN_KEY = "wasa_token";
export const USERNAME_KEY = "wasa_username";

const instance = axios.create({
  baseURL: __API_URL__,
  timeout: 10000,
});

// Attach Authorization header (skip /session)
instance.interceptors.request.use((cfg) => {
  const noAuth = /\/session$/.test(cfg.url || "");
  if (!noAuth) {
    const token = localStorage.getItem(TOKEN_KEY);
    if (token) {
      cfg.headers.Authorization = `Bearer ${token}`;
    }
  }
  return cfg;
});

// Handle 401 globally → reset storage and go back to login
instance.interceptors.response.use(
  (res) => res,
  (err) => {
    if (err?.response?.status === 401) {
      localStorage.removeItem(TOKEN_KEY);
      localStorage.removeItem(USERNAME_KEY);
      if (location.hash !== "#/login") {
        location.hash = "#/login";
      }
    }
    return Promise.reject(err);
  }
);

/**
 * Helper: save username as both token and display name.
 * Call this right after a successful /session login,
 * or after PATCH /user/name.
 */
export function setAuthUser(username) {
  localStorage.setItem(TOKEN_KEY, username);
  localStorage.setItem(USERNAME_KEY, username);
}

export default instance;
