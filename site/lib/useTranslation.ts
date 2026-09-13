"use client";
// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0


import { useMemo } from "react";
import enDict from "@/lib/i18n/en.json";
import idDict from "@/lib/i18n/id.json";
import type { Lang } from "@/lib/site";

type Dict = Record<string, string>;

const DICTS: Record<Lang, Dict> = {
  en: enDict as Dict,
  id: idDict as Dict,
};

export function useTranslation(lang: Lang) {
  return useMemo(() => {
    const dict = DICTS[lang] ?? DICTS.en;
    const t = (key: string, fallback?: string): string =>
      dict[key] ?? DICTS.en[key] ?? fallback ?? key;
    return { t, dict };
  }, [lang]);
}
