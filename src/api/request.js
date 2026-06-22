import { message } from "antd";

const baseURL = "/api";
const ACCESS_TOKEN_KEY = "access_token";

export function getAccessToken() {
  return localStorage.getItem(ACCESS_TOKEN_KEY) || "";
}

export function setAccessToken(token) {
  if (token) {
    localStorage.setItem(ACCESS_TOKEN_KEY, token);
  }
}

export function clearAccessToken() {
  localStorage.removeItem(ACCESS_TOKEN_KEY);
}

export function authHeaders() {
  const token = getAccessToken();
  return token ? { Authorization: `Bearer ${token}` } : {};
}

function buildUrl(path, params) {
  const url = new URL(`${baseURL}${path}`, window.location.origin);

  Object.entries(params || {}).forEach(([key, value]) => {
    if (value !== undefined && value !== null && value !== "") {
      url.searchParams.set(key, value);
    }
  });

  return url.toString();
}

export async function request(path, options = {}) {
  const { method = "GET", data, params, headers, mock, signal } = options;

  if (mock) {
    return Promise.resolve(typeof mock === "function" ? mock() : mock);
  }

  try {
    const response = await fetch(buildUrl(path, params), {
      method,
      signal,
      headers: {
        "Content-Type": "application/json",
        ...authHeaders(),
        ...headers,
      },
      body: data !== undefined ? JSON.stringify(data) : undefined,
    });

    const result = await response.json().catch(() => ({}));

    if (!response.ok || result?.code > 0) {
      if (response.status === 401 && path !== "/auth/login") {
        clearAccessToken();
        redirectToLogin();
      }
      throw new Error(result?.message || "请求失败");
    }

    if (Object.prototype.hasOwnProperty.call(result, "data")) {
      return result.data;
    }

    return result;
  } catch (error) {
    if (error?.name === "AbortError") {
      throw error;
    }
    message.error(error.message || "网络请求异常");
    throw error;
  }
}

function redirectToLogin() {
  if (window.location.pathname === "/login") {
    return;
  }
  const next = `${window.location.pathname}${window.location.search}`;
  window.location.assign(`/login?next=${encodeURIComponent(next)}`);
}

export const api = {
  get: (path, params, options) => request(path, { ...options, params }),
  post: (path, data, options) => request(path, { ...options, method: "POST", data }),
  put: (path, data, options) => request(path, { ...options, method: "PUT", data }),
  delete: (path, data, options) => request(path, { ...options, method: "DELETE", data }),
};
