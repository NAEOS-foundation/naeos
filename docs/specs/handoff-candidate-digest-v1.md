# NAEOS Handoff Candidate Digest v1

**Status:** Proposed interoperability profile; not yet implemented in the production runtime.  
**Profile ID:** `naeos-candidate-jcs-sha256-v1`  
**Related contract:** NAEOS-Handoff-Interop-Test-001

## Purpose

Different runtimes must compute the same candidate digest before an authorization decision can be bound to an action. Go's `encoding/json` output is not, by itself, a cross-language canonicalization contract. This profile specifies a portable representation for interoperability testing.

## Candidate envelope

The digest input is exactly one JSON object with these required top-level members:

- `tool`: non-empty string naming the proposed tool/action.
- `input`: JSON value describing the complete arguments that will be executed.

The digest envelope contains no other members. Authorization metadata (subject, scope, policy version, expiry, decision ID) is validated separately and must not be silently folded into or omitted from the candidate payload. The exact tool and input used for execution must be the ones that were digested.

## Canonicalization and digest

1. Parse input as JSON while rejecting duplicate object member names. Do not accept NaN, Infinity, or -Infinity.
2. Validate that strings are valid Unicode and numbers are representable under the JSON/I-JSON constraints required by RFC 8785. Implementations must not silently round values that cannot be represented consistently as IEEE-754 binary64.
3. Canonicalize the candidate envelope using **JSON Canonicalization Scheme (JCS), RFC 8785**. Object member ordering follows JCS; array order is preserved. Do not Unicode-normalize strings.
4. Compute SHA-256 over the UTF-8 bytes of the domain-separation prefix `NAEOS-CANDIDATE-DIGEST-V1` followed by one LF byte (`0x0A), then immediately followed by the canonical JSON bytes. There is no whitespace or newline after the canonical JSON.
5. Encode the 32-byte hash as 64 lowercase hexadecimal characters. The profile identifier must accompany the digest in any cross-runtime protocol; never infer the profile from the digest alone.

Formally:

`digest = lowercase_hex(SHA256(UTF8("NAEOS-CANDIDATE-DIGEST-V1" || LF) || UTF8(JCS(candidate))))`

## Required rejection behavior

Reject the candidate rather than inventing a digest when any of these apply:

- `tool` is absent, empty, or not a string.
- `input` is absent.
- The JSON contains duplicate member names, invalid Unicode, or non-JSON numeric values.
- A numeric value falls outside the interoperable numeric constraints of the implementation/profile.
- The implementation cannot guarantee RFC 8785 canonicalization.

Do not use a fallback serializer, silently coerce types, normalize strings, or execute when digest generation fails.

## Golden vectors

The machine-readable vectors are in `handoff-candidate-digest-v1-vectors.json`. Each vector records the canonical bytes and expected digest. Implementations in every supported language must match these bytes and hashes exactly before being considered conformant to this profile.

## Security and compatibility notes

- Any change to envelope fields, numeric constraints, canonicalization, prefix, or hash algorithm requires a new profile version.
- Candidate digest equality is necessary but not sufficient authorization. The enforcement point must also validate decision ID, subject, scope/capabilities, current policy version, expiry, and replay/idempotency state.
- A digest is not proof that the action actually executed as described. Runtime evidence must correlate the digest with the authorization decision and observed outcome.
- This document is a proposed specification. The existing Go mock harness still uses its own `encoding/json` representation and is **not yet conformant** to this profile. The interop test package now has a separate JCS-based digest helper and conformance tests, but that helper is test-only and is not wired into the mock harness or production enforcement path. Do not claim production or cross-language conformance until the runtime integration and independent implementation checks are complete.
- Reference: RFC 8785, JSON Canonicalization Scheme: https://www.rfc-editor.org/rfc/rfc8785
