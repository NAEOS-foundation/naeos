# NAEOS-PRO-005: Profile API & CLI Reference

## Status
- Status: Draft
- Version: 1.0
- Owner: NAEOS Foundation
- Last Updated: 2026-07-10

---

## 1. Executive Summary

Dokumentasi ini menyediakan referensi lengkap untuk berinteraksi dengan Profile System NAEOS melalui:
- **CLI** (Command Line Interface) - untuk automation dan scripting
- **Go API** - untuk integration dalam Go applications
- **REST API** - untuk integration dengan external tools
- **SDK** - untuk berbagai bahasa pemrograman

---

## 2. CLI Reference

### 2.1 Installation

```bash
# Install NAEOS CLI
curl -fsSL https://install.naeos.io/cli.sh | bash

# Or using package manager
brew install naeos-cli        # macOS
apt install naeos-cli         # Ubuntu/Debian
choco install naeos-cli       # Windows
```

### 2.2 Global Options

```bash
# All commands support these options:
naeos [command] [options]

--config FILE            # Path to config file (default: ~/.naeos/config.yaml)
--verbose, -v           # Enable verbose output
--debug                 # Enable debug output
--format FORMAT         # Output format: yaml, json, table (default: table)
--output FILE, -o FILE  # Write output to file
--help, -h             # Show help
--version              # Show version
```

### 2.3 Profile Commands

#### 2.3.1 naeos profile list

List semua available profiles.

```bash
# List all profiles
naeos profile list

# List with filter
naeos profile list --filter tag:enterprise
naeos profile list --filter type:organization
naeos profile list --filter status:stable

# JSON output
naeos profile list --format json

# Output:
# ID                VERSION   STATUS   TYPE          TAGS
# base              1.0.0     stable   foundation    universal
# enterprise        1.2.0     stable   organization  governance
# startup           1.0.0     stable   organization  early-stage
# fintech           1.5.0     stable   industry      fintech, regulated
# healthcare        1.3.0     stable   industry      healthcare, hipaa
```

#### 2.3.2 naeos profile show

Show detail profile.

```bash
# Show profile info
naeos profile show enterprise

# Show full details (including policies)
naeos profile show enterprise --full

# Output:
# ID: enterprise
# Name: Enterprise Profile
# Version: 1.2.0
# Status: Stable
# Owner: enterprise-architecture-team
# 
# Inherits:
#   - base
#
# Policies:
#   Quality Gates: 2
#   Security: 5
#   Testing: 4
#   Documentation: 3
#   DevOps: 4
```

#### 2.3.3 naeos profile create

Create profile baru.

```bash
# Interactive creation
naeos profile create

# Create dari template
naeos profile create \
  --id my-org-base \
  --name "MyOrg Base Profile" \
  --version 1.0.0 \
  --template base

# Create dari file
naeos profile create --file profiles/my-profile.yaml
```

#### 2.3.4 naeos profile update

Update profile existing.

```bash
# Update dari file
naeos profile update --file profiles/enterprise-v2.yaml

# Update specific field
naeos profile update enterprise \
  --set description="Updated description" \
  --set version=1.3.0

# Update policies
naeos profile update enterprise \
  --add-policy policies/security-new.yaml \
  --remove-policy sec-old-standard
```

#### 2.3.5 naeos profile delete

Delete profile.

```bash
# Delete profile
naeos profile delete enterprise

# Delete with confirmation
naeos profile delete enterprise --force

# Delete dengan backup
naeos profile delete enterprise --backup backup.yaml
```

#### 2.3.6 naeos profile tree

Show inheritance tree.

```bash
# Show tree
naeos profile tree fintech

# Output:
# fintech
# ├── inherits: enterprise
# │   ├── inherits: base
# │   │   └── (no parent)
# │   └── extends: enterprise-extensions
# └── extends: compliance-pci, compliance-gdpr

# Detailed tree
naeos profile tree fintech --detailed

# JSON output
naeos profile tree fintech --format json
```

#### 2.3.7 naeos profile diff

Compare dua profiles.

```bash
# Compare two profiles
naeos profile diff enterprise startup

# Compare versions
naeos profile diff enterprise:1.0.0 enterprise:1.2.0

# Show detailed diff
naeos profile diff enterprise startup --detailed

# Output (summary):
# Changed Policies: 5
# + Added: 3 new policies
# - Removed: 1 policy
# ~ Modified: 1 policy
```

#### 2.3.8 naeos profile validate

Validate profile.

```bash
# Validate syntax
naeos profile validate profiles/enterprise.yaml

# Validate dengan inheritance check
naeos profile validate profiles/enterprise.yaml --check-inheritance

# Validate dengan conflict check
naeos profile validate \
  profiles/enterprise.yaml \
  profiles/fintech.yaml \
  --check-conflicts

# Output:
# ✓ Syntax valid
# ✓ Inheritance valid
# ⚠ Conflicts: 1
#   - Policy 'coverage-threshold' conflicts between enterprise and fintech
#   Resolution: fintech wins (higher priority)
```

#### 2.3.9 naeos profile compile

Compile profile dengan inheritance resolution.

```bash
# Compile single profile
naeos profile compile --profile enterprise --output compiled-enterprise.yaml

# Compile multiple profiles
naeos profile compile \
  --profiles enterprise,fintech,microservices \
  --output compiled.yaml

# Compile dengan conflict report
naeos profile compile \
  --profiles enterprise,fintech \
  --report-conflicts \
  --output report.json

# Output:
# ✓ Compiled successfully
# Policies: 45
# Conflicts resolved: 2
# Report saved: report.json
```

#### 2.3.10 naeos profile register

Register profile ke registry.

```bash
# Register local profile
naeos profile register --file profiles/enterprise.yaml

# Register dengan specific registry
naeos profile register \
  --file profiles/enterprise.yaml \
  --registry https://registry.naeos.io

# Register dengan authentication
naeos profile register \
  --file profiles/enterprise.yaml \
  --registry https://registry.company.com \
  --token <token>

# Output:
# ✓ Profile registered
# ID: enterprise
# Version: 1.2.0
# Registry: https://registry.naeos.io
```

#### 2.3.11 naeos profile unregister

Unregister profile dari registry.

```bash
# Unregister profile
naeos profile unregister enterprise

# Unregister specific version
naeos profile unregister enterprise:1.2.0

# Unregister from specific registry
naeos profile unregister \
  enterprise \
  --registry https://registry.company.com
```

#### 2.3.12 naeos profile check-compliance

Check project compliance terhadap profile.

```bash
# Check compliance
naeos profile check-compliance \
  --profile enterprise \
  --project ./my-project

# Check compliance dengan detailed output
naeos profile check-compliance \
  --profile enterprise \
  --project ./my-project \
  --verbose

# Output:
# Profile: enterprise
# Project: my-project
# 
# Compliance: 92% (42/45 policies)
# 
# ✓ Passed:
#   - gate-code-review
#   - gate-test-coverage
#   - sec-no-secrets
#   ...
# 
# ✗ Failed:
#   - sec-sast-scan (tool not configured)
#   - doc-api-openapi (no openapi.yaml found)
#   - deploy-approval-required (not yet tested)
```

#### 2.3.13 naeos profile compliance-report

Generate detailed compliance report.

```bash
# Generate HTML report
naeos profile compliance-report \
  --profile enterprise \
  --project ./my-project \
  --format html \
  --output report.html

# Generate PDF report
naeos profile compliance-report \
  --profile enterprise \
  --project ./my-project \
  --format pdf \
  --output report.pdf

# Generate JSON report (for parsing)
naeos profile compliance-report \
  --profile enterprise \
  --project ./my-project \
  --format json \
  --output report.json

# Include trends (time series)
naeos profile compliance-report \
  --profile enterprise \
  --project ./my-project \
  --include-trends \
  --days 30 \
  --format pdf \
  --output compliance-trend.pdf
```

#### 2.3.14 naeos profile export

Export profile dalam berbagai format.

```bash
# Export as YAML (default)
naeos profile export enterprise --output enterprise.yaml

# Export as JSON
naeos profile export enterprise --format json --output enterprise.json

# Export with dependencies (tree)
naeos profile export enterprise \
  --include-dependencies \
  --output enterprise-full.yaml

# Export untuk external use
naeos profile export enterprise \
  --minimal \
  --output enterprise-minimal.yaml
```

#### 2.3.15 naeos profile import

Import profile dari external source.

```bash
# Import dari file
naeos profile import --file external-profile.yaml

# Import dari URL
naeos profile import --url https://registry.company.com/profiles/enterprise.yaml

# Import dan merge dengan existing
naeos profile import \
  --file external-profile.yaml \
  --merge-with enterprise \
  --output merged-profile.yaml
```

### 2.4 Policy Commands

#### 2.4.1 naeos policy list

List semua available policies.

```bash
# List all policies
naeos policy list

# Filter by type
naeos policy list --type security
naeos policy list --type testing
naeos policy list --type quality_gate

# Filter by profile
naeos policy list --profile enterprise

# Output:
# ID                         NAME                       PROFILE    TYPE
# gate-code-review          Code Review Required       enterprise quality_gate
# gate-test-coverage        Test Coverage 80%          enterprise quality_gate
# sec-no-secrets            No Secrets in Code         base       security
# std-readme                README Required            base       standard
```

#### 2.4.2 naeos policy show

Show policy detail.

```bash
# Show policy
naeos policy show gate-test-coverage

# Show from specific profile
naeos policy show gate-test-coverage --profile enterprise

# Detailed output
naeos policy show gate-test-coverage --verbose

# Output:
# ID: gate-test-coverage
# Name: Test Coverage 80%
# Profile: enterprise
# Type: quality_gate
# Version: 1.0.0
# Status: active
# 
# Description:
#   Code must have minimum 80% test coverage
# 
# Configuration:
#   threshold: 80
#   metric: coverage_percent
#   enforced: true
# 
# Tool: coverage.py
# Link: https://wiki.company.com/testing-requirements
# Owner: testing-team
```

#### 2.4.3 naeos policy validate

Validate policy.

```bash
# Validate policy file
naeos policy validate policies/my-policy.yaml

# Validate policy in profile
naeos policy validate --profile enterprise

# Output:
# ✓ Policy syntax valid
# ✓ Policy configuration valid
# ✓ Dependencies satisfied
# ✓ No conflicts detected
```

### 2.5 Practical Examples

#### Example 1: Setup Enterprise Profile

```bash
#!/bin/bash

# Create enterprise profile
naeos profile create \
  --id myorg-enterprise \
  --name "MyOrg Enterprise" \
  --version 1.0.0 \
  --template enterprise

# Register to registry
naeos profile register \
  --file profiles/myorg-enterprise.yaml \
  --registry https://registry.company.com

# Validate
naeos profile validate profiles/myorg-enterprise.yaml

# Show result
naeos profile show myorg-enterprise
```

#### Example 2: Combine Multiple Profiles

```bash
#!/bin/bash

# Create specification with multiple profiles
cat > specification.yaml << EOF
project:
  name: payment-gateway
  version: 1.0.0

profiles:
  - enterprise
  - fintech
  - microservices
  - saas-product
EOF

# Compile and check conflicts
naeos profile compile \
  --profiles enterprise,fintech,microservices,saas-product \
  --report-conflicts \
  --verbose \
  --output compiled-profile.yaml

# Check compliance
naeos profile check-compliance \
  --profile enterprise \
  --project ./ \
  --verbose
```

#### Example 3: Monitor Compliance

```bash
#!/bin/bash

# Run compliance check daily
schedule_compliance_check() {
  while true; do
    echo "Checking compliance at $(date)"
    
    naeos profile compliance-report \
      --profile enterprise \
      --project ./ \
      --format json \
      --output compliance-$(date +%Y%m%d-%H%M%S).json
    
    # Sleep 24 hours
    sleep 86400
  done
}

schedule_compliance_check
```

---

## 3. Go API Reference

### 3.1 Installation

```go
import "github.com/naeos-foundation/naeos/pkg/profile"
```

### 3.2 Loading Profiles

```go
package main

import (
  "github.com/naeos-foundation/naeos/pkg/profile"
)

// Load profile from file
profile, err := profile.LoadFromFile("profiles/enterprise.yaml")
if err != nil {
  // Handle error
}

// Load profile from registry
registryClient := profile.NewRegistryClient("https://registry.naeos.io")
profile, err := registryClient.Get("enterprise", "1.2.0")
if err != nil {
  // Handle error
}
```

### 3.3 Profile Operations

```go
// Get profile info
id := profile.ID
name := profile.Name
version := profile.Version

// Check inheritance
parents := profile.Inherits()
children := profile.InheritedBy()

// Get policies
policies := profile.Policies()
securityPolicies := profile.PoliciesByType("security")

// Check if policy exists
if profile.HasPolicy("gate-test-coverage") {
  policy := profile.GetPolicy("gate-test-coverage")
}
```

### 3.4 Profile Composition

```go
// Combine multiple profiles
profiles := []profile.Profile{
  enterpriseProfile,
  fintechProfile,
  microservicesProfile,
}

composed, err := profile.Compose(profiles)
if err != nil {
  // Handle composition error
}

// Check for conflicts
conflicts := composed.CheckConflicts()
if len(conflicts) > 0 {
  for _, conflict := range conflicts {
    fmt.Printf("Conflict: %s\n", conflict.Message)
  }
}
```

### 3.5 Policy Management

```go
// Add policy
policy := profile.NewPolicy("my-policy")
policy.SetDescription("My custom policy")
profile.AddPolicy(policy)

// Remove policy
profile.RemovePolicy("old-policy")

// Update policy
policy, _ := profile.GetPolicy("gate-test-coverage")
policy.SetThreshold(85)  // Upgrade from 80 to 85

// Validate policies
if err := profile.ValidatePolicies(); err != nil {
  // Handle validation error
}
```

### 3.6 Compliance Checking

```go
// Check project compliance
checker := profile.NewComplianceChecker(profile)

report, err := checker.Check("./my-project")
if err != nil {
  // Handle error
}

// Review report
fmt.Printf("Compliance: %.0f%%\n", report.CompliancePercentage)
for _, result := range report.Results {
  fmt.Printf("- %s: %s\n", result.PolicyID, result.Status)
}

// Export report
report.ExportToJSON("compliance-report.json")
report.ExportToHTML("compliance-report.html")
```

---

## 4. REST API Reference

### 4.1 Base URL

```
https://api.naeos.io/v1
```

### 4.2 Authentication

```bash
# Bearer token authentication
curl -H "Authorization: Bearer <token>" \
  https://api.naeos.io/v1/profiles
```

### 4.3 Endpoints

#### 4.3.1 List Profiles

```http
GET /profiles
```

Query parameters:
- `filter`: Filter profiles (e.g., `tag:enterprise`)
- `limit`: Limit results (default: 50)
- `offset`: Offset for pagination

Response:
```json
{
  "profiles": [
    {
      "id": "enterprise",
      "name": "Enterprise Profile",
      "version": "1.2.0",
      "status": "stable",
      "owner": "platform-team"
    }
  ],
  "total": 42,
  "limit": 50,
  "offset": 0
}
```

#### 4.3.2 Get Profile

```http
GET /profiles/{id}
GET /profiles/{id}/{version}
```

Response:
```json
{
  "id": "enterprise",
  "name": "Enterprise Profile",
  "version": "1.2.0",
  "description": "Profile untuk large organization",
  "inherits": ["base"],
  "policies": [
    {
      "id": "gate-test-coverage",
      "name": "Test Coverage 80%",
      "type": "quality_gate"
    }
  ]
}
```

#### 4.3.3 Create Profile

```http
POST /profiles
Content-Type: application/yaml
```

Request body:
```yaml
id: myorg-enterprise
name: "MyOrg Enterprise"
version: 1.0.0
description: "Enterprise profile for MyOrg"
inherits:
  - base
policies:
  quality_gates:
    - id: gate-coverage
      threshold: 80
```

Response:
```json
{
  "id": "myorg-enterprise",
  "version": "1.0.0",
  "created": "2026-07-10T12:34:56Z",
  "url": "https://api.naeos.io/v1/profiles/myorg-enterprise"
}
```

#### 4.3.4 Update Profile

```http
PUT /profiles/{id}
PATCH /profiles/{id}
```

#### 4.3.5 Delete Profile

```http
DELETE /profiles/{id}
```

#### 4.3.6 Check Compliance

```http
POST /profiles/{id}/check-compliance
Content-Type: application/json
```

Request body:
```json
{
  "project_path": "/path/to/project"
}
```

Response:
```json
{
  "profile_id": "enterprise",
  "compliance_percentage": 92,
  "results": [
    {
      "policy_id": "gate-test-coverage",
      "status": "pass"
    },
    {
      "policy_id": "sec-sast",
      "status": "fail",
      "message": "SAST tool not configured"
    }
  ]
}
```

#### 4.3.7 Generate Compliance Report

```http
POST /profiles/{id}/compliance-report
Content-Type: application/json
```

Request body:
```json
{
  "project_path": "/path/to/project",
  "format": "pdf",
  "include_trends": true,
  "days": 30
}
```

---

## 5. SDK Examples

### 5.1 JavaScript/Node.js

```javascript
const { Profile, ProfileClient } = require('@naeos/sdk');

// Load from registry
const client = new ProfileClient({
  registryUrl: 'https://registry.naeos.io'
});

const profile = await client.get('enterprise', '1.2.0');

// Check compliance
const complianceChecker = new profile.ComplianceChecker();
const report = await complianceChecker.check('./my-project');

console.log(`Compliance: ${report.compliancePercentage}%`);
```

### 5.2 Python

```python
from naeos import Profile, ProfileClient

# Load from registry
client = ProfileClient('https://registry.naeos.io')
profile = client.get('enterprise', '1.2.0')

# Check compliance
compliance = profile.check_compliance('./my-project')
print(f'Compliance: {compliance.percentage}%')

# Export report
compliance.export_to_json('compliance-report.json')
```

### 5.3 Java

```java
import io.naeos.profile.Profile;
import io.naeos.profile.ProfileClient;

// Load from registry
ProfileClient client = new ProfileClient("https://registry.naeos.io");
Profile profile = client.get("enterprise", "1.2.0");

// Check compliance
ComplianceReport report = profile.checkCompliance("./my-project");
System.out.println("Compliance: " + report.getCompliancePercentage() + "%");
```

---

## 6. Error Handling

### 6.1 Common Errors

| Error | Meaning | Solution |
|-------|---------|----------|
| `E001` | Profile not found | Check profile ID and registry |
| `E002` | Invalid profile syntax | Validate profile YAML/JSON |
| `E003` | Circular dependency | Review profile inheritance |
| `E004` | Policy conflict | Resolve conflicting policies |
| `E005` | Compliance check failed | Fix non-compliant issues |

### 6.2 Error Handling Examples

```bash
# CLI error handling
naeos profile show enterprise || echo "Error: Profile not found"

# Check exit code
if naeos profile validate profiles/enterprise.yaml; then
  echo "Profile valid"
else
  echo "Profile validation failed (exit code: $?)"
fi
```

```go
// Go error handling
profile, err := profile.LoadFromFile("profiles/enterprise.yaml")
if err != nil {
  switch err.(type) {
  case profile.NotFoundError:
    fmt.Println("Profile not found")
  case profile.ValidationError:
    fmt.Printf("Validation failed: %v\n", err)
  default:
    fmt.Printf("Error: %v\n", err)
  }
  return
}
```

---

## 7. Rate Limiting

REST API applies rate limits:
- 1000 requests per hour (per token)
- 10 concurrent requests

---

## 8. References

- [NAEOS-PRO-001.md](NAEOS-PRO-001.md) - Profile System Specification
- [NAEOS-PRO-002.md](NAEOS-PRO-002.md) - Profile Implementation & Setup
- [NAEOS-PRO-006.md](NAEOS-PRO-006.md) - Migration & Upgrade Guide
- [NAEOS-PRO-007.md](NAEOS-PRO-007.md) - Troubleshooting & FAQ
