import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import rehypeRaw from "rehype-raw";
import rehypeSlug from "rehype-slug";
import rehypeAutolinkHeadings from "rehype-autolink-headings";
import rehypeHighlight from "rehype-highlight";
import Link from "next/link";
import { CodeBlock, MermaidDiagram } from "./ContentEffects";

interface MarkdownProps {
  content: string;
}

function extractLang(node: unknown): string {
  const className = (node as { properties?: { className?: unknown } })?.properties?.className;
  if (!Array.isArray(className)) return "";
  const langClass = className.find((c) => typeof c === "string" && c.startsWith("language-"));
  return typeof langClass === "string" ? langClass.replace("language-", "") : "";
}

export default function Markdown({ content }: MarkdownProps) {
  return (
    <ReactMarkdown
      remarkPlugins={[remarkGfm]}
      rehypePlugins={[
        rehypeRaw,
        rehypeSlug,
        [
          rehypeAutolinkHeadings,
          { behavior: "append", properties: { className: "anchor-link", ariaHidden: true } },
        ],
        [rehypeHighlight, { detect: true, ignoreMissing: true }],
      ]}
      components={{
        a({ href, children, ...props }) {
          if (!href) return <a {...props}>{children}</a>;
          const internal = href.startsWith("/") && !href.startsWith("//");
          if (internal) {
            return (
              <Link href={href.replace(/\/$/, "") === "" ? "/" : href.replace(/\/$/, "")}>
                {children}
              </Link>
            );
          }
          if (/^https?:\/\//.test(href)) {
            return (
              <a href={href} target="_blank" rel="noopener noreferrer" {...props}>
                {children}
              </a>
            );
          }
          return <a href={href} {...props}>{children}</a>;
        },
        img(props) {
          // eslint-disable-next-line @next/next/no-img-element
          return <img src={props.src} alt={props.alt ?? ""} loading="lazy" />;
        },
        pre({ children, ...props }) {
          void props;
          const child = Array.isArray(children) ? children[0] : children;
          const lang = extractLang(child);
          const codeText = extractText(child);
          if (lang === "mermaid") {
            return <MermaidDiagram chart={codeText.trim()} />;
          }
          return <CodeBlock lang={lang}>{children}</CodeBlock>;
        },
      }}
    >
      {content}
    </ReactMarkdown>
  );
}

function extractText(node: unknown): string {
  if (node == null || typeof node !== "object") return typeof node === "string" ? node : "";
  const el = node as { type?: string; value?: string; children?: unknown[] };
  if (el.type === "text") return el.value ?? "";
  if (Array.isArray(el.children)) return el.children.map(extractText).join("");
  return "";
}
