import { api, clearAccessToken, setAccessToken } from "./request";

export async function login(credentials) {
  const result = await api.post("/auth/login", credentials);
  if (result?.access_token) {
    setAccessToken(result.access_token);
  }
  return result;
}

export const getMe = () => api.get("/auth/me");

export function logout() {
  clearAccessToken();
}
