import { Button, Card, Select, Space, Table, Tag } from "antd";
import PageHeader from "../components/PageHeader";
import { questionsMock } from "../api/mockData";

function QuestionBank() {
  const columns = [
    { title: "科目", dataIndex: "subject" },
    { title: "一级题型", dataIndex: "level1" },
    { title: "二级题型", dataIndex: "level2" },
    { title: "题干", dataIndex: "title" },
    { title: "答案", dataIndex: "answer" },
    { title: "难度", dataIndex: "difficulty" },
    { title: "标签", dataIndex: "tags", render: (tags) => tags.map((tag) => <Tag key={tag}>{tag}</Tag>) },
    { title: "操作", render: () => <Space><Button size="small">详情</Button><Button size="small">编辑</Button><Button size="small">加入错题</Button><Button size="small" danger>删除</Button></Space> },
  ];

  return (
    <div className="page-grid">
      <PageHeader eyebrow="Question Bank" title="题库管理" desc="行测与申论题库列表，支持分类筛选、编辑和加入错题库。" />
      <Card className="glass-card" bordered={false}>
        <Space wrap>
          <Select allowClear placeholder="科目" style={{ width: 140 }} options={["行测", "申论"].map((value) => ({ value, label: value }))} />
          <Select allowClear placeholder="一级题型" style={{ width: 180 }} options={["常识判断", "言语理解与表达", "数量关系", "判断推理", "资料分析", "归纳概括题", "综合分析题", "提出对策题", "应用文写作题", "文章论述题", "公文写作题"].map((value) => ({ value, label: value }))} />
          <Select allowClear placeholder="二级题型" style={{ width: 160 }} options={["增长率", "图形推理", "材料概括"].map((value) => ({ value, label: value }))} />
          <Select mode="tags" placeholder="标签筛选" style={{ minWidth: 180 }} />
        </Space>
      </Card>
      <Card className="glass-card" bordered={false}>
        <Table rowKey="id" columns={columns} dataSource={questionsMock} scroll={{ x: 1100 }} />
      </Card>
    </div>
  );
}

export default QuestionBank;
