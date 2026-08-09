import { describe, expect, it } from "@jest/globals";
import { HttpResponse, http } from "msw";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { renderApp } from "../../../test/render-app.js";
import { server } from "../../../test/server.js";

const meter = {
  id: "550e8400-e29b-41d4-a716-446655440000",
  name: "Meter",
  description: "Measures throughput",
  created_at: "2026-08-09T12:00:00.000Z",
  updated_at: "2026-08-09T12:00:00.000Z",
  version: 1,
} as const;

describe("widget list", () => {
  it("renders loading and then successful widget links", async () => {
    renderApp();

    expect(screen.getByRole("status")).toHaveTextContent("Loading widgets");
    expect(await screen.findByRole("link", { name: "Meter" })).toBeVisible();
    expect(screen.getByRole("link", { name: "Gauge" })).toBeVisible();
  });

  it("renders a recoverable error", async () => {
    server.use(
      http.get("*/widgets", () =>
        HttpResponse.json(
          { title: "Unavailable", status: 503, type: "/problems/unavailable" },
          { status: 503 },
        ),
      ),
    );

    renderApp();

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "Widgets could not be loaded",
    );
    expect(screen.getByRole("button", { name: "Retry" })).toBeEnabled();
  });

  it("loads the next cursor page", async () => {
    const seenCursors: string[] = [];
    server.use(
      http.get("*/widgets", ({ request }) => {
        const cursor = new URL(request.url).searchParams.get("cursor") ?? "";
        seenCursors.push(cursor);
        return HttpResponse.json(
          cursor.length === 0
            ? { items: [meter], next_cursor: "page-two" }
            : {
                items: [
                  {
                    ...meter,
                    id: "6ba7b810-9dad-41d1-80b4-00c04fd430c8",
                    name: "Gauge",
                  },
                ],
                next_cursor: "",
              },
        );
      }),
    );
    const user = userEvent.setup();
    renderApp();
    await screen.findByRole("link", { name: "Meter" });

    await user.click(screen.getByRole("button", { name: "Load more" }));

    expect(await screen.findByRole("link", { name: "Gauge" })).toBeVisible();
    expect(seenCursors).toEqual(["", "page-two"]);
  });
});

describe("widget mutations", () => {
  it("associates validation errors and focuses the invalid field", async () => {
    const user = userEvent.setup();
    renderApp();
    await screen.findByRole("link", { name: "Meter" });

    await user.click(screen.getByRole("button", { name: "Create widget" }));

    expect(screen.getByRole("alert")).toHaveTextContent(
      "Correct the highlighted fields",
    );
    expect(screen.getByLabelText("Name")).toHaveAccessibleDescription(
      "Enter a widget name",
    );
    expect(screen.getByLabelText("Name")).toHaveFocus();
  });

  it("shows an optimistic create and invalidates the list", async () => {
    const response = deferred();
    let listRequests = 0;
    server.use(
      http.get("*/widgets", () => {
        listRequests += 1;
        return HttpResponse.json({ items: [meter], next_cursor: "" });
      }),
      http.post("*/widgets", async ({ request }) => {
        const body = (await request.json()) as Record<string, unknown>;
        await response.promise;
        return HttpResponse.json({
          ...meter,
          id: body.id,
          name: body.name,
          description: body.description,
        });
      }),
    );
    const user = userEvent.setup();
    renderApp();
    await screen.findByRole("link", { name: "Meter" });

    await user.type(screen.getByLabelText("Name"), "Valve");
    await user.type(screen.getByLabelText("Description"), "Controls flow");
    await user.click(screen.getByRole("button", { name: "Create widget" }));

    expect(await screen.findByRole("link", { name: "Valve" })).toBeVisible();
    expect(screen.getByRole("button", { name: "Creating…" })).toBeDisabled();
    response.resolve();
    await waitFor(() => {
      expect(listRequests).toBe(2);
    });
  });

  it("removes a widget through the delete mutation", async () => {
    const user = userEvent.setup();
    renderApp();
    await screen.findByRole("link", { name: "Meter" });

    await user.click(screen.getByRole("button", { name: "Delete Meter" }));

    await waitFor(() => {
      expect(
        screen.queryByRole("link", { name: "Meter" }),
      ).not.toBeInTheDocument();
    });
  });
});

function deferred(): Readonly<{
  promise: Promise<undefined>;
  resolve: () => void;
}> {
  let resolvePromise: ((value: undefined) => void) | undefined;
  const promise = new Promise<undefined>((resolve) => {
    resolvePromise = resolve;
  });
  return {
    promise,
    resolve: () => {
      if (resolvePromise === undefined) {
        throw new Error("Deferred promise was not initialized");
      }
      resolvePromise(undefined);
    },
  };
}
