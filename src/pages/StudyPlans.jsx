import { useEffect, useState } from "react";
import { Button, Card, DatePicker, Form, Input, InputNumber, Select, Space, Table, Tag, message } from "antd";
import { RobotOutlined } from "@ant-design/icons";
import PageHeader from "../components/PageHeader";
import { completePlan, createPlan, deletePlan, getPlans } from "../api/plans";

function StudyPlans() {
  const [form] = Form.useForm();
  const [plans, setPlans] = useState([]);
  const [loading, setLoading] = useState(false);

  const loadPlans = async () => {
    setLoading(true);
    try {
      setPlans((await getPlans()) || []);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadPlans();
  }, []);

  const handleCreate = async () => {
    const values = await form.validateFields();
    await createPlan({
      ...values,
      start_date: values.start_date?.toISOString(),
      due_date: values.due_date?.toISOString(),
      status: "进行中",
    });
    message.success("计划已创建");
    form.resetFields();
    loadPlans();
  };

  const columns = [
    { title: "标题", dataIndex: "title" },
    { title: "计划类型", dataIndex: "plan_type" },
    { title: "科目", dataIndex: "subject" },
    { title: "目标时长", dataIndex: "target_min", render: (value) => `${value || 0}m` },
    { title: "截止日期", dataIndex: "due_date", render: (value) => value ? new Date(value).toLocaleDateString("zh-CN") : "-" },
    { title: "完成状态", dataIndex: "status", render: (value) => <Tag color={value === "已完成" ? "green" : "blue"}>{value || "进行中"}</Tag> },
    { title: "操作", render: (_, record) => <Space><Button size="small" onClick={async () => { await completePlan(record.id); loadPlans(); }}>完成</Button><Button size="small" danger onClick={async () => { await deletePlan(record.id); loadPlans(); }}>删除</Button></Space> },
  ];

  return (
    <div className="page-grid">
      <PageHeader eyebrow="Plans" title="学习计划" desc="每日、每周、阶段计划与 AI 生成入口。" extra={<Button icon={<RobotOutlined />}>AI 生成计划</Button>} />
      <Card className="glass-card" title="新建计划" bordered={false}>
        <Form form={form} layout="inline">
          <Form.Item name="title" label="标题" rules={[{ required: true, message: "请输入标题" }]}><Input /></Form.Item>
          <Form.Item name="plan_type" label="类型"><Select style={{ width: 150 }} options={["每日计划", "每周计划", "阶段计划", "随手任务"].map((value) => ({ value, label: value }))} /></Form.Item>
          <Form.Item name="subject" label="科目"><Select style={{ width: 120 }} options={["行测", "申论"].map((value) => ({ value, label: value }))} /></Form.Item>
          <Form.Item name="target_min" label="目标分钟"><InputNumber min={0} /></Form.Item>
          <Form.Item name="due_date" label="截止"><DatePicker /></Form.Item>
          <Form.Item><Button type="primary" onClick={handleCreate}>保存</Button></Form.Item>
        </Form>
      </Card>
      <Card className="glass-card" bordered={false}>
        <Table rowKey="id" columns={columns} dataSource={plans} loading={loading} pagination={false} />
      </Card>
    </div>
  );
}

export default StudyPlans;
