import { api } from "./request";
import { questionsMock } from "./mockData";

export const getQuestions = () => api.get("/questions", null, { mock: questionsMock });
export const saveQuestion = (data) => api.post("/questions", data, { mock: data });
