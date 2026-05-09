import { Button, Card, DatePicker, Select, Space, Table, Tag } from "antd";
import PageHeader from "../components/PageHeader";
import { mistakesMock } from "../api/mockData";

function MistakeBook() {
  const columns = [
    { title: "错题", dataIndex: "title" },
    { title: "题型", dataIndex: "type" },
    { title: "错误原因", dataIndex: "reason", render: (value) => <Tag color="pink">{value}</Tag> },
    { title: "掌握程度", dataIndex: "mastery", render: (value) => <Tag color={value === "未掌握" ? "red" : "blue"}>{value}</Tag> },
    { title: "复习次数", dataIndex: "reviews" },
    { title: "下次复习", dataIndex: "nextReview" },
    { title: "操作", render: () => <Space><Button size="small">详情</Button><Button size="small">编辑原因</Button><Button size="small">更新掌握</Button></Space> },
  ];

  return (
    <div className="page-grid">
      <PageHeader eyebrow="Mistakes" title="错题库" desc="按错误原因和掌握程度组织复习节奏。" />
      <Card className="glass-card" bordered={false}>
        <Space wrap>
          <Select allowClear placeholder="题型筛选" style={{ width: 180 }} options={["资料分析", "言语理解与表达", "判断推理"].map((value) => ({ value, label: value }))} />
          <Select allowClear placeholder="错误原因" style={{ width: 180 }} options={["粗心", "知识点不会", "题意理解错误", "计算错误", "逻辑判断错误", "时间不够", "选项陷阱", "材料定位错误"].map((value) => ({ value, label: value }))} />
          <Select allowClear placeholder="掌握程度" style={{ width: 160 }} options={["未掌握", "模糊", "基本掌握", "熟练"].map((value) => ({ value, label: value }))} />
          <DatePicker placeholder="下次复习时间" />
        </Space>
      </Card>
      <Card className="glass-card" bordered={false}>
        <Table rowKey="id" columns={columns} dataSource={mistakesMock} />
      </Card>
    </div>
  );
}

export default MistakeBook;
