// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

export type ExternalPublisherTarget = {
  system: string;
  resource: string;
};

export type ExternalPublisherAction = {
  proposalId: string;
  action: string;
  target: ExternalPublisherTarget;
  policyVersion: string;
  decisionId: string;
};

export type ExternalPublisherAuthorization = {
  contractVersion: string;
  authorized: boolean;
};

export type ExternalPublishRequest = {
  executionId: string;
  provider: string;
  authorization: ExternalPublisherAuthorization;
  action: ExternalPublisherAction;
};

export type ExternalPublishPreparation = {
  contractVersion: "external-publisher-v1";
  ready: boolean;
  externalEffect: false;
  executionId: string;
  provider: string;
  proposalId: string;
  action: string;
  target: ExternalPublisherTarget;
  policyVersion: string;
  decisionId: string;
  reasons: string[];
};

export type ExternalProviderReceipt = {
  provider: string;
  receiptId: string;
  status: "accepted" | "rejected" | "unknown";
  observedAt: string;
  externalId?: string;
};

export interface ExternalPublisherAdapter {
  publish(request: ExternalPublishRequest): Promise<ExternalProviderReceipt>;
}

function present(value: string): boolean {
  return value.trim().length > 0;
}

export function prepareExternalPublish(
  input: ExternalPublishRequest,
): ExternalPublishPreparation {
  const reasons: string[] = [];

  if (!present(input.executionId)) reasons.push("execution-id-missing");
  if (!present(input.provider)) reasons.push("provider-missing");
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
    contractVersion: "external-publisher-v1",
    ready: reasons.length === 0,
    externalEffect: false,
    executionId: input.executionId,
    provider: input.provider,
    proposalId: input.action.proposalId,
    action: input.action.action,
    target: input.action.target,
    policyVersion: input.action.policyVersion,
    decisionId: input.action.decisionId,
    reasons,
  };
}

export function isProviderReceiptValid(
  receipt: ExternalProviderReceipt,
  expectedProvider: string,
): boolean {
  return (
    present(expectedProvider) &&
    receipt.provider === expectedProvider &&
    present(receipt.receiptId) &&
    present(receipt.observedAt) &&
    receipt.status !== "unknown"
  );
}
