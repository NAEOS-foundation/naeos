"use client";

import { useEffect, useRef, useState } from "react";

export default function HeroEffects() {
  const containerRef = useRef<HTMLDivElement>(null);
  const [wsReady, setWsReady] = useState(false);

  useEffect(() => {
    const particles = containerRef?.current;
    if (!particles) return;
    if (window.matchMedia("(prefers-reduced-motion: reduce)").matches) return;
    for (let i = 0; i < 20; i++) {
      const particle = document.createElement("div");
      particle.className = "hero-particle";
      particle.style.left = `${Math.random() * 100}%`;
      particle.style.width = `${2 + Math.random() * 3}px`;
      particle.style.height = particle.style.width;
      particle.style.animationDelay = `${Math.random() * 8}s`;
      particle.style.animationDuration = `${6 + Math.random() * 6}s`;
      particles.appendChild(particle);
    }
    return () => {
      particles.innerHTML = "";
    };
  }, []);

  useEffect(() => {
    document.querySelectorAll<HTMLElement>(".terminal-line").forEach((line, i) => {
      line.style.animationDelay = `${i * 0.4 + 0.5}s`;
    });
  }, []);

  useEffect(() => {
    const wsUrl =
      (document.documentElement.dataset.wsUrl as string | undefined) ?? "";
    const target = document.getElementById("interactive-terminal");
    if (!wsUrl || !target || wsUrl === "disabled") return;

    let cancelled = false;
    let cleanup: (() => void) | undefined;

    import("@/components/home/terminalBridge")
      .then(({ startInteractiveTerminal }) => {
        if (cancelled) return;
        cleanup = startInteractiveTerminal(target, wsUrl, () => setWsReady(true));
      })
      .catch(() => {});

    return () => {
      cancelled = true;
      cleanup?.();
    };
  }, []);

  void wsReady;
  return (
    <div
      ref={containerRef}
      className="hero-particles"
      aria-hidden="true"
    />
  );
}
