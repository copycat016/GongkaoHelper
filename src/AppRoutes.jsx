import { lazy, Suspense } from "react";
import { Spin } from "antd";
import { BrowserRouter, Navigate, Route, Routes } from "react-router-dom";
import AppLayout from "./components/AppLayout";
import { MusicProvider } from "./components/MusicProvider";
import ProtectedRoute from "./components/ProtectedRoute";

const AISettings = lazy(() => import("./pages/AISettings"));
const Dashboard = lazy(() => import("./pages/Dashboard"));
const EssayReview = lazy(() => import("./pages/EssayReview"));
const IntakePage = lazy(() => import("./pages/IntakePage"));
const LLMSettings = lazy(() => import("./pages/LLMSettings"));
const Login = lazy(() => import("./pages/Login"));
const MusicPlayer = lazy(() => import("./pages/MusicPlayer"));
const PomodoroPage = lazy(() => import("./pages/PomodoroPage"));
const PromptSettings = lazy(() => import("./pages/PromptSettings"));
const StudyCenter = lazy(() => import("./pages/StudyCenter"));

function LazyPage({ component: Component }) {
  return (
    <Suspense fallback={<div className="route-loading"><Spin /></div>}>
      <Component />
    </Suspense>
  );
}

function AppRoutes() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="login" element={<LazyPage component={Login} />} />
        <Route
          element={(
            <ProtectedRoute>
              <MusicProvider>
                <AppLayout />
              </MusicProvider>
            </ProtectedRoute>
          )}
        >
          <Route index element={<LazyPage component={Dashboard} />} />
          <Route path="intake" element={<LazyPage component={IntakePage} />} />
          <Route path="ocr" element={<Navigate to="/intake" replace />} />
          <Route path="questions" element={<Navigate to="/intake" replace />} />
          <Route path="questions/:id" element={<Navigate to="/intake" replace />} />
          <Route path="mistakes" element={<Navigate to="/intake" replace />} />
          <Route path="essay" element={<LazyPage component={EssayReview} />} />
          <Route path="pdf" element={<Navigate to="/ai?tab=pdf" replace />} />
          <Route path="pomodoro" element={<LazyPage component={PomodoroPage} />} />
          <Route path="music" element={<LazyPage component={MusicPlayer} />} />
          <Route path="study" element={<LazyPage component={StudyCenter} />} />
          <Route path="logs" element={<Navigate to="/study" replace />} />
          <Route path="plans" element={<Navigate to="/study" replace />} />
          <Route path="calendar" element={<Navigate to="/study" replace />} />
          <Route path="ai" element={<LazyPage component={AISettings} />} />
          <Route path="llm" element={<LazyPage component={LLMSettings} />} />
          <Route path="prompts" element={<LazyPage component={PromptSettings} />} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}

export default AppRoutes;
