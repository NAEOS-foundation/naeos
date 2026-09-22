# Evidence Runtime Event Binding v5.2

This deterministic repository experiment verifies that completion evidence cannot become valid from metadata alone. Each lifecycle evidence record must reference an observed runtime event with the same run identity, event type, sequence, and payload digest.

## Scenarios

- valid runtime event + evidence binding -> PASS
- missing event reference -> BLOCK
- payload mismatch -> BLOCK
- event from another run -> BLOCK
- event after the completion boundary -> BLOCK

## Claim discipline

This is a repository-level deterministic control experiment. It does not claim production distributed event durability, independent trusted hardware, cryptographic key management, or a penetration test.
