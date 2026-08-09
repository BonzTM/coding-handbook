import { useEffect, type ReactNode } from "react";
import { WidgetForm } from "../features/widgets/components/widget-form.js";
import { WidgetList } from "../features/widgets/components/widget-list.js";

export function WidgetListRoute(): ReactNode {
  usePageTitle("Widgets | Widget administration");
  return (
    <>
      <h1>Widget administration</h1>
      <WidgetForm />
      <WidgetList />
    </>
  );
}

export function usePageTitle(title: string): void {
  useEffect(() => {
    const previous = document.title;
    document.title = title;
    return () => {
      document.title = previous;
    };
  }, [title]);
}
