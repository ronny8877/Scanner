import { animate, stagger } from 'motion';

function prefersReducedMotion(): boolean {
  if (typeof window === 'undefined') return false;
  return window.matchMedia('(prefers-reduced-motion: reduce)').matches;
}

/**
 * Svelte action (`use:motionCard={{ delay: 0.05 }}`) that animates a bento card
 * with Motion One (`motion`) while respecting reduced-motion settings.
 */
export function motionCard(node: HTMLElement, opts: { delay?: number; y?: number } = {}) {
  if (prefersReducedMotion()) return;

  const y = opts.y ?? 14;
  const delay = opts.delay ?? 0;

  node.style.opacity = '0';
  node.style.transform = `translateY(${y}px)`;

  const controls = animate(
    node,
    { opacity: [0, 1], transform: [`translateY(${y}px)`, 'translateY(0px)'] },
    { duration: 0.38, delay, easing: [0.22, 1, 0.36, 1] }
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
      { opacity: [0.35, 1], transform: ['translateY(6px)', 'translateY(0px)'] },
      { duration: 0.26, delay: stagger(0.025), easing: 'ease-out' }
    );
  };

  run();
}
