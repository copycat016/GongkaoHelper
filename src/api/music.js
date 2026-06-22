import { api, authHeaders } from "./request";

export const getPlaylists = (options) => api.get("/music/playlists", undefined, options);
export const createPlaylist = (data) => api.post("/music/playlists", data);
export const deletePlaylist = (id) => api.delete(`/music/playlists/${id}`);
export const getTracks = (options) => api.get("/music/tracks", undefined, options);
export const getPlaylistTracks = (playlistId, options) => api.get(`/music/playlists/${playlistId}/tracks`, undefined, options);
export const addTrackToPlaylist = (playlistId, trackId) => api.post(`/music/playlists/${playlistId}/tracks/${trackId}`);
export const removeTrackFromPlaylist = (playlistId, trackId) => api.delete(`/music/playlists/${playlistId}/tracks/${trackId}`);
export const deleteTrack = (id) => api.delete(`/music/tracks/${id}`);
export const getTrackPlaylists = (trackId) => api.get(`/music/tracks/${trackId}/playlists`);
export const lookupTrackMetadata = (trackId) => api.post(`/music/tracks/${trackId}/metadata/lookup`, {});
export const applyTrackMetadata = (trackId, data) => api.put(`/music/tracks/${trackId}/metadata`, data);
export const fetchTrackLyrics = (trackId) => api.post(`/music/tracks/${trackId}/lyrics/fetch`, {});

export async function uploadTrack({ file, playlistId }) {
  const form = new FormData();
  form.append("file", file);
  if (playlistId) form.append("playlist_id", String(playlistId));

  const response = await fetch("/api/music/tracks", {
    method: "POST",
    headers: authHeaders(),
    body: form,
  });
  const result = await response.json();
  if (!response.ok || result?.code > 0) {
    throw new Error(result?.message || "上传失败");
  }
  return result.data;
}
