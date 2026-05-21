import { BrowserRouter, Navigate, Route, Routes } from "react-router-dom";
import AppLayout from "./components/AppLayout";
import { MusicProvider } from "./components/MusicProvider";
import AISettings from "./pages/AISettings";
import Dashboard from "./pages/Dashboard";
import EssayReview from "./pages/EssayReview";
import IntakePage from "./pages/IntakePage";
import LLMSettings from "./pages/LLMSettings";
import MusicPlayer from "./pages/MusicPlayer";
import PomodoroPage from "./pages/PomodoroPage";
import PromptSettings from "./pages/PromptSettings";
import StudyCenter from "./pages/StudyCenter";

function AppRoutes() {
  return (
    <BrowserRouter>
      <MusicProvider>
        <Routes>
          <Route element={<AppLayout />}>
            <Route index element={<Dashboard />} />
            <Route path="intake" element={<IntakePage />} />
            <Route path="ocr" element={<Navigate to="/intake" replace />} />
            <Route path="questions" element={<Navigate to="/intake" replace />} />
            <Route path="questions/:id" element={<Navigate to="/intake" replace />} />
            <Route path="mistakes" element={<Navigate to="/intake" replace />} />
            <Route path="essay" element={<EssayReview />} />
            <Route path="pdf" element={<Navigate to="/ai?tab=pdf" replace />} />
            <Route path="pomodoro" element={<PomodoroPage />} />
            <Route path="music" element={<MusicPlayer />} />
            <Route path="study" element={<StudyCenter />} />
            <Route path="logs" element={<Navigate to="/study" replace />} />
            <Route path="plans" element={<Navigate to="/study" replace />} />
            <Route path="calendar" element={<Navigate to="/study" replace />} />
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
