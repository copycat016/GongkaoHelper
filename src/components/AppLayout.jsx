import { useEffect, useMemo, useState } from "react";
import { Outlet, useLocation, useNavigate } from "react-router-dom";
import { Badge, Button, Layout } from "antd";
import {
  ClockCircleOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  NotificationOutlined,
  ThunderboltOutlined,
} from "@ant-design/icons";
import Sidebar, { menuItems } from "./Sidebar";
import ThemePanel from "./ThemePanel";
import GlobalMusicBar from "./GlobalMusicBar";
import { getTodayPomodoroStats } from "../api/pomodoro";

const { Sider, Header, Content } = Layout;

const pageMeta = {
  "/": { title: "首页总览", desc: "今日任务与学习状态" },
  "/ocr": { title: "OCR 识题", desc: "图片识别与结构化" },
  "/questions": { title: "题库管理", desc: "分类与检索" },
  "/mistakes": { title: "错题库", desc: "复习与掌握度" },
  "/essay": { title: "申论批改", desc: "答案评估" },
  "/pomodoro": { title: "番茄钟", desc: "专注计时" },
  "/music": { title: "音乐播放器", desc: "本地音乐" },
  "/study": { title: "学习管理", desc: "计划、日志与日历" },
  "/ai": { title: "配置", desc: "模型、Prompt、OCR、解析与备份" },
  "/llm": { title: "LLM 配置", desc: "模型与服务商" },
  "/prompts": { title: "Prompt 配置", desc: "提示词模板" },
};

function AppLayout() {
  const [collapsed, setCollapsed] = useState(false);
  const [isMobile, setIsMobile] = useState(false);
  const [pomodoroStats, setPomodoroStats] = useState({
    focus_count: 0,
    focus_minutes: 0,
  });
  const navigate = useNavigate();
  const location = useLocation();

  const selectedKey = useMemo(() => {
    if (["/logs", "/plans", "/calendar"].includes(location.pathname)) {
      return "/study";
    }
    if (location.pathname.startsWith("/questions/")) {
      return "/questions";
    }
    const exact = menuItems.find((item) => item.key === location.pathname);
    return exact ? exact.key : "/";
  }, [location.pathname]);

  const current = pageMeta[selectedKey] || pageMeta["/"];

  useEffect(() => {
    const loadStats = () => {
      getTodayPomodoroStats()
      .then((stats) => setPomodoroStats(stats || { focus_count: 0, focus_minutes: 0 }))
      .catch(() => {});
    };

    loadStats();
    window.addEventListener("pomodoro:updated", loadStats);

    return () => window.removeEventListener("pomodoro:updated", loadStats);
  }, [location.pathname]);

  useEffect(() => {
    const media = window.matchMedia("(max-width: 767px)");
    const handleChange = (event) => {
      setIsMobile(event.matches);
      if (event.matches) {
        setCollapsed(true);
      }
    };

    handleChange(media);
    media.addEventListener("change", handleChange);

    return () => media.removeEventListener("change", handleChange);
  }, []);

  return (
    <Layout className="app-layout">
      <Sider
        className="app-sider"
        width={248}
        collapsedWidth={isMobile ? 0 : 82}
        collapsed={collapsed}
        breakpoint="xl"
        onBreakpoint={(broken) => setCollapsed(broken)}
        trigger={null}
      >
        <Sidebar collapsed={collapsed} selectedKey={selectedKey} onSelect={navigate} />
      </Sider>
      <Layout className="main-layout">
        <Header className="top-header">
          <div className="header-left">
            <Button
              type="text"
              className="collapse-button"
              icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
              onClick={() => setCollapsed((value) => !value)}
            />
            <div>
              <h1 className="page-title">{current.title}</h1>
              <div className="page-desc">{current.desc}</div>
            </div>
          </div>
          <div className="header-right">
            <div className="header-stat">
              <ClockCircleOutlined />
              今日 {formatMinutes(pomodoroStats.focus_minutes || 0)}
            </div>
            <div className="header-stat">
              <ThunderboltOutlined />
              {pomodoroStats.focus_count || 0} 个番茄钟
            </div>
            <ThemePanel />
            <Badge dot>
              <Button shape="circle" icon={<NotificationOutlined />} />
            </Badge>
          </div>
        </Header>
        <Content className="app-content">
          <div className="page-shell">
            <Outlet />
          </div>
        </Content>
        <GlobalMusicBar collapsed={collapsed} isMobile={isMobile} />
      </Layout>
    </Layout>
  );
}

function formatMinutes(minutes) {
  if (!minutes) return "0m";
  const hours = Math.floor(minutes / 60);
  const rest = minutes % 60;
  if (hours === 0) return `${rest}m`;
  if (rest === 0) return `${hours}h`;
  return `${hours}h ${rest}m`;
}

export default AppLayout;
