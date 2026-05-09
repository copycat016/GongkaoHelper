import { useEffect, useState } from "react";
import { Button, Card, Col, Form, Input, Row, Select, Space, Tag, Upload, message } from "antd";
import { CloudUploadOutlined, RobotOutlined, SaveOutlined, ScanOutlined } from "@ant-design/icons";
import PageHeader from "../components/PageHeader";
import { correctOcrText, getOcrMonthUsage, getOcrScenes, runOcr } from "../api/ocr";

function OCRQuestion() {
  const [ocrText, setOcrText] = useState("");
  const [fixedText, setFixedText] = useState("");
  const [scene, setScene] = useState("printed");
  const [scenes, setScenes] = useState([]);
  const [usage, setUsage] = useState();
  const [fileList, setFileList] = useState([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    getOcrScenes().then((items) => setScenes(items || [])).catch(() => {});
    getOcrMonthUsage().then(setUsage).catch(() => {});
  }, []);

  const handleOcr = async () => {
    const file = fileList[0]?.originFileObj;
    if (!file) {
      message.warning("请先上传图片");
      return;
    }
    setLoading(true);
    try {
      const result = await runOcr({ file, scene });
      setOcrText(result.text || "");
      if (result.from_cache) {
        message.success("已从缓存读取 OCR 结果");
      } else {
        message.success("OCR 识别完成");
      }
      getOcrMonthUsage().then(setUsage).catch(() => {});
    } finally {
      setLoading(false);
    }
  };

  const handleCorrect = async () => {
    const result = await correctOcrText(ocrText);
    setFixedText(result.text);
  };

  return (
    <div className="page-grid">
      <PageHeader eyebrow="OCR" title="OCR 识题" desc="上传图片、识别原文、AI 修正后保存为结构化题目。" />
      <Row gutter={[16, 16]}>
        <Col xs={24} lg={10}>
          <Card className="glass-card" title="图片上传" bordered={false}>
            <Space direction="vertical" size="middle" className="ocr-upload-stack">
              <Select
                value={scene}
                onChange={setScene}
                options={(scenes.length ? scenes : fallbackScenes).map((item) => ({
                  value: item.key,
                  label: item.label,
                  disabled: item.key === "essay_handwriting",
                }))}
                className="full-input"
              />
              {usage && (
                <Tag color="blue">
                  本系统本月 OCR：{usage.used} / {usage.local_limit || "不限"}
                </Tag>
              )}
            </Space>
            <Upload.Dragger
              beforeUpload={() => false}
              maxCount={1}
              fileList={fileList}
              onChange={({ fileList: nextFileList }) => setFileList(nextFileList)}
            >
              <p className="ant-upload-drag-icon"><CloudUploadOutlined /></p>
              <p>拖拽或点击上传题目图片</p>
            </Upload.Dragger>
            <Space wrap style={{ marginTop: 16 }}>
              <Button type="primary" icon={<ScanOutlined />} onClick={handleOcr} loading={loading}>开始 OCR</Button>
              <Button icon={<RobotOutlined />} onClick={handleCorrect} disabled={!ocrText}>AI 修正</Button>
            </Space>
          </Card>
          <Card className="glass-card" title="OCR 原文" bordered={false}>
            <Input.TextArea rows={8} value={ocrText} onChange={(event) => setOcrText(event.target.value)} />
          </Card>
          <Card className="glass-card" title="AI 修正文稿" bordered={false}>
            <Input.TextArea rows={8} value={fixedText} onChange={(event) => setFixedText(event.target.value)} />
          </Card>
        </Col>
        <Col xs={24} lg={14}>
          <Card className="glass-card" title="题目结构化" bordered={false}>
            <Form layout="vertical">
              <Row gutter={12}>
                <Col xs={24} sm={8}><Form.Item label="科目"><Select options={["行测", "申论"].map((value) => ({ value, label: value }))} /></Form.Item></Col>
                <Col xs={24} sm={8}><Form.Item label="一级题型"><Input /></Form.Item></Col>
                <Col xs={24} sm={8}><Form.Item label="二级题型"><Input /></Form.Item></Col>
              </Row>
              <Form.Item label="题干"><Input.TextArea rows={5} value={fixedText} /></Form.Item>
              <Row gutter={12}>
                {["A", "B", "C", "D"].map((key) => <Col xs={24} sm={12} key={key}><Form.Item label={`选项 ${key}`}><Input /></Form.Item></Col>)}
              </Row>
              <Row gutter={12}>
                <Col xs={24} sm={8}><Form.Item label="正确答案"><Input /></Form.Item></Col>
                <Col xs={24} sm={8}><Form.Item label="难度"><Select options={["简单", "中等", "偏难"].map((value) => ({ value, label: value }))} /></Form.Item></Col>
                <Col xs={24} sm={8}><Form.Item label="标签"><Select mode="tags" /></Form.Item></Col>
              </Row>
              <Form.Item label="解析"><Input.TextArea rows={4} /></Form.Item>
              <Button type="primary" icon={<SaveOutlined />} onClick={() => message.success("已保存为 mock 题目")}>保存为题目</Button>
            </Form>
          </Card>
        </Col>
      </Row>
    </div>
  );
}

const fallbackScenes = [
  { key: "printed", label: "印刷体题目" },
  { key: "printed_accurate", label: "高精度印刷体" },
  { key: "handwriting", label: "手写内容" },
];

export default OCRQuestion;
