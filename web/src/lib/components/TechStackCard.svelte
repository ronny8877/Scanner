<script lang="ts">
  import type { TechStackTelemetry } from '../types';
  import StudioIcon from './StudioIcon.svelte';

  let {
    techStack = undefined,
    title = 'Detected Frontend Tech Stack, UI Design System & Infrastructure',
  }: {
    techStack?: TechStackTelemetry;
    title?: string;
  } = $props();

  function categoryBadge(cat: string): { label: string; bg: string } {
    switch (cat) {
      case 'FRAMEWORK':
        return { label: 'Framework', bg: 'bg-[#dffc78] text-[#19231f]' };
      case 'UI_DESIGN':
        return { label: 'UI / CSS System', bg: 'bg-[#d9d6fc] text-[#19231f]' };
      case 'BUILD_MOTION':
        return { label: 'Motion / Bundler', bg: 'bg-[#ffc3a5] text-[#19231f]' };
      case 'PLATFORM_AUTH':
        return { label: 'Platform / SDK', bg: 'bg-[#5366e8] text-[#fffdf8]' };
      case 'EDGE_HOSTING':
        return { label: 'Edge / Cloud', bg: 'bg-[#19231f] text-[#dffc78]' };
      default:
        return { label: cat, bg: 'bg-[#f4f1e9] text-[#19231f]' };
    }
  }
</script>

{#if techStack}
  <div class="bento-card p-6 space-y-5">
    <div class="flex flex-wrap items-start justify-between gap-4">
      <div>
        <div class="inline-flex items-center gap-2">
          <span class="rounded-full bg-[#19231f] px-2.5 py-0.5 font-display text-[10px] font-bold uppercase tracking-wider text-[#dffc78]">
            Tech Stack Fingerprint
          </span>
          <span class="font-mono text-xs text-[#48534e]">
            DOM + Linked Stylesheet CSS Tokens + HTTP Headers
          </span>
        </div>
        <h3 class="mt-2 font-display text-lg font-bold text-[#19231f]">
          {title}
        </h3>
        <p class="mt-1 text-xs text-[#48534e] leading-relaxed">
          {techStack.summary}
        </p>
      </div>

      <div class="flex flex-wrap items-center gap-2">
        <span class="rounded-xl border border-[#19231f] bg-[#dffc78] px-3.5 py-1.5 font-display text-xs font-bold text-[#19231f]">
          {techStack.totalDetected} Technologies Detected
        </span>
      </div>
    </div>

    <!-- Top 3 Summary Pills: Primary Framework, UI / CSS Design System, Edge Hosting -->
    <div class="grid grid-cols-1 gap-3 sm:grid-cols-3">
      <div class="rounded-2xl border border-[#19231f]/15 bg-[#f4f1e9]/75 p-3.5">
        <div class="flex items-center gap-1.5 text-[11px] font-display font-bold uppercase tracking-wider text-[#48534e]">
          <StudioIcon name="cpu" size={13} />
          <span>Frontend Framework</span>
        </div>
        <p class="mt-1.5 font-display text-sm font-bold text-[#19231f]">
          {techStack.primaryFramework}
        </p>
      </div>

      <div class="rounded-2xl border border-[#19231f]/15 bg-[#f4f1e9]/75 p-3.5">
        <div class="flex items-center gap-1.5 text-[11px] font-display font-bold uppercase tracking-wider text-[#48534e]">
          <StudioIcon name="layers" size={13} />
          <span>UI & CSS System (Tailwind / shadcn / DaisyUI)</span>
        </div>
        <p class="mt-1.5 font-display text-sm font-bold text-[#5366e8]">
          {techStack.uiSystemSummary}
        </p>
      </div>

      <div class="rounded-2xl border border-[#19231f]/15 bg-[#f4f1e9]/75 p-3.5">
        <div class="flex items-center gap-1.5 text-[11px] font-display font-bold uppercase tracking-wider text-[#48534e]">
          <StudioIcon name="globe" size={13} />
          <span>Edge & Cloud Delivery</span>
        </div>
        <p class="mt-1.5 font-display text-sm font-bold text-[#19231f]">
          {techStack.edgePlatform}
        </p>
      </div>
    </div>

    <!-- Detailed Technology Cards -->
    {#if techStack.technologies && techStack.technologies.length > 0}
      <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
        {#each techStack.technologies as tech}
          {@const badge = categoryBadge(tech.category)}
          <div class="flex flex-col justify-between rounded-2xl border border-[#19231f]/15 bg-[#fffdf8] p-4 shadow-xs">
            <div>
              <div class="flex items-center justify-between gap-2">
                <span class="font-display text-sm font-bold text-[#19231f]">
                  {tech.name}
                </span>
                <span class="rounded-full border border-[#19231f]/20 px-2.5 py-0.5 font-display text-[10px] font-bold {badge.bg}">
                  {badge.label}
                </span>
              </div>
              <p class="mt-1.5 text-xs text-[#48534e] leading-relaxed">
                {tech.description}
              </p>
            </div>

            <div class="mt-3 flex items-center justify-between gap-2 border-t border-[#19231f]/10 pt-2.5 text-[11px]">
              <span class="truncate font-mono text-[#19231f]/70" title={tech.matchedBy}>
                {tech.matchedBy}
              </span>
              <span class="shrink-0 rounded bg-[#f4f1e9] px-1.5 py-0.5 font-mono text-[10px] font-bold text-[#19231f]">
                {tech.confidence}
              </span>
            </div>
          </div>
        {/each}
      </div>
    {/if}
  </div>
{/if}
