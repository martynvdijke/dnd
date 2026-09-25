/**
 * Ambient soundscapes synthesized with the Web Audio API.
 *
 * No audio asset files are shipped: each track is a filtered noise bed (plus an
 * optional tonal drone) generated at runtime, mirroring the approach used for
 * dice SFX in ts/dice.ts. Tracks are chosen by name, broadcast over the WS hub,
 * and played by every connected screen so the table shares one soundscape.
 */
import { expose } from './expose';

export type AmbienceTrack = 'rain' | 'fire' | 'wind' | 'tavern' | 'combat' | 'forest' | 'dungeon' | 'waves';

interface TrackSpec {
  label: string;
  icon: string;
  filter: BiquadFilterType;
  freq: number;
  q: number;
  gain: number;
  osc?: number;
  oscType?: OscillatorType;
  pulse?: number; // Hz amplitude LFO for rhythmic beds (drums, heartbeat)
}

const TRACKS: Record<AmbienceTrack, TrackSpec> = {
  rain:    { label: 'Rain',          icon: 'fa-cloud-rain', filter: 'lowpass',  freq: 1400, q: 0.7, gain: 0.10 },
  fire:    { label: 'Campfire',      icon: 'fa-fire',       filter: 'bandpass', freq: 600,  q: 0.6, gain: 0.12, osc: 110, oscType: 'sawtooth' },
  wind:    { label: 'Wind',          icon: 'fa-wind',       filter: 'lowpass',  freq: 500,  q: 0.4, gain: 0.11 },
  tavern:  { label: 'Tavern',        icon: 'fa-mug-hot',    filter: 'lowpass',  freq: 900,  q: 0.5, gain: 0.07, osc: 196, oscType: 'triangle' },
  combat:  { label: 'Combat Drums',  icon: 'fa-drum',       filter: 'bandpass', freq: 800,  q: 0.8, gain: 0.09, osc: 70,  oscType: 'square', pulse: 1.6 },
  forest:  { label: 'Forest',        icon: 'fa-tree',       filter: 'highpass', freq: 2000, q: 0.5, gain: 0.06 },
  dungeon: { label: 'Dungeon',       icon: 'fa-dungeon',    filter: 'lowpass',  freq: 300,  q: 0.6, gain: 0.10, osc: 55,  oscType: 'sine' },
  waves:   { label: 'Ocean',         icon: 'fa-water',      filter: 'lowpass',  freq: 700,  q: 0.3, gain: 0.13, pulse: 0.12 },
};

interface ActiveNodes {
  source: AudioBufferSourceNode;
  filter: BiquadFilterNode;
  gain: GainNode;
  osc?: OscillatorNode;
  lfo?: OscillatorNode;
  lfoGain?: GainNode;
}

let ctx: AudioContext | null = null;
let master: GainNode | null = null;
let active: ActiveNodes | null = null;
let currentTrack: AmbienceTrack | null = null;

export function listAmbienceTracks(): Array<{ key: AmbienceTrack; label: string; icon: string }> {
  return (Object.keys(TRACKS) as AmbienceTrack[]).map((key) => ({
    key, label: TRACKS[key].label, icon: TRACKS[key].icon,
  }));
}

export function getCurrentAmbience(): AmbienceTrack | null {
  return currentTrack;
}

function getCtx(): AudioContext | null {
  if (!ctx) {
    try {
      ctx = new (window.AudioContext || (window as any).webkitAudioContext)();
    } catch {
      return null;
    }
  }
  if (ctx.state === 'suspended') ctx.resume();
  return ctx;
}

/** Generates a short looping white-noise buffer. */
function noiseBuffer(audio: AudioContext): AudioBuffer {
  const seconds = 2;
  const buf = audio.createBuffer(1, audio.sampleRate * seconds, audio.sampleRate);
  const data = buf.getChannelData(0);
  for (let i = 0; i < data.length; i++) data[i] = Math.random() * 2 - 1;
  return buf;
}

export function playAmbience(track: string, volume = 0.5): void {
  const spec = TRACKS[track as AmbienceTrack];
  if (!spec) return;

  // Update state even when audio is unavailable so the UI/e2e can observe it.
  currentTrack = track as AmbienceTrack;
  document.body.dataset.ambience = track;

  stopNodes();

  const audio = getCtx();
  if (!audio) return;
  if (!master) {
    master = audio.createGain();
    master.connect(audio.destination);
  }
  master.gain.value = Math.max(0, Math.min(1, volume)) * 1.4;

  const source = audio.createBufferSource();
  source.buffer = noiseBuffer(audio);
  source.loop = true;

  const filter = audio.createBiquadFilter();
  filter.type = spec.filter;
  filter.frequency.value = spec.freq;
  filter.Q.value = spec.q;

  const gain = audio.createGain();
  gain.gain.value = spec.gain;

  source.connect(filter);
  filter.connect(gain);
  gain.connect(master);
  source.start();

  const nodes: ActiveNodes = { source, filter, gain };

  if (spec.osc) {
    const osc = audio.createOscillator();
    osc.type = spec.oscType || 'sine';
    osc.frequency.value = spec.osc;
    const oscGain = audio.createGain();
    oscGain.gain.value = spec.gain * 0.35;
    osc.connect(oscGain);
    oscGain.connect(gain);
    osc.start();
    nodes.osc = osc;
  }

  if (spec.pulse) {
    // Low-frequency amplitude wobble for drums / waves.
    const lfo = audio.createOscillator();
    lfo.frequency.value = spec.pulse;
    const lfoGain = audio.createGain();
    lfoGain.gain.value = spec.gain * 0.8;
    lfo.connect(lfoGain);
    lfoGain.connect(gain.gain);
    lfo.start();
    nodes.lfo = lfo;
    nodes.lfoGain = lfoGain;
  }

  active = nodes;
}

function stopNodes(): void {
  if (!active) return;
  try { active.source.stop(); } catch { /* already stopped */ }
  try { active.osc?.stop(); } catch { /* already stopped */ }
  try { active.lfo?.stop(); } catch { /* already stopped */ }
  try { active.source.disconnect(); } catch { /* ignore */ }
  active = null;
}

export function stopAmbience(): void {
  stopNodes();
  currentTrack = null;
  delete document.body.dataset.ambience;
}

expose('playAmbience', playAmbience);
expose('stopAmbience', stopAmbience);
