import pino, { type Logger } from "pino";

export type LogLevel = "debug" | "info" | "warn" | "error";

export function createLogger(level: LogLevel): Logger {
  return pino({
    level,
    base: { service: "typescript-exampleservice", version: "0.1.0" },
    redact: {
      paths: [
        "req.headers.authorization",
        "req.headers.cookie",
        "authorization",
        "config.databaseUrl",
        "config.authToken",
        "*.password",
        "*.token",
      ],
      censor: "[REDACTED]",
    },
  });
}

export function createAuditLogger(level: LogLevel): Logger {
  return pino(
    {
      level,
      base: { service: "typescript-exampleservice", stream: "security-audit" },
      redact: {
        paths: ["*.password", "*.token", "authorization"],
        censor: "[REDACTED]",
      },
    },
    pino.destination({ dest: 2, sync: true }),
  );
}
