<script lang="ts">
  import { onMount } from 'svelte';
  import { animate } from 'motion';

  interface Props {
    title: string;
    target: string;
    workers: number;
    steps: string[];
  }

  let { title, target, workers, steps }: Props = $props();
  let progressBarEl = $state<HTMLElement | null>(null);
  let activeStepIndex = $state(0);
  let elapsedMs = $state(0);

  onMount(() => {
    const start = performance.now();
    const timer = setInterval(() => {
      elapsedMs = Math.round(performance.now() - start);
      activeStepIndex = Math.min(steps.length - 1, Math.floor(elapsedMs / 650));
    }, 80);

    let animControl: ReturnType<typeof animate> | undefined;
    if (progressBarEl && !window.matchMedia('(prefers-reduced-motion: reduce)').matches) {
      animControl = animate(
        progressBarEl,
        { width: ['14%', '88%'] },
        { duration: 2.8, easing: [0.22, 1, 0.36, 1] }
      );
    }

    return () => {
      clearInterval(timer);
      animControl?.stop();
    };
  });
</script>

<div
  role="status"
  aria-live="polite"
  class="bento-card-ink p-5 sm:p-6 space-y-4 shadow-xl border-2 border-[#19231f]"
>
  <div class="flex flex-wrap items-center justify-between gap-3">
    <div class="flex items-center gap-3">
      <span class="relative flex h-3 w-3">
        <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-[#dffc78] opacity-75"></span>
        <span class="relative inline-flex rounded-full h-3 w-3 bg-[#dffc78]"></span>
      </span>
      <div>
        <div class="text-xs font-mono uppercase tracking-widest text-[#dffc78]">
          Active Go Worker Pool · {workers} Parallel Goroutines
        </div>
        <h3 class="font-display font-bold text-lg text-[#fffdf8]">
          {title} — <span class="font-serif-editorial font-normal text-xl text-[#dffc78]">{target}</span>
        </h3>
      </div>
    </div>

    <div class="px-3.5 py-1 rounded-full bg-[#fffdf8]/10 border border-[#fffdf8]/20 font-mono text-xs text-[#dffc78]">
      {(elapsedMs / 1000).toFixed(2)}s elapsed
    </div>
  </div>

  <!-- Motion One Animated Progress Bar -->
  <div class="h-2.5 w-full rounded-full bg-[#fffdf8]/15 overflow-hidden p-0.5">
    <div
      bind:this={progressBarEl}
      class="h-full rounded-full bg-[#dffc78]"
      style="width: 42%"
    ></div>
  </div>

  <!-- Step-by-Step Pipeline Telemetry -->
  <div class="grid grid-cols-1 sm:grid-cols-3 gap-2.5 pt-1">
    {#each steps as step, idx}
      {@const isDone = idx < activeStepIndex}
      {@const isCurrent = idx === activeStepIndex}
      <div
        class="rounded-xl px-3.5 py-2 border text-xs font-mono flex items-center gap-2 transition-colors {isCurrent
          ? 'bg-[#dffc78] text-[#19231f] border-[#dffc78] font-bold'
          : isDone
            ? 'bg-[#fffdf8]/10 text-[#fffdf8] border-[#fffdf8]/20'
            : 'bg-transparent text-[#fffdf8]/45 border-[#fffdf8]/10'}"
      >
        <span>{isDone ? '✓' : isCurrent ? '◈' : '○'}</span>
        <span class="truncate">{step}</span>
      </div>
    {/each}
  </div>
</div>
