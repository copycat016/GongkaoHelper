import { useEffect, useMemo, useState } from "react";
import { Outlet, useLocation, useNavigate } from "react-router-dom";
import { Button, Layout } from "antd";
import {
  MenuUnfoldOutlined,
} from "@ant-design/icons";
import Sidebar from "./Sidebar";
import { menuItems } from "./sidebarItems";

const { Sider, Content } = Layout;

function AppLayout() {
  const [collapsed, setCollapsed] = useState(false);
  const [isMobile, setIsMobile] = useState(false);
  const navigate = useNavigate();
  const location = useLocation();

  const selectedKey = useMemo(() => {
    if (["/logs", "/plans", "/calendar"].includes(location.pathname)) {
      return "/study";
    }
    if (["/ocr", "/questions", "/mistakes"].includes(location.pathname) || location.pathname.startsWith("/questions/")) {
      return "/intake";
    }
    const exact = menuItems.find((item) => item.key === location.pathname);
    return exact ? exact.key : "/";
  }, [location.pathname]);

  useEffect(() => {
    const media = window.matchMedia("(max-width: 767px)");
    const handleChange = (event) => {
      setIsMobile(event.matches);
      if (event.matches) {
        setCollapsed(true);
      }
    };

    handleChange(media);
    media.addEventListener("change", handleChange);

    return () => media.removeEventListener("change", handleChange);
  }, []);

  return (
    <Layout className="app-layout">
      <Sider
        className="app-sider"
        width={248}
        collapsedWidth={isMobile ? 0 : 82}
        collapsed={collapsed}
        breakpoint="xl"
        onBreakpoint={(broken) => setCollapsed(broken)}
        trigger={null}
      >
        <Sidebar
          collapsed={collapsed}
          selectedKey={selectedKey}
          onSelect={navigate}
          onToggleCollapse={() => setCollapsed((value) => !value)}
        />
      </Sider>
      <Layout className="main-layout">
        {collapsed && isMobile && (
          <Button
            type="text"
            className="sidebar-open-button"
            icon={<MenuUnfoldOutlined />}
            onClick={() => setCollapsed(false)}
            aria-label="打开导航"
          />
        )}
        <Content className="app-content">
          <div className="page-shell">
            <Outlet />
          </div>
        </Content>
      </Layout>
    </Layout>
  );
}

export default AppLayout;
