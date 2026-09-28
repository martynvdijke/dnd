import { describe, it, expect } from 'vitest';
import { segmentIntersection, computeVisibilityPolygon, pointInPolygon, polygonPath } from './vision';

describe('vision geometry', () => {
  it('finds the intersection of crossing segments', () => {
    const hit = segmentIntersection({ x: 0, y: 0 }, { x: 10, y: 0 }, { x: 5, y: -5 }, { x: 5, y: 5 });
    expect(hit).toEqual({ x: 5, y: 0 });
  });

  it('returns null for parallel segments', () => {
    expect(segmentIntersection({ x: 0, y: 0 }, { x: 10, y: 0 }, { x: 0, y: 1 }, { x: 10, y: 1 })).toBeNull();
  });

  it('returns null when segments do not cross within their bounds', () => {
    expect(segmentIntersection({ x: 0, y: 0 }, { x: 1, y: 0 }, { x: 5, y: -5 }, { x: 5, y: 5 })).toBeNull();
  });

  it('bounds the polygon by the radius when there are no walls', () => {
    const poly = computeVisibilityPolygon({ x: 0, y: 0 }, [], 100, 8);
    expect(poly).toHaveLength(8);
    for (const p of poly) {
      expect(Math.hypot(p.x, p.y)).toBeCloseTo(100, 5);
    }
  });

  it('clips the polygon at a wall', () => {
    // Wall at x=50 blocks the ray pointing right.
    const walls = [{ x1: 50, y1: -100, x2: 50, y2: 100 }];
    const poly = computeVisibilityPolygon({ x: 0, y: 0 }, walls, 100, 4);
    const right = poly[0]; // angle 0
    expect(right.x).toBeCloseTo(50, 5);
    expect(right.y).toBeCloseTo(0, 5);
  });

  it('returns an empty polygon for a non-positive radius', () => {
    expect(computeVisibilityPolygon({ x: 0, y: 0 }, [], 0)).toEqual([]);
  });

  it('tests point-in-polygon', () => {
    const square = [{ x: 0, y: 0 }, { x: 10, y: 0 }, { x: 10, y: 10 }, { x: 0, y: 10 }];
    expect(pointInPolygon({ x: 5, y: 5 }, square)).toBe(true);
    expect(pointInPolygon({ x: 15, y: 5 }, square)).toBe(false);
  });

  it('builds SVG path data', () => {
    expect(polygonPath([{ x: 0, y: 0 }, { x: 1, y: 1 }])).toBe('M 0 0 L 1 1 Z');
    expect(polygonPath([])).toBe('');
  });
});
