import {
  useInfiniteQuery,
  useMutation,
  useQuery,
  useQueryClient,
  type InfiniteData,
} from "@tanstack/react-query";
import type {
  CreateWidgetInput,
  Widget,
  WidgetPage,
} from "../api/widget-schemas.js";
import { useWidgetsApi } from "../widgets-context.js";

const widgetListKey = ["widgets", "list", { pageSize: 20 }] as const;

export const widgetKeys = {
  all: ["widgets"] as const,
  list: () => widgetListKey,
  detail: (id: string) => ["widgets", "detail", id] as const,
};

export function useWidgets() {
  const api = useWidgetsApi();
  return useInfiniteQuery({
    queryKey: widgetKeys.list(),
    queryFn: ({ pageParam, signal }) => api.list(pageParam, signal),
    initialPageParam: "",
    getNextPageParam: (page) =>
      page.next_cursor.length > 0 ? page.next_cursor : undefined,
  });
}

export function useWidget(id: string) {
  const api = useWidgetsApi();
  return useQuery({
    queryKey: widgetKeys.detail(id),
    queryFn: ({ signal }) => api.get(id, signal),
  });
}

export function useCreateWidget() {
  const api = useWidgetsApi();
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: CreateWidgetInput) => api.create(input),
    onMutate: async (input) => {
      await queryClient.cancelQueries({ queryKey: widgetKeys.list() });
      const previous = queryClient.getQueryData<InfiniteData<WidgetPage>>(
        widgetKeys.list(),
      );
      queryClient.setQueryData<InfiniteData<WidgetPage>>(
        widgetKeys.list(),
        (current) => addOptimisticWidget(current, input),
      );
      return { previous };
    },
    onError: (_error, _input, context) => {
      if (context?.previous !== undefined) {
        queryClient.setQueryData(widgetKeys.list(), context.previous);
      }
    },
    onSuccess: (widget) => {
      queryClient.setQueryData(widgetKeys.detail(widget.id), widget);
    },
    onSettled: async () => {
      await queryClient.invalidateQueries({ queryKey: widgetKeys.list() });
    },
  });
}

export function useDeleteWidget() {
  const api = useWidgetsApi();
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.delete(id),
    onMutate: async (id) => {
      await queryClient.cancelQueries({ queryKey: widgetKeys.list() });
      const previous = queryClient.getQueryData<InfiniteData<WidgetPage>>(
        widgetKeys.list(),
      );
      queryClient.setQueryData<InfiniteData<WidgetPage>>(
        widgetKeys.list(),
        (current) => removeWidget(current, id),
      );
      return { previous };
    },
    onError: (_error, _id, context) => {
      if (context?.previous !== undefined) {
        queryClient.setQueryData(widgetKeys.list(), context.previous);
      }
    },
    onSuccess: (_data, id) => {
      queryClient.removeQueries({ queryKey: widgetKeys.detail(id) });
    },
    onSettled: async () => {
      await queryClient.invalidateQueries({ queryKey: widgetKeys.list() });
    },
  });
}

function addOptimisticWidget(
  current: InfiniteData<WidgetPage> | undefined,
  input: CreateWidgetInput,
): InfiniteData<WidgetPage> | undefined {
  const firstPage = current?.pages[0];
  if (current === undefined || firstPage === undefined) {
    return current;
  }
  const now = new Date(0).toISOString();
  const optimistic: Widget = {
    id: input.id,
    name: input.name,
    description: input.description,
    created_at: now,
    updated_at: now,
    version: 1,
  };
  return {
    ...current,
    pages: [
      { ...firstPage, items: [optimistic, ...firstPage.items] },
      ...current.pages.slice(1),
    ],
  };
}

function removeWidget(
  current: InfiniteData<WidgetPage> | undefined,
  id: string,
): InfiniteData<WidgetPage> | undefined {
  if (current === undefined) {
    return undefined;
  }
  return {
    ...current,
    pages: current.pages.map((page) => ({
      ...page,
      items: page.items.filter((widget) => widget.id !== id),
    })),
  };
}
