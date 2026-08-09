import { expect, it } from "@jest/globals";
import { HttpResponse, http } from "msw";
import { z } from "zod/v4";
import { server } from "../../test/server.js";
import type { ApiError } from "./errors.js";
import { ApiClient } from "./http-client.js";

const client = new ApiClient(new URL("http://localhost/"));

it("rejects a success response that fails its Zod contract", async () => {
  server.use(http.get("*/malformed", () => HttpResponse.json({ id: 42 })));

  await expect(
    client.requestJson("malformed", z.object({ id: z.uuid() })),
  ).rejects.toMatchObject({ kind: "invalid-response" });
});

it("maps an HTTP problem without exposing unchecked response fields", async () => {
  server.use(
    http.get("*/problem", () =>
      HttpResponse.json(
        {
          type: "/problems/forbidden",
          title: "Forbidden",
          status: 403,
          detail: "You cannot read this widget",
          internal_stack: "secret",
        },
        { status: 403 },
      ),
    ),
  );

  const request = client.requestJson("problem", z.object({ ok: z.boolean() }));

  await expect(request).rejects.toEqual(
    expect.objectContaining({
      kind: "http",
      message: "You cannot read this widget",
      status: 403,
    } satisfies Partial<ApiError>),
  );
});

it("maps a caller abort separately from network failure", async () => {
  const controller = new AbortController();
  controller.abort();

  await expect(
    client.requestJson("widgets", z.object({}), { signal: controller.signal }),
  ).rejects.toMatchObject({ kind: "aborted" });
});
