function PageHeader({ extra }) {
  if (!extra) return null;

  return <div className="page-header page-header-actions">{extra}</div>;
}

export default PageHeader;
