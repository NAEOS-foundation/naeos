# NAEOS Campaign Experiment Pack

This directory contains execution metadata for the 30 Days of AI Agent Governance campaign.

The campaign is evidence-first: it maps public content to existing deterministic NAEOS experiments instead of creating marketing-only demonstrations.

## Source of truth
- docs/experiments/NAEOS-30-DAY-GOVERNANCE-CAMPAIGN.md
- experiments/README.md
- .github/ISSUE_TEMPLATE/marketing_experiment.md

## Day 01 acceptance test

Run:

    go run ./experiments/governance-lifecycle

Expected:
- exit code 0;
- lifecycle assertions pass;
- output is machine-readable JSON;
- the run demonstrates intent → policy → authorized action → execution → observation → evidence → independent verification.

## Publication rule
Do not publish a result until the experiment has been executed in the publication environment. Include the observed result and repository path. Do not turn a characterization result into a blanket production-security claim.

## Campaign principle
> Build the evidence. Publish the evidence. Invite independent reproduction.
