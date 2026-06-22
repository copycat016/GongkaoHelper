import {
  BookOutlined,
  ClockCircleOutlined,
  CloudUploadOutlined,
  CustomerServiceOutlined,
  DashboardOutlined,
  EditOutlined,
  RobotOutlined,
} from "@ant-design/icons";

export const menuItems = [
  { key: "/", icon: <DashboardOutlined />, label: "首页总览" },
  { key: "/intake", icon: <CloudUploadOutlined />, label: "录入器" },
  { key: "/essay", icon: <EditOutlined />, label: "申论批改" },
  { key: "/pomodoro", icon: <ClockCircleOutlined />, label: "番茄钟" },
  { key: "/music", icon: <CustomerServiceOutlined />, label: "音乐播放器" },
  { key: "/study", icon: <BookOutlined />, label: "学习管理" },
  { key: "/ai", icon: <RobotOutlined />, label: "配置" },
];
