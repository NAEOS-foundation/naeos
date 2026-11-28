// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

import type { Metadata } from "next";
import FoundingPartnerReviewDashboard from "@/components/partners/FoundingPartnerReviewDashboard";

export async function generateMetadata(props: { params: Promise<{ lang: string }> }): Promise<Metadata> {
  const { lang } = await props.params;
  const id = lang === "id";
  return {
    title: id ? "Review Founding Partner — NAEOS" : "Founding Partner Review — NAEOS",
    description: id ? "Workspace review internal untuk founder NAEOS." : "Restricted NAEOS founding partner review workspace.",
    robots: { index: false, follow: false },
  };
}

export default async function FoundingPartnerReviewPage(props: { params: Promise<{ lang: string }> }) {
  const { lang } = await props.params;
  return <FoundingPartnerReviewDashboard lang={lang === "id" ? "id" : "en"} />;
}
