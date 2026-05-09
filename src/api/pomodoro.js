import { api } from "./request";

export const savePomodoroSession = (data) => api.post("/pomodoro/sessions", data);
export const getTodayPomodoroStats = () => api.get("/pomodoro/stats/today");
