import { Tabs } from "antd";
import { useEffect, useMemo } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import CalendarPage from "./CalendarPage";
import StudyLogs from "./StudyLogs";
import StudyPlans from "./StudyPlans";

const tabItems = [
  {
    key: "plans",
    label: "学习计划",
    children: <StudyPlans />,
  },
  {
    key: "logs",
    label: "学习日志",
    children: <StudyLogs />,
  },
  {
    key: "calendar",
    label: "日历视图",
    children: <CalendarPage />,
  },
];

const validTabs = tabItems.map((item) => item.key);

function StudyCenter() {
  const location = useLocation();
  const navigate = useNavigate();

  const activeTab = useMemo(() => {
    const query = new URLSearchParams(location.search);
    const tab = query.get("tab");
    return validTabs.includes(tab) ? tab : "plans";
  }, [location.search]);

  useEffect(() => {
    if (!location.search) {
      navigate("/study?tab=plans", { replace: true });
    }
  }, [location.search, navigate]);

  return (
    <div className="study-center">
      <Tabs
        className="study-center-tabs"
        activeKey={activeTab}
        items={tabItems}
        onChange={(key) => navigate(`/study?tab=${key}`)}
      />
    </div>
  );
}

export default StudyCenter;
