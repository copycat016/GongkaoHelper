import { memo, useCallback, useEffect, useMemo, useRef, useState } from "react";
import {
  Button, Card, Col, Dropdown, Empty, Form, Image, Input, List, Modal, Pagination,
  Popconfirm, Row, Segmented, Select, Slider, Spin, Tag, Tooltip, Upload, message,
} from "antd";
import {
  BackwardOutlined, CustomerServiceOutlined, DeleteOutlined, EllipsisOutlined,
  ForwardOutlined, FolderAddOutlined, MinusCircleOutlined, PauseCircleOutlined,
  PlayCircleOutlined, PlusOutlined, ReloadOutlined, SearchOutlined,
  SoundOutlined, UploadOutlined, ReadOutlined,
} from "@ant-design/icons";
import {
  addTrackToPlaylist, applyTrackMetadata, createPlaylist, deletePlaylist,
  deleteTrack, fetchTrackLyrics, getPlaylistTracks, getPlaylists, getTracks, lookupTrackMetadata,
  removeTrackFromPlaylist, uploadTrack,
} from "../api/music";
import { useMusic } from "../components/MusicProvider";

const acceptTypes = ".mp3,.flac,.wav,.ogg,.m4a,.aac,audio/*";
const ALL_TRACKS_PLAYLIST = "__all_tracks__";
const DEFAULT_TRACKS_PAGE_SIZE = 8;
const TRACK_ROW_HEIGHT = 82;
const MIN_VISIBLE_LYRIC_LINES = 3;

function notifyError(error, fallback) {
  message.error(error?.message || fallback);
}

function useAutoFitText(value, { max = 42, min = 24 } = {}) {
  const ref = useRef(null);

  useEffect(() => {
    const node = ref.current;
    if (!node) return undefined;

    const fit = () => {
      node.style.fontSize = `${max}px`;
      let nextSize = max;
      while ((node.scrollWidth > node.clientWidth || node.scrollHeight > node.clientHeight) && nextSize > min) {
        nextSize -= 1;
        node.style.fontSize = `${nextSize}px`;
      }
    };

    fit();
    const observer = new ResizeObserver(fit);
    observer.observe(node);
    return () => observer.disconnect();
  }, [value, max, min]);

  return ref;
}

// 预定义菜单图标，避免每次 render 都创建新的 JSX 元素导致 Dropdown 不必要的重渲染
const MENU_ICON_SEARCH = <SearchOutlined />;
const MENU_ICON_FOLDER_ADD = <FolderAddOutlined />;
const MENU_ICON_MINUS_CIRCLE = <MinusCircleOutlined />;
const MENU_ICON_DELETE = <DeleteOutlined />;

function MusicPlayer() {
  const {
    tracks, setTracks, currentIndex, setCurrentIndex, currentTrack,
    playing, loopMode, setLoopMode, volume, setVolume,
    currentTime, duration, percent, audioError,
    togglePlay, playAt, nextTrack, prevTrack, seek,
  } = useMusic();

  const [playlists, setPlaylists] = useState([]);
  const [playlistId, setPlaylistId] = useState(ALL_TRACKS_PLAYLIST);
  const [playlistLoading, setPlaylistLoading] = useState(false);
  const [playlistError, setPlaylistError] = useState("");
  const [tracksLoading, setTracksLoading] = useState(false);
  const [tracksError, setTracksError] = useState("");
  const [trackPage, setTrackPage] = useState(1);
  const [trackPageSize, setTrackPageSize] = useState(DEFAULT_TRACKS_PAGE_SIZE);
  const [autoMetadataLoading, setAutoMetadataLoading] = useState(false);
  const [lyricsLoading, setLyricsLoading] = useState(false);
  const [volumeOpen, setVolumeOpen] = useState(false);
  const [playlistOpen, setPlaylistOpen] = useState(false);
  const [metadataOpen, setMetadataOpen] = useState(false);
  const [metadataTrack, setMetadataTrack] = useState(null);
  const [metadataCandidates, setMetadataCandidates] = useState([]);
  const [metadataLoading, setMetadataLoading] = useState(false);
  // 添加到歌单弹窗
  const [addToPlaylistOpen, setAddToPlaylistOpen] = useState(false);
  const [addToPlaylistTrack, setAddToPlaylistTrack] = useState(null);
  const [addToPlaylistLoading, setAddToPlaylistLoading] = useState(false);

  const [playlistForm] = Form.useForm();
  const [metadataForm] = Form.useForm();

  const tracksRequestRef = useRef(0);
  const trackListRef = useRef(null);
  const tracksAbortRef = useRef(null);
  const titleRef = useAutoFitText(currentTrack?.title || "选择歌单播放", { max: 42, min: 26 });

  const isPlaylistView = playlistId !== ALL_TRACKS_PLAYLIST;

  const mergeTrack = useCallback((updatedTrack) => {
    if (!updatedTrack?.id) return;
    setTracks((prev) => prev.map((track) => (track.id === updatedTrack.id ? { ...track, ...updatedTrack } : track)));
    setMetadataTrack((prev) => (prev?.id === updatedTrack.id ? { ...prev, ...updatedTrack } : prev));
  }, [setTracks]);

  // --- 数据加载 ---

  const fetchPlaylists = useCallback(async () => {
    setPlaylistLoading(true);
    setPlaylistError("");
    try {
      const items = await getPlaylists();
      setPlaylists(items || []);
      return items || [];
    } catch (error) {
      setPlaylistError(error?.message || "歌单加载失败");
      return [];
    } finally {
      setPlaylistLoading(false);
    }
  }, []);

  const fetchTracks = useCallback(async (targetPlaylistId, { resetIndex = true } = {}) => {
    tracksAbortRef.current?.abort();
    const controller = new AbortController();
    tracksAbortRef.current = controller;

    if (!targetPlaylistId) {
      setTracks([]);
      if (resetIndex) setCurrentIndex(0);
      return;
    }
    const requestId = ++tracksRequestRef.current;
    setTracksLoading(true);
    setTracksError("");
    try {
      const items = targetPlaylistId === ALL_TRACKS_PLAYLIST
        ? await getTracks({ signal: controller.signal })
        : await getPlaylistTracks(targetPlaylistId, { signal: controller.signal });
      if (requestId !== tracksRequestRef.current || controller.signal.aborted) return;
      setTracks(items || []);
      if (resetIndex) setCurrentIndex(0);
    } catch (error) {
      if (requestId !== tracksRequestRef.current || controller.signal.aborted) return;
      setTracksError(error?.message || "曲目加载失败");
      setTracks([]);
    } finally {
      if (requestId === tracksRequestRef.current && !controller.signal.aborted) {
        setTracksLoading(false);
      }
    }
  }, [setTracks, setCurrentIndex]);

  useEffect(() => {
    fetchPlaylists();
    fetchTracks(ALL_TRACKS_PLAYLIST);
  }, [fetchPlaylists, fetchTracks]);

  useEffect(() => () => {
    tracksAbortRef.current?.abort();
  }, []);

  const prevPlaylistIdRef = useRef(playlistId);
  useEffect(() => {
    if (prevPlaylistIdRef.current === playlistId) return;
    prevPlaylistIdRef.current = playlistId;
    setTrackPage(1);
    fetchTracks(playlistId);
  }, [playlistId, fetchTracks]);

  useEffect(() => {
    const maxPage = Math.max(1, Math.ceil(tracks.length / trackPageSize));
    if (trackPage > maxPage) {
      // 使用 queueMicrotask 避免在 effect 中同步调用 setState 触发级联渲染
      queueMicrotask(() => setTrackPage(maxPage));
    }
  }, [trackPage, trackPageSize, tracks.length]);

  useEffect(() => {
    const node = trackListRef.current;
    if (!node) return undefined;

    const updatePageSize = () => {
      const height = node.clientHeight;
      if (!height) return;
      const nextSize = Math.max(3, Math.min(50, Math.floor(height / TRACK_ROW_HEIGHT)));
      setTrackPageSize((prev) => (prev === nextSize ? prev : nextSize));
    };

    updatePageSize();
    const observer = new ResizeObserver(updatePageSize);
    observer.observe(node);
    window.addEventListener("resize", updatePageSize);
    return () => {
      observer.disconnect();
      window.removeEventListener("resize", updatePageSize);
    };
  }, []);

  // --- 乐观更新辅助函数 ---
  const adjustPlaylistCount = useCallback((targetId, delta) => {
    setPlaylists((prev) => prev.map((pl) =>
      pl.id === targetId ? { ...pl, track_count: Math.max(0, (pl.track_count || 0) + delta) } : pl
    ));
  }, []);

  // --- 上传 ---
  const beforeUpload = useCallback(async (file) => {
    try {
      await uploadTrack({ file, playlistId: isPlaylistView ? playlistId : undefined });
      message.success("音乐已上传");
      await fetchTracks(playlistId, { resetIndex: false });
      if (isPlaylistView) adjustPlaylistCount(playlistId, 1);
      else fetchPlaylists();
    } catch (error) {
      notifyError(error, "上传失败");
    }
    return false;
  }, [isPlaylistView, playlistId, fetchTracks, adjustPlaylistCount, fetchPlaylists]);

  // --- 歌单操作 ---
  const handleCreatePlaylist = useCallback(async () => {
    try {
      const values = await playlistForm.validateFields();
      const playlist = await createPlaylist({ ...values, owner_role: "root", enabled: true });
      message.success("歌单已创建");
      setPlaylistOpen(false);
      playlistForm.resetFields();
      await fetchPlaylists();
      setPlaylistId(playlist.id);
    } catch (error) {
      notifyError(error, "创建歌单失败");
    }
  }, [playlistForm, fetchPlaylists]);

  const handleDeletePlaylist = useCallback(async () => {
    if (!playlistId || !isPlaylistView) return;
    try {
      await deletePlaylist(playlistId);
      message.success("歌单已删除");
      setPlaylistId(ALL_TRACKS_PLAYLIST);
      await fetchPlaylists();
    } catch (error) {
      notifyError(error, "删除歌单失败");
    }
  }, [playlistId, isPlaylistView, fetchPlaylists]);

  // --- 从歌单移除（不删除文件） ---
  const handleRemoveFromPlaylist = useCallback(async (trackId) => {
    if (!isPlaylistView) return;
    try {
      await removeTrackFromPlaylist(playlistId, trackId);
      message.success("已从歌单移除");
      const newTracks = tracks.filter((t) => t.id !== trackId);
      setTracks(newTracks);
      adjustPlaylistCount(playlistId, -1);
      // 使用函数式更新避免闭包问题
      setCurrentIndex((prevIndex) => {
        const deletedIndex = tracks.findIndex((t) => t.id === trackId);
        if (deletedIndex === prevIndex) {
          if (newTracks.length === 0) return 0;
          if (deletedIndex >= newTracks.length) return newTracks.length - 1;
          return prevIndex;
        }
        if (deletedIndex < prevIndex) return Math.max(0, prevIndex - 1);
        return prevIndex;
      });
    } catch (error) {
      notifyError(error, "移除失败");
    }
  }, [isPlaylistView, playlistId, tracks, setTracks, setCurrentIndex, adjustPlaylistCount]);

  // --- 彻底删除（删除文件 + 从所有歌单移除） ---
  const handleDeleteTrack = useCallback(async (trackId) => {
    try {
      await deleteTrack(trackId);
      message.success("曲目已彻底删除");
      const newTracks = tracks.filter((t) => t.id !== trackId);
      setTracks(newTracks);
      fetchPlaylists(); // 影响所有歌单计数，需要全量刷新
      setCurrentIndex((prevIndex) => {
        const deletedIndex = tracks.findIndex((t) => t.id === trackId);
        if (deletedIndex === prevIndex) {
          if (newTracks.length === 0) return 0;
          if (deletedIndex >= newTracks.length) return newTracks.length - 1;
          return prevIndex;
        }
        if (deletedIndex < prevIndex) return Math.max(0, prevIndex - 1);
        return prevIndex;
      });
    } catch (error) {
      notifyError(error, "删除曲目失败");
    }
  }, [tracks, setTracks, setCurrentIndex, fetchPlaylists]);

  // --- 添加到歌单 ---
  const handleOpenAddToPlaylist = useCallback((track) => {
    setAddToPlaylistTrack(track);
    setAddToPlaylistOpen(true);
  }, []);

  const handleAddToPlaylist = useCallback(async (targetPlaylistId) => {
    if (!addToPlaylistTrack) return;
    setAddToPlaylistLoading(true);
    try {
      await addTrackToPlaylist(targetPlaylistId, addToPlaylistTrack.id);
      message.success("已添加到歌单");
      setAddToPlaylistOpen(false);
      setAddToPlaylistTrack(null);
      adjustPlaylistCount(targetPlaylistId, 1);
    } catch (error) {
      notifyError(error, "添加失败");
    } finally {
      setAddToPlaylistLoading(false);
    }
  }, [addToPlaylistTrack, adjustPlaylistCount]);

  // --- 元数据 ---
  const handleLookupMetadata = useCallback(async (track) => {
    setMetadataTrack(track);
    setMetadataOpen(true);
    setMetadataCandidates([]);
    metadataForm.setFieldsValue({
      title: track.title,
      artist: track.artist,
      album: track.album,
      year: track.year || track.release_date,
      genre: track.genre,
      cover_url: track.cover_url,
    });
    setMetadataLoading(true);
    try {
      const candidates = await lookupTrackMetadata(track.id);
      setMetadataCandidates(candidates || []);
      if (!candidates?.length) message.info("未找到候选元数据");
    } catch (error) {
      notifyError(error, "查询元数据失败");
    } finally {
      setMetadataLoading(false);
    }
  }, [metadataForm]);

  // 合并 handleApplyMetadata 和 handleApplyManualMetadata
  const applyMetadata = useCallback(async (candidate, isManual = false) => {
    if (!metadataTrack) return;
    try {
      const payload = isManual
        ? { source: "manual", external_id: "", ...candidate }
        : candidate;
      const updatedTrack = await applyTrackMetadata(metadataTrack.id, payload);
      mergeTrack(updatedTrack);
      message.success("元数据已更新");
      setMetadataOpen(false);
    } catch (error) {
      notifyError(error, "更新元数据失败");
    }
  }, [metadataTrack, mergeTrack]);

  const applyBestMetadata = useCallback(async (track) => {
    const candidates = await lookupTrackMetadata(track.id);
    const bestCandidate = pickMetadataCandidate(candidates, track);
    if (!bestCandidate) return false;
    const updatedTrack = await applyTrackMetadata(track.id, bestCandidate);
    mergeTrack(updatedTrack);
    return true;
  }, [mergeTrack]);

  const handleAutoMetadata = useCallback(async (track) => {
    setMetadataLoading(true);
    try {
      const changed = await applyBestMetadata(track);
      message[changed ? "success" : "info"](changed ? "元数据已自动应用" : "未找到可应用的元数据");
    } catch (error) {
      notifyError(error, "自动补全失败");
    } finally {
      setMetadataLoading(false);
    }
  }, [applyBestMetadata]);

  // 并发批量处理元数据，限制并发数为 3
  const handleAutoMetadataForList = useCallback(async () => {
    if (!tracks.length) return;
    setAutoMetadataLoading(true);
    let updatedCount = 0;
    const CONCURRENCY = 3;

    const runBatch = async (batch) => {
      const results = await Promise.allSettled(batch.map((track) => applyBestMetadata(track)));
      return results.filter((r) => r.status === "fulfilled" && r.value).length;
    };

    try {
      for (let i = 0; i < tracks.length; i += CONCURRENCY) {
        const batch = tracks.slice(i, i + CONCURRENCY);
        updatedCount += await runBatch(batch);
      }
      message[updatedCount ? "success" : "info"](
        updatedCount ? `已补全 ${updatedCount} 首曲目元数据` : "没有找到可应用的元数据"
      );
    } catch (error) {
      notifyError(error, "批量补全失败");
    } finally {
      setAutoMetadataLoading(false);
    }
  }, [tracks, applyBestMetadata]);

  const handleFetchLyrics = useCallback(async (track = currentTrack) => {
    if (!track?.id) return;
    setLyricsLoading(true);
    try {
      const updatedTrack = await fetchTrackLyrics(track.id);
      mergeTrack(updatedTrack);
      message.success(updatedTrack?.lyrics_type === "lrc" ? "已拉取同步歌词" : "已拉取歌词");
    } catch (error) {
      notifyError(error, "歌词拉取失败");
    } finally {
      setLyricsLoading(false);
    }
  }, [currentTrack, mergeTrack]);

  // --- 计算属性 ---
  const totalSize = useMemo(() => tracks.reduce((sum, item) => sum + (item.size_bytes || 0), 0), [tracks]);
  const playlistOptions = useMemo(() => [
    { value: ALL_TRACKS_PLAYLIST, label: "全部歌曲" },
    ...playlists.map((item) => ({
      value: item.id,
      label: `${item.name}${Number.isFinite(item.track_count) ? ` (${item.track_count})` : ""}`,
    })),
  ], [playlists]);
  const pageStart = (trackPage - 1) * trackPageSize;
  const pagedTracks = useMemo(
    () => tracks.slice(pageStart, pageStart + trackPageSize),
    [tracks, pageStart, trackPageSize],
  );

  // 可添加到的歌单列表（排除当前正在查看的歌单）
  const addablePlaylistOptions = useMemo(
    () => playlists.filter((item) => item.id !== playlistId),
    [playlists, playlistId],
  );

  const coverNode = useMemo(() => {
    if (currentTrack?.cover_url) {
      return <Image className="music-cover" src={currentTrack.cover_url} preview={false} fallback="data:image/gif;base64,R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7" />;
    }
    return (
      <div className="music-cover music-cover-placeholder">
        <CustomerServiceOutlined style={{ fontSize: 48, opacity: 0.4 }} />
      </div>
    );
  }, [currentTrack?.cover_url]);

  // --- 构建曲目操作菜单 ---
  // 使用 useMemo 缓存每首曲目的菜单配置，避免每次渲染都重建导致 Dropdown 重渲染
  const trackMenuMap = useMemo(() => {
    const map = new Map();
    for (const item of tracks) {
      const menuItems = [
        {
          key: "metadata",
          icon: MENU_ICON_SEARCH,
          label: "查询元数据",
          onClick: () => handleLookupMetadata(item),
        },
        {
          key: "addToPlaylist",
          icon: MENU_ICON_FOLDER_ADD,
          label: "添加到歌单",
          onClick: () => handleOpenAddToPlaylist(item),
          disabled: !playlists.length,
        },
      ];
      if (isPlaylistView) {
        menuItems.push({
          key: "removeFromPlaylist",
          icon: MENU_ICON_MINUS_CIRCLE,
          label: "从歌单移除",
          onClick: () => handleRemoveFromPlaylist(item.id),
        });
      }
      menuItems.push({ type: "divider" });
      menuItems.push({
        key: "delete",
        icon: MENU_ICON_DELETE,
        label: "彻底删除",
        danger: true,
      });
      map.set(item.id, menuItems);
    }
    return map;
  }, [tracks, isPlaylistView, playlists.length, handleLookupMetadata, handleOpenAddToPlaylist, handleRemoveFromPlaylist]);

  // --- 渲染 ---
  return (
    <div className="page-grid music-page">
      <Row gutter={[18, 18]} className="music-layout-row">
        {/* 播放器面板 */}
        <Col xs={24} xl={15}>
          <Card className="glass-card music-player-card" bordered={false}>
            <div className="music-player-grid">
              <div className="music-now music-studio">
                <div className="music-focus-stage">
                  <div className="music-cover-column">
                    <div className="music-cover-frame">
                      {coverNode}
                      <span className={`music-play-state${playing ? " playing" : ""}`}>{playing ? "Playing" : "Ready"}</span>
                    </div>
                    <div className="music-title-stack">
                      <div className="music-track-kicker">
                        {currentTrack?.lyrics ? <Tag color="purple">{currentTrack.lyrics_type === "lrc" ? "同步歌词" : "歌词"}</Tag> : <Tag>未匹配歌词</Tag>}
                      </div>
                      <h2 ref={titleRef}>{currentTrack?.title || "选择歌单播放"}</h2>
                      <p>{
                        audioError
                          ? <span style={{ color: "#ff4d4f" }}>{audioError}</span>
                          : currentTrack
                            ? trackMetaLine(currentTrack)
                            : "从右侧选择歌单开始播放"
                      }</p>
                    </div>
                  </div>

                  <LyricsPanel track={currentTrack} currentTime={currentTime} loading={lyricsLoading} onFetch={() => handleFetchLyrics()} />
                </div>

                <div className="music-console">
                  <div className="music-console-top">
                    <div className="music-console-summary">
                      <span>曲目</span>
                      <AutoScrollText text={currentTrack?.title || "未选择"} />
                    </div>
                    <div className="music-transport">
                      <Button shape="circle" icon={<BackwardOutlined />} onClick={prevTrack} disabled={!tracks.length} />
                      <Button
                        type="primary"
                        shape="circle"
                        size="large"
                        icon={playing ? <PauseCircleOutlined /> : <PlayCircleOutlined />}
                        onClick={togglePlay}
                        disabled={!currentTrack}
                      />
                      <Button shape="circle" icon={<ForwardOutlined />} onClick={nextTrack} disabled={!tracks.length} />
                    </div>
                    <div className="music-console-actions">
                      <Segmented
                        className="music-loop-switch"
                        value={loopMode}
                        onChange={setLoopMode}
                        options={[
                          { label: "顺序", value: "none" },
                          { label: "列表", value: "list" },
                          { label: "单曲", value: "single" },
                        ]}
                      />
                      <div className={`music-volume-pop ${volumeOpen ? "open" : ""}`}>
                        <Button
                          shape="circle"
                          icon={<SoundOutlined />}
                          onClick={() => setVolumeOpen((value) => !value)}
                          aria-label="音量"
                        />
                        {volumeOpen && (
                          <div className="music-volume-panel">
                            <span>{volume}</span>
                            <Slider vertical value={volume} onChange={setVolume} min={0} max={100} tooltip={{ formatter: null }} />
                          </div>
                        )}
                      </div>
                    </div>
                  </div>

                  <div className="music-progress-panel" aria-label="播放进度">
                    <span>{formatSeconds(currentTime)}</span>
                    <Slider
                      value={percent}
                      onChange={seek}
                      disabled={!currentTrack}
                      tooltip={{ formatter: null }}
                    />
                    <span>{formatSeconds(duration)}</span>
                  </div>
                </div>
              </div>
            </div>
          </Card>
        </Col>

        {/* 歌单管理面板 */}
        <Col xs={24} xl={9}>
          <Card
            className="glass-card music-library-card"
            title={(
              <div className="music-library-title">
                <span>服务器歌单</span>
                <small>{tracks.length} 首 · {formatFileSize(totalSize)} · 自动 {trackPageSize} 首/页</small>
              </div>
            )}
            bordered={false}
          >
            <div className="music-library-toolbar">
              <Select
                value={playlistId}
                onChange={setPlaylistId}
                placeholder={playlistError || "选择歌单"}
                status={playlistError ? "error" : undefined}
                className="music-playlist-select"
                options={playlistOptions}
                loading={playlistLoading}
                notFoundContent={playlistLoading ? <Spin size="small" /> : <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无歌单" />}
              />
              <div className="music-library-actions">
                <Tooltip title="新建歌单">
                  <Button icon={<PlusOutlined />} onClick={() => setPlaylistOpen(true)} />
                </Tooltip>
                <Upload beforeUpload={beforeUpload} accept={acceptTypes} multiple showUploadList={false}>
                  <Tooltip title={isPlaylistView ? "上传到当前歌单" : "上传到服务器曲库"}>
                    <Button icon={<UploadOutlined />} disabled={!playlistId} />
                  </Tooltip>
                </Upload>
                <Tooltip title="补全当前歌单元数据">
                  <Button
                    icon={<ReloadOutlined />}
                    disabled={!tracks.length}
                    loading={autoMetadataLoading}
                    onClick={handleAutoMetadataForList}
                  />
                </Tooltip>
                <Popconfirm
                  title="确定删除歌单？"
                  description="歌单内曲目不会从服务器删除。"
                  onConfirm={handleDeletePlaylist}
                  okText="删除"
                  cancelText="取消"
                  disabled={!isPlaylistView}
                >
                  <Tooltip title="删除当前歌单">
                    <Button danger icon={<DeleteOutlined />} disabled={!isPlaylistView} />
                  </Tooltip>
                </Popconfirm>
              </div>
            </div>

            {playlistError && (
              <div className="music-error-row">
                <span>歌单加载失败：{playlistError}</span>
                <Button type="link" size="small" onClick={fetchPlaylists}>重试</Button>
              </div>
            )}

            {tracksError && (
              <div className="music-error-row">
                <span>{tracksError}</span>
                <Button type="link" size="small" onClick={() => fetchTracks(playlistId)}>重试</Button>
              </div>
            )}

            <div className="music-track-list-shell" ref={trackListRef}>
              <Spin spinning={tracksLoading}>
                <List
                  className="music-track-list"
                  dataSource={pagedTracks}
                  locale={{ emptyText: <Empty description={tracksError ? "曲目加载失败" : "当前歌单还没有音乐"} /> }}
                  renderItem={(item, index) => {
                    const trackIndex = pageStart + index;
                    const isActive = trackIndex === currentIndex;
                    const menuItems = trackMenuMap.get(item.id) || [];
                    return (
                      <List.Item
                        className={`music-track${isActive ? " active" : ""}`}
                        onClick={() => playAt(trackIndex)}
                      >
                        <div className="music-track-row">
                          <TrackCover track={item} size="large" />
                          <div className="music-track-info">
                            <div className="music-track-title">
                              {isActive && playing && <span className="music-playing-dot" />}
                              <span>{item.title || "未知曲目"}</span>
                            </div>
                            <div className="music-track-meta">
                              {item.artist || "未知艺术家"}
                              {item.album ? ` · ${item.album}` : ""}
                            </div>
                            <div className="music-track-sub">
                              {item.duration_sec ? formatSeconds(item.duration_sec) : ""}
                              {item.duration_sec && item.size_bytes ? " · " : ""}
                              {item.size_bytes ? formatFileSize(item.size_bytes) : ""}
                              {(item.year || item.genre) ? " · " : ""}
                              {[item.year, item.genre].filter(Boolean).join(" · ")}
                            </div>
                          </div>
                          <div className="music-track-actions" onClick={(e) => e.stopPropagation()}>
                            <Dropdown
                              menu={{
                                items: menuItems,
                                onClick: ({ key, domEvent }) => {
                                  domEvent.stopPropagation();
                                  if (key === "delete") {
                                    // 由 Popconfirm 处理，不在这里直接执行
                                  }
                                },
                              }}
                              trigger={["click"]}
                              dropdownRender={(menu) => (
                                <div>
                                  {/* 渲染除了 delete 之外的菜单项 */}
                                  {menu}
                                </div>
                              )}
                            >
                              <Button type="text" icon={<EllipsisOutlined />} title="更多操作" />
                            </Dropdown>
                            <Popconfirm
                              title="彻底删除曲目"
                              description="将从服务器永久删除该文件，且会从所有歌单中移除。"
                              onConfirm={() => handleDeleteTrack(item.id)}
                              okText="删除"
                              cancelText="取消"
                              okButtonProps={{ danger: true }}
                            >
                              <Tooltip title="彻底删除">
                                <Button type="text" danger icon={<DeleteOutlined />} size="small" />
                              </Tooltip>
                            </Popconfirm>
                          </div>
                        </div>
                      </List.Item>
                    );
                  }}
                />
              </Spin>
            </div>
            {tracks.length > trackPageSize && (
              <Pagination
                className="music-track-pagination"
                current={trackPage}
                pageSize={trackPageSize}
                total={tracks.length}
                showSizeChanger={false}
                size="small"
                showTotal={(total, range) => `${range[0]}-${range[1]} / ${total}`}
                onChange={setTrackPage}
              />
            )}
          </Card>
        </Col>
      </Row>

      {/* 新建歌单弹窗 */}
      <Modal
        title="新建歌单"
        open={playlistOpen}
        onCancel={() => setPlaylistOpen(false)}
        onOk={handleCreatePlaylist}
        destroyOnClose
      >
        <Form form={playlistForm} layout="vertical">
          <Form.Item
            name="name"
            label="歌单名称"
            rules={[{ required: true, message: "请输入歌单名称" }, { max: 160, message: "名称过长" }]}
          >
            <Input placeholder="例如：我的收藏" />
          </Form.Item>
          <Form.Item name="description" label="说明">
            <Input.TextArea rows={3} placeholder="可选描述" />
          </Form.Item>
        </Form>
      </Modal>

      {/* 添加到歌单弹窗 */}
      <Modal
        title={`添加「${addToPlaylistTrack?.title || "曲目"}」到歌单`}
        open={addToPlaylistOpen}
        onCancel={() => { setAddToPlaylistOpen(false); setAddToPlaylistTrack(null); }}
        footer={null}
        destroyOnClose
      >
        {addablePlaylistOptions.length === 0 ? (
          <Empty description="暂无可添加的歌单，请先创建歌单" />
        ) : (
          <List
            dataSource={addablePlaylistOptions}
            loading={addToPlaylistLoading}
            renderItem={(pl) => (
              <List.Item
                className="music-add-playlist-item"
                actions={[
                  <Button
                    key="add"
                    type="primary"
                    size="small"
                    loading={addToPlaylistLoading}
                    onClick={() => handleAddToPlaylist(pl.id)}
                  >
                    添加
                  </Button>,
                ]}
              >
                <List.Item.Meta
                  title={pl.name}
                  description={pl.description || `${pl.track_count ?? 0} 首曲目`}
                />
              </List.Item>
            )}
          />
        )}
      </Modal>

      {/* 元数据弹窗 */}
      <Modal
        title="在线查询元数据"
        open={metadataOpen}
        onCancel={() => setMetadataOpen(false)}
        footer={[
          <Button key="cancel" onClick={() => setMetadataOpen(false)}>取消</Button>,
          <Button key="manual" type="primary" onClick={async () => {
            try {
              const values = await metadataForm.validateFields();
              await applyMetadata(values, true);
            } catch {
              // 表单校验失败，不关闭弹窗
            }
          }}>保存手动编辑</Button>,
        ]}
        width={720}
        destroyOnClose
      >
        <Form form={metadataForm} layout="vertical" className="music-metadata-form">
          <Form.Item name="title" label="曲名" rules={[{ required: true, message: "请输入曲名" }]}>
            <Input />
          </Form.Item>
          <Form.Item name="artist" label="艺术家">
            <Input />
          </Form.Item>
          <Form.Item name="album" label="专辑">
            <Input />
          </Form.Item>
          <Row gutter={12}>
            <Col span={12}>
              <Form.Item name="year" label="年份">
                <Input />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="genre" label="流派">
                <Input />
              </Form.Item>
            </Col>
          </Row>
          <Form.Item name="cover_url" label="封面 URL">
            <Input />
          </Form.Item>
        </Form>
        <List
          header="候选结果"
          loading={metadataLoading}
          dataSource={metadataCandidates}
          locale={{ emptyText: <Empty description="暂无候选结果，建议手动编辑" /> }}
          renderItem={(item) => (
            <List.Item
              actions={[
                <Button key="apply" type="primary" onClick={() => applyMetadata(item)}>应用</Button>,
              ]}
            >
              <List.Item.Meta
                avatar={item.cover_url ? <Image width={56} height={56} src={item.cover_url} preview={false} /> : null}
                title={`${item.title || "未知曲目"} · ${item.artist || "未知艺术家"}`}
                description={[item.album, item.year || item.release_date, item.genre, item.source].filter(Boolean).join(" · ")}
              />
            </List.Item>
          )}
        />
        {metadataTrack && (
          <Button
            block
            icon={<ReloadOutlined />}
            loading={metadataLoading}
            onClick={() => handleAutoMetadata(metadataTrack)}
          >
            自动应用最佳匹配
          </Button>
        )}
      </Modal>
    </div>
  );
}

// --- 辅助函数与子组件 ---

function pickMetadataCandidate(candidates = [], track = {}) {
  const applicableCandidates = candidates.filter((item) => item.source !== "local");
  const scored = applicableCandidates
    .map((item) => ({ item, score: metadataScore(item, track) }))
    .sort((a, b) => b.score - a.score);
  return scored[0]?.score >= 60 ? scored[0].item : null;
}

function metadataScore(candidate, track) {
  if (candidate.source === "filename" && (candidate.title || candidate.artist)) return 85;
  const sourceText = normalizeMetaText([track.original_name, track.title, track.artist].filter(Boolean).join(" "));
  const candidateTitle = normalizeMetaText(candidate.title);
  const candidateArtist = normalizeMetaText(candidate.artist);
  const candidateAlbum = normalizeMetaText(candidate.album);
  let score = 0;
  if (candidateTitle && sourceText.includes(candidateTitle)) score += 50;
  if (candidateArtist && sourceText.includes(candidateArtist)) score += 30;
  if (candidateAlbum && sourceText.includes(candidateAlbum)) score += 15;
  if (candidate.cover_url) score += 5;
  return score;
}

function normalizeMetaText(value = "") {
  return value.toLowerCase().replace(/\.[a-z0-9]+$/i, "").replace(/[【】[\]()（）「」『』_-]+/g, " ").replace(/\s+/g, " ").trim();
}

function formatSeconds(seconds) {
  if (!Number.isFinite(seconds) || seconds < 0) return "00:00";
  const minutes = Math.floor(seconds / 60);
  const rest = Math.floor(seconds % 60);
  return `${String(minutes).padStart(2, "0")}:${String(rest).padStart(2, "0")}`;
}

function formatFileSize(size) {
  if (!size) return "0 MB";
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`;
  return `${(size / 1024 / 1024).toFixed(1)} MB`;
}

function trackMetaLine(track) {
  return [
    track.artist || "未知艺术家",
    track.album,
    track.year,
    track.genre,
    track.duration_sec ? formatSeconds(track.duration_sec) : "",
  ].filter(Boolean).join(" · ");
}

function loopModeLabel(mode) {
  if (mode === "single") return "单曲循环";
  if (mode === "list") return "列表循环";
  return "顺序播放";
}

function lyricsSourceLabel(source) {
  if (source === "netease") return "网易云音乐";
  return source || "本地歌词";
}

const TrackCover = memo(function TrackCover({ track, size = "small" }) {
  const cls = size === "large" ? "music-track-cover-lg" : "music-track-cover";
  if (track.cover_url) {
    return <Image className={cls} src={track.cover_url} preview={false} fallback="data:image/gif;base64,R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7" />;
  }
  return (
    <div className={`${cls} music-cover-placeholder`}>
      <CustomerServiceOutlined style={size === "large" ? { fontSize: 22 } : undefined} />
    </div>
  );
});

const AutoScrollText = memo(function AutoScrollText({ text }) {
  const shellRef = useRef(null);
  const textRef = useRef(null);
  const [overflow, setOverflow] = useState(false);

  useEffect(() => {
    const shell = shellRef.current;
    const node = textRef.current;
    if (!shell || !node) return undefined;

    const update = () => {
      const distance = Math.max(0, node.scrollWidth - shell.clientWidth);
      shell.style.setProperty("--scroll-distance", `${distance}px`);
      setOverflow(distance > 1);
    };

    update();
    const observer = new ResizeObserver(update);
    observer.observe(shell);
    observer.observe(node);
    return () => observer.disconnect();
  }, [text]);

  return (
    <strong ref={shellRef} className={`music-scroll-title${overflow ? " is-overflow" : ""}`}>
      <span ref={textRef}>{text}</span>
    </strong>
  );
});

const LyricsPanel = memo(function LyricsPanel({ track, currentTime = 0, loading, onFetch }) {
  const scrollRef = useRef(null);
  const [visibleLineCount, setVisibleLineCount] = useState(7);
  const lines = useMemo(() => (track?.lyrics ? parseLyrics(track.lyrics, track.lyrics_type) : []), [track?.lyrics, track?.lyrics_type]);
  const activeIndex = useMemo(() => activeLyricIndex(lines, currentTime), [lines, currentTime]);
  const visibleLines = useMemo(
    () => lyricWindow(lines, activeIndex, track?.lyrics_type, visibleLineCount),
    [lines, activeIndex, track?.lyrics_type, visibleLineCount],
  );

  useEffect(() => {
    const node = scrollRef.current;
    if (!node) return undefined;

    const updateVisibleLineCount = () => {
      const height = node.clientHeight;
      if (!height) return;
      const sample = node.querySelector("p");
      const sampleHeight = sample?.getBoundingClientRect().height || Number.parseFloat(getComputedStyle(node).lineHeight) || 44;
      const nextCount = Math.max(MIN_VISIBLE_LYRIC_LINES, Math.floor(height / sampleHeight));
      setVisibleLineCount((prev) => (prev === nextCount ? prev : nextCount));
    };

    updateVisibleLineCount();
    const observer = new ResizeObserver(updateVisibleLineCount);
    observer.observe(node);
    return () => observer.disconnect();
  }, []);

  if (!track) {
    return (
      <div className="music-lyrics-panel empty">
        <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="选择曲目后查看歌词" />
      </div>
    );
  }

  if (!track.lyrics) {
    return (
      <div className="music-lyrics-panel empty">
        <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="当前曲目还没有歌词" />
        <Button icon={<ReadOutlined />} loading={loading} onClick={onFetch}>从网易云拉取歌词</Button>
      </div>
    );
  }

  return (
    <div className="music-lyrics-panel">
      <div className="music-lyrics-head">
        <span>{lyricsSourceLabel(track.lyrics_source)}</span>
        <Tag color={track.lyrics_type === "lrc" ? "purple" : "default"}>{track.lyrics_type === "lrc" ? "同步" : "纯文本"}</Tag>
      </div>
      <div className="music-lyrics-scroll" ref={scrollRef} aria-live="polite">
        {visibleLines.map(({ line, index }) => (
          <p key={`${line.time ?? "plain"}-${index}`} data-lyric-index={index} className={index === activeIndex ? "active" : ""}>
            {line.text || "♪"}
          </p>
        ))}
      </div>
    </div>
  );
});

function parseLyrics(value = "", type = "plain") {
  const rawLines = value.split(/\r?\n/).map((line) => line.trim()).filter(Boolean);
  if (type !== "lrc") {
    return rawLines.map((text) => ({ text }));
  }
  const timeRe = /^\[(\d{1,2}):(\d{2})(?:\.(\d{1,3}))?]\s*(.*)$/;
  return rawLines.map((line) => {
    const match = line.match(timeRe);
    if (!match) return { text: line };
    const minutes = Number(match[1] || 0);
    const seconds = Number(match[2] || 0);
    const millis = Number((match[3] || "0").padEnd(3, "0"));
    return { time: minutes * 60 + seconds + millis / 1000, text: match[4] || "" };
  }).filter((line) => line.text || Number.isFinite(line.time));
}

function activeLyricIndex(lines, currentTime) {
  let active = -1;
  for (let index = 0; index < lines.length; index += 1) {
    if (!Number.isFinite(lines[index].time)) continue;
    if (lines[index].time <= currentTime) active = index;
    if (lines[index].time > currentTime) break;
  }
  return active;
}

function lyricWindow(lines, activeIndex, type = "plain", visibleLineCount = MIN_VISIBLE_LYRIC_LINES) {
  if (!lines.length) return [];
  const count = Math.max(MIN_VISIBLE_LYRIC_LINES, visibleLineCount);
  if (type !== "lrc") {
    return lines.slice(0, count).map((line, index) => ({ line, index }));
  }
  const center = activeIndex >= 0 ? activeIndex : 0;
  const before = Math.floor((count - 1) / 2);
  const after = count - before - 1;
  let start = Math.max(0, center - before);
  let end = Math.min(lines.length, center + after + 1);
  if (end - start < count) {
    start = Math.max(0, end - count);
    end = Math.min(lines.length, start + count);
  }
  return lines.slice(start, end).map((line, offset) => ({ line, index: start + offset }));
}

export default MusicPlayer;
