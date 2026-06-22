import { Button, Slider, Tooltip } from "antd";
import {
  BackwardOutlined,
  CustomerServiceOutlined,
  ForwardOutlined,
  MenuUnfoldOutlined,
  PauseCircleOutlined,
  PlayCircleOutlined,
  SoundOutlined,
} from "@ant-design/icons";
import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useMusic } from "./musicContext";

function GlobalMusicBar({ collapsed = false, isMobile = false }) {
  const [open, setOpen] = useState(false);
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
    volume,
    setVolume,
    audioError,
  } = useMusic();

  if (!currentTrack) return null;

  return (
    <div className={`global-music-widget ${collapsed ? "collapsed" : ""} ${open ? "open" : ""} ${isMobile ? "mobile" : ""}`}>
      {open && (
        <div className="global-music-popover">
          <div className="global-music-popover-head" onClick={() => navigate("/music")}>
            <div className="global-music-icon large">
              {currentTrack.cover_url ? (
                <img src={currentTrack.cover_url} alt="" onError={(e) => { e.target.style.display = "none"; }} />
              ) : (
                <CustomerServiceOutlined />
              )}
            </div>
            <div className="global-music-meta">
              <strong>{currentTrack.title || "未知曲目"}</strong>
              <span>{audioError || trackMetaLine(currentTrack)}</span>
            </div>
          </div>
          <div className="global-music-controls">
            <Tooltip title="上一首">
              <Button shape="circle" size="small" icon={<BackwardOutlined />} onClick={(e) => { e.stopPropagation(); prevTrack(); }} />
            </Tooltip>
            <Tooltip title={playing ? "暂停" : "播放"}>
              <Button type="primary" shape="circle" icon={playing ? <PauseCircleOutlined /> : <PlayCircleOutlined />} onClick={(e) => { e.stopPropagation(); togglePlay(); }} />
            </Tooltip>
            <Tooltip title="下一首">
              <Button shape="circle" size="small" icon={<ForwardOutlined />} onClick={(e) => { e.stopPropagation(); nextTrack(); }} />
            </Tooltip>
          </div>
          <div className="global-music-progress" onClick={(e) => e.stopPropagation()}>
            <span>{formatSeconds(currentTime)}</span>
            <Slider value={percent} onChange={seek} tooltip={{ formatter: null }} />
            <span>{formatSeconds(duration)}</span>
          </div>
          <div className="global-music-volume" onClick={(e) => e.stopPropagation()}>
            <SoundOutlined />
            <Slider value={volume} onChange={setVolume} min={0} max={100} tooltip={{ formatter: null }} />
          </div>
          <Button block className="soft-button" icon={<MenuUnfoldOutlined />} onClick={() => navigate("/music")}>
            打开播放器
          </Button>
        </div>
      )}

      <button type="button" className="global-music-mini" onClick={() => setOpen((value) => !value)}>
        <div className="global-music-icon">
          {currentTrack.cover_url ? (
            <img src={currentTrack.cover_url} alt="" onError={(e) => { e.target.style.display = "none"; }} />
          ) : (
            <CustomerServiceOutlined />
          )}
        </div>
        <div className="global-music-meta">
          <strong>{currentTrack.title || "未知曲目"}</strong>
          <span>{audioError || trackMetaLine(currentTrack)}</span>
        </div>
        <span className={playing ? "global-music-pulse playing" : "global-music-pulse"} />
      </button>
    </div>
  );
}

function trackMetaLine(track) {
  return [track.artist || "未知艺术家", track.album].filter(Boolean).join(" · ");
}

function formatSeconds(seconds) {
  if (!Number.isFinite(seconds) || seconds < 0) return "00:00";
  const minutes = Math.floor(seconds / 60);
  const rest = Math.floor(seconds % 60);
  return `${String(minutes).padStart(2, "0")}:${String(rest).padStart(2, "0")}`;
}

export default GlobalMusicBar;
