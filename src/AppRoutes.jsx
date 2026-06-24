import { lazy, Suspense, useEffect } from "react";
import { Spin } from "antd";
import { BrowserRouter, Navigate, Route, Routes } from "react-router-dom";
import AppLayout from "./components/AppLayout";
import { MusicProvider } from "./components/MusicProvider";
import ProtectedRoute from "./components/ProtectedRoute";

// 每个页面单独打包成 chunk（首屏轻量）；importer 函数同时用于 lazy() 和后台预取。
const importers = {
  AISettings: () => import("./pages/AISettings"),
  Dashboard: () => import("./pages/Dashboard"),
  EssayReview: () => import("./pages/EssayReview"),
  IntakePage: () => import("./pages/IntakePage"),
  LLMSettings: () => import("./pages/LLMSettings"),
  Login: () => import("./pages/Login"),
  MusicPlayer: () => import("./pages/MusicPlayer"),
  PomodoroPage: () => import("./pages/PomodoroPage"),
  PromptSettings: () => import("./pages/PromptSettings"),
  StudyCenter: () => import("./pages/StudyCenter"),
};

const AISettings = lazy(importers.AISettings);
const Dashboard = lazy(importers.Dashboard);
const EssayReview = lazy(importers.EssayReview);
const IntakePage = lazy(importers.IntakePage);
const LLMSettings = lazy(importers.LLMSettings);
const Login = lazy(importers.Login);
const MusicPlayer = lazy(importers.MusicPlayer);
const PomodoroPage = lazy(importers.PomodoroPage);
const PromptSettings = lazy(importers.PromptSettings);
const StudyCenter = lazy(importers.StudyCenter);

// 应用挂载后趁空闲把其余页面 chunk 预取进缓存，点击切换页面时即时呈现。
function usePrefetchRoutes() {
  useEffect(() => {
    const prefetch = () => {
      Object.values(importers).forEach((load) => {
        load().catch(() => {});
      });
    };
    if (typeof window.requestIdleCallback === "function") {
      const id = window.requestIdleCallback(prefetch, { timeout: 2000 });
      return () => window.cancelIdleCallback(id);
    }
    const timer = window.setTimeout(prefetch, 1500);
    return () => window.clearTimeout(timer);
  }, []);
}

function LazyPage({ component: Component }) {
  return (
    <Suspense fallback={<div className="route-loading"><Spin /></div>}>
      <Component />
    </Suspense>
  );
}

function AppRoutes() {
  usePrefetchRoutes();
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
