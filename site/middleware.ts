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

  if (
    pathname === "/id" ||
    pathname.startsWith("/id/") ||
    pathname === "/en" ||
    pathname.startsWith("/en/")
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
