import type { Principal } from "../core/auth.js";

declare module "fastify" {
  interface FastifyRequest {
    principal?: Principal;
  }
}
