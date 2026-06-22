import { useEffect, useMemo, useState } from "react";
import { Button, Form, Input, message } from "antd";
import { LockOutlined, LoginOutlined, UserOutlined } from "@ant-design/icons";
import { useNavigate, useSearchParams } from "react-router-dom";
import { getAccessToken } from "../api/request";
import { login } from "../api/auth";

function Login() {
  const [loading, setLoading] = useState(false);
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();

  const next = useMemo(() => {
    const value = searchParams.get("next") || "/";
    return value.startsWith("/") && !value.startsWith("//") ? value : "/";
  }, [searchParams]);

  useEffect(() => {
    if (getAccessToken()) {
      navigate(next, { replace: true });
    }
  }, [navigate, next]);

  const handleFinish = async (values) => {
    setLoading(true);
    try {
      await login(values);
      message.success("登录成功");
      navigate(next, { replace: true });
    } finally {
      setLoading(false);
    }
  };

  return (
    <main className="login-page">
      <section className="login-panel">
        <div className="login-brand">
          <div className="login-mark">G</div>
          <div>
            <h1>GongkaoHelper</h1>
            <p>登录继续使用</p>
          </div>
        </div>
        <Form
          layout="vertical"
          initialValues={{ username: "admin" }}
          onFinish={handleFinish}
          requiredMark={false}
        >
          <Form.Item name="username" label="账号" rules={[{ required: true, message: "请输入账号" }]}>
            <Input prefix={<UserOutlined />} placeholder="请输入账号" autoComplete="username" />
          </Form.Item>
          <Form.Item name="password" label="密码" rules={[{ required: true, message: "请输入密码" }]}>
            <Input.Password prefix={<LockOutlined />} placeholder="请输入密码" autoComplete="current-password" />
          </Form.Item>
          <Button
            className="login-submit"
            type="primary"
            htmlType="submit"
            icon={<LoginOutlined />}
            loading={loading}
            block
          >
            登录
          </Button>
        </Form>
      </section>
    </main>
  );
}

export default Login;
