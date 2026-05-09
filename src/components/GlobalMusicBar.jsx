import { Button, Slider, Tag } from "antd";
import {
  BackwardOutlined,
  CustomerServiceOutlined,
  ForwardOutlined,
  PauseCircleOutlined,
  PlayCircleOutlined,
} from "@ant-design/icons";
import { useNavigate } from "react-router-dom";
import { useMusic } from "./MusicProvider";

function GlobalMusicBar() {
  const navigate = useNavigate();
  const {
    currentTrack,
    playing,
    percent,
    currentTime,
    duration,
    togglePlay,
    nextTrack,
    prevTrack,
    seek,
  } = useMusic();

  if (!currentTrack) return null;

  return (
    <div className="global-music-bar">
      <div className="global-music-main" onClick={() => navigate("/music")}>
        <div className="global-music-icon"><CustomerServiceOutlined /></div>
        <div className="global-music-meta">
          <strong>{currentTrack.title}</strong>
          <span>{trackMetaLine(currentTrack)}</span>
        </div>
      </div>
      <div className="global-music-controls">
        <Button shape="circle" size="small" icon={<BackwardOutlined />} onClick={prevTrack} />
        <Button type="primary" shape="circle" icon={playing ? <PauseCircleOutlined /> : <PlayCircleOutlined />} onClick={togglePlay} />
        <Button shape="circle" size="small" icon={<ForwardOutlined />} onClick={nextTrack} />
      </div>
      <div className="global-music-progress">
        <span>{formatSeconds(currentTime)}</span>
        <Slider value={percent} onChange={seek} tooltip={{ formatter: null }} />
        <span>{formatSeconds(duration)}</span>
      </div>
      <Tag className="global-music-tag">后台播放</Tag>
    </div>
  );
}

function trackMetaLine(track) {
  return [track.artist || "未知艺术家", track.album].filter(Boolean).join(" · ");
}

function formatSeconds(seconds) {
  if (!Number.isFinite(seconds)) return "00:00";
  const minutes = Math.floor(seconds / 60);
  const rest = Math.floor(seconds % 60);
  return `${String(minutes).padStart(2, "0")}:${String(rest).padStart(2, "0")}`;
}

export default GlobalMusicBar;
