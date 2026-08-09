import pino, { type Logger } from "pino";

export function buildLogger(level: string): Logger {
  return pino({
    level,
    base: { service: "exampleworker" },
    redact: {
      paths: ["password", "token", "authorization", "*.password", "*.token", "*.authorization"],
      censor: "[REDACTED]",
    },
  });
}
