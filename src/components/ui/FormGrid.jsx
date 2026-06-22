import { Col, Row } from "antd";

// 表单两列栅格母题：统一 gutter（消除各页 12 vs 14 的不一致）。
// 用法：<FormGrid><FormCol><Form.Item/></FormCol>...</FormGrid>
export function FormGrid({ children, gutter = 16, className = "" }) {
  return (
    <Row gutter={[gutter, gutter]} className={className}>
      {children}
    </Row>
  );
}

// 默认两列（窄屏单列）。需要整行时传 sm={24}；其余 antd Col 属性经 ...rest 透传。
export function FormCol({ children, xs = 24, sm = 12, ...rest }) {
  return (
    <Col xs={xs} sm={sm} {...rest}>
      {children}
    </Col>
  );
}
