import type { ReactNode } from "react";
import { Link, useParams } from "react-router-dom";
import { z } from "zod/v4";
import { AsyncStatus } from "../components/async-status.js";
import { useWidget } from "../features/widgets/hooks/widget-queries.js";
import { usePageTitle } from "./widget-list-route.js";

export default function WidgetDetailRoute(): ReactNode {
  const params = useParams();
  const parsedId = z.uuid().safeParse(params.widgetId);
  if (!parsedId.success) {
    return (
      <>
        <h1>Widget not found</h1>
        <p role="alert">The widget address is invalid.</p>
        <Link to="/">Return to widgets</Link>
      </>
    );
  }
  return <WidgetDetail id={parsedId.data} />;
}

function WidgetDetail({ id }: Readonly<{ id: string }>): ReactNode {
  const query = useWidget(id);
  usePageTitle(
    query.data === undefined
      ? "Widget | Widget administration"
      : `${query.data.name} | Widget administration`,
  );
  if (query.isPending) {
    return <AsyncStatus>Loading widget…</AsyncStatus>;
  }
  if (query.isError) {
    return (
      <AsyncStatus kind="alert">The widget could not be loaded.</AsyncStatus>
    );
  }
  return (
    <article>
      <h1>{query.data.name}</h1>
      <p>{query.data.description ?? "No description"}</p>
      <p>Version {query.data.version}</p>
      <Link to="/">Return to widgets</Link>
    </article>
  );
}
