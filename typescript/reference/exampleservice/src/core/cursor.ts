import { z } from "zod/v4";
import { InvalidCursorError } from "./errors.js";
import { parseWidgetId, type WidgetId } from "./widget-id.js";

const MAX_CURSOR_BYTES = 512;

const cursorWireSchema = z.strictObject({
  created_at: z.iso.datetime({ offset: true }),
  id: z.uuid(),
});

export type WidgetCursor = Readonly<{
  createdAt: Date;
  id: WidgetId;
}>;

export function encodeCursor(cursor: WidgetCursor | undefined): string {
  if (cursor === undefined) {
    return "";
  }
  const wire = JSON.stringify({
    created_at: cursor.createdAt.toISOString(),
    id: cursor.id,
  });
  return Buffer.from(wire, "utf8").toString("base64url");
}

export function decodeCursor(token: string): WidgetCursor | undefined {
  if (token.length === 0) {
    return undefined;
  }
  if (Buffer.byteLength(token, "utf8") > MAX_CURSOR_BYTES) {
    throw new InvalidCursorError();
  }
  try {
    const decoded = Buffer.from(token, "base64url").toString("utf8");
    const input: unknown = JSON.parse(decoded);
    const wire = cursorWireSchema.parse(input);
    return {
      createdAt: new Date(wire.created_at),
      id: parseWidgetId(wire.id),
    };
  } catch (error: unknown) {
    if (error instanceof InvalidCursorError) {
      throw error;
    }
    throw new InvalidCursorError(error);
  }
}
