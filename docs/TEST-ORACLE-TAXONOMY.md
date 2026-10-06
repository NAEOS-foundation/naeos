# Test Oracle Taxonomy

External review found that several policy-bypass experiment failures were test-oracle defects rather than governance bypasses. NAEOS now classifies these separately.

## Categories

| Classification | Meaning | Release impact |
|---|---|---|
| POLICY_BYPASS | An unauthorized side effect occurred or an execution path crossed the gateway contrary to policy. | Blocking |
| ORACLE_DEFECT | The system behavior is correct, but the experiment asserted the wrong exit code, status, fixture, or expected state. | Blocking for the test, not evidence of a security bypass |
| INFRA_FAILURE | The experiment could not execute or verify because of environment/tooling failure. | Blocking until reproduced |
| EXPECTED_DENIAL | The gateway correctly denied or blocked execution. | Pass |
| VERIFICATION_FAILURE | Execution/evidence occurred, but independent verification could not establish the claimed result. | Blocking |

## Rule

A failed experiment must not be labeled POLICY_BYPASS solely because its process exited non-zero. The oracle must first inspect:

1. authorization decision;
2. whether execution crossed the gateway;
3. whether the protected side effect occurred;
4. evidence integrity;
5. independent verification;
6. process exit code as a separate assertion.

Exit-code mismatches are therefore ORACLE_DEFECT unless the exit code itself is the contract under test.

## Reporting

Every adversarial experiment should emit classification, expected, observed, and evidence fields. CI may fail on any non-pass classification, but reports must preserve the distinction between a security failure and an invalid oracle.
