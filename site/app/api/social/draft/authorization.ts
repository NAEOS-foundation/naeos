// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

// Node native TypeScript runner requires the explicit extension.
// @ts-ignore TS5097: runtime resolution is handled by Node's strip-types loader.
import { evaluateHumanApproval } from './human-approval.ts';

export type AuthorizationDecision = "authorized" | "deny";

export type AuthorizationTarget = {
  system: string;
  resource: string;
};

export type AuthorizationApproval = {
  approvalId: string;
  approverId: string;
  approverIdentity: string;
  decision: "approve" | "reject";
  approvedAt: string;
  proposalId: string;
  action: string;
  target: AuthorizationTarget;
  policyVersion: string;
  decisionId: string;
};

export type AuthorizationRequest = {
  proposalId: string;
  action: string;
  target: AuthorizationTarget;
  policyVersion: string;
  requiredPolicyVersion: string;
  policyDecision: "review_required" | "deny";
  decisionId: string;
  approval?: AuthorizationApproval;
};

export type AuthorizationResult = {
  contractVersion: "authorization-v1";
  decision: AuthorizationDecision;
  authorized: boolean;
  reasons: string[];
};

function sameTarget(left: AuthorizationTarget, right: AuthorizationTarget): boolean {
  return left.system === right.system && left.resource === right.resource;
}

export function evaluateAuthorization(input: AuthorizationRequest): AuthorizationResult {
  const reasons: string[] = [];

  if (input.policyDecision !== "review_required") {
    reasons.push("policy-denied");
  }

  if (input.policyVersion !== input.requiredPolicyVersion) {
    reasons.push("policy-version-mismatch");
  }

  if (!input.approval) {
    reasons.push("approval-missing");
  } else {
    const humanApproval = evaluateHumanApproval(input.approval);
    if (!humanApproval.accepted) {
      reasons.push("human-approval-invalid", ...humanApproval.reasons);
    }
    if (input.approval.proposalId !== input.proposalId) reasons.push("approval-proposal-mismatch");
    if (input.approval.action !== input.action) reasons.push("approval-action-mismatch");
    if (!sameTarget(input.approval.target, input.target)) reasons.push("approval-target-mismatch");
    if (input.approval.policyVersion !== input.policyVersion) reasons.push("approval-policy-version-mismatch");
    if (input.approval.decisionId !== input.decisionId) reasons.push("approval-decision-mismatch");
  }

  const authorized = reasons.length === 0;

  return {
    contractVersion: "authorization-v1",
    decision: authorized ? "authorized" : "deny",
    authorized,
    reasons,
  };
}
