import { api } from "./request";

export const getCalendarEvents = (params) => api.get("/calendar/events", params);
