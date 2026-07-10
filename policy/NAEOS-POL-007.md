# NAEOS-POL-007: Policy Troubleshooting & FAQ

## Status
- Status: Draft
- Version: 1.0
- Owner: NAEOS Foundation
- Last Updated: 2026-07-10

---

## 1. Troubleshooting Guide

### 1.1 Policy Not Found

**Error:**
```
Error: Policy 'sec-tls-13' not found in registry
```

**Causes:**
1. Policy belum diregister
2. Policy ID salah
3. Registry tidak accessible

**Solutions:**

```bash
# List available policies
naeos policy list

# Search untuk policy
naeos policy list | grep tls

# Check registry connection
naeos policy registry-ping https://registry.naeos.io

# Register policy if missing
naeos policy register --file policies/sec-tls-13.yaml

# Verify registration
naeos policy show sec-tls-13
```

---

### 1.2 Policy Evaluation Fails

**Error:**
```
Error: Policy evaluation failed
  Policy: sec-tls-13
  Reason: Evaluation tool not found: testssl.sh
```

**Causes:**
1. Evaluation tool not installed
2. Tool not in PATH
3. Tool version incompatible

**Solutions:**

```bash
# Check tool availability
which testssl.sh

# Install missing tool
brew install testssl

# Or using package manager
apt install testssl-ng

# Verify tool version
testssl.sh --version

# Configure tool path
naeos policy config sec-tls-13 \
  --tool testssl.sh \
  --path /usr/local/bin/testssl.sh

# Re-run evaluation
naeos policy eval sec-tls-13 --target api-gateway
```

---

### 1.3 Circular Dependency

**Error:**
```
Error: Circular policy dependency detected
  Dependency chain: policy-a → policy-b → policy-c → policy-a
```

**Cause:**
Policy inheritance membentuk cycle.

**Solution:**

```bash
# Visualize dependency
naeos policy tree sec-tls-13 --show-all-dependencies

# Review policies involved
naeos policy show policy-a
naeos policy show policy-b
naeos policy show policy-c

# Remove circular dependency
# Edit one of the policies to break cycle
naeos policy update policy-a --remove-dependency policy-b

# Verify no cycles
naeos policy validate policy-a --check-cycles
```

---

### 1.4 Policy Conflict

**Error:**
```
Error: Policy conflict detected
  Policy 1: sec-high (tls_version: 1.3)
  Policy 2: sec-legacy (tls_version: 1.2)
  Conflicting field: tls_version
```

**Solution:**

```bash
# Analyze conflict
naeos policy diff sec-high sec-legacy

# Resolve by choosing which policy wins
# Option 1: Reorder policies (last wins)
specification:
  policies:
    - sec-legacy
    - sec-high  # Will override sec-legacy

# Option 2: Merge policies
naeos policy merge sec-high sec-legacy \
  --strategy=prefer-strict \
  --output merged-policy.yaml

# Option 3: Create custom policy
naeos policy create \
  --id sec-custom \
  --rule tls_version:1.3 \
  --extends sec-base
```

---

### 1.5 Evaluation Takes Too Long

**Error:**
```
Warning: Policy evaluation took 5.2 seconds (expected < 1 second)
```

**Causes:**
1. Slow evaluation tools
2. Large dataset being evaluated
3. Network latency (registry calls)

**Solutions:**

```bash
# Profile evaluation performance
naeos policy eval sec-tls-13 --target api-gateway --profile

# Output:
# Evaluation timing:
#   - Load policy: 50ms
#   - Initialize tool: 150ms
#   - Run evaluation: 4200ms
#   - Parse results: 100ms
#   - Generate report: 150ms
#   Total: 4650ms

# Optimize by:
# 1. Caching policy
naeos policy eval sec-tls-13 --cache --cache-ttl 3600

# 2. Parallel evaluation
naeos policy eval --parallel --workers 4 \
  sec-tls-13 sec-mfa sec-no-secrets

# 3. Batch evaluation
naeos policy eval --profile enterprise \
  --batch \
  --batch-size 10
```

---

### 1.6 Policy Not Applied

**Error:**
```
Policy 'sec-tls-13' evaluated but appears to have no effect
```

**Causes:**
1. Policy not enabled
2. Conditions don't match
3. Enforcement not enabled

**Solutions:**

```bash
# Check if policy is enabled
naeos policy show sec-tls-13 | grep status

# Enable policy
naeos policy enable sec-tls-13

# Check conditions
naeos policy show sec-tls-13 --verbose | grep conditions

# Verify conditions match target
naeos policy eval sec-tls-13 --target payment-api --debug

# Check enforcement settings
naeos policy config sec-tls-13 --show enforcement

# Enable enforcement
naeos policy config sec-tls-13 \
  --enforcement-mode blocking \
  --enforcement-level critical
```

---

### 1.7 Auto-Remediation Fails

**Error:**
```
Error: Auto-remediation failed for policy sec-tls-13
  Reason: Permission denied when updating configuration
```

**Solutions:**

```bash
# Check remediation permission
naeos policy check-permissions sec-tls-13 remediate

# Run remediation with elevated privileges
sudo naeos policy remediate sec-tls-13 --target api-gateway

# Or configure service account
naeos policy config sec-tls-13 \
  --remediation-service-account platform-admin

# Test remediation in dry-run mode
naeos policy remediate sec-tls-13 \
  --target api-gateway \
  --dry-run \
  --verbose
```

---

## 2. FAQ

### 2.1 Policy Basics

**Q: Apa perbedaan policy dan rule?**

A:
- **Policy**: High-level intent (e.g., "All connections must be secure")
- **Rule**: Concrete implementation of policy (e.g., "Use TLS 1.3")

Policy dipecah menjadi rules untuk evaluasi.

---

**Q: Bisakah saya membuat custom policy?**

A: Ya. Lihat [NAEOS-POL-002.md](NAEOS-POL-002.md) untuk cara membuat custom policy.

```bash
naeos policy create \
  --id my-custom-policy \
  --template security \
  --output my-policy.yaml

# Edit my-policy.yaml, kemudian

naeos policy register --file my-policy.yaml
```

---

**Q: Bagaimana policy di-evaluate?**

A:
1. Load policy definition
2. Check if conditions match
3. Run evaluation tool/method
4. Compare results against baseline
5. Generate report
6. Log audit trail

---

### 2.2 Organization & Management

**Q: Berapa banyak policies yang harus saya buat?**

A: Tergantung kebutuhan. Typical:
- **Base policies**: 5-10 (universal standards)
- **Category policies**: 20-50 (per category)
- **Organization policies**: 10-30 (org-specific)
- **Total**: 50-100 policies adalah normal

---

**Q: Bagaimana organize policies?**

A: Organize by directory structure:
```
policy/
├── security/
├── testing/
├── code_quality/
├── compliance/
└── deployment/
```

Lihat [NAEOS-POL-004.md](NAEOS-POL-004.md) Section 3.1 untuk lebih detail.

---

**Q: Bagaimana version policies?**

A: Gunakan semantic versioning:
```yaml
policy:
  version: "1.2.3"  # MAJOR.MINOR.PATCH
```

---

### 2.3 Evaluation & Enforcement

**Q: Kapan policies harus di-evaluate?**

A:
- **Pre-commit**: Formatting, secrets detection
- **CI/CD**: Security, testing, quality
- **Pre-deployment**: Compliance checks
- **Production**: Continuous monitoring

Lihat [NAEOS-POL-006.md](NAEOS-POL-006.md) Section 2 untuk strategies.

---

**Q: Bagaimana handle policy exceptions?**

A:
```yaml
exceptions:
  - id: exception-legacy-system
    applies_to:
      services: [legacy-api]
    valid_until: "2027-12-31"
    approved_by: "ciso"
```

---

**Q: Bisakah auto-remediate policy violations?**

A: Untuk beberapa policies, ya. Contoh:
- Code formatting: Auto-fix with gofmt/prettier
- File requirements: Auto-create README.md
- Configuration: Auto-update config

Tapi untuk security/compliance, usually auto-remediation tidak recommended. Better to alert dan require manual fix.

---

### 2.4 Integration

**Q: Bagaimana integrate policies dengan CI/CD?**

A: Lihat [NAEOS-POL-006.md](NAEOS-POL-006.md) Section 5 untuk CI/CD integration examples.

Ringkas:
1. Add policy check stage
2. Fail pipeline if critical policies fail
3. Generate reports
4. Comment on PR dengan results

---

**Q: Bagaimana enforce policies di Kubernetes?**

A: Gunakan admission webhook. Lihat [NAEOS-POL-006.md](NAEOS-POL-006.md) Section 6 untuk contoh.

---

### 2.5 Performance & Optimization

**Q: Bagaimana optimize policy evaluation?**

A:
1. **Caching**: Cache policy definitions dan results
2. **Parallel**: Evaluate multiple policies in parallel
3. **Short-circuit**: Stop on first critical failure
4. **Batch**: Evaluate multiple targets at once
5. **Local tools**: Use local tools instead of remote APIs

```bash
# Enable caching
naeos policy eval --cache --cache-ttl 3600

# Parallel evaluation
naeos policy eval --parallel --workers 8

# Batch evaluation
naeos policy eval --batch --batch-size 20
```

---

**Q: Bagaimana monitor compliance over time?**

A:
```bash
# Generate daily compliance report
naeos policy report \
  --profile enterprise \
  --format json \
  --output compliance-$(date +%Y%m%d).json

# Analyze trend
naeos policy trend \
  --profile enterprise \
  --days 30 \
  --output trend.html
```

---

### 2.6 Troubleshooting

**Q: Bagaimana debug policy evaluation?**

A:
```bash
# Enable debug mode
naeos policy eval sec-tls-13 \
  --target api-gateway \
  --debug

# Enable verbose logging
naeos policy eval sec-tls-13 \
  --target api-gateway \
  --verbose

# Show evaluation steps
naeos policy eval sec-tls-13 \
  --target api-gateway \
  --show-steps
```

---

**Q: Bagaimana troubleshoot policy conflicts?**

A:
```bash
# Analyze conflict
naeos policy diff policy1 policy2

# Check inheritance
naeos policy tree policy1

# Validate policy
naeos policy validate policy1 --check-cycles
```

---

### 2.7 Best Practices

**Q: Best practices untuk policy authoring?**

A: Lihat [NAEOS-POL-004.md](NAEOS-POL-004.md) untuk comprehensive best practices.

Ringkas:
1. **Single responsibility**: One policy, one purpose
2. **Clear documentation**: Always explain why
3. **Measurable**: Define clear success criteria
4. **Exceptions**: Allow structured exceptions
5. **Testing**: Test policies in staging first
6. **Communication**: Announce before enforcement

---

## 3. Common Issues Reference

| Issue | Error | Solution | Link |
|-------|-------|----------|------|
| Policy not found | E001 | Check registry, list policies | 1.1 |
| Eval fails | E002 | Install missing tools | 1.2 |
| Circular dependency | E003 | Remove cycle in inherits | 1.3 |
| Policy conflict | E004 | Resolve with strategy | 1.4 |
| Slow evaluation | Perf | Cache, parallelize | 1.5 |
| Policy not applied | E005 | Enable, check conditions | 1.6 |
| Remediation fails | E006 | Check permissions | 1.7 |

---

## 4. Getting Help

### 4.1 Documentation

- [NAEOS-POL-001.md](NAEOS-POL-001.md) - Policy Compiler overview
- [NAEOS-POL-002.md](NAEOS-POL-002.md) - How to write policies
- [NAEOS-POL-003.md](NAEOS-POL-003.md) - Policy examples
- [NAEOS-POL-004.md](NAEOS-POL-004.md) - Best practices
- [NAEOS-POL-005.md](NAEOS-POL-005.md) - Technical deep dive
- [NAEOS-POL-006.md](NAEOS-POL-006.md) - Evaluation & enforcement

### 4.2 Support Channels

- **Issues**: https://github.com/NAEOS-foundation/naeos/issues
- **Discussions**: https://github.com/NAEOS-foundation/naeos/discussions
- **Community Slack**: https://slack.naeos.io
- **Email**: support@naeos.io

### 4.3 Debug Information

When reporting issues, include:

```bash
# Collect diagnostic info
naeos policy diagnose sec-tls-13 \
  --output diagnostics.json

# Include in issue:
# - diagnostics.json
# - NAEOS version: naeos version
# - Policy version: naeos policy show sec-tls-13
# - Steps to reproduce
# - Expected vs actual behavior
```

---

## 5. References

- [NAEOS-POL-001.md](NAEOS-POL-001.md) - Policy Compiler
- [NAEOS-POL-002.md](NAEOS-POL-002.md) - Policy Definition & Authoring
- [NAEOS-POL-003.md](NAEOS-POL-003.md) - Policy Examples & Templates
- [NAEOS-POL-004.md](NAEOS-POL-004.md) - Policy Best Practices
- [NAEOS-POL-005.md](NAEOS-POL-005.md) - Policy Compiler & Engine
- [NAEOS-POL-006.md](NAEOS-POL-006.md) - Policy Evaluation & Enforcement
- [docs/NES-012-Policy.md](docs/NES-012-Policy.md) - Policy Model Specification
