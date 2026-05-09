import { useEffect, useMemo, useState } from "react";
import { Button, Card, Col, List, Row, Space, Tag } from "antd";
import {
  CloudUploadOutlined,
  FormOutlined,
  PlusCircleOutlined,
  ScheduleOutlined,
  ThunderboltOutlined,
} from "@ant-design/icons";
import StatCard from "../components/StatCard";
import { dashboardMock } from "../api/mockData";
import { getTodayPomodoroStats } from "../api/pomodoro";
import { getPlans } from "../api/plans";

function Dashboard() {
  const [pomodoroStats, setPomodoroStats] = useState(null);
  const [plans, setPlans] = useState([]);

  useEffect(() => {
    const loadStats = () => {
      getTodayPomodoroStats()
        .then((stats) => setPomodoroStats(stats))
        .catch(() => {});
    };

    loadStats();
    getPlans().then((items) => setPlans(items || [])).catch(() => {});
    window.addEventListener("pomodoro:updated", loadStats);

    return () => window.removeEventListener("pomodoro:updated", loadStats);
  }, []);

  const stats = useMemo(() => {
    if (!pomodoroStats) return dashboardMock.stats;

    return dashboardMock.stats.map((item) => {
      if (item.label === "今日学习") {
        return {
          ...item,
          value: formatMinutes(pomodoroStats.focus_minutes || 0),
          hint: "来自番茄钟记录",
        };
      }

      if (item.label === "番茄钟") {
        return {
          ...item,
          value: String(pomodoroStats.focus_count || 0),
          hint: "今日完成专注轮次",
        };
      }

      return item;
    });
  }, [pomodoroStats]);

  const shortcuts = [
    { icon: <CloudUploadOutlined />, label: "OCR 识题目", path: "/ocr", desc: "拍照录题" },
    { icon: <ThunderboltOutlined />, label: "开始番茄钟", path: "/pomodoro", desc: "绑定任务" },
    { icon: <ScheduleOutlined />, label: "新建学习计划", path: "/study?tab=plans", desc: "安排今天" },
    { icon: <PlusCircleOutlined />, label: "新增错题", path: "/mistakes", desc: "手动记录" },
    { icon: <FormOutlined />, label: "申论批改", path: "/essay", desc: "练一篇" },
  ];

  const taskGroups = useMemo(() => groupPlans(plans), [plans]);

  return (
    <div className="page-grid">
      <Row gutter={[16, 16]}>
        <Col xs={24} xl={14}>
          <TaskListCard title="每日任务" tag="Today" items={taskGroups.daily} />
        </Col>
        <Col xs={24} xl={10}>
          <Card className="glass-card dashboard-tools" title="工具入口" bordered={false}>
            <Row gutter={[10, 10]}>
              {shortcuts.map((item) => (
                <Col xs={24} sm={12} key={item.label}>
                  <Button block icon={item.icon} className="soft-button dashboard-tool-btn" href={item.path}>
                    <span>{item.label}</span>
                    <small>{item.desc}</small>
                  </Button>
                </Col>
              ))}
            </Row>
          </Card>
        </Col>
      </Row>
      <Row gutter={[16, 16]}>
        {stats.map((item) => (
          <Col xs={24} sm={12} xl={6} key={item.label}>
            <StatCard {...item} />
          </Col>
        ))}
      </Row>
      <Row gutter={[16, 16]}>
        <Col xs={24} md={12} xl={6}>
          <TaskListCard title="未完成每周任务" tag="Week" items={taskGroups.weekly} />
        </Col>
        <Col xs={24} md={12} xl={6}>
          <TaskListCard title="随手任务" tag="Someday" items={taskGroups.someday} />
        </Col>
        <Col xs={24} md={12} xl={8}>
          <Card className="glass-card" title="最近 OCR 识别" bordered={false}>
            <List dataSource={dashboardMock.ocrRecords} renderItem={(item) => <List.Item>{item}</List.Item>} />
          </Card>
        </Col>
        <Col xs={24} md={12} xl={4}>
          <Card className="glass-card" title="最近申论批改" bordered={false}>
            <List dataSource={dashboardMock.essayRecords} renderItem={(item) => <List.Item>{item}</List.Item>} />
          </Card>
        </Col>
      </Row>
      <Card className="glass-card dashboard-completed" title="已完成任务" bordered={false}>
        <List
          dataSource={taskGroups.completed}
          locale={{ emptyText: "今天还没有完成任务" }}
          renderItem={(item) => <List.Item>{item.title}</List.Item>}
        />
      </Card>
    </div>
  );
}

function TaskListCard({ title, tag, items }) {
  return (
    <Card
      className="glass-card dashboard-task-card"
      title={<Space><span>{title}</span><Tag>{tag}</Tag></Space>}
      bordered={false}
    >
      <List
        dataSource={items}
        locale={{ emptyText: "暂无任务" }}
        renderItem={(item) => (
          <List.Item>
            <List.Item.Meta
              title={item.title}
              description={`${item.subject || "未分类"} · ${item.target_min || 0}m · ${item.status || "进行中"}`}
            />
          </List.Item>
        )}
      />
    </Card>
  );
}

function groupPlans(plans) {
  const active = plans.filter((item) => item.status !== "已完成");
  return {
    daily: active.filter((item) => item.plan_type === "每日计划"),
    weekly: active.filter((item) => item.plan_type === "每周计划"),
    someday: active.filter((item) => !item.due_date || item.plan_type === "随手任务" || item.plan_type === "不限时间"),
    completed: plans.filter((item) => item.status === "已完成"),
  };
}

function formatMinutes(minutes) {
  if (!minutes) return "0m";
  const hours = Math.floor(minutes / 60);
  const rest = minutes % 60;
  if (hours === 0) return `${rest}m`;
  if (rest === 0) return `${hours}h`;
  return `${hours}h ${rest}m`;
}

export default Dashboard;
