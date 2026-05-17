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
        <div className={collapsed ? "brand-mark collapsed" : "brand-mark"}>
          <img src="/app-icon.png" alt="" className="brand-icon" />
          {!collapsed && <span className="brand-name">Masiro</span>}
        </div>
      </div>
      <nav className="app-menu" aria-label="主导航">
        {menuItems.map((item) => {
          const selected = item.key === selectedKey;
          return (
            <button
              key={item.key}
              type="button"
              className={`app-menu-item${selected ? " selected" : ""}`}
              aria-current={selected ? "page" : undefined}
              title={collapsed ? item.label : undefined}
              onPointerUp={(event) => {
                if (event.pointerType === "mouse" && event.button !== 0) return;
                onSelect(item.key);
              }}
              onKeyDown={(event) => {
                if (event.key === "Enter" || event.key === " ") {
                  event.preventDefault();
                  onSelect(item.key);
                }
              }}
            >
              <span className="app-menu-icon">{item.icon}</span>
              {!collapsed && <span className="app-menu-label">{item.label}</span>}
            </button>
          );
        })}
      </nav>
    </>
  );
}

export default Sidebar;
