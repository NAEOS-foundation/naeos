// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

import { NextResponse, type NextRequest } from "next/server";
import { DEFAULT_LANG } from "@/lib/site";

const PUBLIC_PREFIXES = ["/_next", "/api", "/images", "/downloads"];

function isPublicAsset(pathname: string): boolean {
  if (PUBLIC_PREFIXES.some((p) => pathname === p || pathname.startsWith(`${p}/`))) {
    return true;
  }
  const last = pathname.split("/").pop() ?? "";
  return last.includes(".");
}

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;

  // Locale-prefixed routes are real App Router routes. Do not redirect /en
  // back to / because the root route is internally rewritten to /en below.
  // Redirecting here can make the internal rewrite re-enter middleware and
  // produce a root/en routing loop in OpenNext/Cloudflare.
  if (
    pathname === "/en" ||
    pathname.startsWith("/en/") ||
    pathname === "/id" ||
    pathname.startsWith("/id/")
  ) {
    return NextResponse.next();
  }

  if (isPublicAsset(pathname)) {
    return NextResponse.next();
  }

  const url = request.nextUrl.clone();
  url.pathname = `/${DEFAULT_LANG}${pathname === "/" ? "" : pathname}`;
  return NextResponse.rewrite(url);
}

export const config = {
  matcher: ["/((?!_next/static|_next/image).*)"],
};
