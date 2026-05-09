import { useEffect, useMemo, useState } from "react";
import {
  Alert,
  Button,
  Card,
  Col,
  Form,
  Input,
  List,
  Row,
  Select,
  Space,
  Statistic,
  Steps,
  Table,
  Tag,
  Upload,
  message,
} from "antd";
import {
  ApartmentOutlined,
  CheckCircleOutlined,
  CloudUploadOutlined,
  FileSearchOutlined,
  RobotOutlined,
  ScissorOutlined,
} from "@ant-design/icons";
import PageHeader from "../components/PageHeader";
import {
  assembleEssayQuestions,
  classifyEssayChunks,
  createEssayDocument,
  getEssayChunks,
  getEssayDocuments,
  getEssayQuestions,
  parseEssayDocument,
  reviewEssayQuestion,
} from "../api/essay";
import { getModels } from "../api/llm";

const chunkTypeMeta = {
  material: { label: "材料", color: "blue" },
  question: { label: "题目", color: "green" },
  reference_answer: { label: "参考答案", color: "purple" },
  scoring_rule: { label: "评分规则", color: "orange" },
  explanation: { label: "解析说明", color: "cyan" },
  unknown: { label: "未知", color: "default" },
};

const documentRoleOptions = [
  { value: "combined", label: "混合卷：材料 / 题目 / 答案都在一个 PDF" },
  { value: "question_paper", label: "题目卷：材料与题目" },
  { value: "answer_key", label: "答案卷：参考答案与评分规则" },
  { value: "explanation", label: "解析卷：解析说明" },
];

const documentRoleLabels = {
  combined: "混合卷",
  question_paper: "题目卷",
  answer_key: "答案卷",
  explanation: "解析卷",
};

function EssayReview() {
  const [uploadForm] = Form.useForm();
  const [reviewForm] = Form.useForm();
  const [documents, setDocuments] = useState([]);
  const [selectedDocumentId, setSelectedDocumentId] = useState();
  const [chunks, setChunks] = useState([]);
  const [questions, setQuestions] = useState([]);
  const [models, setModels] = useState([]);
  const [activeStep, setActiveStep] = useState(0);
  const [loading, setLoading] = useState(false);
  const [reviewResult, setReviewResult] = useState(null);
  const [operations, setOperations] = useState([]);

  const selectedDocument = useMemo(
    () => documents.find((item) => item.id === selectedDocumentId),
    [documents, selectedDocumentId]
  );

  const fastModelOptions = useMemo(() => {
    const enabled = models.filter((item) => item.enabled);
    const fast = enabled.filter((item) => item.use_fast || item.speed_level === "高");
    return (fast.length ? fast : enabled).map(modelOption);
  }, [models]);

  const reviewModelOptions = useMemo(() => {
    const enabled = models.filter((item) => item.enabled);
    const essay = enabled.filter((item) => item.use_essay || item.use_quality);
    return (essay.length ? essay : enabled).map(modelOption);
  }, [models]);

  useEffect(() => {
    loadDocuments();
    getModels().then((items) => setModels(items || [])).catch(() => {});
  }, []);

  useEffect(() => {
    if (!selectedDocumentId) return;
    Promise.all([
      getEssayChunks(selectedDocumentId).catch(() => []),
      getEssayQuestions(selectedDocumentId).catch(() => []),
    ]).then(([nextChunks, nextQuestions]) => {
      setChunks(nextChunks || []);
      setQuestions(nextQuestions || []);
      setActiveStep(stepFromState(selectedDocument, nextChunks || [], nextQuestions || []));
    });
  }, [selectedDocumentId, selectedDocument]);

  const loadDocuments = async () => {
    const items = await getEssayDocuments().catch(() => []);
    setDocuments(items || []);
    if (!selectedDocumentId && items?.[0]) {
      setSelectedDocumentId(items[0].id);
    }
  };

  const runOperation = async (label, action) => {
    const start = Date.now();
    const key = `${start}-${label}`;
    setOperations((items) => [
      { key, label, status: "running", startedAt: new Date(start).toLocaleTimeString(), duration: 0 },
      ...items.slice(0, 5),
    ]);
    setLoading(true);
    try {
      const result = await action();
      const duration = Date.now() - start;
      setOperations((items) => items.map((item) => (
        item.key === key ? { ...item, status: "success", duration, message: "完成" } : item
      )));
      return result;
    } catch (error) {
      const duration = Date.now() - start;
      setOperations((items) => items.map((item) => (
        item.key === key ? { ...item, status: "error", duration, message: error.message || "操作失败" } : item
      )));
      message.error(error.message || "操作失败");
      throw error;
    } finally {
      setLoading(false);
    }
  };

  const handleCreateDocument = async () => {
    const values = await uploadForm.validateFields();
    const file = values.file?.[0]?.originFileObj;
    try {
      await runOperation("上传并自动解析 PDF", async () => {
        const result = await createEssayDocument({
          file,
          title: values.title,
          rawText: values.raw_text,
          documentRole: values.document_role,
          sourceGroup: values.source_group,
        });
        const document = result.document || result;
        message.success(result.chunks ? "文档已上传并自动解析" : "文档已上传");
        uploadForm.resetFields();
        await loadDocuments();
        setSelectedDocumentId(document.id);
        if (result.chunks) setChunks(result.chunks);
        setActiveStep(result.chunks ? 1 : 0);
      });
    } catch {
      await loadDocuments();
    }
  };

  const handleParse = async () => {
    if (!selectedDocumentId) return;
    try {
      await runOperation("重新解析 PDF", async () => {
        const result = await parseEssayDocument(selectedDocumentId, {});
        setChunks(result.chunks || []);
        await loadDocuments();
        setActiveStep(1);
        message.success("已解析并切分 chunk");
      });
    } catch {
      await loadDocuments();
    }
  };

  const handleClassify = async () => {
    if (!selectedDocumentId) return;
    const modelId = uploadForm.getFieldValue("class_model_id");
    try {
      await runOperation("调用分类模型 / 规则分类", async () => {
        const result = await classifyEssayChunks(selectedDocumentId, { model_id: modelId });
        setChunks(result || []);
        await loadDocuments();
        setActiveStep(2);
        message.success("已完成 chunk 分类骨架");
      });
    } catch {
      await loadDocuments();
    }
  };

  const handleAssemble = async () => {
    if (!selectedDocumentId) return;
    try {
      await runOperation("组装 essay_questions", async () => {
        const result = await assembleEssayQuestions(selectedDocumentId);
        setQuestions(result || []);
        await loadDocuments();
        setActiveStep(3);
        message.success("已根据分类结果组装题目");
      });
    } catch {
      await loadDocuments();
    }
  };

  const handleReview = async () => {
    const values = await reviewForm.validateFields();
    try {
      await runOperation("调用高质量模型批改", async () => {
        const result = await reviewEssayQuestion(values.question_id, {
          model_id: values.review_model_id,
          user_answer: values.user_answer,
        });
        setReviewResult(result);
        message.success("批改结果已生成");
      });
    } catch {
      await loadDocuments();
    }
  };

  const chunkStats = useMemo(() => summarizeChunks(chunks), [chunks]);

  return (
    <div className="page-grid essay-workbench">
      <PageHeader
        eyebrow="Essay Pipeline"
        title="申论 PDF 结构化"
        desc="先解析 PDF、切分 chunk、分类内容，再组装题目并进入高质量模型批改。"
      />

      <Steps
        className="essay-steps"
        current={activeStep}
        items={[
          { title: "PDF 文档", icon: <CloudUploadOutlined /> },
          { title: "Chunk 分类", icon: <ScissorOutlined /> },
          { title: "题目组装", icon: <ApartmentOutlined /> },
          { title: "答案批改", icon: <RobotOutlined /> },
        ]}
      />

      <Row gutter={[18, 18]} align="stretch">
        <Col xs={24} xl={8}>
          <Space direction="vertical" size="middle" className="essay-result-stack">
            <Card className="glass-card" title="上传与解析" bordered={false}>
              <Alert
                type="info"
                showIcon
                message="上传后会自动解析"
                description="文本型 PDF 会立即切分为 chunk；扫描件或解析失败会显示错误并记录在文档状态里。题目卷和答案卷可以用同一个卷套标识归档。"
              />
              <Form form={uploadForm} layout="vertical" className="essay-upload-form">
                <Form.Item name="title" label="文档标题">
                  <Input placeholder="例如 2025 省考申论 A 卷" />
                </Form.Item>
                <Form.Item name="source_group" label="卷套标识">
                  <Input placeholder="例如 2025-省考-A，用于关联题目卷和答案卷" />
                </Form.Item>
                <Form.Item name="document_role" label="PDF 类型" initialValue="combined">
                  <Select options={documentRoleOptions} />
                </Form.Item>
                <Form.Item
                  name="file"
                  label="申论 PDF"
                  valuePropName="fileList"
                  getValueFromEvent={(event) => event?.fileList || []}
                >
                  <Upload beforeUpload={() => false} maxCount={1} accept="application/pdf">
                    <Button icon={<CloudUploadOutlined />}>选择 PDF</Button>
                  </Upload>
                </Form.Item>
                <Form.Item name="raw_text" label="开发期原始文本">
                  <Input.TextArea rows={5} placeholder="可选：粘贴 PDF 提取文本，用空行分段，分页可用 --- page 1 ---" />
                </Form.Item>
                <Form.Item name="class_model_id" label="PDF 分类模型">
                  <Select allowClear placeholder="选择快速/廉价模型" options={fastModelOptions} />
                </Form.Item>
                <Space wrap>
                  <Button type="primary" loading={loading} onClick={handleCreateDocument}>上传并自动解析</Button>
                  <Button icon={<FileSearchOutlined />} disabled={!selectedDocumentId} loading={loading} onClick={handleParse}>重新解析</Button>
                  <Button icon={<RobotOutlined />} disabled={!chunks.length} loading={loading} onClick={handleClassify}>分类 chunk</Button>
                  <Button icon={<ApartmentOutlined />} disabled={!chunks.length} loading={loading} onClick={handleAssemble}>组装题目</Button>
                </Space>
              </Form>
            </Card>

            <Card className="glass-card" title="操作状态" bordered={false}>
              <List
                dataSource={operations}
                locale={{ emptyText: "暂无操作" }}
                renderItem={(item) => (
                  <List.Item className={`essay-operation ${item.status}`}>
                    <List.Item.Meta
                      title={<Space><OperationTag status={item.status} /><span>{item.label}</span></Space>}
                      description={`${item.startedAt} · ${formatDuration(item.duration)}${item.message ? ` · ${item.message}` : ""}`}
                    />
                  </List.Item>
                )}
              />
            </Card>

            <Card className="glass-card" title="文档列表" bordered={false}>
              <List
                dataSource={documents}
                locale={{ emptyText: "暂无申论 PDF" }}
                renderItem={(item) => (
                  <List.Item
                    className={item.id === selectedDocumentId ? "essay-document active" : "essay-document"}
                    onClick={() => setSelectedDocumentId(item.id)}
                  >
                    <List.Item.Meta
                      title={<Space wrap><span>{item.title}</span><Tag>{documentRoleLabels[item.document_role] || item.document_role || "混合卷"}</Tag><StatusTag status={item.status} /></Space>}
                      description={`${item.source_group ? `${item.source_group} · ` : ""}${item.page_count || 0} 页 · ${item.chunk_count || 0} chunks${item.note ? ` · ${item.note}` : ""}`}
                    />
                  </List.Item>
                )}
              />
            </Card>
          </Space>
        </Col>

        <Col xs={24} xl={16}>
          <Space direction="vertical" size="middle" className="essay-result-stack">
            <Row gutter={[16, 16]}>
              {Object.entries(chunkStats).map(([type, count]) => (
                <Col xs={12} md={8} xl={4} key={type}>
                  <Card className="glass-card essay-mini-stat" bordered={false}>
                    <Statistic title={chunkTypeMeta[type]?.label || type} value={count} />
                  </Card>
                </Col>
              ))}
            </Row>

            <Card
              className="glass-card"
              title="Chunk 原文与分类"
              extra={<Tag color="default">当前为规则分类，后续接快速模型</Tag>}
              bordered={false}
            >
              <Alert
                type="info"
                showIcon
                message="分类依据会显示在每个 chunk 上"
                description="如果 PDF 原文已经是乱码，后端会尽量拦截并提示走 OCR；已入库的历史乱码文档建议删除后重新上传。"
              />
              <List
                className="essay-chunk-list"
                dataSource={chunks}
                pagination={{ pageSize: 5 }}
                locale={{ emptyText: "暂无 chunk" }}
                renderItem={(item) => (
                  <List.Item className="essay-chunk-card">
                    <div className="essay-chunk-head">
                      <Space wrap>
                        <Tag>第 {item.page_no} 页</Tag>
                        <Tag>#{item.chunk_index}</Tag>
                        <TypeTag value={item.chunk_type} />
                        {item.confidence ? <Tag color="blue">置信度 {Math.round(item.confidence * 100)}%</Tag> : null}
                      </Space>
                      <span>{item.classification_note || "尚未分类"}</span>
                    </div>
                    <div className="essay-chunk-text">{item.content}</div>
                  </List.Item>
                )}
              />
            </Card>

            <Card className="glass-card" title="组装出的 essay_questions" bordered={false}>
              <Table
                rowKey="id"
                dataSource={questions}
                pagination={false}
                scroll={{ x: 720 }}
                columns={[
                  { title: "题目", dataIndex: "title" },
                  { title: "题型", dataIndex: "question_type", width: 130 },
                  { title: "满分", dataIndex: "max_score", width: 90 },
                  { title: "字数", dataIndex: "word_limit", width: 90 },
                ]}
              />
            </Card>

            <Card className="glass-card" title="提交答案批改" bordered={false}>
              <Form form={reviewForm} layout="vertical">
                <Row gutter={[12, 12]}>
                  <Col xs={24} md={12}>
                    <Form.Item name="question_id" label="选择题目" rules={[{ required: true, message: "请选择题目" }]}>
                      <Select
                        placeholder="从组装出的题目中选择"
                        options={questions.map((item) => ({ value: item.id, label: item.title }))}
                      />
                    </Form.Item>
                  </Col>
                  <Col xs={24} md={12}>
                    <Form.Item name="review_model_id" label="高质量批改模型" rules={[{ required: true, message: "请选择批改模型" }]}>
                      <Select placeholder="选择 use_essay / use_quality 模型" options={reviewModelOptions} />
                    </Form.Item>
                  </Col>
                </Row>
                <Form.Item name="user_answer" label="我的答案" rules={[{ required: true, message: "请输入答案" }]}>
                  <Input.TextArea rows={7} placeholder="输入作答内容。后续会连同关联材料、参考答案、评分规则一起传给高质量模型。" />
                </Form.Item>
                <Button type="primary" icon={<CheckCircleOutlined />} loading={loading} onClick={handleReview}>提交批改</Button>
              </Form>
              {reviewResult && (
                <div className="essay-review-result">
                  <Statistic title="骨架评分" value={reviewResult.score} suffix={`/ ${reviewResult.max_score}`} />
                  <p>{reviewResult.summary}</p>
                  <Space wrap>
                    {(reviewResult.suggestions || []).map((item) => <Tag color="blue" key={item}>{item}</Tag>)}
                  </Space>
                </div>
              )}
            </Card>
          </Space>
        </Col>
      </Row>
    </div>
  );
}

function modelOption(item) {
  return {
    value: item.id,
    label: `${item.alias || item.name}${item.provider ? ` · ${item.provider}` : ""}`,
  };
}

function stepFromState(document, chunks, questions) {
  if (questions.length) return 3;
  if (chunks.some((item) => item.chunk_type && item.chunk_type !== "unknown")) return 2;
  if (chunks.length || document?.status === "parsed") return 1;
  return 0;
}

function summarizeChunks(chunks) {
  const initial = {
    material: 0,
    question: 0,
    reference_answer: 0,
    scoring_rule: 0,
    explanation: 0,
    unknown: 0,
  };
  chunks.forEach((item) => {
    const type = item.chunk_type || "unknown";
    initial[type] = (initial[type] || 0) + 1;
  });
  return initial;
}

function TypeTag({ value }) {
  const meta = chunkTypeMeta[value] || chunkTypeMeta.unknown;
  return <Tag color={meta.color}>{meta.label}</Tag>;
}

function StatusTag({ status }) {
  const map = {
    uploaded: { label: "已上传", color: "default" },
    parsing: { label: "解析中", color: "processing" },
    parsed: { label: "已解析", color: "blue" },
    parse_failed: { label: "解析失败", color: "red" },
    classified: { label: "已分类", color: "purple" },
    assembled: { label: "已组装", color: "green" },
  };
  const meta = map[status] || { label: status || "未知", color: "default" };
  return <Tag color={meta.color}>{meta.label}</Tag>;
}

function OperationTag({ status }) {
  const map = {
    running: { label: "进行中", color: "processing" },
    success: { label: "成功", color: "green" },
    error: { label: "失败", color: "red" },
  };
  const meta = map[status] || map.running;
  return <Tag color={meta.color}>{meta.label}</Tag>;
}

function formatDuration(duration) {
  if (!duration) return "0ms";
  if (duration < 1000) return `${duration}ms`;
  return `${(duration / 1000).toFixed(2)}s`;
}

export default EssayReview;
