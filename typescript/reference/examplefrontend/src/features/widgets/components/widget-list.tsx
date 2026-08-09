import type { ReactNode } from "react";
import { Link } from "react-router-dom";
import { AsyncStatus } from "../../../components/async-status.js";
import { useDeleteWidget, useWidgets } from "../hooks/widget-queries.js";

export function WidgetList(): ReactNode {
  const query = useWidgets();
  const deletion = useDeleteWidget();

  if (query.isPending) {
    return <AsyncStatus>Loading widgets…</AsyncStatus>;
  }
  if (query.isError) {
    return (
      <AsyncStatus kind="alert">
        Widgets could not be loaded.{" "}
        <button
          onClick={() => {
            void query.refetch();
          }}
        >
          Retry
        </button>
      </AsyncStatus>
    );
  }
  const widgets = query.data.pages.flatMap((page) => page.items);
  if (widgets.length === 0) {
    return <p>No widgets yet.</p>;
  }

  return (
    <section aria-labelledby="widgets-heading">
      <h2 id="widgets-heading">Widgets</h2>
      {deletion.isError ? (
        <p role="alert">The widget could not be deleted. Try again.</p>
      ) : null}
      <ul>
        {widgets.map((widget) => (
          <li key={widget.id}>
            <Link to={`/widgets/${widget.id}`}>{widget.name}</Link>{" "}
            <button
              type="button"
              onClick={() => {
                deletion.mutate(widget.id);
              }}
              disabled={deletion.isPending}
              aria-label={`Delete ${widget.name}`}
            >
              Delete
            </button>
          </li>
        ))}
      </ul>
      {query.hasNextPage ? (
        <button
          type="button"
          disabled={query.isFetchingNextPage}
          onClick={() => {
            void query.fetchNextPage();
          }}
        >
          {query.isFetchingNextPage ? "Loading more…" : "Load more"}
        </button>
      ) : null}
    </section>
  );
}
