// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

/**
 * Resolve a server-side Control Plane origin from an environment binding.
 *
 * Requiring an explicitly configured HTTPS origin prevents PR previews from
 * silently falling back to the production Control Plane.
 */
export function resolveControlPlaneUpstream(value: unknown): string | null {
  if (typeof value !== "string" || value.trim() === "") return null;

  try {
    const url = new URL(value.trim());
    if (
      url.protocol !== "https:" ||
      !url.hostname ||
      url.username ||
      url.password ||
      url.pathname !== "/" ||
      url.search ||
      url.hash
    ) {
      return null;
    }

    return url.origin;
  } catch {
    return null;
  }
}
