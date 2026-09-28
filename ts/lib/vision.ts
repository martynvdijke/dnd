/**
 * Line-of-sight geometry for the battlemap. Pure functions in an arbitrary
 * coordinate space (the battlemap uses map pixels); no DOM, no dependencies.
 */
export type Point = { x: number; y: number };
export type Segment = { x1: number; y1: number; x2: number; y2: number };

/**
 * Intersection point of segment a→b with segment c→d, or null when they are
 * parallel or do not cross within both segments.
 */
export function segmentIntersection(a: Point, b: Point, c: Point, d: Point): Point | null {
  const rx = b.x - a.x;
  const ry = b.y - a.y;
  const sx = d.x - c.x;
  const sy = d.y - c.y;
  const denom = rx * sy - ry * sx;
  if (Math.abs(denom) < 1e-9) return null;
  const t = ((c.x - a.x) * sy - (c.y - a.y) * sx) / denom;
  const u = ((c.x - a.x) * ry - (c.y - a.y) * rx) / denom;
  if (t < 0 || t > 1 || u < 0 || u > 1) return null;
  return { x: a.x + t * rx, y: a.y + t * ry };
}

/**
 * Visibility polygon around `origin`, bounded by `radius` and clipped by walls.
 * Casts `rayCount` rays evenly around the origin and takes the nearest wall hit.
 *
 * ponytail: fixed ray fan (default 180). Upgrade to an angular sweep over wall
 * endpoints if the fan looks coarse or perf matters.
 */
export function computeVisibilityPolygon(
  origin: Point,
  walls: Segment[],
  radius: number,
  rayCount = 180,
): Point[] {
  if (radius <= 0) return [];
  const n = Math.max(8, Math.floor(rayCount));
  const pts: Point[] = [];
  for (let i = 0; i < n; i++) {
    const ang = (i / n) * Math.PI * 2;
    const dx = Math.cos(ang);
    const dy = Math.sin(ang);
    const far: Point = { x: origin.x + dx * radius, y: origin.y + dy * radius };
    let best = radius;
    for (const w of walls) {
      const hit = segmentIntersection(origin, far, { x: w.x1, y: w.y1 }, { x: w.x2, y: w.y2 });
      if (hit) {
        const dist = Math.hypot(hit.x - origin.x, hit.y - origin.y);
        if (dist < best) best = dist;
      }
    }
    pts.push({ x: origin.x + dx * best, y: origin.y + dy * best });
  }
  return pts;
}

/** Ray-casting point-in-polygon test. */
export function pointInPolygon(p: Point, poly: Point[]): boolean {
  let inside = false;
  for (let i = 0, j = poly.length - 1; i < poly.length; j = i++) {
    const xi = poly[i].x;
    const yi = poly[i].y;
    const xj = poly[j].x;
    const yj = poly[j].y;
    const crosses = yi > p.y !== yj > p.y;
    if (crosses && p.x < ((xj - xi) * (p.y - yi)) / (yj - yi) + xi) inside = !inside;
  }
  return inside;
}

/** SVG path data for a closed polygon. */
export function polygonPath(poly: Point[]): string {
  if (poly.length === 0) return '';
  return `M ${poly.map((p) => `${p.x} ${p.y}`).join(' L ')} Z`;
}
