import { useEffect, useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import {
  Button,
  Card,
  Col,
  Input,
  Row,
  Select,
  Space,
  Statistic,
  Tabs,
  Tag,
  Upload,
  message,
} from "antd";
import {
  ClearOutlined,
  CloudUploadOutlined,
  CopyOutlined,
  EditOutlined,
  FilePdfOutlined,
  ScanOutlined,
  SendOutlined,
} from "@ant-design/icons";
import PageHeader from "../components/PageHeader";
import { getOcrMonthUsage, getOcrScenes } from "../api/ocr";
import { intakeFromImage, intakeFromPdf, intakeFromText } from "../api/intake";

function IntakePage() {
  const navigate = useNavigate();
  const [activeTab, setActiveTab] = useState("image");
  const [imageFiles, setImageFiles] = useState([]);
  const [pdfFiles, setPdfFiles] = useState([]);
  const [scene, setScene] = useState("printed");
  const [scenes, setScenes] = useState([]);
  const [usage, setUsage] = useState();
  const [pastedText, setPastedText] = useState("");
  const [intakeResult, setIntakeResult] = useState(null);
  const [editableText, setEditableText] = useState("");
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    getOcrScenes().then((items) => setScenes(items || [])).catch(() => {});
    getOcrMonthUsage().then(setUsage).catch(() => {});
  }, []);

  const quality = intakeResult?.quality || {};
  const currentText = editableText || "";
  const stats = useMemo(() => ({
    chars: Array.from(currentText).length,
    lines: currentText.split("\n").filter((line) => line.trim()).length,
  }), [currentText]);

  const receiveResult = (result) => {
    setIntakeResult(result);
    setEditableText(result.editable_text || result.text || "");
  };

  const handleImageOcr = async () => {
    const file = imageFiles[0]?.originFileObj;
    if (!file) {
      message.warning("请先上传图片");
      return;
    }
    setLoading(true);
    try {
      receiveResult(await intakeFromImage({ file, scene }));
      getOcrMonthUsage().then(setUsage).catch(() => {});
      message.success("OCR 识别完成");
    } finally {
      setLoading(false);
    }
  };

  const handlePdfParse = async () => {
    const file = pdfFiles[0]?.originFileObj;
    if (!file) {
      message.warning("请先选择 PDF");
      return;
    }
    setLoading(true);
    try {
      receiveResult(await intakeFromPdf(file));
      message.success("PDF 文本解析完成");
    } finally {
      setLoading(false);
    }
  };

  const handlePasteText = () => {
    if (!pastedText.trim()) {
      message.warning("请先粘贴文本");
      return;
    }
    receiveResult(intakeFromText(pastedText));
    message.success("文本已进入预览区");
  };

  const handleCopy = async () => {
    await navigator.clipboard.writeText(currentText);
    message.success("已复制文本");
  };

  const handleSendToEssay = () => {
    sessionStorage.setItem("intake_text", currentText);
    navigate("/essay?from=intake");
  };

  const handleClear = () => {
    setIntakeResult(null);
    setEditableText("");
    setPastedText("");
    setImageFiles([]);
    setPdfFiles([]);
  };

  return (
    <div className="page-grid intake-page">
      <PageHeader
        eyebrow="Intake"
        title="录入器"
        desc="把图片、PDF、粘贴文本统一转成可检查、可修正、可流转的纯文本。"
      />

      <Row gutter={[18, 18]} align="stretch">
        <Col xs={24} xl={8}>
          <Card className="glass-card intake-source-card" title="录入来源" bordered={false}>
            <Tabs
              activeKey={activeTab}
              onChange={setActiveTab}
              items={[
                {
                  key: "image",
                  label: "图片 OCR",
                  children: (
                    <Space direction="vertical" size="middle" className="intake-stack">
                      <Select
                        value={scene}
                        onChange={setScene}
                        options={(scenes.length ? scenes : fallbackScenes).map((item) => ({
                          value: item.key,
                          label: item.label,
                          disabled: item.key === "essay_handwriting",
                        }))}
                      />
                      {usage && (
                        <Tag color="blue">本月 OCR：{usage.used} / {usage.local_limit || "不限"}</Tag>
                      )}
                      <Upload.Dragger
                        beforeUpload={() => false}
                        maxCount={1}
                        accept="image/*"
                        fileList={imageFiles}
                        onChange={({ fileList }) => setImageFiles(fileList)}
                      >
                        <p className="ant-upload-drag-icon"><CloudUploadOutlined /></p>
                        <p>拖拽或点击上传图片</p>
                      </Upload.Dragger>
                      <Button type="primary" icon={<ScanOutlined />} onClick={handleImageOcr} loading={loading}>
                        开始 OCR
                      </Button>
                    </Space>
                  ),
                },
                {
                  key: "pdf",
                  label: "PDF 文件",
                  children: (
                    <Space direction="vertical" size="middle" className="intake-stack">
                      <Upload.Dragger
                        beforeUpload={() => false}
                        maxCount={1}
                        accept="application/pdf"
                        fileList={pdfFiles}
                        onChange={({ fileList }) => setPdfFiles(fileList)}
                      >
                        <p className="ant-upload-drag-icon"><FilePdfOutlined /></p>
                        <p>选择带文本层的 PDF</p>
                      </Upload.Dragger>
                      <Button type="primary" icon={<FilePdfOutlined />} onClick={handlePdfParse} loading={loading}>
                        解析 PDF
                      </Button>
                    </Space>
                  ),
                },
                {
                  key: "paste",
                  label: "粘贴文本",
                  children: (
                    <Space direction="vertical" size="middle" className="intake-stack">
                      <Input.TextArea
                        rows={12}
                        value={pastedText}
                        onChange={(event) => setPastedText(event.target.value)}
                        placeholder="粘贴题面、材料、答案或自己的作答文本"
                      />
                      <Button type="primary" icon={<EditOutlined />} onClick={handlePasteText}>
                        生成预览
                      </Button>
                    </Space>
                  ),
                },
              ]}
            />
          </Card>
        </Col>

        <Col xs={24} xl={16}>
          <Card
            className="glass-card intake-preview-card"
            title="统一预览"
            bordered={false}
            extra={
              <Space wrap>
                <Button icon={<CopyOutlined />} disabled={!currentText} onClick={handleCopy}>复制文本</Button>
                <Button type="primary" icon={<SendOutlined />} disabled={!currentText} onClick={handleSendToEssay}>发送到申论</Button>
                <Button disabled>写入思源</Button>
                <Button icon={<ClearOutlined />} disabled={!intakeResult && !pastedText} onClick={handleClear}>清空</Button>
              </Space>
            }
          >
            {intakeResult ? (
              <Space direction="vertical" size="middle" className="intake-stack">
                <Row gutter={[12, 12]}>
                  <Col xs={12} md={6}><Statistic title="来源" value={sourceLabel(intakeResult.source)} /></Col>
                  <Col xs={12} md={6}><Statistic title="字符数" value={stats.chars} /></Col>
                  <Col xs={12} md={6}><Statistic title="行数" value={stats.lines} /></Col>
                  <Col xs={12} md={6}>
                    <div className="intake-quality-box">
                      <span>质量状态</span>
                      <Tag color={quality.ok ? "green" : "orange"}>{quality.ok ? "可用" : "需检查"}</Tag>
                    </div>
                  </Col>
                </Row>
                <Space wrap>
                  {intakeResult.file_name && <Tag>{intakeResult.file_name}</Tag>}
                  {intakeResult.source_engine && <Tag color="geekblue">{intakeResult.source_engine}</Tag>}
                  <Tag color={quality.ok ? "green" : "orange"}>{quality.reason || "待检查"}</Tag>
                </Space>
                <div className="intake-text-grid">
                  <div>
                    <div className="intake-pane-title">原始文本</div>
                    <Input.TextArea rows={16} readOnly value={intakeResult.text || ""} />
                  </div>
                  <div>
                    <div className="intake-pane-title">可编辑修正文稿</div>
                    <Input.TextArea rows={16} value={editableText} onChange={(event) => setEditableText(event.target.value)} />
                  </div>
                </div>
              </Space>
            ) : (
              <div className="intake-empty">
                <CloudUploadOutlined />
                <span>选择一种来源开始录入，结果会统一出现在这里。</span>
              </div>
            )}
          </Card>
        </Col>
      </Row>
    </div>
  );
}

function sourceLabel(source) {
  const map = {
    image_ocr: "图片 OCR",
    pdf_file: "PDF",
    pasted_text: "粘贴文本",
    ocr: "OCR",
    raw_text: "文本",
  };
  return map[source] || source || "未知";
}

const fallbackScenes = [
  { key: "printed", label: "印刷体题目" },
  { key: "printed_accurate", label: "高精度印刷体" },
  { key: "handwriting", label: "手写内容" },
];

export default IntakePage;
