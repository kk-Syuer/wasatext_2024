// webui/src/services/axios.js
import axios from "axios";

export const TOKEN_KEY = "wasa_token";

const instance = axios.create({
  baseURL: __API_URL__,
  timeout: 10000
});

// Attach token (skip /session)
instance.interceptors.request.use(cfg => {
  const noAuth = /\/session$/.test(cfg.url || "");
  if (!noAuth) {
    const t = localStorage.getItem(TOKEN_KEY);
    if (t) cfg.headers.Authorization = `Bearer ${t}`;
  }
  return cfg;
});

// Optional: handle 401 globally
instance.interceptors.response.use(
  r => r,
  err => {
    if (err?.response?.status === 401) {
      localStorage.removeItem(TOKEN_KEY);
      // go to login without blowing up SPA
      if (location.hash !== "#/login") location.hash = "#/login";
    }
    return Promise.reject(err);
  }
);

export default instance;
