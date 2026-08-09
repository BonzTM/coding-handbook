# Recipe: Add gRPC Method

Use this when a versioned gRPC service adds or changes one RPC.

## Files To Touch

- `api/<service>/v<N>/<service>.proto`
- generated Python modules/type stubs under the repo's generated-code path
- `src/<app>/api/grpc/` servicer and error mapping
- core use case and transport/core tests

## Steps

1. Add the RPC/messages under a versioned package; assign new field numbers and reserve removed names/numbers.
2. Regenerate with the repo-pinned `grpcio-tools` command and review the complete diff.
3. Keep the servicer thin: validate/map protobuf values, call core once, map domain errors to stable status/details, map response.
4. Propagate context cancellation/deadline and impose a server work bound when the caller supplies none.
5. Confirm interceptors cover recovery, auth, logging, tracing, metrics, and message-size limits.

## Invariants To Preserve

- Generated/protobuf types stay out of core; generated files are never hand-edited.
- Existing field numbers are never renumbered or reused.
- Errors are mapped and logged once at the transport boundary.
- Metric labels use method and status only.

## Proof

```bash
uv run python -m grpc_tools.protoc <repo-pinned-arguments>
git diff --exit-code -- <generated-path>
uv run pytest tests/api/grpc -k '<method>'
grpcurl -plaintext -d '{<request>}' localhost:<port> <package>.<Service>/<Method>
make verify
```

Include compatibility, cancellation/deadline, auth rejection, and error-mapping tests. Governing doc: [gRPC services](../services/grpc-services.md).
