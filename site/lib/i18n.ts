import en from "./i18n/en.json";
import id from "./i18n/id.json";
import type { Lang } from "./site";

const dictionaries: Record<Lang, Record<string, string>> = { en, id };

type InterpolateValues = Record<string, string | number>;

export function t(
  lang: Lang,
  key: string,
  values?: InterpolateValues,
): string {
  let str = dictionaries[lang][key] ?? dictionaries.en[key] ?? key;
  if (values) {
    for (const [k, v] of Object.entries(values)) {
      const val = String(v);
      str = str.replaceAll(`{{.${k}}}`, val);
      str = str.replaceAll(`{{ .${k} }}`, val);
      str = str.replaceAll(`{{ ${k} }}`, val);
    }
  }
  return str;
}

export function hasTranslation(lang: Lang, key: string): boolean {
  return key in dictionaries[lang];
}

export function createTranslator(lang: Lang) {
  return (key: string, values?: InterpolateValues) => t(lang, key, values);
}
