import type { FastifyRequest } from "fastify";

export async function withRequestSignal<T>(
  request: FastifyRequest,
  lifetimeSignal: AbortSignal,
  operation: (signal: AbortSignal) => Promise<T>,
): Promise<T> {
  const requestController = new AbortController();
  const abort = (): void => {
    requestController.abort(new Error("client disconnected"));
  };
  request.raw.once("aborted", abort);
  try {
    const signal = AbortSignal.any([requestController.signal, lifetimeSignal]);
    signal.throwIfAborted();
    return await operation(signal);
  } finally {
    request.raw.off("aborted", abort);
  }
}
