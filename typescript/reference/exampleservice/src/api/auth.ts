import { timingSafeEqual } from "node:crypto";
import type { Principal, Role } from "../core/auth.js";
import { UnauthenticatedError } from "../core/errors.js";

export interface Authenticator {
  authenticate(authorization: string | undefined): Principal;
}

export class LocalDevAuthenticator implements Authenticator {
  readonly #expectedToken: Buffer | undefined;
  readonly #principal: Principal;

  constructor(token: string | undefined) {
    this.#expectedToken = token === undefined ? undefined : Buffer.from(token);
    this.#principal = Object.freeze({
      subject: "local-dev",
      tenantId: "local-dev",
      roles: new Set<Role>(["widgets.reader", "widgets.writer"]),
    });
  }

  authenticate(authorization: string | undefined): Principal {
    if (this.#expectedToken === undefined) {
      return this.#principal;
    }
    const token = extractBearerToken(authorization);
    const candidate = Buffer.from(token);
    if (
      candidate.length !== this.#expectedToken.length ||
      !timingSafeEqual(candidate, this.#expectedToken)
    ) {
      throw new UnauthenticatedError();
    }
    return this.#principal;
  }
}

function extractBearerToken(authorization: string | undefined): string {
  if (authorization === undefined) {
    throw new UnauthenticatedError();
  }
  const [scheme, token, extra] = authorization.split(" ");
  if (
    scheme?.toLowerCase() !== "bearer" ||
    token === undefined ||
    extra !== undefined
  ) {
    throw new UnauthenticatedError();
  }
  return token;
}
