// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

import {
  prepareExternalPublish,
  type ExternalProviderReceipt,
  type ExternalPublishRequest,
} from "./external-publisher.ts";

export const LINKEDIN_PROVIDER = "linkedin";
export const LINKEDIN_API_VERSION = "202608";
export const LINKEDIN_POSTS_ENDPOINT = "https://api.linkedin.com/rest/posts";

export type LinkedInCredentialProvider = {
  getAccessToken(): Promise<string>;
};

export type LinkedInPostInput = ExternalPublishRequest & {
  authorUrn: string;
  commentary: string;
};

export type LinkedInPublisherDependencies = {
  credentials: LinkedInCredentialProvider;
  fetchImpl?: typeof fetch;
  revalidateAuthorization: (
    request: ExternalPublishRequest,
  ) => Promise<boolean>;
};

function present(value: string): boolean {
  return value.trim().length > 0;
}

function isLinkedInOrganizationUrn(value: string): boolean {
  return /^urn:li:organization:\d+$/.test(value);
}

function toReceipt(
  input: LinkedInPostInput,
  response: Response,
): ExternalProviderReceipt {
  const receiptId =
    response.headers.get("x-restli-id") ??
    response.headers.get("x-restli-id".toLowerCase()) ??
    "";

  return {
    provider: LINKEDIN_PROVIDER,
    receiptId,
    status:
      response.status >= 200 && response.status < 300
        ? "accepted"
        : response.status >= 400
          ? "rejected"
          : "unknown",
    observedAt: new Date().toISOString(),
    ...(receiptId ? { externalId: receiptId } : {}),
  };
}

export async function publishLinkedInPost(
  input: LinkedInPostInput,
  dependencies: LinkedInPublisherDependencies,
): Promise<ExternalProviderReceipt> {
  const preparation = prepareExternalPublish({
    ...input,
    provider: LINKEDIN_PROVIDER,
  });

  if (!preparation.ready) {
    return {
      provider: LINKEDIN_PROVIDER,
      receiptId: "",
      status: "rejected",
      observedAt: new Date().toISOString(),
    };
  }

  if (!isLinkedInOrganizationUrn(input.authorUrn)) {
    return {
      provider: LINKEDIN_PROVIDER,
      receiptId: "",
      status: "rejected",
      observedAt: new Date().toISOString(),
    };
  }

  if (!present(input.commentary)) {
    return {
      provider: LINKEDIN_PROVIDER,
      receiptId: "",
      status: "rejected",
      observedAt: new Date().toISOString(),
    };
  }

  // Last-controllable-boundary revalidation: an earlier authorization result
  // is not sufficient evidence to perform an irreversible provider call.
  if (!(await dependencies.revalidateAuthorization(input))) {
    return {
      provider: LINKEDIN_PROVIDER,
      receiptId: "",
      status: "rejected",
      observedAt: new Date().toISOString(),
    };
  }

  const accessToken = await dependencies.credentials.getAccessToken();
  if (!present(accessToken)) {
    return {
      provider: LINKEDIN_PROVIDER,
      receiptId: "",
      status: "rejected",
      observedAt: new Date().toISOString(),
    };
  }

  const fetchImpl = dependencies.fetchImpl ?? fetch;
  const response = await fetchImpl(LINKEDIN_POSTS_ENDPOINT, {
    method: "POST",
    headers: {
      Authorization: `Bearer ${accessToken}`,
      "Content-Type": "application/json",
      "Linkedin-Version": LINKEDIN_API_VERSION,
      "X-Restli-Protocol-Version": "2.0.0",
    },
    body: JSON.stringify({
      author: input.authorUrn,
      commentary: input.commentary,
      visibility: "PUBLIC",
      distribution: {
        feedDistribution: "MAIN_FEED",
        targetEntities: [],
        thirdPartyDistributionChannels: [],
      },
      lifecycleState: "PUBLISHED",
      isReshareDisabledByAuthor: false,
    }),
  });

  return toReceipt(input, response);
}
