import { useEffect, useMemo, useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { Alert, AutoComplete, Button, Card, Checkbox, Form, Input, InputNumber, List, Radio, Row, Col, Space, Switch, Tabs, Tag, Typography, Upload, message, Progress } from "antd";
import LLMSettings from "./LLMSettings";
import PromptSettings from "./PromptSettings";
import { getOcrConfig, getOcrMonthUsage, updateOcrConfig } from "../api/ocr";
import { exportBackup } from "../api/backup";
import { getModels } from "../api/llm";
import { parsePdfTest } from "../api/pdf";

const { Text } = Typography;

const OCR_CONFIG_KEY = "aozora-ocr-config";

function AISettings() {
  const location = useLocation();
  const navigate = useNavigate();
  const activeTab = new URLSearchParams(location.search).get("tab") || "llm";

  return (
    <Tabs
      className="ai-settings-tabs"
      activeKey={activeTab}
      onChange={(key) => navigate(`/ai?tab=${key}`, { replace: true })}
      items={[
        {
          key: "llm",
          label: "LLM 提供方与模型",
          children: <LLMSettings />,
        },
        {
          key: "prompts",
          label: "Prompt 配置",
          children: <PromptSettings />,
        },
        {
          key: "ocr",
          label: "OCR 配置",
          children: <OCRConfig />,
        },
        {
          key: "pdf",
          label: "PDF 解析",
          children: <PDFCapability />,
        },
        {
          key: "backup",
          label: "数据备份",
          children: <BackupExportPanel />,
        },
      ]}
    />
  );
}

function OCRConfig() {
  const [form] = Form.useForm();
  const [saving, setSaving] = useState(false);
  const [masked, setMasked] = useState({});
  const [usage, setUsage] = useState();
  const [llmModels, setLlmModels] = useState([]);

  const llmModelOptions = useMemo(() => llmModels
    .filter((item) => item.enabled !== false)
    .map((item) => ({
      value: item.name,
      label: [item.alias || item.name, item.provider].filter(Boolean).join(" · "),
    })), [llmModels]);

  useEffect(() => {
    const raw = localStorage.getItem(OCR_CONFIG_KEY);
    form.setFieldsValue(raw ? JSON.parse(raw) : defaultOCRConfig);
    getOcrConfig()
      .then((config) => {
        setMasked(config || {});
        form.setFieldsValue({
          baiduEnabled: config.enabled,
          baiduMonthlyLimit: config.monthly_limit,
          baiduTimeoutSeconds: config.timeout_seconds,
          baiduSource: config.source,
          engineLimits: config.engine_limits || {},
        });
      })
      .catch(() => {});
    getOcrMonthUsage()
      .then((nextUsage) => {
        setUsage(nextUsage);
        form.setFieldsValue({ usageText: `${nextUsage.used} 次 · ${nextUsage.month}` });
      })
      .catch(() => {});
    getModels()
      .then((items) => setLlmModels(items || []))
      .catch(() => {});
  }, [form]);

  const handleChange = (_, values) => {
    localStorage.setItem(OCR_CONFIG_KEY, JSON.stringify(values));
  };

  const handleSaveBaidu = async () => {
    const values = form.getFieldsValue();
    setSaving(true);
    try {
      const config = await updateOcrConfig({
        enabled: values.baiduEnabled,
        api_key: values.baiduApiKey,
        secret_key: values.baiduSecretKey,
        monthly_limit: values.baiduMonthlyLimit,
        engine_limits: values.engineLimits || {},
        timeout_seconds: values.baiduTimeoutSeconds,
      });
      setMasked(config || {});
      form.setFieldsValue({ baiduApiKey: "", baiduSecretKey: "", baiduSource: config.source });
      message.success("百度 OCR 配置已保存到服务器本地");
    } finally {
      setSaving(false);
    }
  };

  return (
    <Card className="glass-card" title="OCR 识别方式" bordered={false}>
      <Alert
        type="info"
        showIcon
        message="百度 OCR AK/SK 保存在后端服务器本地"
        description="这里填写后会写入 backend/data/ocr_config.json，本地文件已加入 .gitignore。当前没有登录权限系统，未来公开服务时需要只允许 admin/root 修改。百度不同 OCR 能力额度独立，下面可以为每个能力设置本地上限。"
      />
      <Form form={form} layout="vertical" initialValues={defaultOCRConfig} onValuesChange={handleChange}>
        <Form.Item name="mode" label="识别模式">
          <Radio.Group>
            <Radio.Button value="baidu">百度 OCR</Radio.Button>
            <Radio.Button value="local_paddle">本地 PaddleOCR</Radio.Button>
            <Radio.Button value="vl_model">VL 大模型</Radio.Button>
          </Radio.Group>
        </Form.Item>
        <Form.Item noStyle shouldUpdate={(prev, cur) => prev.mode !== cur.mode}>
          {({ getFieldValue }) => {
            const mode = getFieldValue("mode");
            if (mode === "baidu") {
              return (
                <>
                  <Card className="glass-card settings-sub-card" bordered={false}>
                    <Row gutter={[14, 14]}>
                      <Col xs={24} md={8}>
                        <Form.Item name="baiduEnabled" label="启用百度 OCR" valuePropName="checked">
                          <Switch />
                        </Form.Item>
                      </Col>
                      <Col xs={24} md={8}>
                        <Form.Item name="usageText" label="本地调用统计">
                          <Input readOnly placeholder="后端启动后自动读取" />
                        </Form.Item>
                      </Col>
                      <Col xs={24} md={8}>
                        <Form.Item name="baiduSource" label="配置来源">
                          <Input readOnly placeholder="env / local_file" />
                        </Form.Item>
                      </Col>
                      <Col xs={24} md={12}>
                        <Form.Item label={`API Key ${masked.api_key_masked ? `(${masked.api_key_masked})` : ""}`} name="baiduApiKey">
                          <Input.Password placeholder="留空则保持当前值" autoComplete="new-password" />
                        </Form.Item>
                      </Col>
                      <Col xs={24} md={12}>
                        <Form.Item label={`Secret Key ${masked.secret_masked ? `(${masked.secret_masked})` : ""}`} name="baiduSecretKey">
                          <Input.Password placeholder="留空则保持当前值" autoComplete="new-password" />
                        </Form.Item>
                      </Col>
                      <Col xs={24} md={12}>
                        <Form.Item name="baiduMonthlyLimit" label="本地月上限">
                          <InputNumber min={0} className="full-input" />
                        </Form.Item>
                      </Col>
                      <Col xs={24} md={12}>
                        <Form.Item name="baiduTimeoutSeconds" label="请求超时秒数">
                          <InputNumber min={5} max={120} className="full-input" />
                        </Form.Item>
                      </Col>
                    </Row>
                    <Button type="primary" onClick={handleSaveBaidu} loading={saving}>
                      保存百度 OCR 配置
                    </Button>
                  </Card>
                  <Card className="glass-card settings-sub-card" title="百度 OCR 能力额度" bordered={false}>
                    <Row gutter={[14, 14]}>
                      {(usage?.engines || []).map((item) => (
                        <Col xs={24} md={12} xl={8} key={item.key}>
                          <div className="ocr-quota-card">
                            <div className="ocr-quota-head">
                              <strong>{item.label}</strong>
                              <span>{item.used} / {item.limit || "不限"}</span>
                            </div>
                            <Progress
                              percent={item.limit ? Math.min(100, Math.round((item.used / item.limit) * 100)) : 0}
                              showInfo={false}
                              size="small"
                            />
                            <Form.Item name={["engineLimits", item.key]} label="本地月上限">
                              <InputNumber min={0} className="full-input" placeholder="0 表示不限" />
                            </Form.Item>
                            <Text type="secondary">{item.description}</Text>
                          </div>
                        </Col>
                      ))}
                    </Row>
                  </Card>
                  <Space direction="vertical">
                    <Text>印刷体题目：默认走通用文字识别。</Text>
                    <Text>高精度印刷体：默认走高精度文字识别。</Text>
                    <Text>手写内容：默认走手写文字识别。</Text>
                    <Text type="secondary">申论手写作文为异步接口，当前已预留，后续补轮询。</Text>
                  </Space>
                </>
              );
            }
            if (mode === "vl_model") {
              return (
                <>
                  <Form.Item name="vlModelId" label="VL 模型">
                    <AutoComplete
                      placeholder="从 LLM 配置中选择模型，或直接手填"
                      options={llmModelOptions}
                      filterOption={(inputValue, option) => (
                        String(option?.value || "").toLowerCase().includes(inputValue.toLowerCase())
                        || String(option?.label || "").toLowerCase().includes(inputValue.toLowerCase())
                      )}
                    />
                  </Form.Item>
                  <Alert type="info" showIcon message="这里读取已启用的 LLM 模型作为候选；如果模型未入库，也可以直接手填。" />
                </>
              );
            }

            return (
              <>
                <Space align="start" wrap>
                  <Form.Item name="localHost" label="PaddleOCR Host">
                    <Input style={{ width: 220 }} placeholder="localhost" />
                  </Form.Item>
                  <Form.Item name="localPort" label="端口">
                    <InputNumber min={21000} max={65535} />
                  </Form.Item>
                </Space>
                <Form.Item name="endpoint" label="完整 OCR Endpoint">
                  <Input placeholder="http://localhost:21090/ocr" />
                </Form.Item>
              </>
            );
          }}
        </Form.Item>
        <Form.Item name="useLLMCorrection" label="LLM 纠错" valuePropName="checked">
          <Switch />
        </Form.Item>
      </Form>
    </Card>
  );
}

function PDFCapability() {
  const [form] = Form.useForm();
  const [testing, setTesting] = useState(false);
  const [testResult, setTestResult] = useState(null);

  useEffect(() => {
    const raw = localStorage.getItem("aozora-pdf-config");
    form.setFieldsValue(raw ? JSON.parse(raw) : defaultPDFConfig);
  }, [form]);

  const handleChange = (_, values) => {
    localStorage.setItem("aozora-pdf-config", JSON.stringify(values));
  };

  const handleTestPDF = async (file) => {
    setTesting(true);
    try {
      const result = await parsePdfTest(file);
      setTestResult(result);
      message.success("PDF 解析测试完成");
    } catch (error) {
      setTestResult(null);
      message.error(error.message || "PDF 解析测试失败");
    } finally {
      setTesting(false);
    }
  };

  return (
    <Card className="glass-card" title="PDF 解析配置" bordered={false}>
      <Space direction="vertical" size="middle" className="backup-panel">
        <Form form={form} layout="vertical" initialValues={defaultPDFConfig} onValuesChange={handleChange}>
          <Card className="glass-card settings-sub-card" title="PDF 解析测试端口" bordered={false}>
            <Alert
              type="success"
              showIcon
              message="Go PDF 文本解析测试"
              description="这个接口不入库、不切 chunk、不分类，只返回 Go PDF 解析库从每页抽出的文本，方便判断 PDF 文本层到底是什么情况。"
            />
            <Row gutter={[14, 14]}>
              <Col xs={24} md={8}>
                <Form.Item name="testHost" label="测试 Host">
                  <Input placeholder="localhost" />
                </Form.Item>
              </Col>
              <Col xs={24} md={8}>
                <Form.Item name="testPort" label="测试端口">
                  <InputNumber min={21000} max={65535} className="full-input" />
                </Form.Item>
              </Col>
              <Col xs={24} md={8}>
                <Form.Item name="testEndpoint" label="测试 Endpoint">
                  <Input readOnly placeholder="/api/pdf/parse-test" />
                </Form.Item>
              </Col>
            </Row>
            <Upload beforeUpload={(file) => { handleTestPDF(file); return false; }} showUploadList={false} accept="application/pdf">
              <Button type="primary" loading={testing}>上传 PDF 并输出文字</Button>
            </Upload>
          </Card>
        </Form>

        {testResult && (
          <Card className="glass-card settings-sub-card" title="测试输出" bordered={false}>
            <Space wrap>
              <Tag color={testResult.quality?.ok ? "green" : "red"}>{testResult.quality?.ok ? "文本层可用" : "文本层异常"}</Tag>
              <Tag>{testResult.page_count || 0} 页</Tag>
              <Tag>{testResult.total_chars || 0} 字符</Tag>
              {testResult.quality?.reason && <Tag color="blue">{testResult.quality.reason}</Tag>}
            </Space>
            <List
              className="pdf-test-pages"
              dataSource={testResult.pages || []}
              pagination={{ pageSize: 3 }}
              renderItem={(page) => (
                <List.Item className="pdf-test-page">
                  <div className="pdf-test-page-title">第 {page.PageNo || page.page_no} 页</div>
                  <pre>{page.Text || page.text || ""}</pre>
                </List.Item>
              )}
            />
          </Card>
        )}
      </Space>
    </Card>
  );
}

function BackupExportPanel() {
  const [includeSecrets, setIncludeSecrets] = useState(false);
  const [exporting, setExporting] = useState(false);

  const handleExport = async () => {
    setExporting(true);
    try {
      const blob = await exportBackup({ includeSecrets });
      const url = URL.createObjectURL(blob);
      const link = document.createElement("a");
      link.href = url;
      link.download = `gkweb-backup-${new Date().toISOString().slice(0, 19).replace(/[:T]/g, "-")}.json`;
      document.body.appendChild(link);
      link.click();
      link.remove();
      URL.revokeObjectURL(url);
      message.success("备份已导出");
    } finally {
      setExporting(false);
    }
  };

  return (
    <Card className="glass-card" title="数据备份导出" bordered={false}>
      <Space direction="vertical" size="middle" className="backup-panel">
        <Alert
          type="info"
          showIcon
          message="导出为 JSON 备份"
          description="当前导出数据库中的业务表数据，不包含上传的音乐、图片等二进制文件。默认不包含密钥，适合日常备份。"
        />
        <Checkbox checked={includeSecrets} onChange={(event) => setIncludeSecrets(event.target.checked)}>
          包含敏感配置，例如 LLM Provider API Key
        </Checkbox>
        <Button type="primary" onClick={handleExport} loading={exporting}>
          导出 JSON 备份
        </Button>
      </Space>
    </Card>
  );
}

const defaultOCRConfig = {
  mode: "baidu",
  localHost: "localhost",
  localPort: 21090,
  endpoint: "http://localhost:21090/ocr",
  useLLMCorrection: true,
};

const defaultPDFConfig = {
  testHost: "localhost",
  testPort: 21080,
  testEndpoint: "http://localhost:21080/api/essay/documents/:id/parse",
};

export default AISettings;
