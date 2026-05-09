import { useEffect, useState } from "react";
import { Button, Card, Form, Input, Modal, Select, Space, Switch, Table, Tag, message } from "antd";
import { ExperimentOutlined, PlusOutlined } from "@ant-design/icons";
import PageHeader from "../components/PageHeader";
import { createPrompt, deletePrompt, getPrompts, updatePrompt } from "../api/prompts";

const taskTypes = ["OCR 自动纠错", "题目结构化", "行测题目解析", "错题原因分析", "申论批改", "申论答案润色", "申论范文生成", "学习计划生成", "学习日志总结", "PDF 内容提取"];

function PromptSettings() {
  const [type, setType] = useState();
  const [open, setOpen] = useState(false);
  const [data, setData] = useState([]);
  const [loading, setLoading] = useState(false);
  const [editingPrompt, setEditingPrompt] = useState(null);
  const [form] = Form.useForm();

  const loadData = async (nextType = type) => {
    setLoading(true);
    try {
      const prompts = await getPrompts(nextType ? { task_type: nextType } : undefined);
      setData(prompts || []);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, []);

  const handleTypeChange = (value) => {
    setType(value);
    loadData(value);
  };

  const openCreateModal = () => {
    setEditingPrompt(null);
    form.setFieldsValue({
      task_type: type,
      version: "v1.0",
      enabled: true,
    });
    setOpen(true);
  };

  const openEditModal = (record) => {
    setEditingPrompt(record);
    form.setFieldsValue(record);
    setOpen(true);
  };

  const handleSave = async () => {
    const values = await form.validateFields();
    if (editingPrompt) {
      await updatePrompt(editingPrompt.id, values);
      message.success("Prompt 已更新");
    } else {
      await createPrompt(values);
      message.success("Prompt 已新增");
    }
    setOpen(false);
    form.resetFields();
    loadData();
  };

  const handleDelete = async (id) => {
    await deletePrompt(id);
    message.success("Prompt 已删除");
    loadData();
  };

  const columns = [
    { title: "任务类型", dataIndex: "task_type" },
    { title: "Prompt 名称", dataIndex: "name" },
    { title: "默认模型", dataIndex: "default_model" },
    { title: "版本", dataIndex: "version" },
    { title: "状态", dataIndex: "enabled", render: (value) => <Tag color={value ? "blue" : "default"}>{value ? "启用" : "停用"}</Tag> },
    { title: "操作", render: (_, record) => <Space><Button size="small" onClick={() => openEditModal(record)}>编辑</Button><Button size="small" icon={<ExperimentOutlined />}>测试</Button><Button size="small" danger onClick={() => handleDelete(record.id)}>删除</Button></Space> },
  ];

  return (
    <div className="page-grid">
      <PageHeader eyebrow="Prompt" title="Prompt 配置" desc="按任务类型维护长提示词，预留测试和启停入口。" />
      <Card className="glass-card" bordered={false}>
        <Space wrap>
          <Select allowClear placeholder="按任务类型筛选" value={type} onChange={handleTypeChange} options={taskTypes.map((value) => ({ value, label: value }))} style={{ width: 220 }} />
          <Button type="primary" icon={<PlusOutlined />} onClick={openCreateModal}>新增 Prompt</Button>
        </Space>
      </Card>
      <Card className="glass-card" bordered={false}>
        <Table rowKey="id" columns={columns} dataSource={data} loading={loading} pagination={false} />
      </Card>
      <Modal width={860} title="Prompt 编辑" open={open} onCancel={() => setOpen(false)} onOk={handleSave}>
        <Form form={form} layout="vertical">
          <Form.Item name="task_type" label="任务类型" rules={[{ required: true, message: "请选择任务类型" }]}><Select options={taskTypes.map((value) => ({ value, label: value }))} /></Form.Item>
          <Form.Item name="name" label="Prompt 名称" rules={[{ required: true, message: "请输入 Prompt 名称" }]}><Input /></Form.Item>
          <Form.Item name="system_prompt" label="System Prompt"><Input.TextArea rows={6} /></Form.Item>
          <Form.Item name="user_prompt" label="User Prompt 模板"><Input.TextArea rows={8} /></Form.Item>
          <Form.Item name="variables" label="变量说明"><Input.TextArea rows={4} /></Form.Item>
          <Form.Item name="default_model" label="默认模型"><Input /></Form.Item>
          <Form.Item name="version" label="版本号"><Input placeholder="v1.0" /></Form.Item>
          <Form.Item name="enabled" label="是否启用" valuePropName="checked"><Switch defaultChecked /></Form.Item>
        </Form>
      </Modal>
    </div>
  );
}

export default PromptSettings;
