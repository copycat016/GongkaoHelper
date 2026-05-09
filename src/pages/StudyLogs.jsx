import { useEffect, useState } from "react";
import { Card, Col, DatePicker, Row, Segmented, Space, Table, Timeline, Typography } from "antd";
import dayjs from "dayjs";
import PageHeader from "../components/PageHeader";
import StatCard from "../components/StatCard";
import { getLogs, getLogStats } from "../api/logs";

const { Text } = Typography;

const scopeOptions = [
  { label: "按日", value: "day" },
  { label: "按周", value: "week" },
  { label: "按月", value: "month" },
];

function StudyLogs() {
  const [scope, setScope] = useState("day");
  const [date, setDate] = useState(dayjs());
  const [logs, setLogs] = useState([]);
  const [stats, setStats] = useState({});
  const [loading, setLoading] = useState(false);

  const loadData = async (nextDate = date, nextScope = scope) => {
    setLoading(true);
    try {
      const params = {
        date: (nextDate || dayjs()).format("YYYY-MM-DD"),
        scope: nextScope,
      };
      const [logList, statData] = await Promise.all([getLogs(params), getLogStats(params)]);
      setLogs(logList || []);
      setStats(statData || {});
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, []);

  const handleScopeChange = (value) => {
    setScope(value);
    loadData(date, value);
  };

  const handleDateChange = (value) => {
    const nextDate = value || dayjs();
    setDate(nextDate);
    loadData(nextDate, scope);
  };

  const columns = [
    { title: "开始", dataIndex: "start_time", render: formatTime },
    { title: "结束", dataIndex: "end_time", render: formatTime },
    { title: "持续", dataIndex: "duration_min", render: (value) => `${value || 0}m` },
    { title: "学习类型", dataIndex: "study_type" },
    { title: "科目", dataIndex: "subject" },
    { title: "题型", dataIndex: "question_type" },
    { title: "来源", dataIndex: "source" },
    { title: "备注", dataIndex: "note" },
  ];

  return (
    <div className="page-grid">
      <PageHeader eyebrow="Logs" title="学习日志" desc="查看某一天、某周或某月的学习时长、题型统计和番茄钟完成情况。" />
      <Card className="glass-card log-filter-card" bordered={false}>
        <Row gutter={[16, 16]} align="middle" justify="space-between">
          <Col xs={24} md={12}>
            <Space direction="vertical" size={4}>
              <Text strong>统计范围</Text>
              <Text type="secondary">{periodLabel(date, scope)}</Text>
            </Space>
          </Col>
          <Col xs={24} md={12}>
            <Space wrap className="log-filter-actions">
              <Segmented options={scopeOptions} value={scope} onChange={handleScopeChange} />
              <DatePicker
                value={date}
                picker={scope === "day" ? "date" : scope}
                allowClear={false}
                onChange={handleDateChange}
              />
            </Space>
          </Col>
        </Row>
      </Card>
      <Row gutter={[16, 16]}>
        <Col xs={24} md={6}><StatCard label="总时长" value={formatMinutes(stats.total_minutes || 0)} hint={scopeHint(scope)} /></Col>
        <Col xs={24} md={6}><StatCard label="番茄钟" value={String(stats.pomodoro_count || 0)} hint="完成次数" /></Col>
        <Col xs={24} md={6}><StatCard label="中断次数" value={String(stats.interruptions || 0)} hint="预留统计" /></Col>
        <Col xs={24} md={6}><StatCard label="主要科目" value={stats.main_subject || "-"} hint="按时长统计" /></Col>
      </Row>
      <Row gutter={[16, 16]}>
        <Col xs={24} lg={8}>
          <Card className="glass-card" title="学习时间轴" bordered={false}>
            <Timeline items={logs.map((item) => ({ children: `${formatTime(item.start_time)}-${formatTime(item.end_time)} ${item.subject || item.study_type || "学习"} · ${item.note || ""}` }))} />
          </Card>
        </Col>
        <Col xs={24} lg={16}>
          <Card className="glass-card" title="日志表格" bordered={false}>
            <Table rowKey="id" columns={columns} dataSource={logs} loading={loading} pagination={false} scroll={{ x: 800 }} />
          </Card>
        </Col>
      </Row>
    </div>
  );
}

function formatTime(value) {
  if (!value) return "-";
  return new Date(value).toLocaleTimeString("zh-CN", { hour: "2-digit", minute: "2-digit" });
}

function formatMinutes(minutes) {
  const hours = Math.floor(minutes / 60);
  const rest = minutes % 60;
  if (!hours) return `${rest}m`;
  return rest ? `${hours}h ${rest}m` : `${hours}h`;
}

function scopeHint(scope) {
  if (scope === "week") return "本周累计";
  if (scope === "month") return "本月累计";
  return "当日累计";
}

function periodLabel(date, scope) {
  if (!date) return "-";
  if (scope === "week") {
    const day = date.day() || 7;
    const start = date.subtract(day - 1, "day");
    return `${start.format("YYYY-MM-DD")} 至 ${start.add(6, "day").format("YYYY-MM-DD")}`;
  }
  if (scope === "month") return date.format("YYYY 年 MM 月");
  return date.format("YYYY 年 MM 月 DD 日");
}

export default StudyLogs;
