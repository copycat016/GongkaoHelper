import { createContext, useCallback, useContext, useEffect, useMemo, useRef, useState } from "react";

const MusicContext = createContext(null);

function getPathFromSrc(src) {
  try {
    return new URL(src, window.location.origin).pathname;
  } catch {
    return src;
  }
}

export function MusicProvider({ children }) {
  const audioRef = useRef(null);
  const lastTimeUpdateRef = useRef(0);
  const [tracks, setTracks] = useState([]);
  const [currentIndex, setCurrentIndex] = useState(0);
  const [playing, setPlaying] = useState(false);
  const [loopMode, setLoopMode] = useState("list");
  const [volume, setVolume] = useState(70);
  const [currentTime, setCurrentTime] = useState(0);
  const [duration, setDuration] = useState(0);
  const [audioError, setAudioError] = useState(null);

  const safeIndex = Math.min(currentIndex, Math.max(tracks.length - 1, 0));
  const currentTrack = tracks[safeIndex];
  const percent = duration ? Math.round((currentTime / duration) * 100) : 0;

  useEffect(() => {
    if (currentIndex >= tracks.length && tracks.length > 0) {
      setCurrentIndex(0);
    }
  }, [tracks.length, currentIndex]);

  useEffect(() => {
    const audio = audioRef.current;
    if (audio) audio.volume = volume / 100;
  }, [volume]);

  useEffect(() => {
    const audio = audioRef.current;
    if (!audio || !currentTrack) return;

    const currentPath = getPathFromSrc(audio.src);
    const targetPath = currentTrack.public_url;

    if (currentPath === targetPath) {
      if (playing) {
        audio.play().catch((err) => {
          setAudioError(err.message);
          setPlaying(false);
        });
      }
      return;
    }

    setAudioError(null);
    audio.src = currentTrack.public_url;
    audio.load();
    setCurrentTime(0);
    if (playing) {
      audio.play().catch((err) => {
        setAudioError(err.message);
        setPlaying(false);
      });
    }
  }, [currentTrack, playing]);

  const play = useCallback(() => {
    const audio = audioRef.current;
    if (!audio || !currentTrack) return;
    setAudioError(null);
    audio.play().then(() => setPlaying(true)).catch((err) => {
      setAudioError(err.message);
      setPlaying(false);
    });
  }, [currentTrack]);

  const pause = useCallback(() => {
    audioRef.current?.pause();
    setPlaying(false);
  }, []);

  const togglePlay = useCallback(() => {
    if (playing) pause();
    else play();
  }, [playing, pause, play]);

  const playAt = useCallback((index) => {
    if (!tracks[index]) return;
    if (index === safeIndex && playing) {
      pause();
      return;
    }
    setCurrentIndex(index);
    setPlaying(true);
  }, [tracks, safeIndex, playing, pause]);

  const nextTrack = useCallback(() => {
    if (!tracks.length) return;
    setCurrentIndex((prev) => (prev + 1) % tracks.length);
    setPlaying(true);
  }, [tracks.length]);

  const prevTrack = useCallback(() => {
    if (!tracks.length) return;
    setCurrentIndex((prev) => (prev - 1 + tracks.length) % tracks.length);
    setPlaying(true);
  }, [tracks.length]);

  const seek = useCallback((value) => {
    const audio = audioRef.current;
    if (!audio || !duration) return;
    const nextTime = (value / 100) * duration;
    audio.currentTime = nextTime;
    setCurrentTime(nextTime);
  }, [duration]);

  const handleEnded = useCallback(() => {
    if (loopMode === "single") {
      const audio = audioRef.current;
      if (audio) {
        audio.currentTime = 0;
        audio.play().catch(() => setPlaying(false));
      }
      return;
    }
    if (currentIndex < tracks.length - 1) {
      setCurrentIndex((prev) => prev + 1);
      setPlaying(true);
      return;
    }
    if (loopMode === "list" && tracks.length > 0) {
      setCurrentIndex(0);
      setPlaying(true);
      return;
    }
    setPlaying(false);
  }, [loopMode, currentIndex, tracks.length]);

  const handleError = useCallback(() => {
    setAudioError("音频加载失败，文件可能不存在或格式不支持");
    setPlaying(false);
  }, []);

  const handleTimeUpdate = useCallback((event) => {
    const nextTime = event.currentTarget.currentTime;
    const now = performance.now();
    setCurrentTime((prev) => {
      if (now - lastTimeUpdateRef.current < 250 && Math.abs(nextTime - prev) < 0.5) {
        return prev;
      }
      lastTimeUpdateRef.current = now;
      return nextTime;
    });
  }, []);

  const value = useMemo(() => ({
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
    audioError,
    play,
    pause,
    togglePlay,
    playAt,
    nextTrack,
    prevTrack,
    seek,
  }), [tracks, currentIndex, currentTrack, playing, loopMode, volume, currentTime, duration, percent, audioError, play, pause, togglePlay, playAt, nextTrack, prevTrack, seek]);

  return (
    <MusicContext.Provider value={value}>
      {children}
      <audio
        ref={audioRef}
        onTimeUpdate={handleTimeUpdate}
        onLoadedMetadata={(event) => setDuration(event.currentTarget.duration || 0)}
        onEnded={handleEnded}
        onError={handleError}
      />
    </MusicContext.Provider>
  );
}

export function useMusic() {
  const context = useContext(MusicContext);
  if (!context) {
    throw new Error("useMusic must be used inside MusicProvider");
  }
  return context;
}
