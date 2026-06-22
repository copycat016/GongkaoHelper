import { useEffect, useState } from "react";

const STORAGE_KEY = "gkweb:preview-banner-dismissed";

function PreviewBanner() {
  const releaseChannel = import.meta.env.VITE_RELEASE_CHANNEL;
  const version = import.meta.env.VITE_APP_VERSION || "";
  const feedbackUrl =
    import.meta.env.VITE_FEEDBACK_URL ||
    "https://github.com/anomalyco/GongkaoHelper/issues/new";

  const [dismissed, setDismissed] = useState(() => {
    if (typeof window === "undefined") return false;
    return window.sessionStorage.getItem(STORAGE_KEY) === "1";
  });

  useEffect(() => {
    if (dismissed && typeof window !== "undefined") {
      window.sessionStorage.setItem(STORAGE_KEY, "1");
    }
  }, [dismissed]);

  if (releaseChannel !== "preview" || dismissed) {
    return null;
  }

  return (
    <div className="preview-banner" role="status">
      <span className="preview-banner__tag">PREVIEW</span>
      <span className="preview-banner__text">
        当前为预览版{version ? ` ${version}` : ""}，可能存在缺陷与数据不兼容风险，欢迎{" "}
        <a href={feedbackUrl} target="_blank" rel="noopener noreferrer">
          反馈问题
        </a>
        。
      </span>
      <button
        type="button"
        className="preview-banner__close"
        aria-label="关闭预览提示"
        onClick={() => setDismissed(true)}
      >
        ×
      </button>
    </div>
  );
}

export default PreviewBanner;
