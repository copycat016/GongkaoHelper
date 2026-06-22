import { useEffect, useState } from "react";
import { Button, Drawer, Form, Input, message, Select, Slider, Space, Switch, Upload } from "antd";
import { BgColorsOutlined, UploadOutlined } from "@ant-design/icons";
import { getThemeConfig, saveThemeConfig } from "../api/theme";
import { applyThemeConfig, defaultConfig, palettes } from "../theme/applyThemeConfig";
import { useThemePalette } from "../theme/themeContext";

function ThemePanel({ compact = false }) {
  const [open, setOpen] = useState(false);
  const [form] = Form.useForm();
  const [config, setConfig] = useState(defaultConfig);
  const { setPaletteKey } = useThemePalette();

  useEffect(() => {
    getThemeConfig().then((stored) => {
      const applied = applyThemeConfig(stored);
      setConfig(applied);
      setPaletteKey(applied.palette);
      form.setFieldsValue(applied);
    });
  }, [form, setPaletteKey]);

  const handleValuesChange = async (_, values) => {
    const next = applyThemeConfig(values);
    setConfig(next);
    setPaletteKey(next.palette);
    try {
      await saveThemeConfig(next);
    } catch (error) {
      message.error(error.message);
    }
  };

  const handleUpload = (file) => {
    compressImage(file)
      .then((imageData) => {
      const next = {
        ...form.getFieldsValue(),
        backgroundEnabled: true,
          backgroundImage: imageData,
      };
      form.setFieldsValue(next);
      handleValuesChange(null, next);
        message.success("背景图已上传");
      })
      .catch(() => message.error("背景图上传失败，请换一张图片"));
    return false;
  };

  const clearBackground = () => {
    const next = {
      ...form.getFieldsValue(),
      backgroundEnabled: false,
      backgroundImage: "",
    };
    form.setFieldsValue(next);
    handleValuesChange(null, next);
  };

  const handleDockUpload = (file) => {
    compressImage(file)
      .then((imageData) => {
        const next = { ...form.getFieldsValue(), dockImage: imageData };
        form.setFieldsValue(next);
        handleValuesChange(null, next);
        message.success("时钟卡背景已更新");
      })
      .catch(() => message.error("背景图上传失败，请换一张图片"));
    return false;
  };

  const clearDock = () => {
    const next = { ...form.getFieldsValue(), dockImage: "" };
    form.setFieldsValue(next);
    handleValuesChange(null, next);
  };

  return (
    <>
      <Button
        icon={<BgColorsOutlined />}
        onClick={() => setOpen(true)}
        aria-label="打开主题设置"
        title="主题"
      >
        {!compact && "主题"}
      </Button>
      <Drawer title="Aozora Desk 主题" open={open} onClose={() => setOpen(false)} width={380}>
        <Form
          form={form}
          layout="vertical"
          initialValues={config}
          onValuesChange={handleValuesChange}
        >
          <Form.Item name="palette" label="配色方案">
            <Select
              options={Object.entries(palettes).map(([value, item]) => ({
                value,
                label: (
                  <Space>
                    <span
                      className="theme-swatch"
                      style={{ background: item.primaryBlue }}
                    />
                    {item.label}
                  </Space>
                ),
              }))}
            />
          </Form.Item>
          <Form.Item name="backgroundEnabled" label="背景图开关" valuePropName="checked">
            <Switch />
          </Form.Item>
          <Form.Item label="上传背景图">
            <Space wrap>
              <Upload beforeUpload={handleUpload} showUploadList={false} accept="image/*">
                <Button icon={<UploadOutlined />}>选择本地图片</Button>
              </Upload>
              <Button onClick={clearBackground}>清除背景</Button>
            </Space>
          </Form.Item>
          <Form.Item name="backgroundImage" hidden>
            <Input />
          </Form.Item>
          <Form.Item name="blur" label="背景模糊度">
            <Slider min={0} max={24} />
          </Form.Item>
          <Form.Item name="brightness" label="背景亮度">
            <Slider min={40} max={130} />
          </Form.Item>
          <Form.Item name="maskOpacity" label="背景遮罩透明度">
            <Slider min={0} max={90} />
          </Form.Item>
          <Form.Item name="cardOpacity" label="内容卡片透明度">
            <Slider min={38} max={96} />
          </Form.Item>
          <Form.Item label="左下角时钟卡背景图" extra="建议用横图，左侧清晰、右侧自动毛玻璃淡出。不上传则用默认占位图。">
            <Space wrap>
              <Upload beforeUpload={handleDockUpload} showUploadList={false} accept="image/*">
                <Button icon={<UploadOutlined />}>上传时钟卡背景</Button>
              </Upload>
              <Button onClick={clearDock}>恢复默认</Button>
            </Space>
          </Form.Item>
          <Form.Item name="dockImage" hidden>
            <Input />
          </Form.Item>
          <Form.Item name="backgroundSize" label="背景缩放方式">
            <Select options={[{ value: "cover", label: "Cover" }, { value: "contain", label: "Contain" }, { value: "auto", label: "Auto" }]} />
          </Form.Item>
          <Form.Item name="backgroundPosition" label="背景位置">
            <Select options={[{ value: "center", label: "居中" }, { value: "top", label: "顶部" }, { value: "bottom", label: "底部" }]} />
          </Form.Item>
        </Form>
      </Drawer>
    </>
  );
}

function compressImage(file) {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onerror = reject;
    reader.onload = () => {
      const image = new Image();
      image.onerror = reject;
      image.onload = () => {
        const maxSize = 1800;
        const scale = Math.min(1, maxSize / Math.max(image.width, image.height));
        const width = Math.round(image.width * scale);
        const height = Math.round(image.height * scale);
        const canvas = document.createElement("canvas");
        canvas.width = width;
        canvas.height = height;
        const context = canvas.getContext("2d");
        context.drawImage(image, 0, 0, width, height);
        resolve(canvas.toDataURL("image/jpeg", 0.86));
      };
      image.src = reader.result;
    };
    reader.readAsDataURL(file);
  });
}

export default ThemePanel;
