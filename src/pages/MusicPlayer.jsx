import { useEffect, useMemo, useState } from "react";
import { Button, Card, Col, Form, Input, List, Modal, Progress, Row, Segmented, Select, Slider, Space, Tag, Upload, message } from "antd";
import { BackwardOutlined, ForwardOutlined, PauseCircleOutlined, PlayCircleOutlined, PlusOutlined, SoundOutlined, UploadOutlined } from "@ant-design/icons";
import { createPlaylist, getPlaylistTracks, getPlaylists, getTracks, uploadTrack } from "../api/music";
import { useMusic } from "../components/MusicProvider";

const acceptTypes = ".mp3,.flac,.wav,.ogg,.m4a,.aac,audio/*";

function MusicPlayer() {
  const [playlists, setPlaylists] = useState([]);
  const [playlistId, setPlaylistId] = useState();
  const {
    tracks,
    setTracks,
    currentIndex,
    setCurrentIndex,
    currentTrack,
    playing,
    loopMode,
    setLoopMode,
    volume,
    setVolume,
    currentTime,
    duration,
    percent,
    togglePlay,
    playAt,
    nextTrack,
    prevTrack,
    seek,
  } = useMusic();
  const [playlistOpen, setPlaylistOpen] = useState(false);
  const [playlistForm] = Form.useForm();

  const loadPlaylists = async () => {
    const items = await getPlaylists();
    setPlaylists(items || []);
    if (!playlistId && items?.[0]) setPlaylistId(items[0].id);
  };

  const loadTracks = async (nextPlaylistId = playlistId) => {
    const items = nextPlaylistId ? await getPlaylistTracks(nextPlaylistId) : await getTracks();
    setTracks(items || []);
    setCurrentIndex(0);
  };

  useEffect(() => {
    loadPlaylists().catch(() => {});
    loadTracks().catch(() => {});
  }, []);

  useEffect(() => {
    loadTracks(playlistId).catch(() => {});
  }, [playlistId]);

  const beforeUpload = async (file) => {
    try {
      await uploadTrack({ file, playlistId });
      message.success("音乐已上传到服务器");
      await loadTracks(playlistId);
    } catch (error) {
      message.error(error.message || "上传失败");
    }
    return false;
  };

  const handleCreatePlaylist = async () => {
    const values = await playlistForm.validateFields();
    const playlist = await createPlaylist({ ...values, owner_role: "root", enabled: true });
    message.success("歌单已创建");
    setPlaylistOpen(false);
    playlistForm.resetFields();
    await loadPlaylists();
    setPlaylistId(playlist.id);
  };

  const totalSize = useMemo(() => tracks.reduce((sum, item) => sum + item.size_bytes, 0), [tracks]);

  return (
    <div className="page-grid">
      <Row gutter={[16, 16]}>
        <Col xs={24} lg={14}>
          <Card className="glass-card music-player-card" bordered={false}>
            <div className="music-now">
              <Tag>{tracks.length ? `${tracks.length} 首` : "空歌单"}</Tag>
              <h2>{currentTrack?.title || "选择服务器歌单"}</h2>
              <p>{currentTrack ? trackMetaLine(currentTrack) : "由 admin/root 上传维护，当前阶段不做权限限制。"}</p>
            </div>
            <Progress percent={percent} showInfo={false} strokeWidth={8} />
            <Slider value={percent} onChange={seek} tooltip={{ formatter: null }} />
            <div className="music-time-row"><span>{formatSeconds(currentTime)}</span><span>{formatSeconds(duration)}</span></div>
            <Space wrap className="music-controls">
              <Button shape="circle" icon={<BackwardOutlined />} onClick={prevTrack} />
              <Button type="primary" shape="circle" size="large" icon={playing ? <PauseCircleOutlined /> : <PlayCircleOutlined />} onClick={togglePlay} />
              <Button shape="circle" icon={<ForwardOutlined />} onClick={nextTrack} />
              <Segmented value={loopMode} onChange={setLoopMode} options={[{ label: "顺序", value: "none" }, { label: "列表循环", value: "list" }, { label: "单曲循环", value: "single" }]} />
            </Space>
            <div className="music-volume"><SoundOutlined /><Slider value={volume} onChange={setVolume} min={0} max={100} /></div>
          </Card>
        </Col>
        <Col xs={24} lg={10}>
          <Card className="glass-card" title="服务器歌单" extra={<Tag>{formatFileSize(totalSize)}</Tag>} bordered={false}>
            <Space.Compact style={{ width: "100%" }}>
              <Select value={playlistId} onChange={setPlaylistId} placeholder="选择歌单" style={{ width: "100%" }} options={playlists.map((item) => ({ value: item.id, label: item.name }))} />
              <Button icon={<PlusOutlined />} onClick={() => setPlaylistOpen(true)} />
            </Space.Compact>
            <Upload beforeUpload={beforeUpload} accept={acceptTypes} multiple showUploadList={false}>
              <Button block icon={<UploadOutlined />} style={{ marginTop: 12 }}>上传到当前歌单</Button>
            </Upload>
            <List className="music-track-list" dataSource={tracks} locale={{ emptyText: "当前歌单还没有音乐" }} renderItem={(item, index) => (
              <List.Item className={index === currentIndex ? "music-track active" : "music-track"} onClick={() => playAt(index)} actions={[<Button key="play" type="text" icon={index === currentIndex && playing ? <PauseCircleOutlined /> : <PlayCircleOutlined />} onClick={(event) => { event.stopPropagation(); playAt(index); }} />]}>
                <List.Item.Meta title={item.title} description={trackListDescription(item)} />
              </List.Item>
            )} />
          </Card>
        </Col>
      </Row>
      <Modal title="新建歌单" open={playlistOpen} onCancel={() => setPlaylistOpen(false)} onOk={handleCreatePlaylist}>
        <Form form={playlistForm} layout="vertical">
          <Form.Item name="name" label="歌单名称" rules={[{ required: true, message: "请输入歌单名称" }]}><Input /></Form.Item>
          <Form.Item name="description" label="说明"><Input.TextArea rows={3} /></Form.Item>
        </Form>
      </Modal>
    </div>
  );
}

function formatSeconds(seconds) {
  if (!Number.isFinite(seconds)) return "00:00";
  const minutes = Math.floor(seconds / 60);
  const rest = Math.floor(seconds % 60);
  return `${String(minutes).padStart(2, "0")}:${String(rest).padStart(2, "0")}`;
}

function formatFileSize(size) {
  if (!size) return "0 MB";
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

function trackListDescription(track) {
  return [
    track.artist || "未知艺术家",
    track.album,
    track.duration_sec ? formatSeconds(track.duration_sec) : "",
    formatFileSize(track.size_bytes),
  ].filter(Boolean).join(" · ");
}

export default MusicPlayer;
