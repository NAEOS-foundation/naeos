// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

import {
  publishLinkedInPost,
  type LinkedInPostInput,
  type LinkedInPublisherDependencies,
} from "./linkedin-publisher.ts";
import type { ExternalProviderReceipt } from "./external-publisher.ts";

export type LinkedInExecutionClaim =
  | { status: "acquired" }
  | { status: "existing"; receipt: ExternalProviderReceipt }
  | { status: "in-flight" };

export type LinkedInExecutionReceiptStore = {
  claim(executionKey: string): Promise<LinkedInExecutionClaim>;
  put(executionKey: string, receipt: ExternalProviderReceipt): Promise<void>;
  release(executionKey: string): Promise<void>;
};

export type LinkedInRuntimeDependencies = LinkedInPublisherDependencies & {
  receiptStore: LinkedInExecutionReceiptStore;
};

export function linkedinExecutionKey(input: LinkedInPostInput): string {
  return [
    "linkedin-execution-v1",
    input.executionId,
    input.provider,
    input.action.proposalId,
    input.action.action,
    input.action.target.system,
    input.action.target.resource,
    input.action.policyVersion,
    input.action.decisionId,
  ].map((value) => JSON.stringify(value)).join(":");
}

/**
 * Last-mile runtime wrapper.
 *
 * The store is intentionally injected: production must provide durable storage
 * appropriate to the deployment. The runtime never treats a provider receipt
 * as authorization and never retries an already-observed execution implicitly.
 *
 * claim must be atomic in the production store. A concurrent caller that
 * cannot observe an existing receipt receives an in-flight result instead of
 * issuing a second provider request.
 */
export async function executeLinkedInPublication(
  input: LinkedInPostInput,
  dependencies: LinkedInRuntimeDependencies,
): Promise<ExternalProviderReceipt> {
  const executionKey = linkedinExecutionKey(input);
  const claim = await dependencies.receiptStore.claim(executionKey);

  if (claim.status === "existing") {
    return claim.receipt;
  }

  if (claim.status === "in-flight") {
    throw new Error(`LinkedIn execution already in progress: ${executionKey}`);
  }

  try {
    const receipt = await publishLinkedInPost(input, dependencies);
    await dependencies.receiptStore.put(executionKey, receipt);
    return receipt;
  } catch (error) {
    await dependencies.receiptStore.release(executionKey);
    throw error;
  }
}
