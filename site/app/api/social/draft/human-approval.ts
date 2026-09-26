// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

export type HumanApprovalDecision = "approve" | "reject";

export type HumanApprovalTarget = {
  system: string;
  resource: string;
};

export type HumanApprovalRequest = {
  approvalId: string;
  approverId: string;
  approverIdentity: string;
  decision: HumanApprovalDecision;
  approvedAt: string;
  proposalId: string;
  action: string;
  target: HumanApprovalTarget;
  policyVersion: string;
  decisionId: string;
};

export type HumanApprovalResult = {
  contractVersion: "human-approval-v1";
  accepted: boolean;
  reasons: string[];
  authenticated: false;
};

function present(value: string): boolean {
  return value.trim().length > 0;
}

export function evaluateHumanApproval(input: HumanApprovalRequest): HumanApprovalResult {
  const reasons: string[] = [];

  if (!present(input.approvalId)) reasons.push("approval-id-missing");
  if (!present(input.approverId)) reasons.push("approver-id-missing");
  if (!present(input.approverIdentity)) reasons.push("approver-identity-missing");
  if (!present(input.approvedAt)) reasons.push("approval-time-missing");
  if (!present(input.proposalId)) reasons.push("proposal-id-missing");
  if (!present(input.action)) reasons.push("action-missing");
  if (!present(input.target.system)) reasons.push("target-system-missing");
  if (!present(input.target.resource)) reasons.push("target-resource-missing");
  if (!present(input.policyVersion)) reasons.push("policy-version-missing");
  if (!present(input.decisionId)) reasons.push("decision-id-missing");

  if (input.decision !== "approve") reasons.push("human-rejected");

  return {
    contractVersion: "human-approval-v1",
    accepted: reasons.length === 0,
    reasons,
    authenticated: false,
  };
}
