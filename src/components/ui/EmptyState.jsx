import { Empty } from "antd";

// 统一空状态母题：图标 + 描述 + 可选操作。替代各页裸用 antd Empty。
export default function EmptyState({
  image,
  description = "暂无数据",
  children,
  className = "",
}) {
  return (
    <div className={`ui-empty ${className}`.trim()}>
      <Empty image={image || Empty.PRESENTED_IMAGE_SIMPLE} description={description} />
      {children}
    </div>
  );
}
