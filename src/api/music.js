import { api } from "./request";

export const getPlaylists = () => api.get("/music/playlists");
export const createPlaylist = (data) => api.post("/music/playlists", data);
export const getTracks = () => api.get("/music/tracks");
export const getPlaylistTracks = (playlistId) => api.get(`/music/playlists/${playlistId}/tracks`);
export const addTrackToPlaylist = (playlistId, trackId) => api.post(`/music/playlists/${playlistId}/tracks/${trackId}`);

export async function uploadTrack({ file, playlistId }) {
  const form = new FormData();
  form.append("file", file);
  if (playlistId) form.append("playlist_id", String(playlistId));

  const response = await fetch("/api/music/tracks", {
    method: "POST",
    body: form,
  });
  const result = await response.json();
  if (!response.ok || result?.code > 0) {
    throw new Error(result?.message || "上传失败");
  }
  return result.data;
}
