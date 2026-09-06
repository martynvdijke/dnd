import { describe, it, expect } from "vitest";
import { distanceInUnits, pixelDistance, snapToGrid, gridSizeFromDrag } from "./map-math";

describe("map-math", () => {
  it("pixelDistance sums segments", () => {
    expect(pixelDistance([{x:0,y:0},{x:3,y:4}])).toBe(5);
    expect(pixelDistance([{x:0,y:0},{x:3,y:4},{x:6,y:4}])).toBe(8);
  });
  it("distanceInUnits fallback to px when gridSize 0", () => {
    expect(distanceInUnits([{x:0,y:0},{x:10,y:0}], 0, "ft")).toBe("10 px");
  });
  it("calibrated shows units", () => {
    const d = distanceInUnits([{x:0,y:0},{x:150,y:0}], 50, "ft");
    expect(d).toContain("ft");
    expect(d).toContain("150");
  });
  it("snapToGrid centers", () => {
    expect(snapToGrid(10,10,50)).toEqual({x:25,y:25});
    expect(snapToGrid(60,60,50)).toEqual({x:75,y:75});
    expect(snapToGrid(5,5,0)).toEqual({x:5,y:5});
  });
  it("gridSizeFromDrag", () => {
    expect(gridSizeFromDrag({x:0,y:0},{x:50,y:0})).toBe(50);
  });
});
