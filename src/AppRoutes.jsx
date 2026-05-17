import { BrowserRouter, Navigate, Route, Routes } from "react-router-dom";
import AppLayout from "./components/AppLayout";
import { MusicProvider } from "./components/MusicProvider";
import AISettings from "./pages/AISettings";
import Dashboard from "./pages/Dashboard";
import EssayReview from "./pages/EssayReview";
import LLMSettings from "./pages/LLMSettings";
import MistakeBook from "./pages/MistakeBook";
import MusicPlayer from "./pages/MusicPlayer";
import OCRQuestion from "./pages/OCRQuestion";
import PomodoroPage from "./pages/PomodoroPage";
import PromptSettings from "./pages/PromptSettings";
import QuestionBank from "./pages/QuestionBank";
import QuestionDetail from "./pages/QuestionDetail";
import StudyCenter from "./pages/StudyCenter";

function AppRoutes() {
  return (
    <BrowserRouter>
      <MusicProvider>
        <Routes>
          <Route element={<AppLayout />}>
            <Route index element={<Dashboard />} />
            <Route path="ocr" element={<OCRQuestion />} />
            <Route path="questions" element={<QuestionBank />} />
            <Route path="questions/:id" element={<QuestionDetail />} />
            <Route path="mistakes" element={<MistakeBook />} />
            <Route path="essay" element={<EssayReview />} />
            <Route path="pdf" element={<Navigate to="/ai?tab=pdf" replace />} />
            <Route path="pomodoro" element={<PomodoroPage />} />
            <Route path="music" element={<MusicPlayer />} />
            <Route path="study" element={<StudyCenter />} />
            <Route path="logs" element={<Navigate to="/study?tab=logs" replace />} />
            <Route path="plans" element={<Navigate to="/study?tab=plans" replace />} />
            <Route path="calendar" element={<Navigate to="/study?tab=calendar" replace />} />
            <Route path="ai" element={<AISettings />} />
            <Route path="llm" element={<LLMSettings />} />
            <Route path="prompts" element={<PromptSettings />} />
          </Route>
        </Routes>
      </MusicProvider>
    </BrowserRouter>
  );
}

export default AppRoutes;
