import { useCallback, useEffect, useMemo, useState } from "react";
import {
  AutoComplete,
  Button,
  Checkbox,
  Form,
  Input,
  Modal,
  Select,
  Space,
  Switch,
  Table,
  Tag,
  message,
} from "antd";
import { PlusOutlined, ReloadOutlined } from "@ant-design/icons";
import { AppCard, FormCol, FormGrid, Page, PageHeader } from "../components/ui";
import {
  createModel,
  createProvider,
  deleteModel,
  deleteProvider,
  fetchProviderModels,
  getModels,
  getProviders,
  updateModel,
  updateProvider,
} from "../api/llm";

const providerTypeOptions = [
  { value: "openai-compatible", label: "OpenAI 兼容接口" },
  { value: "official", label: "官方接口" },
  { value: "local", label: "本地模型服务" },
  { value: "custom", label: "自定义服务" },
];

const levelOptions = ["低", "中", "高"].map((value) => ({ value, label: value }));

const usageOptions = [
  { value: "fast", label: "快速任务" },
  { value: "quality", label: "高质量任务" },
  { value: "ocr", label: "OCR 修正" },
  { value: "question", label: "题目解析" },
  { value: "essay", label: "申论批改" },
  { value: "summary", label: "日志总结" },
  { value: "plan", label: "计划生成" },
];

function LLMSettings() {
  const [providerOpen, setProviderOpen] = useState(false);
  const [modelOpen, setModelOpen] = useState(false);
  const [providers, setProviders] = useState([]);
  const [models, setModels] = useState([]);
  const [loading, setLoading] = useState(false);
  const [remoteModels, setRemoteModels] = useState([]);
  const [remoteModelLoading, setRemoteModelLoading] = useState(false);
  const [editingProvider, setEditingProvider] = useState(null);
  const [editingModel, setEditingModel] = useState(null);
  const [providerForm] = Form.useForm();
  const [modelForm] = Form.useForm();

  const providerOptions = useMemo(() => providers.map((item) => ({
    value: item.id,
    label: item.name,
  })), [providers]);

  const remoteModelOptions = useMemo(() => remoteModels.map((item) => ({
    value: item.id || item.name,
    label: item.name && item.name !== item.id ? `${item.name} (${item.id})` : item.id,
  })), [remoteModels]);

  const loadData = useCallback(async () => {
    setLoading(true);
    try {
      const [providerList, modelList] = await Promise.all([
        getProviders(),
        getModels(),
      ]);
      setProviders(providerList || []);
      setModels(modelList || []);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    queueMicrotask(() => loadData());
  }, [loadData]);

  const openProviderModal = (record) => {
    setEditingProvider(record || null);
    providerForm.setFieldsValue(record ? { ...record, api_key: "" } : {
      provider_type: "openai-compatible",
      enabled: true,
    });
    setProviderOpen(true);
  };

  const openModelModal = (record) => {
    const formValues = modelToForm(record);
    setEditingModel(record || null);
    modelForm.setFieldsValue(formValues);
    if (formValues.provider_id) {
      loadRemoteModels(formValues.provider_id, { silent: true });
    } else {
      setRemoteModels([]);
    }
    setModelOpen(true);
  };

  const loadRemoteModels = async (providerId, options = {}) => {
    if (!providerId) {
      setRemoteModels([]);
      return;
    }
    setRemoteModelLoading(true);
    try {
      const items = await fetchProviderModels(providerId);
      setRemoteModels(items || []);
      if (!options.silent) {
        message.success(`已拉取 ${items?.length || 0} 个可选模型`);
      }
    } catch {
      setRemoteModels([]);
      if (!options.silent) {
        message.warning("模型列表拉取失败，可继续手动填写模型名");
      }
    } finally {
      setRemoteModelLoading(false);
    }
  };

  const handleSaveProvider = async () => {
    const values = await providerForm.validateFields();
    if (editingProvider) {
      await updateProvider(editingProvider.id, values);
      message.success("Provider 已更新");
    } else {
      await createProvider(values);
      message.success("Provider 已新增");
    }
    setProviderOpen(false);
    providerForm.resetFields();
    loadData();
  };

  const handleSaveModel = async () => {
    const values = await modelForm.validateFields();
    const payload = modelFormToPayload(values, providers);
    if (editingModel) {
      await updateModel(editingModel.id, payload);
      message.success("模型已更新");
    } else {
      await createModel(payload);
      message.success("模型已新增");
    }
    setModelOpen(false);
    modelForm.resetFields();
    loadData();
  };

  const handleDeleteProvider = async (id) => {
    await deleteProvider(id);
    message.success("Provider 已删除");
    loadData();
  };

  const handleDeleteModel = async (id) => {
    await deleteModel(id);
    message.success("模型已删除");
    loadData();
  };

  const providerColumns = [
    { title: "服务商", dataIndex: "name" },
    { title: "类型", dataIndex: "provider_type" },
    { title: "Base URL", dataIndex: "base_url" },
    { title: "API Key", render: (_, record) => record.has_api_key ? <Tag>{record.api_key_masked}</Tag> : <Tag>未设置</Tag> },
    { title: "状态", dataIndex: "enabled", render: (value) => <Tag color={value ? "blue" : "default"}>{value ? "启用" : "停用"}</Tag> },
    { title: "备注", dataIndex: "note" },
    { title: "操作", render: (_, record) => <Space><Button size="small" onClick={() => openProviderModal(record)}>编辑</Button><Button size="small" danger onClick={() => handleDeleteProvider(record.id)}>删除</Button></Space> },
  ];

  const modelColumns = [
    { title: "所属服务商", dataIndex: "provider" },
    { title: "模型名称", dataIndex: "name" },
    { title: "别名", dataIndex: "alias" },
    { title: "成本", dataIndex: "cost_level" },
    { title: "速度", dataIndex: "speed_level" },
    { title: "质量", dataIndex: "quality_level" },
    { title: "状态", dataIndex: "enabled", render: (value) => <Tag color={value ? "green" : "default"}>{value ? "启用" : "停用"}</Tag> },
    { title: "用途", render: (_, record) => <Space wrap>{modelUsageTags(record)}</Space> },
    { title: "操作", render: (_, record) => <Space><Button size="small" onClick={() => openModelModal(record)}>编辑</Button><Button size="small" danger onClick={() => handleDeleteModel(record.id)}>删除</Button></Space> },
  ];

  return (
    <Page>
      <PageHeader eyebrow="LLM" title="LLM 配置" description="维护服务商、Base URL、API Key 和模型用途。" />
      <AppCard title="Provider 列表" extra={<Button icon={<PlusOutlined />} onClick={() => openProviderModal()}>新增 Provider</Button>}>
        <Table rowKey="id" columns={providerColumns} dataSource={providers} loading={loading} pagination={false} />
      </AppCard>
      <AppCard title="模型列表" extra={<Button icon={<PlusOutlined />} onClick={() => openModelModal()}>新增模型</Button>}>
        <Table rowKey="id" columns={modelColumns} dataSource={models} loading={loading} pagination={false} scroll={{ x: 980 }} />
      </AppCard>
      <Modal
        title="Provider 表单"
        open={providerOpen}
        onCancel={() => setProviderOpen(false)}
        onOk={handleSaveProvider}
        width={720}
      >
        <Form form={providerForm} layout="vertical">
          <FormGrid>
            <FormCol>
              <Form.Item name="name" label="服务商名称" rules={[{ required: true, message: "请输入服务商名称" }]}>
                <Input placeholder="例如 OpenAI / DeepSeek / Local LLM" />
              </Form.Item>
            </FormCol>
            <FormCol>
              <Form.Item name="provider_type" label="服务商类型" rules={[{ required: true, message: "请选择服务商类型" }]}>
                <Select placeholder="选择接口类型" options={providerTypeOptions} />
              </Form.Item>
            </FormCol>
          </FormGrid>
          <Form.Item name="base_url" label="Base URL" rules={[{ required: true, message: "请输入 Base URL" }]}>
            <Input placeholder="https://api.example.com/v1" />
          </Form.Item>
          <Form.Item
            name="api_key"
            label={editingProvider?.has_api_key ? `API Key（当前 ${editingProvider.api_key_masked}）` : "API Key"}
          >
            <Input.Password placeholder={editingProvider ? "留空则保留当前 API Key" : "请输入 API Key"} />
          </Form.Item>
          <FormGrid>
            <FormCol>
              <Form.Item name="enabled" label="是否启用" valuePropName="checked">
                <Switch defaultChecked />
              </Form.Item>
            </FormCol>
            <FormCol>
              <Form.Item name="usage" label="默认用途">
                <Select
                  mode="multiple"
                  placeholder="可多选"
                  options={usageOptions}
                />
              </Form.Item>
            </FormCol>
          </FormGrid>
          <Form.Item name="note" label="备注"><Input.TextArea rows={3} /></Form.Item>
        </Form>
      </Modal>
      <Modal
        title="模型表单"
        open={modelOpen}
        onCancel={() => {
          setModelOpen(false);
          setRemoteModels([]);
        }}
        onOk={handleSaveModel}
        width={760}
      >
        <Form form={modelForm} layout="vertical">
          <FormGrid>
            <FormCol>
              <Form.Item name="provider_id" label="所属服务商" rules={[{ required: true, message: "请选择 Provider" }]}>
                <Select
                  placeholder="选择 Provider"
                  options={providerOptions}
                  onChange={(value) => loadRemoteModels(value)}
                />
              </Form.Item>
            </FormCol>
            <FormCol>
              <Form.Item
                label="模型名称"
                required
                extra={remoteModelLoading
                  ? "正在从所选 Provider 拉取模型列表..."
                  : remoteModels.length > 0
                    ? `已拉取 ${remoteModels.length} 个模型，也可以直接手填。`
                    : "选择 Provider 后会自动拉取模型；如果服务不支持列表接口，可以直接手填。"}
              >
                <Space.Compact style={{ width: "100%" }}>
                  <Form.Item name="name" noStyle rules={[{ required: true, message: "请输入模型名称" }]}>
                    <AutoComplete
                      options={remoteModelOptions}
                      placeholder="选择模型，或直接输入模型名"
                      filterOption={(inputValue, option) => (
                        String(option?.value || "").toLowerCase().includes(inputValue.toLowerCase())
                        || String(option?.label || "").toLowerCase().includes(inputValue.toLowerCase())
                      )}
                      onFocus={() => {
                        const providerId = modelForm.getFieldValue("provider_id");
                        if (providerId && remoteModels.length === 0) {
                          loadRemoteModels(providerId, { silent: true });
                        }
                      }}
                    />
                  </Form.Item>
                  <Button
                    icon={<ReloadOutlined />}
                    loading={remoteModelLoading}
                    onClick={() => loadRemoteModels(modelForm.getFieldValue("provider_id"))}
                  />
                </Space.Compact>
              </Form.Item>
            </FormCol>
          </FormGrid>
          <Form.Item name="alias" label="模型别名">
            <Input placeholder="例如：高质量解析 / 快速草稿 / 本地离线" />
          </Form.Item>
          <FormGrid>
            <FormCol sm={8}>
              <Form.Item name="cost_level" label="成本等级">
                <Select placeholder="选择成本" options={levelOptions} />
              </Form.Item>
            </FormCol>
            <FormCol sm={8}>
              <Form.Item name="speed_level" label="速度等级">
                <Select placeholder="选择速度" options={levelOptions} />
              </Form.Item>
            </FormCol>
            <FormCol sm={8}>
              <Form.Item name="quality_level" label="质量等级">
                <Select placeholder="选择质量" options={levelOptions} />
              </Form.Item>
            </FormCol>
          </FormGrid>
          <Form.Item name="usage" label="模型用途">
            <Checkbox.Group options={usageOptions} />
          </Form.Item>
          <Form.Item name="enabled" label="是否启用" valuePropName="checked">
            <Switch defaultChecked />
          </Form.Item>
        </Form>
      </Modal>
    </Page>
  );
}

function modelUsageTags(record) {
  const map = [
    ["use_fast", "快速任务"],
    ["use_quality", "高质量任务"],
    ["use_ocr", "OCR"],
    ["use_question", "题目解析"],
    ["use_essay", "申论"],
    ["use_summary", "总结"],
    ["use_plan", "计划"],
  ];

  return map
    .filter(([key]) => record[key])
    .map(([, label]) => <Tag key={label}>{label}</Tag>);
}

function modelToForm(record) {
  if (!record) {
    return {
      enabled: true,
      usage: ["fast"],
    };
  }

  return {
    ...record,
    usage: [
      record.use_fast && "fast",
      record.use_quality && "quality",
      record.use_ocr && "ocr",
      record.use_question && "question",
      record.use_essay && "essay",
      record.use_summary && "summary",
      record.use_plan && "plan",
    ].filter(Boolean),
  };
}

function modelFormToPayload(values, providers) {
  const provider = providers.find((item) => item.id === values.provider_id);
  const usage = values.usage || [];

  return {
    ...values,
    provider: provider?.name || "",
    use_fast: usage.includes("fast"),
    use_quality: usage.includes("quality"),
    use_ocr: usage.includes("ocr"),
    use_question: usage.includes("question"),
    use_essay: usage.includes("essay"),
    use_summary: usage.includes("summary"),
    use_plan: usage.includes("plan"),
  };
}

export default LLMSettings;
