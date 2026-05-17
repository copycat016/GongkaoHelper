import { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import {
  Button,
  Card,
  Col,
  Descriptions,
  Empty,
  Popconfirm,
  Row,
  Space,
  Spin,
  Tag,
  Typography,
  message,
} from "antd";
import {
  ArrowLeftOutlined,
  DeleteOutlined,
  HistoryOutlined,
  ReadOutlined,
} from "@ant-design/icons";
import PageHeader from "../components/PageHeader";
import { createMistake } from "../api/mistakes";
import { questionsMock } from "../api/mockData";
import { deleteQuestion, getQuestion } from "../api/questions";

function QuestionDetail() {
  const { id } = useParams();
  const navigate = useNavigate();
  const [question, setQuestion] = useState(null);
  const [loading, setLoading] = useState(false);

  const loadQuestion = async () => {
    setLoading(true);
    try {
      const item = await getQuestion(id);
      setQuestion(item);
    } catch {
      setQuestion(questionsMock.find((item) => String(item.id) === String(id)) || null);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadQuestion();
  }, [id]);

  const handleAddMistake = async () => {
    if (!question) return;
    await createMistake({
      subject: question.subject,
      question_type: question.level1,
      sub_type: question.level2,
      title: question.title,
      stem: question.stem || question.title,
      correct_answer: question.answer || "",
      error_reason: "待复盘",
      mastery: "未掌握",
      source: question.document || question.source || "题库",
      tags: JSON.stringify(question.tags || []),
      note: question.material_count ? `已关联 ${question.material_count} 段材料` : "",
    });
    message.success("已加入错题库");
  };

  const handleDelete = async () => {
    await deleteQuestion(id);
    message.success("题目已删除");
    navigate("/questions");
  };

  if (loading) {
    return <Spin />;
  }

  if (!question) {
    return (
      <div className="page-grid">
        <PageHeader eyebrow="Question Detail" title="题目详情" desc="未找到题目" />
        <Card className="glass-card" bordered={false}>
          <Empty description="题目不存在或已被删除" />
          <Button icon={<ArrowLeftOutlined />} onClick={() => navigate("/questions")}>返回题库</Button>
        </Card>
      </div>
    );
  }

  return (
    <div className="page-grid question-detail-page">
      <PageHeader eyebrow="Question Detail" title={question.title} desc={question.document || question.source || "题库题目"} />

      <Card className="glass-card question-detail-toolbar" bordered={false}>
        <Space wrap>
          <Button icon={<ArrowLeftOutlined />} onClick={() => navigate("/questions")}>返回题库</Button>
          <Button icon={<HistoryOutlined />} onClick={handleAddMistake}>加入错题库</Button>
          <Popconfirm title="删除这道题？" okText="删除" cancelText="取消" onConfirm={handleDelete}>
            <Button danger icon={<DeleteOutlined />}>删除</Button>
          </Popconfirm>
        </Space>
      </Card>

      <Row gutter={[16, 16]}>
        <Col xs={24} lg={15}>
          <Card className="glass-card question-reading-card" bordered={false} title="题目原文">
            <Space wrap className="question-detail-tags">
              <Tag color={question.subject === "申论" ? "green" : "blue"}>{question.subject}</Tag>
              <Tag>{question.level1 || "未分类"}</Tag>
              {question.question_no ? <Tag color="gold">第 {question.question_no} 题</Tag> : null}
              <Tag color={question.material_count ? "blue" : "default"}>{question.material_count || 0} 段材料</Tag>
            </Space>
            <Typography.Paragraph className="question-main-text">
              {question.stem || question.title}
            </Typography.Paragraph>
          </Card>
        </Col>

        <Col xs={24} lg={9}>
          <Card className="glass-card" bordered={false} title="题目信息">
            <Descriptions column={1} size="small">
              <Descriptions.Item label="来源文档">{question.document || "-"}</Descriptions.Item>
              <Descriptions.Item label="题型">{question.level1 || "-"}</Descriptions.Item>
              <Descriptions.Item label="二级分类">{question.level2 || "-"}</Descriptions.Item>
              <Descriptions.Item label="难度">{question.difficulty || "-"}</Descriptions.Item>
              <Descriptions.Item label="标签">
                {(question.tags || []).filter(Boolean).map((tag) => <Tag key={tag}>{tag}</Tag>)}
              </Descriptions.Item>
            </Descriptions>
          </Card>
        </Col>
      </Row>

      <Card className="glass-card" bordered={false} title={<Space><ReadOutlined />关联材料</Space>}>
        {(question.materials || []).length ? (
          <Space direction="vertical" size={14} className="question-material-list">
            {(question.materials || []).map((material, index) => (
              <section className="question-material-block" key={material.id || `${material.title}-${index}`}>
                <Typography.Title level={5}>{material.title || `材料 ${index + 1}`}</Typography.Title>
                <Typography.Paragraph>{material.content}</Typography.Paragraph>
              </section>
            ))}
          </Space>
        ) : (
          <Empty description="暂无关联材料" />
        )}
      </Card>

      {question.answer ? (
        <Card className="glass-card" bordered={false} title="参考答案">
          <Typography.Paragraph className="question-main-text">{question.answer}</Typography.Paragraph>
        </Card>
      ) : null}
    </div>
  );
}

export default QuestionDetail;
