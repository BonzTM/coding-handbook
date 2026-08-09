export interface TelemetryLifecycle {
  shutdown(signal: AbortSignal): Promise<void>;
}

export class NoopTelemetry implements TelemetryLifecycle {
  shutdown(signal: AbortSignal): Promise<void> {
    signal.throwIfAborted();
    return Promise.resolve();
  }
}
