import { readFile } from "node:fs/promises";
import { resolve } from "node:path";
import { describe, expect, it } from "@jest/globals";
import { buildWidget } from "../testutil/builders.js";
import { widgetResponseSchema } from "./schemas.js";
import { encodeWidget } from "./serialization.js";

describe("widget serialization", () => {
  it("matches the reviewed golden wire representation", async () => {
    const golden = await readFile(
      resolve("src/api/testdata/widget.json"),
      "utf8",
    );
    const encoded = `${JSON.stringify(encodeWidget(buildWidget()), null, 2)}\n`;
    const decoded: unknown = JSON.parse(golden);

    expect(encoded).toBe(golden);
    expect(widgetResponseSchema.parse(decoded)).toEqual(
      encodeWidget(buildWidget()),
    );
  });
});
