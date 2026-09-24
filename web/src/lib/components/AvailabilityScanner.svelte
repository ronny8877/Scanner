<script lang="ts">
  import type { ScanReport, ScanResultItem } from '../types';
  import { motionCard } from '../motion';
  import LoadingProgressBanner from './LoadingProgressBanner.svelte';

  interface Props {
    report: ScanReport | null;
    loading: boolean;
    onRunScan: (opts: {
      keywords: string[];
      tlds: string[];
      mutations: boolean;
      onlyAvailable: boolean;
      minScore: number;
    }) => void;
    onInspectDomain: (domain: string) => void;
    onCheckHistory: (domain: string) => void;
    onRunRecon: (domain: string) => void;
    onCrawlDomain: (domain: string) => void;
    onToggleSave: (domain: string, available: boolean) => void;
    savedDomainsSet: Set<string>;
  }

  let {
    report,
    loading,
    onRunScan,
    onInspectDomain,
    onCheckHistory,
    onRunRecon,
    onCrawlDomain,
    onToggleSave,
    savedDomainsSet,
  }: Props = $props();

  let keywordInput = $state('veltrix, nova');
  let selectedTlds = $state<string[]>(['com', 'ai', 'io', 'dev', 'co', 'app']);
  let mutations = $state(true);
  let onlyAvailable = $state(false);
  let minScore = $state(0);
  let activeItem = $state<ScanResultItem | null>(null);

  const allTlds = ['com', 'ai', 'io', 'dev', 'co', 'app', 'net', 'xyz'];

  const presetSeeds = [
    { label: 'Unclaimed Tech Gems', value: 'veltrix, nexora' },
    { label: 'AI & Compute', value: 'synth, vector' },
    { label: 'Design & Craft', value: 'atelier, forma' },
    { label: 'Fintech & Ledger', value: 'kinetix, mint' },
  ];

  $effect(() => {
    if (report && report.items.length > 0) {
      if (!activeItem || !report.items.some((i) => i.domain === activeItem?.domain)) {
        activeItem = report.items[0];
      }
    }
  });

  function toggleTld(tld: string) {
    if (selectedTlds.includes(tld)) {
      if (selectedTlds.length > 1) {
        selectedTlds = selectedTlds.filter((t) => t !== tld);
      }
    } else {
      selectedTlds = [...selectedTlds, tld];
    }
  }

  function applyPreset(seedStr: string, filterAvail = false) {
    keywordInput = seedStr;
    onlyAvailable = filterAvail;
    const keywords = seedStr
      .split(/[,;\s]+/)
      .map((s) => s.trim())
      .filter(Boolean);
    onRunScan({
      keywords,
      tlds: selectedTlds,
      mutations,
      onlyAvailable: filterAvail,
      minScore,
    });
  }

  function submitScan(e: Event) {
    e.preventDefault();
    const keywords = keywordInput
      .split(/[,;\s]+/)
      .map((s) => s.trim())
      .filter(Boolean);
    onRunScan({
      keywords: keywords.length ? keywords : ['nova'],
      tlds: selectedTlds,
      mutations,
      onlyAvailable,
      minScore,
    });
  }

  const availableRatio = $derived(
    report && report.totalChecked > 0
      ? Math.round((report.availableCount / report.totalChecked) * 100)
      : 0
  );
</script>

<section aria-labelledby="scanner-workbench-heading" class="space-y-6">
  <!-- TOP ASYMMETRIC BENTO: 8-col Interactive Workbench + 4-col Studio Yield & Curated Presets -->
  <div class="grid grid-cols-1 lg:grid-cols-12 gap-6 items-stretch">
    <!-- 8-Column Workbench Card -->
    <form
      use:motionCard
      onsubmit={submitScan}
      class="lg:col-span-8 bento-card p-6 sm:p-7 flex flex-col justify-between gap-6"
    >
      <div class="space-y-4">
        <div class="flex flex-wrap items-center justify-between gap-2">
          <div class="flex items-center gap-2.5">
            <span class="inline-block w-2.5 h-2.5 rounded-full bg-[#19231f]"></span>
            <h2 id="scanner-workbench-heading" class="studio-label">
              01 // Naming Workbench & Parallel Availability Engine
            </h2>
          </div>
          <span class="text-xs font-mono text-[#48534e]">
            16 Worker Goroutines · DNS + RDAP + Past History
          </span>
        </div>

        <div class="flex flex-col sm:flex-row gap-3">
          <div class="flex-1 relative">
            <label for="seed-input" class="sr-only">Seed roots or keywords</label>
            <input
              id="seed-input"
              type="text"
              bind:value={keywordInput}
              placeholder="Enter multiple roots to scan in parallel (e.g. veltrix, nova, atelier)..."
              class="w-full rounded-2xl bg-[#f4f1e9] border-[1.5px] border-[#19231f]/20 focus:border-[#19231f] px-4 py-3.5 text-base font-mono text-[#19231f] placeholder-[#6d7873] transition-colors"
            />
          </div>
          <button
            type="submit"
            disabled={loading}
            class="studio-btn-primary px-7 py-3.5 text-sm tracking-tight cursor-pointer disabled:opacity-50 shrink-0"
          >
            {loading ? 'Scanning in Parallel…' : 'Scan & Appraise →'}
          </button>
        </div>

        <!-- Quick Curated Seed Presets -->
        <div class="flex flex-wrap items-center gap-2 pt-1">
          <span class="text-xs font-medium text-[#6d7873] mr-1">Try curated seeds:</span>
          {#each presetSeeds as preset}
            <button
              type="button"
              onclick={() => applyPreset(preset.value, preset.label.includes('Unclaimed'))}
              class="px-3 py-1 rounded-full text-xs font-medium bg-[#f4f1e9] hover:bg-[#dffc78] text-[#19231f] border border-[#19231f]/15 transition-colors cursor-pointer"
            >
              {preset.label}
            </button>
          {/each}
        </div>
      </div>

      <!-- Bottom Workbench Controls: TLD Matrix & Valuation Filters -->
      <div class="pt-5 border-t border-[#19231f]/10 flex flex-col xl:flex-row xl:items-center justify-between gap-4">
        <div class="flex flex-wrap items-center gap-1.5" role="group" aria-label="TLD extensions">
          <span class="studio-label mr-2">Extensions</span>
          {#each allTlds as tld}
            <button
              type="button"
              onclick={() => toggleTld(tld)}
              aria-pressed={selectedTlds.includes(tld)}
              class="px-3 py-1 rounded-full text-xs font-mono font-medium border transition-all cursor-pointer {selectedTlds.includes(
                tld
              )
                ? 'bg-[#19231f] text-[#fffdf8] border-[#19231f]'
                : 'bg-[#f4f1e9] text-[#48534e] border-[#19231f]/15 hover:border-[#19231f]/40'}"
            >
              .{tld}
            </button>
          {/each}
        </div>

        <div class="flex flex-wrap items-center gap-4 text-xs font-medium text-[#19231f]">
          <label class="inline-flex items-center gap-2 cursor-pointer select-none">
            <input
              type="checkbox"
              bind:checked={mutations}
              class="w-4 h-4 rounded border-[#19231f]/40 accent-[#19231f]"
            />
            <span>Affix studio (+hq, +labs, +flow, get+)</span>
          </label>

          <label
            class="inline-flex items-center gap-2 px-3 py-1 rounded-full border cursor-pointer select-none transition-colors {onlyAvailable
              ? 'bg-[#dffc78] border-[#19231f]'
              : 'bg-[#f4f1e9] border-[#19231f]/15'}"
          >
            <input
              type="checkbox"
              bind:checked={onlyAvailable}
              class="w-3.5 h-3.5 accent-[#19231f]"
            />
            <span class="font-display font-semibold">Only Unclaimed</span>
          </label>

          <div class="flex items-center gap-2 bg-[#f4f1e9] px-3 py-1 rounded-full border border-[#19231f]/15">
            <label for="score-range" class="text-[#48534e]">Min Score</label>
            <input
              id="score-range"
              type="range"
              min="0"
              max="85"
              step="5"
              bind:value={minScore}
              class="w-20 accent-[#19231f] cursor-pointer"
            />
            <span class="font-mono font-semibold w-6 text-right">{minScore}</span>
          </div>
        </div>
      </div>
    </form>

    <!-- 4-Column Asymmetric Editorial Pulse Card -->
    <div
      use:motionCard={{ delay: 0.05 }}
      class="lg:col-span-4 bento-card p-6 sm:p-7 flex flex-col justify-between bg-[#fffdf8] relative overflow-hidden"
    >
      <div class="flex items-center justify-between">
        <span class="studio-label">Registry Yield Snapshot</span>
        {#if report}
          <span class="px-2.5 py-0.5 rounded-full text-xs font-mono bg-[#f4f1e9] border border-[#19231f]/15">
            {report.durationMs}ms
          </span>
        {/if}
      </div>

      {#if report}
        <div class="my-4 space-y-4">
          <div class="flex items-baseline justify-between gap-4">
            <div>
              <div class="text-5xl font-display font-bold tracking-tight text-[#19231f]">
                {report.availableCount}
                <span class="text-2xl font-serif-editorial font-normal text-[#48534e]">unclaimed</span>
              </div>
              <p class="text-xs text-[#48534e] mt-1">
                out of <strong class="font-mono text-[#19231f]">{report.totalChecked}</strong> domain candidates tested
              </p>
            </div>

            <div class="text-right">
              <span class="inline-block px-3 py-1 rounded-full text-xs font-display font-bold bg-[#dffc78] border border-[#19231f]">
                {report.highValueCount} Prime Gems
              </span>
              <div class="text-[11px] text-[#6d7873] mt-1">Score ≥ 74/100</div>
            </div>
          </div>

          <!-- Editorial Proportional Bar -->
          <div class="space-y-1.5">
            <div class="h-3 w-full rounded-full bg-[#f4f1e9] border border-[#19231f]/15 overflow-hidden flex">
              <div
                class="h-full bg-[#dffc78] border-r border-[#19231f] transition-all duration-300"
                style="width: {availableRatio}%"
                title="Available ({availableRatio}%)"
              ></div>
              <div
                class="h-full bg-[#ffc3a5] transition-all duration-300"
                style="width: {100 - availableRatio}%"
                title="Registered ({100 - availableRatio}%)"
              ></div>
            </div>
            <div class="flex justify-between text-[11px] font-mono text-[#48534e]">
              <span class="flex items-center gap-1.5">
                <span class="w-2 h-2 rounded-full bg-[#dffc78] border border-[#19231f]"></span>
                {availableRatio}% Unclaimed
              </span>
              <span class="flex items-center gap-1.5">
                <span class="w-2 h-2 rounded-full bg-[#ffc3a5] border border-[#19231f]"></span>
                {100 - availableRatio}% Registered
              </span>
            </div>
          </div>
        </div>
      {:else}
        <div class="py-8 text-sm text-[#48534e]">
          Run a scan to inspect registry availability and phonetic valuation across TLDs.
        </div>
      {/if}

      <div class="pt-4 border-t border-[#19231f]/10 flex items-center justify-between text-xs">
        <span class="text-[#48534e]">Filter view:</span>
        <div class="flex gap-1.5">
          <button
            type="button"
            onclick={() => {
              onlyAvailable = true;
              onRunScan({
                keywords: keywordInput.split(/[,;\s]+/).filter(Boolean),
                tlds: selectedTlds,
                mutations,
                onlyAvailable: true,
                minScore,
              });
            }}
            class="px-2.5 py-1 rounded-full font-display font-semibold bg-[#dffc78] border border-[#19231f] text-[#19231f] cursor-pointer"
          >
            Isolate Empty Domains
          </button>
        </div>
      </div>
    </div>
  </div>

  <!-- CLEAR LOADING TELEMETRY BANNER -->
  {#if loading}
    <LoadingProgressBanner
      title="Parallel Bulk Availability & Valuation Scan"
      target={keywordInput}
      workers={16}
      steps={[
        `Generating Root & Affix Variations Across ${selectedTlds.length} TLDs`,
        'Probing Live DNS NS & A Record Delegation in Parallel',
        'Computing 4-Axis Phonetic & Commercial Value Scores',
      ]}
    />
  {/if}

  <!-- MAIN ASYMMETRIC BENTO: 7-col Domain Ledger + 5-col Appraisal Dossier -->
  {#if report && !loading}
    <div class="grid grid-cols-1 lg:grid-cols-12 gap-6 items-start">
      <!-- 7-Column Domain Candidates Ledger -->
      <div use:motionCard={{ delay: 0.06 }} class="lg:col-span-7 bento-card overflow-hidden">
        <div class="px-6 py-4 border-b border-[#19231f]/10 flex flex-wrap items-center justify-between gap-2 bg-[#fffdf8]">
          <div>
            <h3 class="font-display font-bold text-base text-[#19231f]">
              Candidate Ledger & Market Valuation
            </h3>
            <p class="text-xs text-[#48534e]">
              Click any domain to appraise, check <strong>⏳ Past History</strong>, or <strong>★ Save</strong> to Vault
            </p>
          </div>
          <span class="px-3 py-1 rounded-full text-xs font-mono bg-[#f4f1e9] border border-[#19231f]/15">
            {report.items.length} results
          </span>
        </div>

        <div class="scroll-panel max-h-[620px] divide-y divide-[#19231f]/10">
          {#each report.items as item (item.domain)}
            {@const isSelected = activeItem?.domain === item.domain}
            {@const isSaved = savedDomainsSet.has(item.domain)}
            <div
              class="px-6 py-4 flex flex-col sm:flex-row sm:items-center justify-between gap-3 transition-colors cursor-pointer {isSelected
                ? 'bg-[#f4f1e9]'
                : 'hover:bg-[#f4f1e9]/60'}"
              role="button"
              tabindex="0"
              onclick={() => (activeItem = item)}
              onkeydown={(e) => e.key === 'Enter' && (activeItem = item)}
            >
              <div class="space-y-1 min-w-0">
                <div class="flex items-center gap-2 flex-wrap">
                  <button
                    type="button"
                    onclick={(e) => {
                      e.stopPropagation();
                      onToggleSave(item.domain, item.available);
                    }}
                    class="w-6 h-6 rounded-full border flex items-center justify-center text-xs transition-colors cursor-pointer {isSaved
                      ? 'bg-[#dffc78] border-[#19231f] text-[#19231f] font-bold'
                      : 'bg-[#fffdf8] border-[#19231f]/20 text-[#6d7873] hover:text-[#19231f]'}"
                    title={isSaved ? 'Saved in Vault (Click to remove)' : 'Save domain to Vault'}
                  >
                    {isSaved ? '★' : '☆'}
                  </button>

                  <span class="font-mono text-lg font-semibold tracking-tight text-[#19231f]">
                    {item.rootName}<span class="text-[#5366e8]">.{item.tld}</span>
                  </span>

                  {#if item.available}
                    <span
                      class="px-2.5 py-0.5 rounded-full text-[11px] font-display font-bold uppercase tracking-wide bg-[#dffc78] text-[#19231f] border border-[#19231f]"
                    >
                      ● Unclaimed
                    </span>
                  {:else}
                    <span
                      class="px-2.5 py-0.5 rounded-full text-[11px] font-display font-semibold uppercase tracking-wide bg-[#ffc3a5]/70 text-[#19231f] border border-[#19231f]/30"
                    >
                      Registered
                    </span>
                  {/if}

                  {#if item.valuation.score >= 85}
                    <span
                      class="px-2.5 py-0.5 rounded-full text-[11px] font-display font-semibold bg-[#d9d6fc] text-[#19231f] border border-[#19231f]/30"
                    >
                      Ultra Premium
                    </span>
                  {/if}
                </div>

                {#if item.valuation.highlights?.length}
                  <p class="text-xs text-[#48534e] truncate pl-8">
                    {item.valuation.highlights.join('  ·  ')}
                  </p>
                {/if}
              </div>

              <div class="flex items-center gap-3 shrink-0 pl-8 sm:pl-0">
                <div class="text-right">
                  <div class="font-display font-bold text-sm text-[#19231f]">
                    {item.valuation.estimatedDisplay}
                  </div>
                  <div class="text-[11px] font-mono text-[#48534e]">
                    Index <strong class="text-[#19231f]">{item.valuation.score}</strong>/100
                  </div>
                </div>

                <div class="flex items-center gap-1.5">
                  <button
                    type="button"
                    onclick={(e) => {
                      e.stopPropagation();
                      onCheckHistory(item.domain);
                    }}
                    class="px-2.5 py-1.5 rounded-full text-xs font-display font-semibold bg-[#fffdf8] hover:bg-[#ffc3a5] text-[#19231f] border border-[#19231f]/25 transition-colors cursor-pointer"
                    title="Check if registered in the past (Wayback Machine + CT Logs)"
                  >
                    ⏳ Past Reg?
                  </button>

                  <button
                    type="button"
                    onclick={(e) => {
                      e.stopPropagation();
                      onInspectDomain(item.domain);
                    }}
                    class="px-2.5 py-1.5 rounded-full text-xs font-display font-semibold bg-[#fffdf8] hover:bg-[#d9d6fc] text-[#19231f] border border-[#19231f]/25 transition-colors cursor-pointer"
                    title="Inspect RDAP Registration Date & DNS"
                  >
                    RDAP
                  </button>
                </div>
              </div>
            </div>
          {/each}
        </div>
      </div>

      <!-- 5-Column Sticky Appraisal Dossier Card (Selective Lilac #d9d6fc Header Accent) -->
      <aside use:motionCard={{ delay: 0.1 }} class="lg:col-span-5 bento-card overflow-hidden lg:sticky lg:top-24">
        {#if activeItem}
          <!-- Selective Lilac Studio Header -->
          <div class="bg-[#d9d6fc] border-b border-[#19231f]/15 p-6">
            <div class="flex items-center justify-between gap-2">
              <span class="studio-label text-[#19231f]">Valuation Specimen</span>
              <div class="flex items-center gap-1.5">
                <button
                  type="button"
                  onclick={() => activeItem && onToggleSave(activeItem.domain, activeItem.available)}
                  class="px-3 py-0.5 rounded-full text-xs font-display font-bold border border-[#19231f] cursor-pointer {savedDomainsSet.has(
                    activeItem.domain
                  )
                    ? 'bg-[#dffc78] text-[#19231f]'
                    : 'bg-[#fffdf8] text-[#19231f] hover:bg-[#dffc78]'}"
                >
                  {savedDomainsSet.has(activeItem.domain) ? '★ Saved' : '☆ Save'}
                </button>
                <span
                  class="px-3 py-0.5 rounded-full text-xs font-mono font-semibold bg-[#fffdf8] border border-[#19231f] text-[#19231f]"
                >
                  Score {activeItem.valuation.score}/100
                </span>
              </div>
            </div>

            <h3 class="mt-3 text-3xl font-display font-bold text-[#19231f] break-all">
              {activeItem.domain}
            </h3>

            <div class="mt-3 flex flex-wrap items-baseline justify-between gap-2 pt-3 border-t border-[#19231f]/15">
              <span class="text-xs font-medium text-[#19231f]/80">
                Tier: <span class="font-serif-editorial text-lg text-[#19231f]">{activeItem.valuation.tier}</span>
              </span>
              <span class="font-display font-bold text-xl text-[#19231f]">
                {activeItem.valuation.estimatedDisplay}
              </span>
            </div>
          </div>

          <!-- 4-Axis Breakdown Body -->
          <div class="p-6 space-y-5 bg-[#fffdf8]">
            <div class="space-y-3.5">
              <div>
                <div class="flex justify-between text-xs font-medium mb-1">
                  <span>Length Scarcity ({activeItem.rootName.length} letters)</span>
                  <span class="font-mono font-semibold">{activeItem.valuation.lengthScore} / 30</span>
                </div>
                <div class="h-2.5 rounded-full bg-[#f4f1e9] border border-[#19231f]/15 overflow-hidden">
                  <div
                    class="h-full bg-[#19231f] rounded-full"
                    style="width: {(activeItem.valuation.lengthScore / 30) * 100}%"
                  ></div>
                </div>
              </div>

              <div>
                <div class="flex justify-between text-xs font-medium mb-1">
                  <span>Extension Weight (.{activeItem.tld})</span>
                  <span class="font-mono font-semibold">{activeItem.valuation.tldScore} / 25</span>
                </div>
                <div class="h-2.5 rounded-full bg-[#f4f1e9] border border-[#19231f]/15 overflow-hidden">
                  <div
                    class="h-full bg-[#5366e8] rounded-full"
                    style="width: {(activeItem.valuation.tldScore / 25) * 100}%"
                  ></div>
                </div>
              </div>

              <div>
                <div class="flex justify-between text-xs font-medium mb-1">
                  <span>Phonetic & Syllable Cadence</span>
                  <span class="font-mono font-semibold">{activeItem.valuation.phoneticScore} / 25</span>
                </div>
                <div class="h-2.5 rounded-full bg-[#f4f1e9] border border-[#19231f]/15 overflow-hidden">
                  <div
                    class="h-full bg-[#19231f] rounded-full"
                    style="width: {(activeItem.valuation.phoneticScore / 25) * 100}%"
                  ></div>
                </div>
              </div>

              <div>
                <div class="flex justify-between text-xs font-medium mb-1">
                  <span>Commercial & Sector Signal</span>
                  <span class="font-mono font-semibold">{activeItem.valuation.keywordScore} / 20</span>
                </div>
                <div class="h-2.5 rounded-full bg-[#f4f1e9] border border-[#19231f]/15 overflow-hidden">
                  <div
                    class="h-full bg-[#19231f] rounded-full"
                    style="width: {(activeItem.valuation.keywordScore / 20) * 100}%"
                  ></div>
                </div>
              </div>
            </div>

            <!-- Editorial Appraisal Notes -->
            <div class="rounded-2xl bg-[#f4f1e9] p-4 border border-[#19231f]/12 space-y-2">
              <div class="studio-label">Why this domain carries weight</div>
              <ul class="space-y-1.5 text-xs text-[#19231f]">
                {#each activeItem.valuation.highlights as signal}
                  <li class="flex items-start gap-2">
                    <span class="font-bold text-[#5366e8]">→</span>
                    <span>{signal}</span>
                  </li>
                {/each}
              </ul>
            </div>

            <!-- Direct Studio Actions -->
            <div class="space-y-2 pt-1">
              <button
                type="button"
                onclick={() => activeItem && onCheckHistory(activeItem.domain)}
                class="w-full py-2.5 px-4 rounded-full font-display font-bold text-xs bg-[#ffc3a5] hover:bg-[#ffb490] text-[#19231f] border-[1.5px] border-[#19231f] transition-colors cursor-pointer"
              >
                ⏳ Check If Registered in Past (Wayback + CT) →
              </button>

              <div class="grid grid-cols-2 gap-2">
                <button
                  type="button"
                  onclick={() => activeItem && onInspectDomain(activeItem.domain)}
                  class="studio-btn-primary py-2.5 px-3 text-xs text-center cursor-pointer"
                >
                  RDAP Dossier →
                </button>
                <button
                  type="button"
                  onclick={() => activeItem && onRunRecon(activeItem.domain)}
                  class="studio-btn-ink py-2.5 px-3 text-xs text-center cursor-pointer"
                >
                  Ports & TLS →
                </button>
              </div>
            </div>
          </div>
        {/if}
      </aside>
    </div>
  {/if}
</section>
