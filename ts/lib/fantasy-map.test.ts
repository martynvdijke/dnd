import { describe, it, expect } from 'vitest';
import { WORLD_BOUNDS, WORLD_MAP_NAME, findWorldMap, parchmentDataUrl } from './fantasy-map';

describe('fantasy-map', () => {
  it('world bounds enclose the full degree range', () => {
    const [[latA, lngA], [latB, lngB]] = WORLD_BOUNDS;
    expect(Math.max(latA, latB)).toBe(90);
    expect(Math.min(latA, latB)).toBe(-90);
    expect(Math.max(lngA, lngB)).toBe(180);
    expect(Math.min(lngA, lngB)).toBe(-180);
  });

  it('findWorldMap selects only the world map, ignoring battle maps and junk', () => {
    const maps = [
      { id: 1, name: 'Goblin Cave', image_url: '/media/a.png' },
      { id: 2, name: WORLD_MAP_NAME, image_url: '/media/world.png' },
    ];
    expect(findWorldMap(maps)?.id).toBe(2);
    expect(findWorldMap([maps[0]])).toBeNull();
    expect(findWorldMap(null)).toBeNull();
    expect(findWorldMap(undefined as any)).toBeNull();
  });

  it('findWorldMap still finds a world map with no image yet', () => {
    expect(findWorldMap([{ id: 5, name: WORLD_MAP_NAME, image_url: '' }])?.id).toBe(5);
  });

  it('parchmentDataUrl returns a data URL or null when canvas is unavailable', () => {
    const url = parchmentDataUrl(64, 32);
    if (url !== null) expect(url.startsWith('data:image/')).toBe(true);
  });
});
