import { Button, Card, Col, List, Row, Steps, Upload } from "antd";
import { FilePdfOutlined, RobotOutlined } from "@ant-design/icons";
import PageHeader from "../components/PageHeader";

function PDFParser() {
  return (
    <div className="page-grid">
      <PageHeader eyebrow="PDF" title="PDF 解析" desc="上传、解析、分页文本、AI 提取题目和确认入库的流程占位。" />
      <Card className="glass-card" bordered={false}>
        <Steps current={1} items={[{ title: "上传 PDF" }, { title: "解析文本" }, { title: "AI 提题" }, { title: "确认入库" }]} />
      </Card>
      <Row gutter={[16, 16]}>
        <Col xs={24} lg={10}>
          <Card className="glass-card" title="上传 PDF" bordered={false}>
            <Upload.Dragger beforeUpload={() => false} maxCount={1}>
              <p className="ant-upload-drag-icon"><FilePdfOutlined /></p>
              <p>拖拽或点击上传 PDF 文件</p>
            </Upload.Dragger>
            <Button type="primary" style={{ marginTop: 16 }}>解析 PDF</Button>
          </Card>
        </Col>
        <Col xs={24} lg={14}>
          <Card className="glass-card" title="页码与文本" extra={<Button icon={<RobotOutlined />}>AI 提取题目</Button>} bordered={false}>
            <List dataSource={["第 1 页：资料分析材料 mock 文本", "第 2 页：题目与选项 mock 文本"]} renderItem={(item) => <List.Item>{item}</List.Item>} />
          </Card>
        </Col>
      </Row>
    </div>
  );
}

export default PDFParser;
