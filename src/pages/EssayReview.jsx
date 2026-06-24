import { useEffect, useMemo, useState, useCallback } from "react";
import { useLocation } from "react-router-dom";
import {
  Alert,
  Button,
  Col,
  Checkbox,
  Collapse,
  Divider,
  Drawer,
  Empty,
  Form,
  Input,
  InputNumber,
  List,
  Modal,
  Popconfirm,
  Progress,
  Row,
  Select,
  Space,
  Statistic,
  Steps,
  Tag,
  Typography,
  Upload,
  message,
} from "antd";
import {
  CheckCircleOutlined,
  CloudUploadOutlined,
  DeleteOutlined,
  EditOutlined,
  ExclamationCircleOutlined,
  FileSearchOutlined,
  HighlightOutlined,
  PlusOutlined,
  RobotOutlined,
  StarOutlined,
} from "@ant-design/icons";
import { AppCard, Page, PageHeader } from "../components/ui";
import {
  assembleEssayQuestions,
  createEssayQuestion,
  createEssayDocument,
  deleteEssayQuestion,
  deleteEssayDocument,
  debugEssayBoundary,
  getEssayDocuments,
  getEssayQuestions,
  getEssaySections,
  parseEssayDocument,
  replaceEssayQuestionRelations,
  reviewEssayQuestion,
  updateEssayQuestion,
  updateEssaySection,
} from "../api/essay";
import { getModels } from "../api/llm";
import { getPrompts } from "../api/prompts";

const { Text, Paragraph } = Typography;

const sectionTypeMeta = {
  material: { label: "材料", color: "blue" },
  question: { label: "题目", color: "green" },
  answer: { label: "答案/评分", color: "purple" },
  analysis: { label: "解析说明", color: "cyan" },
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

const questionTypeOptions = ["归纳概括题", "综合分析题", "提出对策题", "应用文写作题", "文章论述题", "公文写作题", "其他"]
  .map((value) => ({ value, label: value }));

const sectionTypeOptions = Object.entries(sectionTypeMeta).map(([value, meta]) => ({
  value,
  label: meta.label,
}));

function EssayReview() {
  const location = useLocation();
  const [uploadForm] = Form.useForm();
  const [reviewForm] = Form.useForm();
  const [questionForm] = Form.useForm();
  const [sectionForm] = Form.useForm();
  const [documents, setDocuments] = useState([]);
  const [selectedDocumentId, setSelectedDocumentId] = useState();
  const [sections, setSections] = useState([]);
  const [questions, setQuestions] = useState([]);
  const [models, setModels] = useState([]);
  const [prompts, setPrompts] = useState([]);
  const [activeStep, setActiveStep] = useState(0);
  const [loading, setLoading] = useState(false);
  const [reviewResult, setReviewResult] = useState(null);
  const [debugResult, setDebugResult] = useState(null);
  const [operations, setOperations] = useState([]);
  const [selectedQuestionId, setSelectedQuestionId] = useState(null);
  const [questionDrawerOpen, setQuestionDrawerOpen] = useState(false);
  const [editingQuestion, setEditingQuestion] = useState(null);
  const [sectionDrawerOpen, setSectionDrawerOpen] = useState(false);
  const [editingSection, setEditingSection] = useState(null);
  const [materialSelection, setMaterialSelection] = useState([]);
  const [answerSelection, setAnswerSelection] = useState([]);

  const selectedDocument = useMemo(
    () => documents.find((item) => item.id === selectedDocumentId),
    [documents, selectedDocumentId]
  );

  const boundaryModelOptions = useMemo(() => {
    const enabled = models.filter((item) => item.enabled);
    const fast = enabled.filter((item) => item.use_fast || item.speed_level === "高");
    return (fast.length ? fast : enabled).map(modelOption);
  }, [models]);

  const reviewModelOptions = useMemo(() => {
    const enabled = models.filter((item) => item.enabled);
    const essay = enabled.filter((item) => item.use_essay || item.use_quality);
    return (essay.length ? essay : enabled).map(modelOption);
  }, [models]);

  const promptOptions = useMemo(() => (
    prompts
      .filter((item) => item.enabled !== false)
      .map((item) => ({ value: item.id, label: `${item.name}${item.version ? ` · ${item.version}` : ""}` }))
  ), [prompts]);

  // 当选中题目时，从 sections 提取关联的材料和答案
  const selectedQuestion = useMemo(
    () => questions.find((q) => q.id === selectedQuestionId),
    [questions, selectedQuestionId]
  );

  const materialSections = useMemo(
    () => sections.filter((s) => s.section_type === "material"),
    [sections]
  );

  const answerSections = useMemo(
    () => sections.filter((s) => s.section_type === "answer" || s.section_type === "analysis"),
    [sections]
  );

  const relatedMaterials = useMemo(() => {
    if (!selectedQuestion) return [];
    const ids = selectedQuestion.material_section_ids || [];
    if (ids.length > 0) return materialSections.filter((s) => ids.includes(s.id));
    return materialSections;
  }, [materialSections, selectedQuestion]);

  const relatedAnswers = useMemo(() => {
    if (!selectedQuestion) return [];
    const ids = selectedQuestion.answer_section_ids || [];
    if (ids.length > 0) return answerSections.filter((s) => ids.includes(s.id));
    const qNo = selectedQuestion.question_no;
    // 优先查找 related_question_nos 精确匹配的 answer
    const matched = answerSections.filter((s) => {
      const relNos = (s.related_question_nos || "").split(",").map((n) => n.trim()).filter(Boolean);
      return relNos.includes(qNo);
    });
    // 如果精确匹配无结果，返回所有 answer
    return matched.length > 0 ? matched : answerSections;
  }, [answerSections, selectedQuestion]);

  const loadSelectedDocumentData = useCallback(async (documentId = selectedDocumentId) => {
    if (!documentId) return;
    const [nextSections, nextQuestions] = await Promise.all([
      getEssaySections(documentId).catch(() => []),
      getEssayQuestions(documentId).catch(() => []),
    ]);
    setSections(nextSections || []);
    setQuestions(nextQuestions || []);
    setActiveStep(stepFromState(selectedDocument, nextSections || [], nextQuestions || []));
    setReviewResult(null);
  }, [selectedDocumentId, selectedDocument]);

  const loadDocuments = useCallback(async () => {
    const items = await getEssayDocuments().catch(() => []);
    setDocuments(items || []);
    if (!selectedDocumentId && items?.[0]) {
      setSelectedDocumentId(items[0].id);
    }
  }, [selectedDocumentId]);

  useEffect(() => {
    queueMicrotask(() => {
      loadDocuments();
      getModels().then((items) => setModels(items || [])).catch(() => {});
      getPrompts().then((items) => setPrompts(items || [])).catch(() => {});
    });
  }, [loadDocuments]);

  useEffect(() => {
    const params = new URLSearchParams(location.search);
    if (params.get("from") !== "intake") return;
    const intakeText = sessionStorage.getItem("intake_text") || "";
    if (!intakeText.trim()) return;
    uploadForm.setFieldsValue({
      raw_text: intakeText,
      title: uploadForm.getFieldValue("title") || "录入器文本",
      document_role: uploadForm.getFieldValue("document_role") || "combined",
    });
    message.success("已填入录入器文本");
  }, [location.search, uploadForm]);

  useEffect(() => {
    if (!selectedDocumentId) return;
    loadSelectedDocumentData(selectedDocumentId).then(() => {
      setSelectedQuestionId(null);
    });
  }, [selectedDocumentId, loadSelectedDocumentData]);

  useEffect(() => {
    if (!selectedQuestion) {
      setMaterialSelection([]);
      setAnswerSelection([]);
      return;
    }
    const materialIds = selectedQuestion.material_section_ids?.length
      ? selectedQuestion.material_section_ids
      : relatedMaterials.map((item) => item.id);
    const answerIds = selectedQuestion.answer_section_ids?.length
      ? selectedQuestion.answer_section_ids
      : relatedAnswers.map((item) => item.id);
    setMaterialSelection(materialIds);
    setAnswerSelection(answerIds);
  }, [selectedQuestion, relatedMaterials, relatedAnswers]);

  const runOperation = useCallback(async (label, action) => {
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
      setOperations((items) => items.map((item) =>
        item.key === key ? { ...item, status: "success", duration, message: "完成" } : item
      ));
      return result;
    } catch (error) {
      const duration = Date.now() - start;
      setOperations((items) => items.map((item) =>
        item.key === key ? { ...item, status: "error", duration, message: error.message || "操作失败" } : item
      ));
      message.error(error.message || "操作失败");
      throw error;
    } finally {
      setLoading(false);
    }
  }, []);

  const handleCreateDocument = async () => {
    const values = await uploadForm.validateFields();
    const file = values.file?.[0]?.originFileObj;
    try {
      await runOperation("上传并解析", async () => {
        const result = await createEssayDocument({
          file,
          title: values.title,
          rawText: values.raw_text,
          documentRole: values.document_role,
          sourceGroup: values.source_group,
          boundaryModelId: values.boundary_model_id,
        });
        const document = result.document || result;
        const nextSections = result.sections || [];
        message.success(nextSections.length ? "文档已上传并解析完成" : "文档已上传");
        uploadForm.resetFields();
        await loadDocuments();
        setSelectedDocumentId(document.id);
        setSections(nextSections);
        // 自动获取已组装的 questions
        const nextQuestions = await getEssayQuestions(document.id).catch(() => []);
        setQuestions(nextQuestions || []);
        setActiveStep(stepFromState(document, nextSections, nextQuestions || []));
      });
    } catch {
      await loadDocuments();
    }
  };

  const handleParse = async () => {
    if (!selectedDocumentId) return;
    Modal.confirm({
      title: "确认重新解析？",
      content: "重新解析会用当前文档内容重建分段和题目，可能覆盖已经手动修正的题面、分段和关联。",
      okText: "继续解析",
      cancelText: "取消",
      onOk: runParse,
    });
  };

  const runParse = async () => {
    try {
      await runOperation("重新解析", async () => {
        const result = await parseEssayDocument(selectedDocumentId, {
          boundary_model_id: uploadForm.getFieldValue("boundary_model_id"),
        });
        setSections(result.sections || []);
        await loadDocuments();
        const nextQuestions = await getEssayQuestions(selectedDocumentId).catch(() => []);
        setQuestions(nextQuestions || []);
        setActiveStep(stepFromState(selectedDocument, result.sections || [], nextQuestions || []));
        message.success("解析完成");
      });
    } catch {
      await loadDocuments();
    }
  };

  const handleAssemble = async () => {
    if (!selectedDocumentId) return;
    try {
      await runOperation("组装题目", async () => {
        const result = await assembleEssayQuestions(selectedDocumentId);
        setQuestions(result || []);
        await loadDocuments();
        setActiveStep(2);
        message.success("题目组装完成");
      });
    } catch {
      await loadDocuments();
    }
  };

  const handleDebugBoundary = async () => {
    if (!selectedDocumentId) return;
    try {
      await runOperation("调试切分", async () => {
        const result = await debugEssayBoundary(selectedDocumentId, {
          raw_text: uploadForm.getFieldValue("raw_text"),
          boundary_model_id: uploadForm.getFieldValue("boundary_model_id"),
        });
        setDebugResult(result);
        if (result?.parse_error || result?.apply_error) {
          message.warning("调试完成，发现问题");
        } else {
          message.success("调试完成");
        }
      });
    } catch { /* handled */ }
  };

  const handleReview = async () => {
    const values = await reviewForm.validateFields();
    try {
      await runOperation("批改中", async () => {
        const result = await reviewEssayQuestion(values.question_id, {
          model_id: values.review_model_id,
          user_answer: values.user_answer,
        });
        setReviewResult(result);
        message.success("批改完成");
      });
    } catch { /* handled */ }
  };

  const handleDeleteDocument = async (id) => {
    try {
      await runOperation("删除文档", async () => {
        await deleteEssayDocument(id);
        await loadDocuments();
        if (selectedDocumentId === id) {
          setSelectedDocumentId(undefined);
          setSections([]);
          setQuestions([]);
          setActiveStep(0);
        }
        message.success("已删除");
      });
    } catch {
      await loadDocuments();
    }
  };

  const handleQuestionSelect = (questionId) => {
    setSelectedQuestionId(questionId);
    setReviewResult(null);
    reviewForm.setFieldsValue({ question_id: questionId });
  };

  const openQuestionDrawer = (question = null) => {
    setEditingQuestion(question);
    questionForm.setFieldsValue({
      document_id: question?.document_id || selectedDocumentId,
      question_no: question?.question_no || "",
      title: question?.title || "",
      question_type: question?.question_type || "",
      question_text: question?.question_text || "",
      max_score: question?.max_score || 100,
      word_limit: question?.word_limit || 500,
      scoring_rubric: question?.scoring_rubric || "",
      custom_prompt_id: question?.custom_prompt_id || undefined,
    });
    setQuestionDrawerOpen(true);
  };

  const handleSaveQuestion = async () => {
    const values = await questionForm.validateFields();
    await runOperation(editingQuestion ? "保存题目" : "新建题目", async () => {
      const payload = {
        ...values,
        document_id: Number(values.document_id || selectedDocumentId),
        max_score: Number(values.max_score || 0),
        word_limit: Number(values.word_limit || 0),
      };
      const saved = editingQuestion
        ? await updateEssayQuestion(editingQuestion.id, payload)
        : await createEssayQuestion(payload);
      setQuestionDrawerOpen(false);
      setEditingQuestion(null);
      await loadSelectedDocumentData(saved.document_id || selectedDocumentId);
      setSelectedQuestionId(saved.id);
      message.success("题目已保存");
    });
  };

  const handleDeleteQuestion = async (questionId) => {
    await runOperation("删除题目", async () => {
      await deleteEssayQuestion(questionId);
      if (selectedQuestionId === questionId) {
        setSelectedQuestionId(null);
      }
      await loadSelectedDocumentData();
      message.success("题目已删除");
    });
  };

  const openSectionDrawer = (section) => {
    setEditingSection(section);
    sectionForm.setFieldsValue({
      section_type: section.section_type,
      title: section.title,
      content: section.content,
      question_no: section.question_no,
      related_question_nos: section.related_question_nos,
    });
    setSectionDrawerOpen(true);
  };

  const handleSaveSection = async () => {
    const values = await sectionForm.validateFields();
    await runOperation("保存分段", async () => {
      await updateEssaySection(editingSection.id, values);
      setSectionDrawerOpen(false);
      setEditingSection(null);
      await loadSelectedDocumentData();
      message.success("分段已保存");
    });
  };

  const handleSaveRelations = async () => {
    if (!selectedQuestion) return;
    await runOperation("保存关联", async () => {
      const saved = await replaceEssayQuestionRelations(selectedQuestion.id, {
        material_ids: materialSelection,
        answer_ids: answerSelection,
      });
      await loadSelectedDocumentData(saved.document_id || selectedDocumentId);
      setSelectedQuestionId(selectedQuestion.id);
      message.success("关联已保存");
    });
  };

  const sectionStats = useMemo(() => summarizeSections(sections), [sections]);

  return (
    <Page className="essay-workbench">
      <PageHeader
        eyebrow="Essay Review"
        title="申论批改"
        description="上传 PDF，LLM 自动切分材料/题目/答案，选题作答，获得分维度专业批改。"
        icon={<HighlightOutlined />}
      />

      <Steps
        className="essay-steps"
        current={activeStep}
        items={[
          { title: "上传解析", icon: <CloudUploadOutlined /> },
          { title: "选题作答", icon: <FileSearchOutlined /> },
          { title: "批改结果", icon: <RobotOutlined /> },
        ]}
      />

      <Row gutter={[18, 18]} align="stretch">
        {/* ────── 左栏：上传 + 文档列表 ────── */}
        <Col xs={24} xl={7}>
          <Space direction="vertical" size="middle" className="essay-result-stack">
            <AppCard title="上传与解析" bordered={false}>
              <Form form={uploadForm} layout="vertical" className="essay-upload-form">
                <Form.Item name="title" label="文档标题">
                  <Input placeholder="例如 2025 省考申论 A 卷" />
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
                <Form.Item name="boundary_model_id" label="切分模型">
                  <Select
                    allowClear
                    placeholder="选择模型用于自动切分"
                    options={boundaryModelOptions}
                  />
                </Form.Item>
                <Collapse
                  size="small"
                  items={[{
                    key: "advanced",
                    label: "高级选项",
                    children: (
                      <>
                        <Form.Item name="source_group" label="卷套标识">
                          <Input placeholder="例如 2025-省考-A" />
                        </Form.Item>
                        <Form.Item name="raw_text" label="原始文本">
                          <Input.TextArea rows={4} placeholder="可选：粘贴 PDF 提取文本" />
                        </Form.Item>
                      </>
                    ),
                  }]}
                />
                <Space wrap style={{ marginTop: 12 }}>
                  <Button type="primary" loading={loading} onClick={handleCreateDocument}>上传并解析</Button>
                  <Button disabled={!selectedDocumentId} loading={loading} onClick={handleParse}>重新解析</Button>
                  <Button disabled={!selectedDocumentId} loading={loading} onClick={handleDebugBoundary} size="small">调试</Button>
                  {sections.length > 0 && !questions.length && (
                    <Button loading={loading} onClick={handleAssemble} size="small">手动组装</Button>
                  )}
                </Space>
              </Form>
            </AppCard>

            <AppCard title="文档列表" bordered={false} size="small">
              <List
                dataSource={documents}
                locale={{ emptyText: "暂无文档" }}
                renderItem={(item) => (
                  <List.Item
                    className={item.id === selectedDocumentId ? "essay-document active" : "essay-document"}
                    onClick={() => setSelectedDocumentId(item.id)}
                    actions={[
                      <Button
                        key="delete" type="text" danger size="small" icon={<DeleteOutlined />}
                        onClick={(e) => { e.stopPropagation(); handleDeleteDocument(item.id); }}
                      />,
                    ]}
                  >
                    <List.Item.Meta
                      title={<Space wrap size={4}><span>{item.title}</span><Tag>{documentRoleLabels[item.document_role] || "混合卷"}</Tag><StatusTag status={item.status} /></Space>}
                      description={`${item.chunk_count || 0} 段 · ${(item.note || "").slice(0, 40)}`}
                    />
                  </List.Item>
                )}
              />
            </AppCard>

            {operations.length > 0 && (
              <AppCard title="操作日志" bordered={false} size="small">
                <List
                  dataSource={operations}
                  renderItem={(item) => (
                    <List.Item className={`essay-operation ${item.status}`} style={{ padding: "6px 8px" }}>
                      <List.Item.Meta
                        title={<Space size={4}><OperationTag status={item.status} /><span>{item.label}</span></Space>}
                        description={`${formatDuration(item.duration)}${item.message ? ` · ${item.message}` : ""}`}
                      />
                    </List.Item>
                  )}
                />
              </AppCard>
            )}
          </Space>
        </Col>

        {/* ────── 中栏：题目列表 + 题目详情（材料/答案） ────── */}
        <Col xs={24} xl={9}>
          <Space direction="vertical" size="middle" className="essay-result-stack">
            {/* 统计概览 */}
            {sections.length > 0 && (
              <Row gutter={[12, 12]}>
                {Object.entries(sectionStats).filter(([, c]) => c > 0).map(([type, count]) => (
                  <Col xs={12} md={6} key={type}>
                    <AppCard className="essay-mini-stat" bordered={false}>
                      <Statistic title={sectionTypeMeta[type]?.label || type} value={count} />
                    </AppCard>
                  </Col>
                ))}
              </Row>
            )}

            {/* 题目列表 */}
            <AppCard
              title={`题目列表 (${questions.length})`}
              bordered={false}
              extra={
                <Button size="small" icon={<PlusOutlined />} disabled={!selectedDocumentId} onClick={() => openQuestionDrawer()}>
                  新增
                </Button>
              }
            >
              {questions.length === 0 ? (
                <Empty description="暂无题目，请先上传 PDF 并选择切分模型" />
              ) : (
                <List
                  dataSource={questions}
                  renderItem={(q) => (
                    <List.Item
                      className={q.id === selectedQuestionId ? "essay-document active" : "essay-document"}
                      onClick={() => handleQuestionSelect(q.id)}
                      style={{ cursor: "pointer" }}
                      actions={[
                        <Button
                          key="edit"
                          type="text"
                          size="small"
                          icon={<EditOutlined />}
                          onClick={(event) => {
                            event.stopPropagation();
                            openQuestionDrawer(q);
                          }}
                        />,
                        <Popconfirm
                          key="delete"
                          title="删除题目？"
                          description="会同时删除该题关联和历史批改记录。"
                          onConfirm={() => handleDeleteQuestion(q.id)}
                        >
                          <Button
                            type="text"
                            danger
                            size="small"
                            icon={<DeleteOutlined />}
                            onClick={(event) => event.stopPropagation()}
                          />
                        </Popconfirm>,
                      ]}
                    >
                      <List.Item.Meta
                        title={
                          <Space size={6} wrap>
                            <Tag color="green">第 {q.question_no || "?"} 题</Tag>
                            <span>{q.title}</span>
                          </Space>
                        }
                        description={
                          <Space size={8}>
                            <Tag>{q.question_type || "待确认"}</Tag>
                            <span>{q.max_score} 分</span>
                            <span>{q.word_limit} 字</span>
                          </Space>
                        }
                      />
                    </List.Item>
                  )}
                />
              )}
            </AppCard>

            {/* 选中题目的详情：题目原文 + 关联材料 + 参考答案 */}
            {selectedQuestion && (
              <AppCard
                title={`第 ${selectedQuestion.question_no || "?"} 题详情`}
                bordered={false}
                extra={<Button size="small" icon={<EditOutlined />} onClick={() => openQuestionDrawer(selectedQuestion)}>编辑题目</Button>}
              >
                <h4 className="essay-section-title">题目原文</h4>
                <Paragraph className="essay-chunk-text" style={{ maxHeight: 200 }}>
                  {selectedQuestion.question_text}
                </Paragraph>
                {selectedQuestion.scoring_rubric && (
                  <>
                    <h4 className="essay-section-title">评分细则</h4>
                    <Paragraph className="essay-chunk-text" style={{ maxHeight: 160 }}>
                      {selectedQuestion.scoring_rubric}
                    </Paragraph>
                  </>
                )}

                <Divider style={{ margin: "12px 0" }} />
                <div className="essay-relation-editor">
                  <div>
                    <h4 className="essay-section-title">关联材料</h4>
                    <Checkbox.Group
                      className="essay-relation-checks"
                      value={materialSelection}
                      onChange={(values) => setMaterialSelection(values)}
                      options={materialSections.map((item, index) => ({
                        value: item.id,
                        label: item.title || `材料 ${index + 1}`,
                      }))}
                    />
                  </div>
                  <div>
                    <h4 className="essay-section-title">关联答案</h4>
                    <Checkbox.Group
                      className="essay-relation-checks"
                      value={answerSelection}
                      onChange={(values) => setAnswerSelection(values)}
                      options={answerSections.map((item, index) => ({
                        value: item.id,
                        label: item.title || `答案 ${index + 1}`,
                      }))}
                    />
                  </div>
                  <Button size="small" type="primary" onClick={handleSaveRelations} loading={loading}>
                    保存关联
                  </Button>
                </div>

                {relatedMaterials.length > 0 && (
                  <>
                    <Divider style={{ margin: "12px 0" }} />
                    <Collapse
                      size="small"
                      items={relatedMaterials.map((m, i) => ({
                        key: m.id,
                        label: <Space><Tag color="blue">材料</Tag><span>{m.title || `材料 ${i + 1}`}</span></Space>,
                        children: <pre className="essay-chunk-text" style={{ maxHeight: 300 }}>{m.content}</pre>,
                      }))}
                    />
                  </>
                )}

                {relatedAnswers.length > 0 && (
                  <>
                    <Divider style={{ margin: "12px 0" }} />
                    <Collapse
                      size="small"
                      items={relatedAnswers.map((a, i) => ({
                        key: a.id,
                        label: <Space><Tag color="purple">参考答案</Tag><span>{a.title || `答案 ${i + 1}`}</span></Space>,
                        children: <pre className="essay-chunk-text" style={{ maxHeight: 300 }}>{a.content}</pre>,
                      }))}
                    />
                  </>
                )}
              </AppCard>
            )}

            {/* 结构分段（折叠查看） */}
            {sections.length > 0 && (
              <Collapse
                size="small"
                items={[{
                  key: "sections",
                  label: `查看全部结构分段 (${sections.length} 段)`,
                  children: (
                    <List
                      className="essay-chunk-list"
                      dataSource={sections}
                      pagination={{ pageSize: 5, size: "small" }}
                      renderItem={(item) => (
                        <List.Item className="essay-chunk-card">
                          <div className="essay-chunk-head">
                            <Space wrap size={4}>
                              <TypeTag value={item.section_type} />
                              <span>{item.title}</span>
                              {item.confidence ? <Tag color="blue">{Math.round(item.confidence * 100)}%</Tag> : null}
                            </Space>
                            <Button size="small" type="text" icon={<EditOutlined />} onClick={() => openSectionDrawer(item)}>
                              编辑分段
                            </Button>
                          </div>
                          <div className="essay-chunk-text" style={{ maxHeight: 120 }}>{item.content}</div>
                        </List.Item>
                      )}
                    />
                  ),
                }]}
              />
            )}
          </Space>
        </Col>

        {/* ────── 右栏：作答 + 批改结果 ────── */}
        <Col xs={24} xl={8}>
          <Space direction="vertical" size="middle" className="essay-result-stack">
            <AppCard title="作答与批改" bordered={false}>
              <Form form={reviewForm} layout="vertical">
                <Form.Item name="question_id" label="选择题目" rules={[{ required: true, message: "请选择题目" }]}>
                  <Select
                    placeholder="点击左侧题目或在此选择"
                    options={questions.map((item) => ({
                      value: item.id,
                      label: `第${item.question_no || "?"}题 · ${item.title}`,
                    }))}
                    onChange={(v) => setSelectedQuestionId(v)}
                  />
                </Form.Item>
                <Form.Item name="review_model_id" label="批改模型" rules={[{ required: true, message: "请选择模型" }]}>
                  <Select placeholder="选择批改模型" options={reviewModelOptions} />
                </Form.Item>
                <Form.Item name="user_answer" label="我的答案" rules={[{ required: true, message: "请输入答案" }]}>
                  <Input.TextArea
                    rows={8}
                    placeholder={selectedQuestion ? `请根据题目要求作答（${selectedQuestion.word_limit} 字以内）` : "请先选择题目"}
                  />
                </Form.Item>
                <Button type="primary" icon={<CheckCircleOutlined />} loading={loading} onClick={handleReview} block>
                  提交批改
                </Button>
              </Form>
            </AppCard>

            {/* ── 批改结果展示 ── */}
            {reviewResult && <ReviewResultCard result={reviewResult} />}
          </Space>
        </Col>
      </Row>

      {/* ── 调试面板（底部折叠） ── */}
      {debugResult && (
        <AppCard title="LLM 切分调试" bordered={false} style={{ marginTop: 18 }}>
          <Alert
            type={debugResult.parse_error || debugResult.apply_error ? "warning" : "success"}
            showIcon
            message={`${debugResult.block_count || 0} 行文本`}
            description={debugResult.parse_error || debugResult.apply_error || "LLM 返回的 JSON 已成功解析。"}
            style={{ marginBottom: 12 }}
          />
          <Collapse
            size="small"
            items={[
              { key: "prompt", label: "发送给模型的 Prompt", children: <Input.TextArea readOnly rows={12} value={debugResult.prompt || ""} /> },
              { key: "raw", label: "模型原始返回", children: <Input.TextArea readOnly rows={10} value={debugResult.raw_response || ""} /> },
              { key: "json", label: "提取的 JSON", children: <Input.TextArea readOnly rows={10} value={debugResult.extracted_json || JSON.stringify(debugResult.plan || {}, null, 2)} /> },
              { key: "sections", label: "切分结果", children: <Input.TextArea readOnly rows={10} value={JSON.stringify(debugResult.sections || [], null, 2)} /> },
            ]}
          />
        </AppCard>
      )}

      <Drawer
        title={editingQuestion ? "编辑题目" : "新增题目"}
        open={questionDrawerOpen}
        onClose={() => setQuestionDrawerOpen(false)}
        width={560}
        extra={<Button type="primary" loading={loading} onClick={handleSaveQuestion}>保存</Button>}
      >
        <Form form={questionForm} layout="vertical">
          <Form.Item name="document_id" hidden><Input /></Form.Item>
          <Row gutter={12}>
            <Col xs={24} sm={8}>
              <Form.Item name="question_no" label="题号"><Input placeholder="1" /></Form.Item>
            </Col>
            <Col xs={24} sm={16}>
              <Form.Item name="question_type" label="题型"><Select allowClear options={questionTypeOptions} /></Form.Item>
            </Col>
          </Row>
          <Form.Item name="title" label="标题" rules={[{ required: true, message: "请输入标题" }]}>
            <Input />
          </Form.Item>
          <Form.Item name="question_text" label="题面" rules={[{ required: true, message: "请输入题面" }]}>
            <Input.TextArea rows={8} />
          </Form.Item>
          <Row gutter={12}>
            <Col xs={24} sm={12}>
              <Form.Item name="max_score" label="满分"><InputNumber min={1} max={200} className="full-input" /></Form.Item>
            </Col>
            <Col xs={24} sm={12}>
              <Form.Item name="word_limit" label="字数"><InputNumber min={1} max={3000} className="full-input" /></Form.Item>
            </Col>
          </Row>
          <Form.Item name="custom_prompt_id" label="Prompt 模板">
            <Select allowClear options={promptOptions} placeholder="默认按题型模板" />
          </Form.Item>
          <Form.Item name="scoring_rubric" label="评分细则">
            <Input.TextArea rows={5} placeholder="可填写单题评分要点或扣分规则" />
          </Form.Item>
        </Form>
      </Drawer>

      <Drawer
        title="编辑分段"
        open={sectionDrawerOpen}
        onClose={() => setSectionDrawerOpen(false)}
        width={560}
        extra={<Button type="primary" loading={loading} onClick={handleSaveSection}>保存</Button>}
      >
        <Form form={sectionForm} layout="vertical">
          <Row gutter={12}>
            <Col xs={24} sm={12}>
              <Form.Item name="section_type" label="类型" rules={[{ required: true, message: "请选择类型" }]}>
                <Select options={sectionTypeOptions} />
              </Form.Item>
            </Col>
            <Col xs={24} sm={12}>
              <Form.Item name="question_no" label="题号"><Input placeholder="题目分段使用" /></Form.Item>
            </Col>
          </Row>
          <Form.Item name="title" label="标题"><Input /></Form.Item>
          <Form.Item name="related_question_nos" label="关联题号">
            <Input placeholder="例如 1,2。材料/答案分段可用" />
          </Form.Item>
          <Form.Item name="content" label="内容" rules={[{ required: true, message: "请输入内容" }]}>
            <Input.TextArea rows={12} />
          </Form.Item>
        </Form>
      </Drawer>
    </Page>
  );
}

/* ────── 批改结果卡片 ────── */
function ReviewResultCard({ result }) {
  const review = result.review || {};
  const resultData = parseResultJSON(review.result_json);
  const dimensions = resultData?.dimensions || [];
  const highlights = resultData?.highlights || [];
  const scoringDetail = resultData?.scoring_detail || "";
  const suggestions = result.suggestions || resultData?.suggestions || [];
  const summary = result.summary || resultData?.summary || "";
  const score = result.score ?? review.score ?? 0;
  const maxScore = result.max_score ?? review.max_score ?? 100;
  const pct = maxScore > 0 ? Math.round((score / maxScore) * 100) : 0;

  return (
    <AppCard title="批改结果" bordered={false}>
      {/* 总分 */}
      <div style={{ textAlign: "center", marginBottom: 16 }}>
        <Progress
          type="dashboard"
          percent={pct}
          format={() => <span style={{ fontSize: 22, fontWeight: 800 }}>{score}<span style={{ fontSize: 14, fontWeight: 400 }}>/{maxScore}</span></span>}
          strokeColor={pct >= 80 ? "#52c41a" : pct >= 60 ? "#1890ff" : "#ff4d4f"}
          size={120}
        />
      </div>

      {/* 总评 */}
      {summary && (
        <Paragraph className="essay-result-summary">{summary}</Paragraph>
      )}

      {/* 分维度评分 */}
      {dimensions.length > 0 && (
        <>
          <Divider style={{ margin: "12px 0" }}>分维度评分</Divider>
          <div className="essay-dimension-list">
            {dimensions.map((dim, i) => (
              <div key={i} className="essay-dimension">
                <div className="essay-dimension-head">
                  <Text strong>{dim.name}</Text>
                  <Tag color={dim.score >= dim.max_score * 0.8 ? "green" : dim.score >= dim.max_score * 0.6 ? "blue" : "orange"}>
                    {dim.score} / {dim.max_score}
                  </Tag>
                </div>
                {dim.comment && <p>{dim.comment}</p>}
              </div>
            ))}
          </div>
        </>
      )}

      {/* 亮点与问题 */}
      {highlights.length > 0 && (
        <>
          <Divider style={{ margin: "12px 0" }}>亮点与问题</Divider>
          <List
            dataSource={highlights}
            renderItem={(h, i) => (
              <List.Item key={i} style={{ padding: "8px 0", alignItems: "flex-start" }}>
                <List.Item.Meta
                  avatar={h.type === "good"
                    ? <StarOutlined style={{ color: "#52c41a", fontSize: 16 }} />
                    : <ExclamationCircleOutlined style={{ color: "#ff4d4f", fontSize: 16 }} />
                  }
                  title={
                    <Text type={h.type === "good" ? "success" : "danger"} style={{ fontSize: 13 }}>
                      {h.type === "good" ? "亮点" : "问题"}
                    </Text>
                  }
                  description={
                    <>
                      {h.text && <Paragraph style={{ margin: 0, fontStyle: "italic", color: "#555" }}>&ldquo;{h.text}&rdquo;</Paragraph>}
                      {h.comment && <Paragraph style={{ margin: "4px 0 0", color: "#666" }}>{h.comment}</Paragraph>}
                    </>
                  }
                />
              </List.Item>
            )}
          />
        </>
      )}

      {/* 逐要点评分说明 */}
      {scoringDetail && (
        <>
          <Divider style={{ margin: "12px 0" }}>评分详情</Divider>
          <Paragraph className="essay-chunk-text" style={{ maxHeight: 200 }}>{scoringDetail}</Paragraph>
        </>
      )}

      {/* 改进建议 */}
      {suggestions.length > 0 && (
        <>
          <Divider style={{ margin: "12px 0" }}>改进建议</Divider>
          <List
            dataSource={suggestions}
            renderItem={(s, i) => (
              <List.Item key={i} style={{ padding: "4px 0" }}>
                <Space>
                  <HighlightOutlined style={{ color: "var(--primary-blue)" }} />
                  <span>{s}</span>
                </Space>
              </List.Item>
            )}
          />
        </>
      )}
    </AppCard>
  );
}

/* ────── 辅助函数和小组件 ────── */

function parseResultJSON(jsonStr) {
  if (!jsonStr) return null;
  try { return JSON.parse(jsonStr); } catch { return null; }
}

function modelOption(item) {
  return {
    value: item.id,
    label: `${item.alias || item.name}${item.provider ? ` · ${item.provider}` : ""}`,
  };
}

function stepFromState(document, sections, questions) {
  if (questions.length) return 1;
  if (sections.length || document?.status === "parsed") return 1;
  return 0;
}

function summarizeSections(sections) {
  const counts = { material: 0, question: 0, answer: 0, analysis: 0, unknown: 0 };
  sections.forEach((item) => {
    const type = item.section_type || "unknown";
    counts[type] = (counts[type] || 0) + 1;
  });
  return counts;
}

function TypeTag({ value }) {
  const meta = sectionTypeMeta[value] || sectionTypeMeta.unknown;
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
