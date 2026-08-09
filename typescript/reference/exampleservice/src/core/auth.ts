import { ForbiddenError, UnauthenticatedError } from "./errors.js";

export const roles = ["widgets.reader", "widgets.writer"] as const;
export type Role = (typeof roles)[number];

export type Principal = Readonly<{
  subject: string;
  tenantId: string;
  roles: ReadonlySet<Role>;
}>;

export function requireRole(
  principal: Principal | undefined,
  role: Role,
): Principal {
  if (principal === undefined) {
    throw new UnauthenticatedError();
  }
  if (!principal.roles.has(role)) {
    throw new ForbiddenError();
  }
  return principal;
}
