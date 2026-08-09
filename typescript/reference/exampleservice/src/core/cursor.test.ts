import { describe, expect, it } from "@jest/globals";
import { InvalidCursorError } from "./errors.js";
import { decodeCursor, encodeCursor } from "./cursor.js";
import { WIDGET_ID } from "../testutil/builders.js";

describe("widget cursor", () => {
  it("round-trips the stable keyset position", () => {
    const cursor = {
      createdAt: new Date("2026-08-08T12:00:00.000Z"),
      id: WIDGET_ID,
    };

    expect(decodeCursor(encodeCursor(cursor))).toEqual(cursor);
  });

  it("rejects malformed input", () => {
    expect(() => decodeCursor("not-json")).toThrow(InvalidCursorError);
  });
});
