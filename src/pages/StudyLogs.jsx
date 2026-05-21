import { useEffect, useState } from "react";
import {
  Button,
  Card,
  Col,
  DatePicker,
  Form,
  Input,
  InputNumber,
  Modal,
  Row,
  Segmented,
  Select,
  Space,
  Table,
  Timeline,
  Typography,
  message,
} from "antd";
import { PlusOutlined } from "@ant-design/icons";
import dayjs from "dayjs";
import PageHeader from "../components/PageHeader";
import StatCard from "../components/StatCard";
import { getLogs, getLogStats, saveLog } from "../api/logs";

const { Text } = Typography;

const scopeOptions = [
  { label: "按日", value: "day" },
  { label: "按周", value: "week" },
  { label: "按月", value: "month" },
];

const studyTypeOptions = ["申论", "行测", "阅读", "复盘", "整理"];
const subjectOptions = ["申论", "行测", "公共基础", "面试", "其他"];

function StudyLogs() {
  const [manualForm] = Form.useForm();
  const [scope, setScope] = useState("day");
  const [date, setDate] = useState(dayjs());
  const [logs, setLogs] = useState([]);
  const [stats, setStats] = useState({});
  const [loading, setLoading] = useState(false);
  const [manualOpen, setManualOpen] = useState(false);
  const [saving, setSaving] = useState(false);

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

  const openManualLog = () => {
    manualForm.setFieldsValue({
      start_time: date.hour(dayjs().hour()).minute(dayjs().minute()).second(0),
      duration_min: 30,
      study_type: "申论",
      subject: "申论",
    });
    setManualOpen(true);
  };

  const handleSaveManualLog = async () => {
    const values = await manualForm.validateFields();
    const startTime = values.start_time;
    const endTime = values.end_time;
    let durationMin = Number(values.duration_min || 0);
    if (!endTime && durationMin <= 0) {
      message.warning("请填写结束时间或持续分钟");
      return;
    }
    if (endTime) {
      if (!endTime.isAfter(startTime)) {
        message.warning("结束时间必须晚于开始时间");
        return;
      }
      durationMin = Math.max(1, endTime.diff(startTime, "minute"));
    }
    const finalEndTime = endTime || startTime.add(durationMin, "minute");

    setSaving(true);
    try {
      await saveLog({
        start_time: startTime.toISOString(),
        end_time: finalEndTime.toISOString(),
        duration_min: durationMin,
        study_type: values.study_type,
        subject: values.subject,
        question_type: values.question_type,
        source: "manual",
        note: values.note,
      });
      message.success("补登日志已保存");
      setManualOpen(false);
      manualForm.resetFields();
      await loadData(date, scope);
    } finally {
      setSaving(false);
    }
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
              <Button type="primary" icon={<PlusOutlined />} onClick={openManualLog}>补登日志</Button>
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
        <Col xs={24} md={8}><StatCard label="总时长" value={formatMinutes(stats.total_minutes || 0)} hint={scopeHint(scope)} /></Col>
        <Col xs={24} md={8}><StatCard label="番茄钟" value={String(stats.pomodoro_count || 0)} hint="完成次数" /></Col>
        <Col xs={24} md={8}><StatCard label="主要科目" value={stats.main_subject || "-"} hint="按时长统计" /></Col>
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
      <Modal
        title="补登日志"
        open={manualOpen}
        onCancel={() => setManualOpen(false)}
        onOk={handleSaveManualLog}
        confirmLoading={saving}
        destroyOnHidden
      >
        <Form form={manualForm} layout="vertical">
          <Form.Item name="start_time" label="开始时间" rules={[{ required: true, message: "请选择开始时间" }]}>
            <DatePicker showTime format="YYYY-MM-DD HH:mm" className="full-input" />
          </Form.Item>
          <Row gutter={12}>
            <Col xs={24} sm={12}>
              <Form.Item name="end_time" label="结束时间">
                <DatePicker showTime format="YYYY-MM-DD HH:mm" className="full-input" />
              </Form.Item>
            </Col>
            <Col xs={24} sm={12}>
              <Form.Item name="duration_min" label="持续分钟">
                <InputNumber min={1} max={1440} className="full-input" />
              </Form.Item>
            </Col>
          </Row>
          <Row gutter={12}>
            <Col xs={24} sm={12}>
              <Form.Item name="study_type" label="学习类型" rules={[{ required: true, message: "请选择学习类型" }]}>
                <Select options={studyTypeOptions.map((value) => ({ value, label: value }))} />
              </Form.Item>
            </Col>
            <Col xs={24} sm={12}>
              <Form.Item name="subject" label="科目">
                <Select allowClear options={subjectOptions.map((value) => ({ value, label: value }))} />
              </Form.Item>
            </Col>
          </Row>
          <Form.Item name="question_type" label="题型 / 内容">
            <Input placeholder="例如 归纳概括题、材料阅读、资料分析" />
          </Form.Item>
          <Form.Item name="note" label="备注">
            <Input.TextArea rows={3} placeholder="记录这段学习做了什么" />
          </Form.Item>
        </Form>
      </Modal>
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
