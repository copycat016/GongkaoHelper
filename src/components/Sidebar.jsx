import { useEffect, useState } from "react";
import { Button } from "antd";
import {
  PauseOutlined,
  SoundOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  RightOutlined,
} from "@ant-design/icons";
import { menuItems } from "./sidebarItems";
import { useMusic } from "./musicContext";
import ThemePanel from "./ThemePanel";

function Sidebar({
  collapsed,
  selectedKey,
  onSelect,
  onToggleCollapse,
}) {
  const { currentTrack, currentTime, playing, pause } = useMusic();

  return (
    <div className="sidebar-shell">
      <div className="brand-box">
        <div className={collapsed ? "brand-mark collapsed" : "brand-mark"}>
          <img src="/app-icon.png" alt="" className="brand-icon" />
          {!collapsed && <span className="brand-name">Masiro</span>}
        </div>
        <Button
          type="text"
          className="sidebar-collapse-button"
          icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
          onClick={onToggleCollapse}
          aria-label={collapsed ? "展开导航" : "收起导航"}
        />
      </div>
      <nav className="app-menu" aria-label="主导航">
        {menuItems.map((item) => {
          const selected = item.key === selectedKey;
          return (
            <button
              key={item.key}
              type="button"
              className={`app-menu-item${selected ? " selected" : ""}`}
              aria-current={selected ? "page" : undefined}
              title={collapsed ? item.label : undefined}
              onPointerUp={(event) => {
                if (event.pointerType === "mouse" && event.button !== 0) return;
                onSelect(item.key);
              }}
              onKeyDown={(event) => {
                if (event.key === "Enter" || event.key === " ") {
                  event.preventDefault();
                  onSelect(item.key);
                }
              }}
            >
              <span className="app-menu-icon">{item.icon}</span>
              {!collapsed && <span className="app-menu-label">{item.label}</span>}
            </button>
          );
        })}
      </nav>
      <div className="sidebar-bottom">
        <div className={collapsed ? "sidebar-theme-row collapsed" : "sidebar-theme-row"}>
          <ThemePanel compact={collapsed} />
        </div>
        {playing && currentTrack ? (
          <MiniMusicDock
            collapsed={collapsed}
            track={currentTrack}
            currentTime={currentTime}
            onPause={pause}
            onOpen={() => onSelect("/music")}
          />
        ) : (
          <IdleCompanionDock collapsed={collapsed} />
        )}
      </div>
    </div>
  );
}

function IdleCompanionDock({ collapsed }) {
  const now = useNow();
  return (
    <div
      className={collapsed ? "idle-dock collapsed" : "idle-dock"}
      aria-hidden="true"
    >
      <div className="idle-dock-info">
        <strong className="idle-dock-clock">{formatClock(now)}</strong>
        {!collapsed && <span className="idle-dock-date">{formatDockDate(now)}</span>}
      </div>
    </div>
  );
}

function MiniMusicDock({ collapsed, track, currentTime, onPause, onOpen }) {
  const lyric = getLyricLine(track, currentTime);
  return (
    <div className={collapsed ? "mini-music-dock collapsed" : "mini-music-dock"}>
      {collapsed ? (
        <>
          <Button className="mini-music-icon-button" icon={<PauseOutlined />} onClick={onPause} aria-label="暂停音乐" />
          <Button className="mini-music-icon-button" icon={<RightOutlined />} onClick={onOpen} aria-label="进入播放器" />
        </>
      ) : (
        <>
          <div className="mini-music-cover">
            {track?.cover_url ? <img src={track.cover_url} alt="" /> : <SoundOutlined />}
          </div>
          <div className="mini-music-info">
            <span className="mini-music-kicker"><SoundOutlined /> 正在播放</span>
            <strong>{track.title || "未知曲目"}</strong>
            <small>{lyric}</small>
          </div>
          <div className="mini-music-actions">
            <Button className="mini-music-icon-button" icon={<PauseOutlined />} onClick={onPause} aria-label="暂停音乐" />
            <Button className="mini-music-icon-button" icon={<RightOutlined />} onClick={onOpen} aria-label="进入播放器" />
          </div>
        </>
      )}
    </div>
  );
}

function getLyricLine(track, currentTime) {
  const lyrics = track?.lyrics || "";
  if (!lyrics.trim()) return track?.artist || "轻声陪你学习";
  const lines = lyrics.split(/\r?\n/).map((line) => line.trim()).filter(Boolean);
  if (track.lyrics_type !== "lrc") return lines[0] || track?.artist || "轻声陪你学习";

  let active = "";
  const timeRe = /^\[(\d{1,2}):(\d{2})(?:\.(\d{1,3}))?]\s*(.*)$/;
  for (const line of lines) {
    const match = line.match(timeRe);
    if (!match) continue;
    const minutes = Number(match[1] || 0);
    const seconds = Number(match[2] || 0);
    const millis = Number((match[3] || "0").padEnd(3, "0"));
    const time = minutes * 60 + seconds + millis / 1000;
    if (time <= currentTime) active = match[4] || active;
    if (time > currentTime) break;
  }
  return active || lines.map((line) => line.replace(timeRe, "$4")).find(Boolean) || track?.artist || "轻声陪你学习";
}

function useNow() {
  const [now, setNow] = useState(() => new Date());

  useEffect(() => {
    const timer = window.setInterval(() => setNow(new Date()), 1000);
    return () => window.clearInterval(timer);
  }, []);

  return now;
}

function formatClock(date) {
  return `${String(date.getHours()).padStart(2, "0")}:${String(date.getMinutes()).padStart(2, "0")}`;
}

function formatDockDate(date) {
  const weekdays = ["日", "一", "二", "三", "四", "五", "六"];
  return `${date.getMonth() + 1}月${date.getDate()}日 · 周${weekdays[date.getDay()]}`;
}

export default Sidebar;
