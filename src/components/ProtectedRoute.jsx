import { Navigate, useLocation } from "react-router-dom";
import { getAccessToken } from "../api/request";

function ProtectedRoute({ children }) {
  const location = useLocation();

  if (!getAccessToken()) {
    const next = `${location.pathname}${location.search}`;
    return <Navigate to={`/login?next=${encodeURIComponent(next)}`} replace />;
  }

  return children;
}

export default ProtectedRoute;
