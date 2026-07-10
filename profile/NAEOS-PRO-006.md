# NAEOS-PRO-006: Profile Migration & Upgrade Guide

## Status
- Status: Draft
- Version: 1.0
- Owner: NAEOS Foundation
- Last Updated: 2026-07-10

---

## 1. Executive Summary

Dokumentasi ini menjelaskan cara:
- **Upgrade** profile ke versi baru
- **Migrate** dari profile lama ke profile baru
- **Manage** breaking changes
- **Rollback** jika ada masalah
- **Test** upgrade sebelum production

---

## 2. Upgrade Strategies

### 2.1 Semantic Versioning

Profile menggunakan semantic versioning:

```
MAJOR.MINOR.PATCH

- PATCH (x.x.X): Bugfixes, non-breaking changes
- MINOR (x.X.x): New features, non-breaking changes  
- MAJOR (X.x.x): Breaking changes
```

**Upgrade Path**:
- `1.0.0 → 1.0.1`: Safe, automatic
- `1.0.0 → 1.1.0`: Safe, check release notes
- `1.0.0 → 2.0.0`: Breaking changes, requires migration

### 2.2 Compatibility Matrix

```
Profile Version   Go Version   NAEOS Version
1.0.x            >= 1.16      >= 0.1.0
1.5.x            >= 1.20      >= 0.2.0
2.0.x            >= 1.21      >= 0.3.0
```

---

## 3. Patch Upgrade (PATCH version)

### 3.1 Automatic Upgrade

Patch upgrades adalah safe dan non-breaking.

```bash
# Current version
naeos profile show enterprise
# Version: 1.5.2

# Upgrade to latest patch
naeos profile upgrade enterprise --to latest

# Verify
naeos profile show enterprise
# Version: 1.5.3

# Or explicitly
naeos profile upgrade enterprise --to 1.5.3
```

### 3.2 Patch Release Notes

```markdown
## enterprise v1.5.3 (Patch)

### Bug Fixes
- Fixed incorrect threshold in `gate-performance` policy
- Corrected policy description for `sec-mfa`

### Internal Changes
- Optimized policy resolution algorithm
- Improved error messages

### Migration
No changes required. Upgrade is automatic.

### Supported Projects
All projects using 1.5.x are compatible with 1.5.3.
```

---

## 4. Minor Upgrade (MINOR version)

### 4.1 Upgrade Process

Minor upgrades add features tanpa breaking changes.

```bash
# Check current version
naeos profile show enterprise
# Version: 1.4.0

# Check available minor upgrades
naeos profile upgrade enterprise --list-available
# Available:
# - 1.5.0 (new features)
# - 1.5.1 (bugfix)
# - 1.5.2 (bugfix)

# Upgrade to specific version
naeos profile upgrade enterprise --to 1.5.0

# Verify
naeos profile show enterprise
# Version: 1.5.0
```

### 4.2 Release Notes Review

Selalu review release notes sebelum minor upgrade.

```bash
# View release notes
naeos profile changelog enterprise --from 1.4.0 --to 1.5.0

# Output:
# ## enterprise v1.5.0 (Minor)
# 
# ### New Features
# - Added `gate-performance` quality gate
# - Added `deploy-canary` deployment policy
# - Added support for custom validators
# 
# ### Deprecations
# - `std-old-format` is deprecated (will be removed in v2.0.0)
# 
# ### Recommendations
# - Activate `gate-performance` for production services
# - Consider adopting `deploy-canary` for safer deployments
# 
# ### Migration
# No breaking changes. Update is optional.
```

### 4.3 Testing Minor Upgrade

```bash
# Create test specification with new version
cat > specification-test.yaml << EOF
project:
  name: payment-gateway
  version: 1.0.0

profiles:
  - enterprise:1.5.0  # New version
EOF

# Test the upgrade
naeos profile check-compliance \
  --specification specification-test.yaml \
  --project ./test-project \
  --verbose

# If all passes, update production
cat > specification.yaml << EOF
project:
  name: payment-gateway
  version: 1.0.0

profiles:
  - enterprise:1.5.0  # Updated
EOF

# Commit and deploy
git add specification.yaml
git commit -m "upgrade(profile): enterprise 1.4.0 → 1.5.0"
git push origin main
```

---

## 5. Major Upgrade (MAJOR version)

### 5.1 Breaking Changes

Major upgrades mengandung breaking changes.

```markdown
## enterprise v2.0.0 (MAJOR - Breaking Changes)

### Breaking Changes

#### Removed
- ❌ Removed `std-old-format` (deprecated in v1.5.0)
- ❌ Removed `gate-legacy-coverage` (replaced by v2 algorithm)

#### Changed
- 🔄 `gate-test-coverage` threshold increased from 80% to 85%
- 🔄 `sec-encryption` now requires TLS 1.3 (was 1.2)

#### Renamed
- ↪️ `std-readme` → `doc-readme`
- ↪️ `sec-mfa-optional` → `sec-mfa-required`

### Migration Required
Yes. See migration guide below.

### Timeline
- v1.9.0 released: 2026-01-01
- v2.0.0 released: 2026-07-01
- Support for v1.x ends: 2027-01-01
```

### 5.2 Migration Path for Major Upgrade

```
Current           Migration Step             Target
┌─────────┐      ┌──────────────┐      ┌─────────┐
│v1.0.0   │─────→│Deprecation   │─────→│v1.5.0   │
│         │      │release notes │      │prepare  │
└─────────┘      │v1.6-1.8      │      │upgrade  │
                 └──────────────┘      └─────────┘
                                             ↓
                                       ┌─────────┐
                                       │v1.9.0   │
                                       │release  │
                                       └─────────┘
                                             ↓
                                       ┌─────────┐
                                       │v2.0.0   │
                                       │release  │
                                       └─────────┘
                                             ↓
                                       ┌─────────┐
                                       │Migrate  │
                                       │projects │
                                       └─────────┘
```

### 5.3 Major Upgrade Checklist

```bash
#!/bin/bash

# Step 1: Review breaking changes
echo "Step 1: Review breaking changes"
naeos profile changelog enterprise --from 1.8.0 --to 2.0.0

# Step 2: Check impact on projects
echo "Step 2: Check impact on projects"
naeos profile upgrade-impact \
  --profile enterprise \
  --from-version 1.8.0 \
  --to-version 2.0.0 \
  --affected-projects all

# Step 3: Create migration plan
echo "Step 3: Create migration plan"
# Edit migration-plan.md (see example below)

# Step 4: Test with sample projects
echo "Step 4: Test upgrade"
naeos profile test-upgrade \
  --profile enterprise \
  --from-version 1.8.0 \
  --to-version 2.0.0 \
  --test-projects ./test-projects/

# Step 5: Prepare migration branch
echo "Step 5: Create migration branch"
git checkout -b feature/upgrade-enterprise-v2

# Step 6: Update specification
sed -i 's/enterprise:1.8.0/enterprise:2.0.0/g' specification.yaml

# Step 7: Fix migration issues
echo "Step 7: Fix identified issues"
# Edit files based on upgrade-impact report

# Step 8: Validate
echo "Step 8: Validate"
naeos profile validate specification.yaml

# Step 9: Test entire project
echo "Step 9: Test project"
naeos profile check-compliance --specification specification.yaml --project ./

# Step 10: Deploy
echo "Step 10: Deploy"
git add specification.yaml
git commit -m "chore(profile): upgrade enterprise to v2.0.0

This is a major upgrade with breaking changes.
See migration-plan.md for details.

Breaking changes:
- Removed std-old-format
- TLS 1.3 now required
- Coverage threshold raised to 85%"

git push origin feature/upgrade-enterprise-v2

# Step 11: Create PR for review
echo "Step 11: Create PR"
gh pr create \
  --title "chore(profile): upgrade enterprise to v2.0.0" \
  --body "$(cat migration-plan.md)"
```

### 5.4 Migration Plan Example

```markdown
# Migration Plan: enterprise v1.8.0 → v2.0.0

## Summary
This PR upgrades the enterprise profile from v1.8.0 to v2.0.0.
This is a MAJOR upgrade with breaking changes.

## Breaking Changes

### 1. Removed Policies
- `std-old-format`: Use `std-new-format` instead
  - **Action**: Review `.naeos/config.yaml` and use new format
  - **Impact**: 3 projects affected (listed below)

### 2. Changed Thresholds
- `gate-test-coverage`: 80% → 85%
  - **Action**: Increase test coverage in affected projects
  - **Impact**: 5 projects may fail compliance initially

### 3. Security Changes
- TLS 1.3 now required (was 1.2)
  - **Action**: Update server configurations
  - **Impact**: Infrastructure team to coordinate

## Affected Projects

| Project | Issue | Resolution | Owner |
|---------|-------|-----------|-------|
| payment-gateway | Coverage 82% < 85% | Add tests | team-payments |
| auth-service | Coverage 78% < 85% | Add tests | team-identity |
| api-gateway | TLS 1.2 configured | Update infra | platform-team |
| dashboard | Old format used | Update config | team-frontend |
| batch-service | Coverage 75% < 85% | Add tests | team-backend |

## Timeline

- **Week 1**: Notify all teams
- **Week 2**: Teams start remediation
- **Week 3**: Final tests and validation
- **Week 4**: Deploy to production

## Rollback Plan

If critical issues discovered:
```bash
git revert <commit-hash>
naeos profile show enterprise  # Should show v1.8.0
```

## Verification

After deployment:
```bash
# All projects must have 100% compliance
naeos profile compliance-report \
  --profile enterprise \
  --affected-projects all
```

## Sign-off

- [ ] Platform Team Lead
- [ ] Security Team Lead
- [ ] Architecture Team Lead
```

### 5.5 Major Upgrade Execution

```bash
# 1. Pre-flight checks
naeos profile validate specification.yaml
naeos profile check-compliance --specification specification.yaml --project ./

# 2. Backup current profile
naeos profile export enterprise:1.8.0 > backup-enterprise-v1.8.0.yaml

# 3. Upgrade profile
naeos profile upgrade enterprise --to 2.0.0

# 4. Fix issues identified
# ... make necessary changes ...

# 5. Re-validate
naeos profile validate specification.yaml

# 6. Check compliance again
naeos profile check-compliance --specification specification.yaml --project ./

# 7. Final approval
echo "Ready for production deployment? (yes/no)"
read approval

if [ "$approval" = "yes" ]; then
  # 8. Deploy
  git push origin main
  # ... trigger CI/CD pipeline ...
else
  # Rollback
  naeos profile downgrade enterprise --to 1.8.0
fi
```

---

## 6. Profile Switching

### 6.1 Switch to Different Profile

Ganti dari satu profile ke profile lain.

```bash
# Current: enterprise v1.0.0
naeos profile show active
# enterprise v1.0.0

# Switch to startup profile
naeos profile switch startup:1.2.0

# Verify
naeos profile show active
# startup v1.2.0
```

### 6.2 Parallel Profile Testing

Test profile baru sebelum switch.

```bash
# Create test specification
cat > specification-test.yaml << EOF
project:
  name: payment-gateway
  version: 1.0.0

profiles:
  - startup:1.2.0  # New profile to test
EOF

# Run tests with new profile
naeos profile check-compliance \
  --specification specification-test.yaml \
  --project ./test-suite/

# If passes, update production specification
cat > specification.yaml << EOF
project:
  name: payment-gateway
  version: 1.0.0

profiles:
  - startup:1.2.0  # Updated
EOF
```

---

## 7. Rollback Strategies

### 7.1 Rollback Recent Upgrade

```bash
# View version history
naeos profile history enterprise
# 2026-07-10 10:15  1.9.0  (current)
# 2026-07-10 09:30  1.8.0
# 2026-07-09 15:20  1.7.0

# Rollback to previous
naeos profile downgrade enterprise

# Verify rollback
naeos profile show enterprise
# Version: 1.8.0
```

### 7.2 Rollback to Specific Version

```bash
# Explicit rollback
naeos profile downgrade enterprise --to 1.7.0

# Verify
naeos profile show enterprise
# Version: 1.7.0

# Re-validate project
naeos profile check-compliance --specification specification.yaml --project ./
```

### 7.3 Git-based Rollback

```bash
# View commit history
git log --oneline specification.yaml

# Rollback specific file
git revert <commit-hash>

# Or go back to specific commit
git checkout <commit-hash> specification.yaml
git commit -m "revert(profile): downgrade to enterprise v1.7.0"

# Verify
naeos profile show enterprise
```

---

## 8. Multi-Project Migration

### 8.1 Staged Rollout

Upgrade di stages untuk mengurangi risk.

```bash
#!/bin/bash

# Stage 1: Non-production environments
echo "Stage 1: Non-production environments"
deploy_env "development"
deploy_env "staging"
run_tests_stage1

# Stage 2: Less critical production services
echo "Stage 2: Less critical production services"
deploy_env "prod-batch"
deploy_env "prod-reporting"
run_tests_stage2

# Stage 3: Core production services
echo "Stage 3: Core production services"
deploy_env "prod-api"
deploy_env "prod-payments"
run_tests_stage3

# Stage 4: Monitor for 48 hours
echo "Stage 4: Monitor"
monitor_metrics 48h
if [ $? -ne 0 ]; then
  rollback_all
  exit 1
fi

echo "✓ Upgrade complete"
```

### 8.2 Canary Deployment

Upgrade small percentage first.

```yaml
# canary-specification.yaml
project:
  name: payment-gateway
  version: 1.0.0

# 10% of projects use new profile
canary:
  percentage: 10
  strategy: "random-selection"

profiles:
  - enterprise:2.0.0  # New profile

# Other 90% still use old profile
fallback:
  profiles:
    - enterprise:1.8.0
```

### 8.3 Blue-Green Deployment

Maintain two complete environments.

```bash
# Blue environment (current)
BLUE_ENV="prod-blue"
BLUE_PROFILE="enterprise:1.8.0"

# Green environment (new)
GREEN_ENV="prod-green"
GREEN_PROFILE="enterprise:2.0.0"

# Deploy to green
deploy_to_env $GREEN_ENV $GREEN_PROFILE

# Test green
run_smoke_tests $GREEN_ENV

# Switch traffic
switch_traffic_to $GREEN_ENV

# Monitor
monitor_errors 1h

# If good, decommission blue
if [ $? -eq 0 ]; then
  decommission_env $BLUE_ENV
else
  # Switch back
  switch_traffic_to $BLUE_ENV
fi
```

---

## 9. Automation & CI/CD

### 9.1 GitHub Actions Example

```yaml
# .github/workflows/profile-upgrade.yml

name: Profile Upgrade

on:
  workflow_dispatch:
    inputs:
      profile_id:
        description: 'Profile to upgrade'
        required: true
      target_version:
        description: 'Target version'
        required: true

jobs:
  upgrade:
    runs-on: ubuntu-latest
    
    steps:
      - uses: actions/checkout@v3
      
      - name: Install NAEOS CLI
        run: curl -fsSL https://install.naeos.io/cli.sh | bash
      
      - name: Validate new version
        run: |
          naeos profile validate \
            --profile ${{ inputs.profile_id }}:${{ inputs.target_version }}
      
      - name: Check compatibility
        run: |
          naeos profile compatibility-check \
            --profile ${{ inputs.profile_id }} \
            --target-version ${{ inputs.target_version }} \
            --affected-projects all
      
      - name: Update specification
        run: |
          sed -i 's/${{ inputs.profile_id }}:[^/]*/${{ inputs.profile_id }}:${{ inputs.target_version }}/g' \
            specification.yaml
      
      - name: Run compliance checks
        run: |
          naeos profile check-compliance \
            --specification specification.yaml \
            --project ./
      
      - name: Create PR
        uses: peter-evans/create-pull-request@v4
        with:
          commit-message: |
            upgrade(profile): ${{ inputs.profile_id }} → ${{ inputs.target_version }}
          title: |
            upgrade(profile): ${{ inputs.profile_id }} → ${{ inputs.target_version }}
          body: |
            Automated profile upgrade
            
            Profile: ${{ inputs.profile_id }}
            Target Version: ${{ inputs.target_version }}
          branch: upgrade/${{ inputs.profile_id }}-${{ inputs.target_version }}
```

### 9.2 GitLab CI Example

```yaml
# .gitlab-ci.yml

profile_upgrade:
  stage: prepare
  script:
    - naeos profile validate --profile enterprise:$TARGET_VERSION
    - naeos profile compatibility-check --profile enterprise --target-version $TARGET_VERSION
    - sed -i "s/enterprise:[^/]*/enterprise:$TARGET_VERSION/g" specification.yaml
    - naeos profile check-compliance --specification specification.yaml --project ./
  only:
    - schedules  # Run on schedule
```

---

## 10. Monitoring & Alerting

### 10.1 Post-Upgrade Monitoring

```bash
#!/bin/bash

# Monitor metrics after upgrade
monitor_profile_upgrade() {
  local profile=$1
  local duration=${2:-24h}
  
  # Check error rates
  echo "Checking error rates..."
  ERROR_RATE=$(get_metric "error_rate_percentage" $duration)
  if (( $(echo "$ERROR_RATE > 5" | bc -l) )); then
    alert "High error rate: $ERROR_RATE%"
  fi
  
  # Check compliance
  echo "Checking compliance..."
  naeos profile check-compliance --profile $profile --project ./
  
  # Check performance
  echo "Checking performance..."
  PERF=$(get_metric "p95_latency_ms" $duration)
  if (( $(echo "$PERF > 500" | bc -l) )); then
    alert "High latency: ${PERF}ms"
  fi
}

monitor_profile_upgrade enterprise 24h
```

### 10.2 Alert Rules

```yaml
# alert-rules.yaml

alerts:
  - name: ProfileUpgradeFailed
    condition: "profile_compliance < 90"
    action: "page_on_call"
  
  - name: HighErrorRateAfterUpgrade
    condition: "error_rate > 5% AND time_since_deploy < 1h"
    action: "trigger_rollback"
  
  - name: PerformanceDegradation
    condition: "p95_latency_increase > 50%"
    action: "alert_platform_team"
```

---

## 11. References

- [NAEOS-PRO-001.md](NAEOS-PRO-001.md) - Profile System Specification
- [NAEOS-PRO-002.md](NAEOS-PRO-002.md) - Profile Implementation & Setup
- [NAEOS-PRO-004.md](NAEOS-PRO-004.md) - Profile Best Practices
- [NAEOS-PRO-005.md](NAEOS-PRO-005.md) - Profile API & CLI Reference
- [NAEOS-PRO-007.md](NAEOS-PRO-007.md) - Troubleshooting & FAQ
