<script lang="ts">
  import type { SavedExecutiveReport } from '../types';
  import StudioIcon from './StudioIcon.svelte';

  let {
    reports = [],
    activeReportId = null,
    preparingDomain = null,
    onSelectReport,
    onDeleteReport,
  }: {
    reports: SavedExecutiveReport[];
    activeReportId?: string | null;
    preparingDomain?: string | null;
    onSelectReport: (report: SavedExecutiveReport) => void;
    onDeleteReport: (id: string) => void;
  } = $props();

  let isExpanded = $state(true);
</script>

{#if reports.length > 0 || preparingDomain}
  <aside
    aria-label="Generated Domain Dossier Reports Dock"
    class="fixed right-5 bottom-0 z-40 w-80 sm:w-96 rounded-t-2xl border-2 border-b-0 border-[#19231f] bg-[#fffdf8] shadow-[0_-10px_35px_rgba(25,35,31,0.18)] transition-all print:hidden"
  >
    <!-- LinkedIn-Style Dock Header Bar -->
    <div
      role="button"
      tabindex="0"
      onclick={() => (isExpanded = !isExpanded)}
      onkeydown={(e) => {
        if (e.key === 'Enter' || e.key === ' ') isExpanded = !isExpanded;
      }}
      class="flex cursor-pointer items-center justify-between rounded-t-xl bg-[#19231f] px-4 py-3 text-[#fffdf8]"
    >
      <div class="flex items-center gap-2.5">
        <span class="relative flex h-2.5 w-2.5">
          {#if preparingDomain}
            <span class="absolute inline-flex h-full w-full animate-ping rounded-full bg-[#dffc78] opacity-75"></span>
          {/if}
          <span class="relative inline-flex h-2.5 w-2.5 rounded-full bg-[#dffc78]"></span>
        </span>
        <span class="font-display text-xs font-bold tracking-wide">
          Domain Reports ({reports.length})
        </span>
        {#if preparingDomain}
          <span class="rounded-full bg-[#5366e8] px-2 py-0.5 font-display text-[10px] font-bold text-[#fffdf8]">
            Building {preparingDomain}...
          </span>
        {/if}
      </div>

      <span class="rounded-lg bg-[#fffdf8]/10 px-2 py-0.5 font-display text-[10px] font-bold text-[#dffc78]">
        {isExpanded ? 'Minimize' : 'Expand'}
      </span>
    </div>

    {#if isExpanded}
      <div class="max-h-72 overflow-y-auto divide-y divide-[#19231f]/10 p-2">
        {#if preparingDomain}
          <div class="flex items-center gap-3 rounded-xl border border-[#5366e8]/30 bg-[#d9d6fc]/30 p-3">
            <span class="h-3 w-3 shrink-0 animate-spin rounded-full border-2 border-[#19231f] border-t-transparent"></span>
            <div class="min-w-0 flex-1">
              <p class="truncate font-display text-xs font-bold text-[#19231f]">
                Preparing Full Report · {preparingDomain}
              </p>
              <p class="text-[11px] text-[#19231f]/65">
                Running all 7 engines in parallel (Valuation, RDAP, Traffic, Recon, Crawl, Robots, Meta)...
              </p>
            </div>
          </div>
        {/if}

        {#each reports as item (item.id)}
          <div
            class="group flex items-center justify-between gap-2 rounded-xl p-2.5 transition-colors {activeReportId === item.id
              ? 'bg-[#dffc78]/45'
              : 'hover:bg-[#f4f1e9]'}"
          >
            <button
              type="button"
              onclick={() => onSelectReport(item)}
              class="flex min-w-0 flex-1 items-center gap-3 text-left"
            >
              <span class="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl border border-[#19231f] bg-[#19231f] font-display text-xs font-bold text-[#dffc78]">
                {item.suite.recon?.securityGrade || 'A'}
              </span>
              <div class="min-w-0 flex-1">
                <div class="flex items-center gap-1.5">
                  <span class="truncate font-display text-xs font-bold text-[#19231f]">
                    {item.domain}
                  </span>
                  <span class="rounded bg-[#d9d6fc] px-1.5 py-0.5 font-display text-[9px] font-bold uppercase text-[#19231f]">
                    PDF Dossier
                  </span>
                </div>
                <p class="truncate text-[11px] text-[#19231f]/65">
                  {item.suite.inquiry?.valuation?.bandLabel || '$10/yr'} · {item.suite.traffic?.estimatedMonthlyRange || 'Traffic Checked'}
                </p>
              </div>
            </button>

            <div class="flex items-center gap-1">
              <button
                type="button"
                onclick={() => onSelectReport(item)}
                class="rounded-lg border border-[#19231f] bg-[#19231f] px-2.5 py-1 font-display text-[10px] font-bold text-[#fffdf8] hover:bg-[#5366e8]"
              >
                Open
              </button>
              <button
                type="button"
                onclick={() => onDeleteReport(item.id)}
                aria-label="Remove report for {item.domain}"
                class="rounded-lg p-1 text-[#19231f]/40 hover:bg-[#ffc3a5]/50 hover:text-[#19231f]"
              >
                <StudioIcon name="close" class="h-3.5 w-3.5" />
              </button>
            </div>
          </div>
        {/each}
      </div>
    {/if}
  </aside>
{/if}
