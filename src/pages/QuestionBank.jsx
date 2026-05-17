import { useEffect, useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import {
  Button,
  Card,
  Form,
  Input,
  Modal,
  Popconfirm,
  Select,
  Space,
  Table,
  Tag,
  Typography,
  message,
} from "antd";
import {
  DeleteOutlined,
  EditOutlined,
  EyeOutlined,
  HistoryOutlined,
  ReloadOutlined,
} from "@ant-design/icons";
import PageHeader from "../components/PageHeader";
import { questionsMock } from "../api/mockData";
import { deleteQuestion, getQuestions, updateQuestion } from "../api/questions";
import { createMistake } from "../api/mistakes";

const questionTypes = [
  "常识判断",
  "言语理解与表达",
  "数量关系",
  "判断推理",
  "资料分析",
  "归纳概括题",
  "综合分析题",
  "提出对策题",
  "应用文写作题",
  "文章论述题",
  "公文写作题",
  "待确认",
];

function QuestionBank() {
  const navigate = useNavigate();
  const [editForm] = Form.useForm();
  const [questions, setQuestions] = useState([]);
  const [loading, setLoading] = useState(false);
  const [filters, setFilters] = useState({});
  const [editing, setEditing] = useState(null);

  const loadQuestions = async () => {
    setLoading(true);
    try {
      const items = await getQuestions();
      setQuestions(items?.length ? items : questionsMock);
    } catch {
      setQuestions(questionsMock);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadQuestions();
  }, []);

  const filteredQuestions = useMemo(() => {
    const keyword = (filters.keyword || "").trim().toLowerCase();
    return questions.filter((item) => {
      if (filters.subject && item.subject !== filters.subject) return false;
      if (filters.level1 && item.level1 !== filters.level1) return false;
      if (filters.materialState === "linked" && !item.material_count) return false;
      if (filters.materialState === "unlinked" && item.material_count) return false;
      if (!keyword) return true;
      return [item.title, item.stem, item.document, item.level1, item.level2]
        .filter(Boolean)
        .some((value) => String(value).toLowerCase().includes(keyword));
    });
  }, [filters, questions]);

  const openEdit = (record) => {
    setEditing(record);
    editForm.setFieldsValue({
      title: record.title,
      level1: record.level1,
      question_no: record.question_no,
      stem: record.stem || record.title,
    });
  };

  const handleEdit = async () => {
    const values = await editForm.validateFields();
    if (!editing?.id) return;
    await updateQuestion(editing.id, values);
    message.success("题目已更新");
    setEditing(null);
    await loadQuestions();
  };

  const handleAddMistake = async (record) => {
    await createMistake({
      subject: record.subject,
      question_type: record.level1,
      sub_type: record.level2,
      title: record.title,
      stem: record.stem || record.title,
      correct_answer: record.answer || "",
      error_reason: "待复盘",
      mastery: "未掌握",
      source: record.document || record.source || "题库",
      tags: JSON.stringify(record.tags || []),
      note: record.material_count ? `已关联 ${record.material_count} 段材料` : "",
    });
    message.success("已加入错题库");
  };

  const handleDelete = async (record) => {
    if (record.source !== "申论解析") {
      message.info("非申论解析题暂不支持在题库页删除");
      return;
    }
    await deleteQuestion(record.id);
    message.success("题目已删除");
    await loadQuestions();
  };

  const columns = [
    {
      title: "题目",
      dataIndex: "title",
      width: 360,
      render: (_, record) => (
        <Space direction="vertical" size={4} className="question-title-cell">
          <Button type="link" className="question-title-link" onClick={() => navigate(`/questions/${record.id}`)}>
            {record.title}
          </Button>
          <Typography.Text type="secondary" ellipsis>
            {record.stem || "暂无题干"}
          </Typography.Text>
        </Space>
      ),
    },
    { title: "科目", dataIndex: "subject", width: 96, render: (value) => <Tag color={value === "申论" ? "green" : "blue"}>{value || "-"}</Tag> },
    { title: "题型", dataIndex: "level1", width: 140 },
    { title: "来源文档", dataIndex: "document", width: 180, render: (value) => value || "-" },
    { title: "材料", dataIndex: "material_count", width: 96, render: (value) => <Tag color={value ? "blue" : "default"}>{value || 0} 段</Tag> },
    { title: "难度", dataIndex: "difficulty", width: 100 },
    {
      title: "标签",
      dataIndex: "tags",
      width: 180,
      render: (tags = []) => tags.filter(Boolean).slice(0, 3).map((tag) => <Tag key={tag}>{tag}</Tag>),
    },
    {
      title: "操作",
      fixed: "right",
      width: 260,
      render: (_, record) => (
        <Space size={6}>
          <Button size="small" icon={<EyeOutlined />} onClick={() => navigate(`/questions/${record.id}`)}>详情</Button>
          <Button size="small" icon={<EditOutlined />} onClick={() => openEdit(record)}>编辑</Button>
          <Button size="small" icon={<HistoryOutlined />} onClick={() => handleAddMistake(record)}>错题</Button>
          <Popconfirm title="删除这道题？" okText="删除" cancelText="取消" onConfirm={() => handleDelete(record)}>
            <Button size="small" danger icon={<DeleteOutlined />} />
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div className="page-grid question-bank-page">
      <PageHeader eyebrow="Question Bank" title="题库管理" desc="按题目、材料和来源组织申论与行测题目。" />

      <Card className="glass-card question-filter-card" bordered={false}>
        <Space wrap>
          <Input.Search
            allowClear
            placeholder="搜索题干、来源文档、题型"
            style={{ width: 280 }}
            onSearch={(value) => setFilters((prev) => ({ ...prev, keyword: value }))}
            onChange={(event) => setFilters((prev) => ({ ...prev, keyword: event.target.value }))}
          />
          <Select allowClear placeholder="科目" style={{ width: 140 }} options={["行测", "申论"].map((value) => ({ value, label: value }))} onChange={(value) => setFilters((prev) => ({ ...prev, subject: value }))} />
          <Select allowClear showSearch placeholder="题型" style={{ width: 180 }} options={questionTypes.map((value) => ({ value, label: value }))} onChange={(value) => setFilters((prev) => ({ ...prev, level1: value }))} />
          <Select
            allowClear
            placeholder="材料关联"
            style={{ width: 150 }}
            options={[
              { value: "linked", label: "已关联材料" },
              { value: "unlinked", label: "未关联材料" },
            ]}
            onChange={(value) => setFilters((prev) => ({ ...prev, materialState: value }))}
          />
          <Button icon={<ReloadOutlined />} onClick={loadQuestions}>刷新</Button>
        </Space>
      </Card>

      <Card
        className="glass-card"
        bordered={false}
        title={`${filteredQuestions.length} 道题目`}
      >
        <Table
          rowKey={(record) => `${record.subject}-${record.source}-${record.id}`}
          columns={columns}
          dataSource={filteredQuestions}
          loading={loading}
          scroll={{ x: 1360 }}
          pagination={{ pageSize: 8, showSizeChanger: true }}
        />
      </Card>

      <Modal title="编辑题目" open={!!editing} onCancel={() => setEditing(null)} onOk={handleEdit} width={760}>
        <Form form={editForm} layout="vertical">
          <Form.Item name="title" label="标题" rules={[{ required: true, message: "请输入标题" }]}>
            <Input />
          </Form.Item>
          <Space.Compact block>
            <Form.Item name="level1" label="题型" style={{ width: "70%" }}>
              <Select options={questionTypes.map((value) => ({ value, label: value }))} />
            </Form.Item>
            <Form.Item name="question_no" label="题号" style={{ width: "30%" }}>
              <Input />
            </Form.Item>
          </Space.Compact>
          <Form.Item name="stem" label="题目原文" rules={[{ required: true, message: "请输入题目原文" }]}>
            <Input.TextArea rows={8} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
}

export default QuestionBank;
