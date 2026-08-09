export type ErrorCode =
  | "unauthenticated"
  | "forbidden"
  | "widget_invalid"
  | "cursor_invalid"
  | "widget_not_found"
  | "widget_name_conflict"
  | "widget_version_conflict"
  | "idempotency_conflict"
  | "dependency_failed";

export class AppError extends Error {
  readonly code: ErrorCode;

  constructor(code: ErrorCode, message: string, options?: ErrorOptions) {
    super(message, options);
    this.name = "AppError";
    this.code = code;
  }
}

export class UnauthenticatedError extends AppError {
  constructor() {
    super("unauthenticated", "authentication is required");
    this.name = "UnauthenticatedError";
  }
}

export class ForbiddenError extends AppError {
  constructor() {
    super("forbidden", "the principal cannot perform this action");
    this.name = "ForbiddenError";
  }
}

export class InvalidWidgetError extends AppError {
  constructor(message: string) {
    super("widget_invalid", message);
    this.name = "InvalidWidgetError";
  }
}

export class InvalidCursorError extends AppError {
  constructor(cause?: unknown) {
    super("cursor_invalid", "the pagination cursor is invalid", { cause });
    this.name = "InvalidCursorError";
  }
}

export class WidgetNotFoundError extends AppError {
  constructor() {
    super("widget_not_found", "the widget was not found");
    this.name = "WidgetNotFoundError";
  }
}

export class DuplicateWidgetNameError extends AppError {
  constructor() {
    super("widget_name_conflict", "a widget with that name already exists");
    this.name = "DuplicateWidgetNameError";
  }
}

export class WidgetVersionConflictError extends AppError {
  constructor() {
    super("widget_version_conflict", "the widget version has changed");
    this.name = "WidgetVersionConflictError";
  }
}

export class IdempotencyConflictError extends AppError {
  constructor() {
    super(
      "idempotency_conflict",
      "the idempotency key was already used for another request",
    );
    this.name = "IdempotencyConflictError";
  }
}

export class DependencyError extends AppError {
  constructor(message: string, cause: unknown) {
    super("dependency_failed", message, { cause });
    this.name = "DependencyError";
  }
}
