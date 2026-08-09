export type RandomSource = () => number;

export type BackoffPolicy = Readonly<{
  baseDelayMs: number;
  maxDelayMs: number;
  random: RandomSource;
}>;

export function backoffCeilingMs(attempt: number, baseDelayMs: number, maxDelayMs: number): number {
  validateBackoff(attempt, baseDelayMs, maxDelayMs);
  const exponent = Math.min(attempt - 1, 30);
  return Math.min(baseDelayMs * 2 ** exponent, maxDelayMs);
}

export function fullJitterDelayMs(attempt: number, policy: BackoffPolicy): number {
  const random = policy.random();
  if (!Number.isFinite(random) || random < 0 || random > 1) {
    throw new RangeError("random source must return a finite value in [0, 1]");
  }
  const ceiling = backoffCeilingMs(attempt, policy.baseDelayMs, policy.maxDelayMs);
  return Math.floor(ceiling * random);
}

function validateBackoff(attempt: number, baseDelayMs: number, maxDelayMs: number): void {
  if (!Number.isInteger(attempt) || attempt < 1) {
    throw new RangeError("attempt must be a positive integer");
  }
  if (!Number.isInteger(baseDelayMs) || baseDelayMs < 1) {
    throw new RangeError("baseDelayMs must be a positive integer");
  }
  if (!Number.isInteger(maxDelayMs) || maxDelayMs < baseDelayMs) {
    throw new RangeError("maxDelayMs must be an integer at least baseDelayMs");
  }
}

export interface Waiter {
  wait(delayMs: number, signal: AbortSignal): Promise<void>;
}

export class TimerWaiter implements Waiter {
  async wait(delayMs: number, signal: AbortSignal): Promise<void> {
    signal.throwIfAborted();
    await new Promise<void>((resolve, reject) => {
      const timer = setTimeout(finish, delayMs);
      const abort = (): void => {
        finish(signal.reason);
      };

      function finish(reason?: unknown): void {
        clearTimeout(timer);
        signal.removeEventListener("abort", abort);
        if (reason === undefined) resolve();
        else reject(reason instanceof Error ? reason : new Error("wait aborted"));
      }

      signal.addEventListener("abort", abort, { once: true });
    });
  }
}
