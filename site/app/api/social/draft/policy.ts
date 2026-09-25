// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

export type SocialPolicyDecision = "review_required" | "deny";

export type SocialPolicyInput = {
  signals: string[];
  source: "github-live" | "unavailable";
};

export type SocialPolicyResult = {
  decision: SocialPolicyDecision;
  authorized: false;
  reasons: string[];
  policyVersion: "social-v1";
};

const CANDIDATE_SIGNALS = new Set(["release", "merged-pr", "bug", "ci-failure"]);

export function evaluateSocialPolicy(input: SocialPolicyInput): SocialPolicyResult {
  const reasons: string[] = [];

  if (input.source !== "github-live") {
    reasons.push("source-unavailable");
  }

  const supportedSignals = input.signals.filter((signal) => CANDIDATE_SIGNALS.has(signal));
  if (supportedSignals.length === 0) {
    reasons.push("no-publishable-signal");
  }

  if (supportedSignals.length > 0) {
    reasons.push("human-approval-required");
  }

  return {
    decision: supportedSignals.length > 0 && input.source === "github-live" ? "review_required" : "deny",
    authorized: false,
    reasons,
    policyVersion: "social-v1",
  };
}
