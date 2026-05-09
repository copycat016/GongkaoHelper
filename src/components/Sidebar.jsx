import { Menu } from "antd";
import {
  BookOutlined,
  ClockCircleOutlined,
  CloudUploadOutlined,
  DashboardOutlined,
  DatabaseOutlined,
  EditOutlined,
  HistoryOutlined,
  CustomerServiceOutlined,
  RobotOutlined,
} from "@ant-design/icons";

export const menuItems = [
  { key: "/", icon: <DashboardOutlined />, label: "首页总览" },
  { key: "/ocr", icon: <CloudUploadOutlined />, label: "OCR 识题" },
  { key: "/questions", icon: <DatabaseOutlined />, label: "题库管理" },
  { key: "/mistakes", icon: <HistoryOutlined />, label: "错题库" },
  { key: "/essay", icon: <EditOutlined />, label: "申论批改" },
  { key: "/pomodoro", icon: <ClockCircleOutlined />, label: "番茄钟" },
  { key: "/music", icon: <CustomerServiceOutlined />, label: "音乐播放器" },
  { key: "/study", icon: <BookOutlined />, label: "学习管理" },
  { key: "/ai", icon: <RobotOutlined />, label: "配置" },
];

function Sidebar({ collapsed, selectedKey, onSelect }) {
  return (
    <>
      <div className="brand-box">
        <div className="brand-mark">{collapsed ? "A" : "Aozora"}</div>
      </div>
      <Menu
        className="app-menu"
        mode="inline"
        selectedKeys={[selectedKey]}
        items={menuItems}
        onClick={({ key }) => onSelect(key)}
      />
    </>
  );
}

export default Sidebar;
