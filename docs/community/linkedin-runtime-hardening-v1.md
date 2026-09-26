# LinkedIn Runtime Hardening v1

The LinkedIn adapter is provider-facing, but the runtime boundary must also prevent duplicate execution and preserve the observation produced by the provider.

## Execution boundary

authorization-v1
  -> external-publisher-v1
  -> last-boundary revalidation
  -> LinkedIn runtime
  -> LinkedIn adapter
  -> provider
  -> receipt
  -> durable observation store

linkedin-runtime.ts adds an injected receipt store around the provider adapter.

## Idempotency

The execution key binds:

- execution ID
- provider
- proposal ID
- action
- target system/resource
- policy version
- decision ID

The key uses a fixed-field canonical encoding so delimiter characters inside individual fields cannot create an alternate key with the same serialized value.

The receipt store exposes an atomic claim operation. A concurrent caller that cannot observe an existing receipt receives an in-flight result and does not invoke LinkedIn a second time.

If a receipt already exists for the execution key, NAEOS returns that observation and does not invoke LinkedIn again.

The runtime does not silently retry an execution with an existing observation, including an unknown provider outcome. Recovery policy belongs to a higher-level execution/reconciliation mechanism.

## Durability and concurrency boundary

The receipt store is an interface rather than an in-memory implementation. Production wiring must supply durable storage appropriate to the deployment and implement claim atomically.

The runtime therefore does not claim durability merely because the adapter returned a receipt. The store must also provide a recovery policy for claims left in-flight by a crashed worker; that policy is outside this wrapper and must not implicitly reinterpret an unknown provider outcome as authorization or success.

## Authorization boundary

The runtime does not create, upgrade, or infer authorization. The LinkedIn adapter continues to revalidate authorization at the last controllable boundary immediately before the provider HTTP request.

Provider receipts remain observations and never become authorization evidence.

## Testing

The runtime tests use an injected store and mock fetch. They verify:

1. first execution reaches the provider and records the receipt;
2. replay returns the recorded receipt without a second provider call;
3. concurrent execution is rejected while the first execution holds the claim;
4. changing governance identity changes the execution key;
5. delimiter characters do not create ambiguous execution keys.

No production credential or live LinkedIn call is used in CI.
