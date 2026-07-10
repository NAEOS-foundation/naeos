# NAEOS-POL-006: Policy Evaluation & Enforcement

## Status
- Status: Draft
- Version: 1.0
- Owner: NAEOS Foundation
- Last Updated: 2026-07-10

---

## 1. Executive Summary

Panduan lengkap tentang cara mengevaluasi policies terhadap targets dan mengimplementasikan enforcement. Mencakup:
- Evaluation strategies & workflows
- Enforcement mechanisms
- Integration dengan CI/CD
- Real-time vs batch evaluation
- Monitoring & alerting

---

## 2. Evaluation Strategies

### 2.1 On-Demand Evaluation

Evaluate policy ketika diminta:

```bash
# Evaluate single policy
naeos policy eval sec-tls-13 --target api-gateway

# Evaluate multiple policies
naeos policy eval sec-tls-13,sec-mfa --target payment-api

# Evaluate all policies
naeos policy eval --all --target myservice
```

**Use cases:**
- Manual compliance checks
- Pre-deployment validation
- Troubleshooting
- Ad-hoc audits

### 2.2 Pre-Commit Evaluation

Evaluate before code is committed:

```bash
# Git hook: .git/hooks/pre-commit
#!/bin/bash

naeos policy eval code-format --project .
if [ $? -ne 0 ]; then
    echo "Code formatting policy failed"
    exit 1
fi

naeos policy eval test-coverage --project .
if [ $? -ne 0 ]; then
    echo "Test coverage policy failed"
    exit 1
fi
```

**Policies suitable for pre-commit:**
- code-format
- test-coverage (minimum)
- sec-no-secrets
- code-complexity

### 2.3 CI/CD Pipeline Evaluation

Evaluate in CI pipeline:

```yaml
# .github/workflows/policy-check.yml

name: Policy Checks

on: [push, pull_request]

jobs:
  policy-eval:
    runs-on: ubuntu-latest
    
    steps:
      - uses: actions/checkout@v3
      
      - name: Install NAEOS
        run: curl -fsSL https://install.naeos.io/cli.sh | bash
      
      - name: Evaluate policies
        run: |
          naeos policy eval \
            --profile enterprise \
            --project . \
            --exit-on-fail
      
      - name: Generate report
        run: |
          naeos policy report \
            --profile enterprise \
            --project . \
            --format html \
            --output policy-report.html
      
      - name: Upload report
        uses: actions/upload-artifact@v3
        with:
          name: policy-report
          path: policy-report.html
```

### 2.4 Scheduled Evaluation

Evaluate on schedule (daily, weekly):

```yaml
# GitLab CI scheduled pipeline
policy_check_daily:
  schedule:
    - cron: "0 2 * * *"  # 2 AM daily
  
  script:
    - naeos policy eval --all --project .
    - naeos policy report --format json --output report.json
    - curl -X POST https://reporting.company.com/api/reports -d @report.json
```

### 2.5 Continuous Evaluation

Always-on monitoring:

```go
// Go program running in production
func MonitorCompliance() {
    ticker := time.NewTicker(1 * time.Hour)
    defer ticker.Stop()
    
    for range ticker.C {
        result := engine.Evaluate(context)
        
        if !result.Compliant {
            alert.Send(Alert{
                Level: CRITICAL,
                Policy: result.FailedPolicy,
                Message: result.Message,
            })
        }
        
        metrics.RecordCompliance(result)
    }
}
```

---

## 3. Enforcement Mechanisms

### 3.1 Blocking Enforcement

Policy violation prevents action:

```go
// Block if policy fails
if !engine.Evaluate(ctx).Compliant {
    return errors.New("Policy violation: " + ctx.FailedPolicy)
}

// Proceed with action
performAction()
```

**Implementation in different contexts:**

| Context | Block Method |
|---------|--------------|
| Git | Pre-commit hook, branch protection |
| CI/CD | Pipeline stage failure |
| Deployment | Kubernetes admission webhook |
| Runtime | Service enforcer |
| API | Request interceptor |

### 3.2 Gradual Enforcement (Soft → Hard)

Roll out enforcement gradually:

```
Phase 1 (Weeks 1-4): Reporting only
  - Evaluate policies
  - Generate reports
  - Alert teams
  - No blocking

Phase 2 (Weeks 5-8): Soft enforcement
  - Evaluate policies
  - Warning if fail
  - Can override with approval
  - Still not blocking

Phase 3 (Weeks 9+): Hard enforcement
  - Evaluate policies
  - Block if fail
  - Exception only with CISO approval
```

```bash
# Configuration for gradual enforcement
naeos policy config policy-id \
  --enforcement-mode "reporting" \
  --scheduled-upgrade "2026-08-10:soft" \
  --scheduled-upgrade "2026-09-10:blocking"
```

### 3.3 Exception-Based Enforcement

Allow exceptions with proper controls:

```go
// Check if exception applies
if engine.HasException(ctx.PolicyID, ctx.Target) {
    exception := engine.GetException(ctx.PolicyID, ctx.Target)
    
    if exception.IsValid() && !exception.IsExpired() {
        // Proceed with exception
        return nil
    }
}

// No valid exception, evaluate normally
if !engine.Evaluate(ctx).Compliant {
    return errors.New("Policy violation and no valid exception")
}
```

---

## 4. Evaluation Workflow

### 4.1 Simple Workflow

```
Input: Policy, Target
  ↓
Check conditions (does policy apply?)
  ├─ No → Skip
  ├─ Yes → Proceed
  ↓
Load evaluation tools
  ↓
Execute evaluation
  ↓
Compare against baseline
  ├─ Pass → Success
  ├─ Fail → Handle violation
  ↓
Generate report
  ↓
Log audit trail
  ↓
Output: Result
```

### 4.2 Complex Workflow (Multi-Policy)

```
Input: Multiple policies, Target
  ↓
Load all policies
  ↓
Filter applicable policies
  ↓
Prioritize (critical first)
  ↓
For each policy:
  ├─ Execute evaluation
  ├─ Check result
  ├─ If critical fail → Short-circuit
  ├─ Otherwise → Continue
  ↓
Aggregate results
  ↓
Conflict detection (conflicting policies?)
  ├─ Yes → Resolve
  ├─ No → Use combined result
  ↓
Generate comprehensive report
  ↓
Output: Aggregated result
```

---

## 5. Integration with CI/CD

### 5.1 GitHub Actions Integration

```yaml
name: Policy Enforcement

on:
  pull_request:
  push:
    branches: [main]

jobs:
  policy-check:
    runs-on: ubuntu-latest
    
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup NAEOS
        uses: naeos-foundation/setup-naeos@v1
        with:
          version: latest
      
      - name: Run security policies
        id: security
        continue-on-error: true
        run: |
          naeos policy eval sec-* \
            --profile enterprise \
            --project . \
            --format json \
            --output security-report.json
      
      - name: Run testing policies
        id: testing
        continue-on-error: true
        run: |
          naeos policy eval test-* \
            --profile enterprise \
            --project . \
            --format json \
            --output testing-report.json
      
      - name: Run code quality policies
        id: quality
        continue-on-error: true
        run: |
          naeos policy eval code-* \
            --profile enterprise \
            --project . \
            --format json \
            --output quality-report.json
      
      - name: Comment on PR
        if: github.event_name == 'pull_request'
        uses: actions/github-script@v6
        with:
          script: |
            const fs = require('fs');
            const security = JSON.parse(fs.readFileSync('security-report.json'));
            const testing = JSON.parse(fs.readFileSync('testing-report.json'));
            const quality = JSON.parse(fs.readFileSync('quality-report.json'));
            
            const comment = `
            ## Policy Evaluation Results
            
            ### Security: ${security.compliance_percentage}%
            - Passed: ${security.passed}
            - Failed: ${security.failed}
            
            ### Testing: ${testing.compliance_percentage}%
            - Passed: ${testing.passed}
            - Failed: ${testing.failed}
            
            ### Code Quality: ${quality.compliance_percentage}%
            - Passed: ${quality.passed}
            - Failed: ${quality.failed}
            `;
            
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: comment
            });
      
      - name: Fail if critical policies failed
        if: steps.security.outcome == 'failure'
        run: exit 1
```

### 5.2 GitLab CI Integration

```yaml
stages:
  - policy-check
  - build
  - test
  - deploy

policy_evaluation:
  stage: policy-check
  image: naeos:latest
  
  script:
    - naeos policy eval --all --project . --format json -o report.json
  
  artifacts:
    reports:
      policy: report.json
    paths:
      - report.json
    expire_in: 30 days
  
  allow_failure: false

policy_report:
  stage: policy-check
  needs: ["policy_evaluation"]
  image: naeos:latest
  
  script:
    - naeos policy report --profile enterprise --format html -o report.html
  
  artifacts:
    paths:
      - report.html
```

---

## 6. Kubernetes Admission Webhook

Enforce policies at deployment time:

```go
// admission_webhook.go

func HandleAdmissionRequest(w http.ResponseWriter, r *http.Request) {
    // Parse admission review
    admissionReview := parseAdmissionReview(r)
    
    // Extract deployment from request
    deployment := admissionReview.Request.Object
    
    // Evaluate policies
    result := engine.Evaluate(Context{
        Type: "deployment",
        Object: deployment,
    })
    
    // Build admission response
    admissionResponse := &admissionv1.AdmissionResponse{
        UID: admissionReview.Request.UID,
        Allowed: result.Compliant,
    }
    
    if !result.Compliant {
        admissionResponse.Result = &metav1.Status{
            Message: fmt.Sprintf(
                "Policy violation: %s\n%v",
                result.FailedPolicy,
                result.Violations,
            ),
        }
    }
    
    // Send response
    respondWithJSON(w, admissionResponse)
}
```

**Kubernetes webhook configuration:**

```yaml
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingWebhookConfiguration
metadata:
  name: policy-enforcement
webhooks:
  - name: policy.naeos.io
    clientConfig:
      service:
        name: policy-webhook
        namespace: naeos-system
        path: "/admit"
      caBundle: <base64-encoded-ca>
    
    rules:
      - operations: ["CREATE", "UPDATE"]
        apiGroups: ["apps"]
        apiVersions: ["v1"]
        resources: ["deployments"]
    
    admissionReviewVersions: ["v1"]
    sideEffects: None
    failurePolicy: Fail
    timeoutSeconds: 5
```

---

## 7. Remediation

### 7.1 Auto-Remediation

Automatically fix policy violations:

```go
type AutoRemediator struct {
    policies []Policy
}

func (r *AutoRemediator) Remediate(ctx Context) error {
    result := engine.Evaluate(ctx)
    
    for _, violation := range result.Violations {
        policy := r.FindPolicy(violation.PolicyID)
        
        if policy.CanAutoRemediate() {
            if err := policy.Remediation(ctx, violation); err != nil {
                return fmt.Errorf("Remediation failed for %s: %w", 
                    violation.PolicyID, err)
            }
            
            // Verify remediation worked
            if !policy.Evaluate(ctx) {
                return fmt.Errorf("Remediation failed verification for %s",
                    violation.PolicyID)
            }
        }
    }
    
    return nil
}
```

**Example: Auto-remediation for code formatting**

```go
// Policy: code-format
// Remediation:

func RemediateFormatting(ctx Context) error {
    // Run formatter
    cmd := exec.Command("gofmt", "-w", ctx.ProjectPath)
    if err := cmd.Run(); err != nil {
        return err
    }
    
    // Stage changes
    cmd = exec.Command("git", "add", ".")
    if err := cmd.Run(); err != nil {
        return err
    }
    
    // Commit
    cmd = exec.Command("git", "commit", "-m", "fix: apply code formatting")
    if err := cmd.Run(); err != nil {
        // Might fail if no changes, that's OK
    }
    
    return nil
}
```

### 7.2 Assisted Remediation

Guide users to fix violations:

```json
{
  "policy": "test-coverage-80",
  "status": "failed",
  "current_coverage": 75,
  "required_coverage": 80,
  "gap": 5,
  
  "remediation_steps": [
    {
      "step": 1,
      "action": "Write tests for uncovered code",
      "file": "src/payment/processor.go",
      "lines": [45, 67, 89],
      "estimated_effort": "2 hours"
    },
    {
      "step": 2,
      "action": "Run coverage report",
      "command": "coverage run -m pytest"
    },
    {
      "step": 3,
      "action": "Verify coverage",
      "command": "coverage report",
      "expected": ">= 80%"
    }
  ]
}
```

---

## 8. Monitoring & Alerting

### 8.1 Compliance Dashboard

```go
type ComplianceDashboard struct {
    metrics map[string]float64
}

func (d *ComplianceDashboard) Update() {
    policies := engine.GetAllPolicies()
    
    for _, policy := range policies {
        result := engine.Evaluate(policy)
        
        d.metrics[policy.ID] = float64(result.Passed) / 
                               float64(result.Total) * 100
    }
}
```

**Metrics:**
- % compliant by policy
- % compliant by team
- Trend over time
- Time to remediation
- Exception count

### 8.2 Alerts

```yaml
# alert-rules.yaml

alerts:
  - name: PolicyComplianceDrop
    condition: "compliance < 90"
    threshold_duration: "10 minutes"
    action:
      - send_slack: "#alerts"
      - page_on_call: true
  
  - name: CriticalPolicyFailed
    condition: "policy.priority = CRITICAL and status = FAILED"
    threshold_duration: "1 minute"
    action:
      - page_on_call: true
      - create_incident: true
  
  - name: ExceptionExpiringSoon
    condition: "exception.expires_in < 7 days"
    frequency: "daily"
    action:
      - send_email: "compliance-team"
```

---

## 9. Reporting

### 9.1 Compliance Report

```bash
# Generate compliance report
naeos policy report \
  --profile enterprise \
  --format html \
  --output report.html

# Output includes:
# - Compliance percentage
# - Failed policies
# - Remediation recommendations
# - Trend over time
# - Risk assessment
```

### 9.2 Audit Trail Report

```bash
# Export audit trail
naeos policy audit-trail \
  --policy sec-tls-13 \
  --from 2026-06-01 \
  --to 2026-07-10 \
  --format csv \
  --output audit.csv
```

---

## 10. References

- [NAEOS-POL-001.md](NAEOS-POL-001.md) - Policy Compiler
- [NAEOS-POL-002.md](NAEOS-POL-002.md) - Policy Definition
- [NAEOS-POL-003.md](NAEOS-POL-003.md) - Policy Examples
- [NAEOS-POL-004.md](NAEOS-POL-004.md) - Policy Best Practices
- [NAEOS-POL-005.md](NAEOS-POL-005.md) - Policy Compiler & Engine
