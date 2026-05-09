import { createContext, useContext, useEffect, useMemo, useRef, useState } from "react";

const MusicContext = createContext(null);

export function MusicProvider({ children }) {
  const audioRef = useRef(null);
  const [tracks, setTracks] = useState([]);
  const [currentIndex, setCurrentIndex] = useState(0);
  const [playing, setPlaying] = useState(false);
  const [loopMode, setLoopMode] = useState("list");
  const [volume, setVolume] = useState(70);
  const [currentTime, setCurrentTime] = useState(0);
  const [duration, setDuration] = useState(0);

  const currentTrack = tracks[currentIndex];
  const percent = duration ? Math.round((currentTime / duration) * 100) : 0;

  useEffect(() => {
    const audio = audioRef.current;
    if (audio) audio.volume = volume / 100;
  }, [volume]);

  useEffect(() => {
    const audio = audioRef.current;
    if (!audio || !currentTrack) return;
    if (audio.src.endsWith(currentTrack.public_url)) {
      if (playing) audio.play().catch(() => setPlaying(false));
      return;
    }
    audio.src = currentTrack.public_url;
    audio.load();
    setCurrentTime(0);
    if (playing) audio.play().catch(() => setPlaying(false));
  }, [currentTrack, playing]);

  const play = () => {
    const audio = audioRef.current;
    if (!audio || !currentTrack) return;
    audio.play().then(() => setPlaying(true)).catch(() => setPlaying(false));
  };

  const pause = () => {
    audioRef.current?.pause();
    setPlaying(false);
  };

  const togglePlay = () => {
    if (playing) pause();
    else play();
  };

  const playAt = (index) => {
    if (!tracks[index]) return;
    setCurrentIndex(index);
    setPlaying(true);
  };

  const nextTrack = () => {
    if (!tracks.length) return;
    setCurrentIndex((prev) => (prev + 1) % tracks.length);
    setPlaying(true);
  };

  const prevTrack = () => {
    if (!tracks.length) return;
    setCurrentIndex((prev) => (prev - 1 + tracks.length) % tracks.length);
    setPlaying(true);
  };

  const seek = (value) => {
    const audio = audioRef.current;
    if (!audio || !duration) return;
    const nextTime = (value / 100) * duration;
    audio.currentTime = nextTime;
    setCurrentTime(nextTime);
  };

  const handleEnded = () => {
    if (loopMode === "single") {
      audioRef.current.currentTime = 0;
      audioRef.current.play();
      return;
    }
    if (currentIndex < tracks.length - 1) {
      setCurrentIndex((prev) => prev + 1);
      return;
    }
    if (loopMode === "list" && tracks.length > 0) {
      setCurrentIndex(0);
      return;
    }
    setPlaying(false);
  };

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
    play,
    pause,
    togglePlay,
    playAt,
    nextTrack,
    prevTrack,
    seek,
  }), [tracks, currentIndex, currentTrack, playing, loopMode, volume, currentTime, duration, percent]);

  return (
    <MusicContext.Provider value={value}>
      {children}
      <audio
        ref={audioRef}
        onTimeUpdate={(event) => setCurrentTime(event.currentTarget.currentTime)}
        onLoadedMetadata={(event) => setDuration(event.currentTarget.duration || 0)}
        onEnded={handleEnded}
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
