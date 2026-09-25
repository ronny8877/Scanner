import { animate, stagger } from 'motion';

function prefersReducedMotion(): boolean {
  if (typeof window === 'undefined') return false;
  return window.matchMedia('(prefers-reduced-motion: reduce)').matches;
}

/**
 * Svelte action (`use:motionCard={{ delay: 0.05 }}`) that animates a bento card
 * subtly without ever setting opacity to 0 (preventing full-page tab-switch flicker).
 */
export function motionCard(node: HTMLElement, opts: { delay?: number; y?: number } = {}) {
  if (prefersReducedMotion()) return;

  const y = Math.min(opts.y ?? 6, 6);
  const delay = Math.min(opts.delay ?? 0, 0.04);

  const controls = animate(
    node,
    { opacity: [0.96, 1], transform: [`translateY(${y}px)`, 'translateY(0px)'] },
    { duration: 0.18, delay, easing: [0.22, 1, 0.36, 1] }
  );

  return {
    destroy() {
      controls.stop();
    },
  };
}

/**
 * Svelte action (`use:motionStaggerList`) that cascades children rows smoothly using Motion One stagger().
 */
export function motionStaggerList(node: HTMLElement) {
  if (prefersReducedMotion()) return;

  const run = () => {
    const children = Array.from(node.children) as HTMLElement[];
    if (!children.length) return;
    const targetChildren = children.slice(0, 18);
    animate(
      targetChildren,
      { opacity: [0.92, 1], transform: ['translateY(4px)', 'translateY(0px)'] },
      { duration: 0.16, delay: stagger(0.015), easing: 'ease-out' }
    );
  };

  run();
}
