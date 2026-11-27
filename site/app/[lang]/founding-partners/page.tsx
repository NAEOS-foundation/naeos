// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

import type { Metadata } from "next";
import FoundingPartnerIntake from "@/components/partners/FoundingPartnerIntake";

export async function generateMetadata(props: { params: Promise<{ lang: string }> }): Promise<Metadata> {
  const { lang } = await props.params;
  const id = lang === "id";
  return {
    title: id ? "Form Founding Partner NAEOS" : "NAEOS Founding Partner Application",
    description: id ? "Ajukan minat bergabung sebagai founding partner NAEOS." : "Apply to join the NAEOS founding partner team.",
    robots: { index: true, follow: true },
  };
}

export default async function FoundingPartnerPage(props: { params: Promise<{ lang: string }> }) {
  const { lang } = await props.params;
  return <FoundingPartnerIntake lang={lang === "id" ? "id" : "en"} />;
}
