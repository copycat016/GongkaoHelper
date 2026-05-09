import { useEffect, useState } from "react";
import { Button, Drawer, Form, Input, message, Select, Slider, Space, Switch, Upload } from "antd";
import { BgColorsOutlined, UploadOutlined } from "@ant-design/icons";
import { getThemeConfig, saveThemeConfig } from "../api/theme";

const palettes = {
  aozora: {
    label: "Aozora Blue",
    bgMain: "#f5f9ff",
    bgMaskRgb: "245, 249, 255",
    primaryBlue: "#6aa7e8",
    deepBlueGray: "#2f4058",
    lightBlue: "#dceeff",
    accentPurple: "#b9a7f5",
    accentPink: "#f5b7cf",
    textBlueGray: "#5d6f86",
  },
  sakura: {
    label: "Sakura Mist",
    bgMain: "#fff7fb",
    bgMaskRgb: "255, 247, 251",
    primaryBlue: "#7aa8f5",
    deepBlueGray: "#3f4058",
    lightBlue: "#eef3ff",
    accentPurple: "#c6b5ff",
    accentPink: "#f4a9c6",
    textBlueGray: "#75677a",
  },
  matcha: {
    label: "Matcha Morning",
    bgMain: "#f6fbf5",
    bgMaskRgb: "246, 251, 245",
    primaryBlue: "#73b7c8",
    deepBlueGray: "#314c4f",
    lightBlue: "#e2f4f5",
    accentPurple: "#b8c9ef",
    accentPink: "#f2b8bd",
    textBlueGray: "#607b75",
  },
  sumi: {
    label: "Sumi Soft",
    bgMain: "#f7f8fb",
    bgMaskRgb: "247, 248, 251",
    primaryBlue: "#7d9fd8",
    deepBlueGray: "#313949",
    lightBlue: "#e6ecf7",
    accentPurple: "#a99fe2",
    accentPink: "#e8aebc",
    textBlueGray: "#687287",
  },
};

const defaultConfig = {
  palette: "aozora",
  backgroundEnabled: false,
  backgroundImage: "",
  blur: 0,
  brightness: 100,
  maskOpacity: 34,
  backgroundSize: "cover",
  backgroundPosition: "center",
  cardOpacity: 72,
};

export function applyThemeConfig(config) {
  const next = { ...defaultConfig, ...config };
  const palette = palettes[next.palette] || palettes.aozora;
  const root = document.documentElement;
  root.style.setProperty("--bg-main", palette.bgMain);
  root.style.setProperty("--bg-mask-rgb", palette.bgMaskRgb);
  root.style.setProperty("--primary-blue", palette.primaryBlue);
  root.style.setProperty("--deep-blue-gray", palette.deepBlueGray);
  root.style.setProperty("--light-blue", palette.lightBlue);
  root.style.setProperty("--accent-purple", palette.accentPurple);
  root.style.setProperty("--accent-pink", palette.accentPink);
  root.style.setProperty("--text-blue-gray", palette.textBlueGray);
  root.style.setProperty("--line-soft", `${palette.primaryBlue}29`);
  root.style.setProperty("--bg-blur", `${next.blur}px`);
  root.style.setProperty("--bg-brightness", `${next.brightness / 100}`);
  root.style.setProperty("--bg-mask-opacity", `${next.maskOpacity / 100}`);
  root.style.setProperty("--card-opacity", `${next.cardOpacity / 100}`);
  root.style.setProperty("--bg-image", next.backgroundEnabled && next.backgroundImage ? `url("${next.backgroundImage}")` : "none");
  root.style.setProperty("--bg-size", next.backgroundSize);
  root.style.setProperty("--bg-position", next.backgroundPosition);
  return next;
}

function ThemePanel() {
  const [open, setOpen] = useState(false);
  const [form] = Form.useForm();
  const [config, setConfig] = useState(defaultConfig);

  useEffect(() => {
    const stored = applyThemeConfig(getThemeConfig());
    setConfig(stored);
    form.setFieldsValue(stored);
  }, [form]);

  const handleValuesChange = (_, values) => {
    const next = applyThemeConfig(values);
    setConfig(next);
    try {
      saveThemeConfig(next);
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

  return (
    <>
      <Button className="soft-button" icon={<BgColorsOutlined />} onClick={() => setOpen(true)}>
        主题
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
