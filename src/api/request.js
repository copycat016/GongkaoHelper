import { message } from "antd";

const baseURL = "/api";

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
  const { method = "GET", data, params, headers, mock } = options;

  if (mock) {
    return Promise.resolve(typeof mock === "function" ? mock() : mock);
  }

  try {
    const response = await fetch(buildUrl(path, params), {
      method,
      headers: {
        "Content-Type": "application/json",
        ...headers,
      },
      body: data ? JSON.stringify(data) : undefined,
    });

    const result = await response.json().catch(() => ({}));

    if (!response.ok || result?.code > 0) {
      throw new Error(result?.message || "请求失败");
    }

    if (Object.prototype.hasOwnProperty.call(result, "data")) {
      return result.data;
    }

    return result;
  } catch (error) {
    message.error(error.message || "网络请求异常");
    throw error;
  }
}

export const api = {
  get: (path, params, options) => request(path, { ...options, params }),
  post: (path, data, options) => request(path, { ...options, method: "POST", data }),
  put: (path, data, options) => request(path, { ...options, method: "PUT", data }),
  delete: (path, data, options) => request(path, { ...options, method: "DELETE", data }),
};
