// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

export type PublisherTarget = {
  system: string;
  resource: string;
};

export type PublisherAction = {
  proposalId: string;
  action: string;
  target: PublisherTarget;
  policyVersion: string;
  decisionId: string;
};

export type PublisherAuthorization = {
  contractVersion: "authorization-v1";
  authorized: boolean;
};

export type PublisherDryRunRequest = {
  receiptId: string;
  executedAt: string;
  authorization: PublisherAuthorization;
  action: PublisherAction;
};

export type PublisherDryRunReceipt = {
  receiptVersion: "publisher-receipt-v1";
  receiptId: string;
  mode: "dry-run";
  status: "simulated" | "rejected";
  executedAt: string;
  externalEffect: false;
  proposalId: string;
  action: string;
  target: PublisherTarget;
  policyVersion: string;
  decisionId: string;
  authorizationContractVersion: "authorization-v1";
  reasons: string[];
};

function present(value: string): boolean {
  return value.trim().length > 0;
}

export function runPublisherDryRun(input: PublisherDryRunRequest): PublisherDryRunReceipt {
  const reasons: string[] = [];

  if (!present(input.receiptId)) reasons.push("receipt-id-missing");
  if (!present(input.executedAt)) reasons.push("execution-time-missing");
  if (!input.authorization.authorized) reasons.push("authorization-required");
  if (input.authorization.contractVersion !== "authorization-v1") {
    reasons.push("authorization-contract-version-unsupported");
  }
  if (!present(input.action.proposalId)) reasons.push("proposal-id-missing");
  if (!present(input.action.action)) reasons.push("action-missing");
  if (!present(input.action.target.system)) reasons.push("target-system-missing");
  if (!present(input.action.target.resource)) reasons.push("target-resource-missing");
  if (!present(input.action.policyVersion)) reasons.push("policy-version-missing");
  if (!present(input.action.decisionId)) reasons.push("decision-id-missing");

  return {
    receiptVersion: "publisher-receipt-v1",
    receiptId: input.receiptId,
    mode: "dry-run",
    status: reasons.length === 0 ? "simulated" : "rejected",
    executedAt: input.executedAt,
    externalEffect: false,
    proposalId: input.action.proposalId,
    action: input.action.action,
    target: input.action.target,
    policyVersion: input.action.policyVersion,
    decisionId: input.action.decisionId,
    authorizationContractVersion: input.authorization.contractVersion,
    reasons,
  };
}
