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
      dictionaryPack?: string;
      mutations: boolean;
      onlyAvailable: boolean;
      minScore: number;
    }) => void;
    onCancelJob: () => void;
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
    onCancelJob,
    onInspectDomain,
    onCheckHistory,
    onRunRecon,
    onCrawlDomain,
    onToggleSave,
    savedDomainsSet,
  }: Props = $props();

  let keywordInput = $state('veltrix, nova');
  let selectedTlds = $state<string[]>(['com', 'ai', 'io', 'dev', 'co', 'app', 'net', 'xyz']);
  let dictionaryPack = $state<string>('');
  let mutations = $state(true);
  let onlyAvailable = $state(false);
  let minScore = $state(0);
  let activeItem = $state<ScanResultItem | null>(null);
  let showAllTlds = $state(false);

  const primaryTlds = ['com', 'ai', 'io', 'dev', 'co', 'app', 'net', 'org', 'sh', 'xyz', 'me', 'vc'];
  const extraTlds = ['gg', 'so', 'cloud', 'tech', 'studio', 'design', 'tools', 'codes', 'finance', 'store', 'shop', 'build'];
  const visibleTlds = $derived(showAllTlds ? [...primaryTlds, ...extraTlds] : primaryTlds);

  const dictionaryOptions = [
    { id: '', label: 'Custom Seeds Only (No Dictionary Pack)' },
    { id: 'short_english', label: '📖 Short 4–5 Letter English Words (loom, cove, kiln, helm, arch...)' },
    { id: 'ai_neural', label: '🧠 AI, Agents & Compute Dictionary (agent, neural, cortex, tensor...)' },
    { id: 'devtools_infra', label: '⚙️ Systems, DevTools & Cloud Infra (deploy, socket, kernel, vault...)' },
    { id: 'fintech_capital', label: '🏛️ Fintech, Treasury & Commerce (treasury, yield, settle, escrow...)' },
    { id: 'design_atelier', label: '✦ Design Studio & Editorial Craft (atelier, foundry, folio, serif...)' },
  ];

  const presetSeeds = [
    { label: 'Unclaimed Tech Gems', value: 'veltrix, nexora', dict: '' },
    { label: 'Short English Dictionary', value: 'nova', dict: 'short_english' },
    { label: 'AI & Autonomous Agents', value: 'pulse', dict: 'ai_neural' },
    { label: 'Design & Editorial Studio', value: 'forma', dict: 'design_atelier' },
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

  function selectTldGroup(group: 'popular' | 'tech' | 'classics' | 'all') {
    if (group === 'popular') selectedTlds = ['com', 'ai', 'io', 'dev', 'co', 'app'];
    else if (group === 'tech') selectedTlds = ['ai', 'io', 'dev', 'app', 'sh', 'cloud', 'tech', 'codes', 'tools'];
    else if (group === 'classics') selectedTlds = ['com', 'net', 'org', 'co', 'me', 'xyz'];
    else {
      showAllTlds = true;
      selectedTlds = [...primaryTlds, ...extraTlds];
    }
  }

  function applyPreset(seedStr: string, dictId: string, filterAvail = false) {
    keywordInput = seedStr;
    dictionaryPack = dictId;
    onlyAvailable = filterAvail;
    const keywords = seedStr
      .split(/[,;\s]+/)
      .map((s) => s.trim())
      .filter(Boolean);
    onRunScan({
      keywords,
      tlds: selectedTlds,
      dictionaryPack: dictId,
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
      dictionaryPack,
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
      class="lg:col-span-8 bento-card p-6 sm:p-7 flex flex-col justify-between gap-5"
    >
      <div class="space-y-4">
        <div class="flex flex-wrap items-center justify-between gap-2">
          <div class="flex items-center gap-2.5">
            <span class="inline-block w-2.5 h-2.5 rounded-full bg-[#19231f]"></span>
            <h2 id="scanner-workbench-heading" class="studio-label">
              01 // Naming Workbench, Dictionary Combinator & 24-TLD Engine
            </h2>
          </div>
          <span class="text-xs font-mono text-[#48534e]">
            18 Parallel Goroutines · Calibrated Reg Fee & Flip Pricing
          </span>
        </div>

        <!-- Seed Input + Dictionary Pack Selector -->
        <div class="grid grid-cols-1 sm:grid-cols-12 gap-3">
          <div class="sm:col-span-6">
            <label for="seed-input" class="block text-[11px] font-mono text-[#48534e] mb-1">
              Seed Roots / Keywords
            </label>
            <input
              id="seed-input"
              type="text"
              bind:value={keywordInput}
              placeholder="e.g. veltrix, nova, supercoloring..."
              class="w-full rounded-2xl bg-[#f4f1e9] border-[1.5px] border-[#19231f]/20 focus:border-[#19231f] px-4 py-3 text-sm font-mono text-[#19231f] placeholder-[#6d7873] transition-colors"
            />
          </div>

          <div class="sm:col-span-6">
            <label for="dict-pack" class="block text-[11px] font-mono text-[#48534e] mb-1">
              Built-in Dictionary Word Generator
            </label>
            <select
              id="dict-pack"
              bind:value={dictionaryPack}
              class="w-full rounded-2xl bg-[#f4f1e9] border-[1.5px] border-[#19231f]/20 focus:border-[#19231f] px-3.5 py-3 text-xs font-display font-semibold text-[#19231f]"
            >
              {#each dictionaryOptions as opt}
                <option value={opt.id}>{opt.label}</option>
              {/each}
            </select>
          </div>
        </div>

        <!-- Submit + Curated Seeds Row -->
        <div class="flex flex-wrap items-center justify-between gap-3 pt-1">
          <div class="flex flex-wrap items-center gap-1.5">
            <span class="text-xs font-medium text-[#6d7873] mr-1">Quick presets:</span>
            {#each presetSeeds as preset}
              <button
                type="button"
                onclick={() => applyPreset(preset.value, preset.dict, preset.label.includes('Unclaimed'))}
                class="px-3 py-1 rounded-full text-xs font-medium bg-[#f4f1e9] hover:bg-[#dffc78] text-[#19231f] border border-[#19231f]/15 transition-colors cursor-pointer"
              >
                {preset.label}
              </button>
            {/each}
          </div>

          <div class="flex items-center gap-2">
            {#if loading}
              <button
                type="button"
                onclick={onCancelJob}
                class="px-5 py-3 rounded-full bg-[#ffc3a5] hover:bg-[#ffb490] text-[#19231f] font-display font-bold text-xs border-[1.5px] border-[#19231f] cursor-pointer"
              >
                ✕ Cancel Scan
              </button>
            {/if}
            <button
              type="submit"
              disabled={loading}
              class="studio-btn-primary px-7 py-3 text-sm tracking-tight cursor-pointer disabled:opacity-50 shrink-0"
            >
              {loading ? 'Scanning in Parallel…' : 'Scan & Appraise →'}
            </button>
          </div>
        </div>
      </div>

      <!-- Bottom Workbench Controls: 24-TLD Matrix & Valuation Filters -->
      <div class="pt-4 border-t border-[#19231f]/10 space-y-3">
        <div class="flex flex-wrap items-center justify-between gap-2">
          <div class="flex items-center gap-2">
            <span class="studio-label">Extensions ({selectedTlds.length} selected)</span>
            <button
              type="button"
              onclick={() => (showAllTlds = !showAllTlds)}
              class="text-xs font-mono text-[#5366e8] hover:underline cursor-pointer"
            >
              {showAllTlds ? 'Show Fewer (12)' : '+ Show All 24 TLDs'}
            </button>
          </div>

          <div class="flex flex-wrap items-center gap-1 text-[11px] font-mono">
            <button
              type="button"
              onclick={() => selectTldGroup('popular')}
              class="px-2 py-0.5 rounded bg-[#f4f1e9] hover:bg-[#d9d6fc] border border-[#19231f]/15 cursor-pointer"
            >
              Popular
            </button>
            <button
              type="button"
              onclick={() => selectTldGroup('tech')}
              class="px-2 py-0.5 rounded bg-[#f4f1e9] hover:bg-[#d9d6fc] border border-[#19231f]/15 cursor-pointer"
            >
              Tech & AI
            </button>
            <button
              type="button"
              onclick={() => selectTldGroup('classics')}
              class="px-2 py-0.5 rounded bg-[#f4f1e9] hover:bg-[#d9d6fc] border border-[#19231f]/15 cursor-pointer"
            >
              Classics
            </button>
            <button
              type="button"
              onclick={() => selectTldGroup('all')}
              class="px-2 py-0.5 rounded bg-[#f4f1e9] hover:bg-[#dffc78] border border-[#19231f]/15 cursor-pointer"
            >
              All 24
            </button>
          </div>
        </div>

        <div class="flex flex-wrap items-center gap-1.5" role="group" aria-label="TLD extensions">
          {#each visibleTlds as tld}
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

        <div class="flex flex-wrap items-center justify-between gap-4 pt-2 border-t border-[#19231f]/10 text-xs font-medium text-[#19231f]">
          <div class="flex flex-wrap items-center gap-4">
            <label class="inline-flex items-center gap-2 cursor-pointer select-none">
              <input
                type="checkbox"
                bind:checked={mutations}
                class="w-4 h-4 rounded border-[#19231f]/40 accent-[#19231f]"
              />
              <span>Affix studio (+hq, +labs, +flow, +studio, get+)</span>
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
          </div>

          <div class="flex items-center gap-2 bg-[#f4f1e9] px-3 py-1 rounded-full border border-[#19231f]/15">
            <label for="score-range" class="text-[#48534e]">Min Quality Score</label>
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
              <div class="text-[11px] text-[#6d7873] mt-1">Score ≥ 73/100</div>
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
          Run a scan to inspect registry availability and calibrated registration / resale valuation across TLDs.
        </div>
      {/if}

      <div class="pt-4 border-t border-[#19231f]/10 flex items-center justify-between text-xs">
        <span class="text-[#48534e]">Pricing Model:</span>
        <span class="font-mono text-[#19231f] font-semibold">1st-Yr Reg Cost + Est. Flip Range</span>
      </div>
    </div>
  </div>

  <!-- CLEAR LOADING TELEMETRY BANNER WITH CANCEL BUTTON -->
  {#if loading}
    <LoadingProgressBanner
      title="Parallel Bulk Availability & Dictionary Scan"
      target={keywordInput + (dictionaryPack ? ` + [${dictionaryPack}]` : '')}
      workers={18}
      onCancel={onCancelJob}
      steps={[
        `Generating Candidates Across ${selectedTlds.length} TLDs`,
        'Probing Live DNS NS & A Delegation in Parallel',
        'Calibrating Registrar Fee & Aftermarket Flip Value',
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
              Candidate Ledger & Calibrated Valuation
            </h3>
            <p class="text-xs text-[#48534e]">
              Unclaimed domains show exact <strong>1st-yr registration cost</strong> + flip estimate; registered show aftermarket value
            </p>
          </div>
          <span class="px-3 py-1 rounded-full text-xs font-mono bg-[#f4f1e9] border border-[#19231f]/15">
            {report.items.length} results
          </span>
        </div>

        <div class="scroll-panel max-h-[640px] divide-y divide-[#19231f]/10">
          {#each report.items as item (item.domain)}
            {@const isSelected = activeItem?.domain === item.domain}
            {@const isSaved = savedDomainsSet.has(item.domain)}
            <div
              class="px-5 sm:px-6 py-4 flex flex-col sm:flex-row sm:items-center justify-between gap-3 transition-colors cursor-pointer {isSelected
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
                    title={isSaved ? 'Saved in Vault' : 'Save domain to Vault'}
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
                      ● Unclaimed ({item.valuation.regFeeDisplay || '$12/yr'})
                    </span>
                  {:else}
                    <span
                      class="px-2.5 py-0.5 rounded-full text-[11px] font-display font-semibold uppercase tracking-wide bg-[#ffc3a5]/70 text-[#19231f] border border-[#19231f]/30"
                    >
                      Registered
                    </span>
                  {/if}

                  {#if item.valuation.isDictionaryWord}
                    <span
                      class="px-2 py-0.5 rounded-full text-[10px] font-mono font-bold bg-[#d9d6fc] text-[#19231f] border border-[#19231f]/30"
                    >
                      Dictionary Word
                    </span>
                  {/if}
                </div>

                {#if item.valuation.highlights?.length}
                  <p class="text-xs text-[#48534e] truncate pl-8">
                    {item.valuation.highlights.join('  ·  ')}
                  </p>
                {/if}
              </div>

              <div class="flex items-center gap-2.5 shrink-0 pl-8 sm:pl-0 flex-wrap">
                <div class="text-right mr-1">
                  <div class="font-display font-bold text-xs sm:text-sm text-[#19231f]">
                    {item.valuation.estimatedDisplay}
                  </div>
                  <div class="text-[11px] font-mono text-[#48534e]">
                    Quality <strong class="text-[#19231f]">{item.valuation.score}</strong>/100
                  </div>
                </div>

                <div class="flex items-center gap-1">
                  <a
                    href={`https://${item.domain}`}
                    target="_blank"
                    rel="noopener noreferrer"
                    onclick={(e) => e.stopPropagation()}
                    class="px-2.5 py-1.5 rounded-full text-xs font-display font-bold bg-[#fffdf8] hover:bg-[#dffc78] text-[#19231f] border border-[#19231f]/25 transition-colors"
                    title="Open live domain in new tab"
                  >
                    ↗ Site
                  </a>

                  <button
                    type="button"
                    onclick={(e) => {
                      e.stopPropagation();
                      onCheckHistory(item.domain);
                    }}
                    class="px-2.5 py-1.5 rounded-full text-xs font-display font-semibold bg-[#fffdf8] hover:bg-[#ffc3a5] text-[#19231f] border border-[#19231f]/25 transition-colors cursor-pointer"
                    title="Check Wayback Machine Snapshots & Past Registration History"
                  >
                    ⏳ History
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
                <a
                  href={`https://${activeItem.domain}`}
                  target="_blank"
                  rel="noopener noreferrer"
                  class="px-2.5 py-0.5 rounded-full text-xs font-display font-bold bg-[#fffdf8] hover:bg-[#dffc78] border border-[#19231f] text-[#19231f]"
                >
                  ↗ Visit Site
                </a>
                <a
                  href={`https://web.archive.org/web/*/${activeItem.domain}`}
                  target="_blank"
                  rel="noopener noreferrer"
                  class="px-2.5 py-0.5 rounded-full text-xs font-display font-bold bg-[#fffdf8] hover:bg-[#ffc3a5] border border-[#19231f] text-[#19231f]"
                >
                  🏛️ Wayback ↗
                </a>
              </div>
            </div>

            <h3 class="mt-3 text-3xl font-display font-bold text-[#19231f] break-all">
              {activeItem.domain}
            </h3>

            <div class="mt-3 flex flex-wrap items-baseline justify-between gap-2 pt-3 border-t border-[#19231f]/15">
              <span class="text-xs font-medium text-[#19231f]/80">
                Tier: <span class="font-serif-editorial text-lg text-[#19231f]">{activeItem.valuation.tier}</span>
              </span>
              <span class="font-display font-bold text-lg text-[#19231f]">
                {activeItem.valuation.estimatedDisplay}
              </span>
            </div>
          </div>

          <!-- 4-Axis Breakdown Body -->
          <div class="p-6 space-y-5 bg-[#fffdf8]">
            <!-- Registrar vs Flip Callout -->
            <div class="grid grid-cols-2 gap-3 text-xs">
              <div class="rounded-xl bg-[#f4f1e9] border border-[#19231f]/15 p-3">
                <div class="studio-label">1st-Yr Registrar Cost</div>
                <div class="mt-1 font-mono font-bold text-base text-[#19231f]">
                  {activeItem.valuation.regFeeDisplay || '$12/yr'}
                </div>
                <div class="text-[11px] text-[#48534e]">
                  {activeItem.available ? 'Available at standard fee' : 'Currently registered'}
                </div>
              </div>

              <div class="rounded-xl bg-[#f4f1e9] border border-[#19231f]/15 p-3">
                <div class="studio-label">Aftermarket Appraisal</div>
                <div class="mt-1 font-mono font-bold text-base text-[#5366e8]">
                  ${activeItem.valuation.estimatedMinUsd.toLocaleString()} – ${activeItem.valuation.estimatedMaxUsd.toLocaleString()}
                </div>
                <div class="text-[11px] text-[#48534e]">Quality Index: {activeItem.valuation.score}/100</div>
              </div>
            </div>

            <div class="space-y-3">
              <div>
                <div class="flex justify-between text-xs font-medium mb-1">
                  <span>Length Scarcity ({activeItem.rootName.length} letters)</span>
                  <span class="font-mono font-semibold">{activeItem.valuation.lengthScore} / 30</span>
                </div>
                <div class="h-2 rounded-full bg-[#f4f1e9] border border-[#19231f]/15 overflow-hidden">
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
                <div class="h-2 rounded-full bg-[#f4f1e9] border border-[#19231f]/15 overflow-hidden">
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
                <div class="h-2 rounded-full bg-[#f4f1e9] border border-[#19231f]/15 overflow-hidden">
                  <div
                    class="h-full bg-[#19231f] rounded-full"
                    style="width: {(activeItem.valuation.phoneticScore / 25) * 100}%"
                  ></div>
                </div>
              </div>

              <div>
                <div class="flex justify-between text-xs font-medium mb-1">
                  <span>Dictionary & Commercial Signal</span>
                  <span class="font-mono font-semibold">{activeItem.valuation.keywordScore} / 20</span>
                </div>
                <div class="h-2 rounded-full bg-[#f4f1e9] border border-[#19231f]/15 overflow-hidden">
                  <div
                    class="h-full bg-[#19231f] rounded-full"
                    style="width: {(activeItem.valuation.keywordScore / 20) * 100}%"
                  ></div>
                </div>
              </div>
            </div>

            <!-- Direct Studio Actions -->
            <div class="space-y-2 pt-1">
              <button
                type="button"
                onclick={() => activeItem && onCheckHistory(activeItem.domain)}
                class="w-full py-2.5 px-4 rounded-full font-display font-bold text-xs bg-[#ffc3a5] hover:bg-[#ffb490] text-[#19231f] border-[1.5px] border-[#19231f] transition-colors cursor-pointer"
              >
                ⏳ Inspect Wayback Snapshots & Past History →
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
                  Ports & Security →
                </button>
              </div>
            </div>
          </div>
        {/if}
      </aside>
    </div>
  {/if}
</section>
