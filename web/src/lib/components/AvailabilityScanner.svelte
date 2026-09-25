<script lang="ts">
  import type { ScanReport, ScanResultItem } from '../types';
  import { motionCard } from '../motion';
  import LoadingProgressBanner from './LoadingProgressBanner.svelte';
  import StudioIcon from './StudioIcon.svelte';

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
    onCheckTraffic?: (domain: string) => void;
    onPrepareReport?: (domain: string) => void;
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
    onCheckTraffic,
    onPrepareReport,
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

  const dictionaryOptions: Array<{
    id: string;
    shortName: string;
    label: string;
    icon: 'book' | 'cpu' | 'terminal' | 'layers' | 'compass';
  }> = [
    { id: '', shortName: 'Custom Only', label: 'Custom Seeds Only (No Dictionary Pack)', icon: 'compass' },
    { id: 'short_english', shortName: 'Short English', label: 'Short 4–5 Letter English Words (loom, cove, kiln, helm, arch...)', icon: 'book' },
    { id: 'ai_neural', shortName: 'AI & Agents', label: 'AI, Agents & Compute Dictionary (agent, neural, cortex, tensor...)', icon: 'cpu' },
    { id: 'devtools_infra', shortName: 'DevTools & Cloud', label: 'Systems, DevTools & Cloud Infra (deploy, socket, kernel, vault...)', icon: 'terminal' },
    { id: 'fintech_capital', shortName: 'Fintech & Capital', label: 'Fintech, Treasury & Commerce (treasury, yield, settle, escrow...)', icon: 'layers' },
    { id: 'design_atelier', shortName: 'Design Studio', label: 'Design Studio & Editorial Craft (atelier, foundry, folio, serif...)', icon: 'compass' },
  ];

  const presetSeeds = [
    { label: 'Unclaimed Tech Gems', value: 'veltrix, nexora', dict: '' },
    { label: 'Short English Dictionary', value: '', dict: 'short_english' },
    { label: 'AI & Autonomous Agents', value: 'agent.co', dict: 'ai_neural' },
    { label: 'Design & Editorial Studio', value: '', dict: 'design_atelier' },
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

  function executeScan(customDict?: string) {
    const activeDict = customDict !== undefined ? customDict : dictionaryPack;
    const keywords = keywordInput
      .split(/[,;\s]+/)
      .map((s) => s.trim())
      .filter(Boolean);

    onRunScan({
      keywords: keywords.length ? keywords : activeDict ? [] : ['veltrix'],
      tlds: selectedTlds,
      dictionaryPack: activeDict,
      mutations,
      onlyAvailable,
      minScore,
    });
  }

  function selectDictionaryPack(packId: string) {
    dictionaryPack = packId;
    executeScan(packId);
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
    executeScan();
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
            18 Workers · Authoritative Port-43 WHOIS + DNS Verification
          </span>
        </div>

        <!-- Seed Input + Dictionary Pack Dropdown -->
        <div class="grid grid-cols-1 sm:grid-cols-12 gap-3">
          <div class="sm:col-span-6">
            <label for="seed-input" class="block text-[11px] font-mono text-[#48534e] mb-1">
              Seed Roots or Exact Domains (Optional with Dictionary)
            </label>
            <input
              id="seed-input"
              type="text"
              bind:value={keywordInput}
              placeholder="e.g. veltrix, nova, agent.co..."
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
              onchange={() => executeScan(dictionaryPack)}
              class="studio-select w-full rounded-2xl bg-[#f4f1e9] border-[1.5px] border-[#19231f]/20 focus:border-[#19231f] py-3 text-xs font-display font-semibold text-[#19231f] cursor-pointer"
            >
              {#each dictionaryOptions as opt}
                <option value={opt.id}>{opt.label}</option>
              {/each}
            </select>
          </div>
        </div>

        <!-- Interactive 1-Click Dictionary Pack Selector Bar (With Vector Icons, Zero Emojis) -->
        <div class="space-y-1.5">
          <div class="text-[11px] font-mono text-[#48534e]">
            Instant Dictionary Word Packs (click to scan dictionary words across selected TLDs):
          </div>
          <div class="flex flex-wrap items-center gap-1.5">
            {#each dictionaryOptions as pack}
              <button
                type="button"
                onclick={() => selectDictionaryPack(pack.id)}
                class="px-3 py-1.5 rounded-full text-xs font-display font-semibold border transition-all cursor-pointer inline-flex items-center gap-1.5 {dictionaryPack ===
                pack.id
                  ? 'bg-[#19231f] text-[#dffc78] border-[#19231f]'
                  : 'bg-[#f4f1e9] text-[#19231f] border-[#19231f]/15 hover:bg-[#d9d6fc]'}"
              >
                <StudioIcon name={pack.icon} class="w-3.5 h-3.5" />
                <span>{pack.shortName}</span>
              </button>
            {/each}
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
                class="px-5 py-3 rounded-full bg-[#ffc3a5] hover:bg-[#ffb490] text-[#19231f] font-display font-bold text-xs border-[1.5px] border-[#19231f] cursor-pointer inline-flex items-center gap-1.5"
              >
                <StudioIcon name="stop" class="w-3.5 h-3.5" />
                <span>Cancel Scan</span>
              </button>
            {/if}
            <button
              type="submit"
              disabled={loading}
              class="studio-btn-primary px-7 py-3 text-sm tracking-tight cursor-pointer disabled:opacity-50 shrink-0"
            >
              {loading ? 'Scanning in Parallel…' : 'Scan & Appraise'}
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
              class="w-24 accent-[#5366e8]"
            />
            <span class="font-mono font-bold text-[#19231f] w-6 text-right">{minScore}</span>
          </div>
        </div>
      </div>
    </form>

    <!-- 4-Column Studio Yield & Calibrated Pricing Summary Card -->
    <div
      use:motionCard={{ delay: 0.05 }}
      class="lg:col-span-4 bento-card p-6 sm:p-7 flex flex-col justify-between bg-[#fffdf8]"
    >
      <div>
        <div class="flex items-center justify-between">
          <span class="studio-label">Yield & Pricing Calibration</span>
          {#if report}
            <span class="px-2.5 py-0.5 rounded-full text-xs font-mono bg-[#f4f1e9] border border-[#19231f]/15">
              {report.durationMs}ms
            </span>
          {/if}
        </div>

        <h3 class="mt-2 text-2xl font-display font-bold text-[#19231f]">
          Registrar Fee vs. <span class="font-serif-editorial font-normal text-3xl">Aftermarket</span>
        </h3>
      </div>

      {#if report}
        <div class="my-4 space-y-3">
          <div class="grid grid-cols-3 gap-2.5">
            <div class="rounded-2xl bg-[#dffc78] border-[1.5px] border-[#19231f] p-3.5">
              <div class="text-[11px] font-display font-bold uppercase tracking-wider text-[#19231f]/80">
                Unclaimed
              </div>
              <div class="mt-1 text-2xl font-display font-bold text-[#19231f]">
                {report.availableCount}
              </div>
            </div>

            <div class="rounded-2xl bg-[#f4f1e9] border border-[#19231f]/15 p-3.5">
              <div class="text-[11px] font-display font-bold uppercase tracking-wider text-[#48534e]">
                Taken
              </div>
              <div class="mt-1 text-2xl font-display font-bold text-[#19231f]">
                {report.takenCount}
              </div>
            </div>

            <div class="rounded-2xl bg-[#d9d6fc] border border-[#19231f]/20 p-3.5">
              <div class="text-[11px] font-display font-bold uppercase tracking-wider text-[#19231f]/80">
                Prime (82+)
              </div>
              <div class="mt-1 text-2xl font-display font-bold text-[#19231f]">
                {report.highValueCount}
              </div>
            </div>
          </div>

          <div class="space-y-1">
            <div class="flex justify-between text-xs font-mono text-[#48534e]">
              <span>Unclaimed Discovery Rate</span>
              <span class="font-semibold text-[#19231f]">{availableRatio}% of {report.totalChecked}</span>
            </div>
            <div class="h-2.5 w-full rounded-full bg-[#f4f1e9] border border-[#19231f]/15 overflow-hidden">
              <div
                class="h-full bg-[#19231f] transition-all duration-500"
                style="width: {availableRatio}%"
              ></div>
            </div>
          </div>
        </div>
      {:else}
        <div class="py-8 text-sm text-[#48534e]">
          Run a scan to inspect registry availability and calibrated registration / resale valuation across TLDs.
        </div>
      {/if}

      <div class="pt-4 border-t border-[#19231f]/10 flex items-center justify-between text-xs">
        <span class="text-[#48534e]">Verification:</span>
        <span class="font-mono text-[#19231f] font-semibold">DNS + Port-43 WHOIS + RDAP</span>
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
        'Probing Live DNS NS/A & Authoritative Port-43 WHOIS in Parallel',
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
              Verified via DNS + WHOIS so parked/squatted domains without A records are accurately marked Registered
            </p>
          </div>
          <span class="px-3 py-1 rounded-full text-xs font-mono bg-[#f4f1e9] border border-[#19231f]/15">
            {report.items.length} results
          </span>
        </div>

        <div class="divide-y divide-[#19231f]/10">
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
                    <StudioIcon name={isSaved ? 'bookmark-solid' : 'bookmark'} class="w-3 h-3" />
                  </button>

                  <span class="font-mono text-lg font-semibold tracking-tight text-[#19231f]">
                    {item.rootName}<span class="text-[#5366e8]">.{item.tld}</span>
                  </span>

                  {#if item.available}
                    <span
                      class="px-2.5 py-0.5 rounded-full text-[11px] font-display font-bold uppercase tracking-wide bg-[#dffc78] text-[#19231f] border border-[#19231f]"
                    >
                      Unclaimed ({item.valuation.regFeeDisplay || '$12/yr'})
                    </span>
                  {:else}
                    <span
                      class="px-2.5 py-0.5 rounded-full text-[11px] font-display font-semibold uppercase tracking-wide bg-[#ffc3a5]/75 text-[#19231f] border border-[#19231f]/30"
                    >
                      Registered{item.registeredAt ? ` · ${item.registeredAt.slice(0, 4)}` : ''}
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

                <p class="text-xs text-[#48534e] truncate pl-8">
                  {#if !item.available && item.registrar}
                    Registrar: {item.registrar}  ·  {item.valuation.highlights?.join('  ·  ') || ''}
                  {:else if item.valuation.highlights?.length}
                    {item.valuation.highlights.join('  ·  ')}
                  {/if}
                </p>
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
                    class="px-2.5 py-1.5 rounded-full text-xs font-display font-bold bg-[#fffdf8] hover:bg-[#dffc78] text-[#19231f] border border-[#19231f]/25 transition-colors inline-flex items-center gap-1"
                    title="Open live domain in new tab"
                  >
                    <span>Site</span>
                    <StudioIcon name="external" class="w-3 h-3" />
                  </a>

                  <button
                    type="button"
                    onclick={(e) => {
                      e.stopPropagation();
                      onCheckHistory(item.domain);
                    }}
                    class="px-2.5 py-1.5 rounded-full text-xs font-display font-semibold bg-[#fffdf8] hover:bg-[#ffc3a5] text-[#19231f] border border-[#19231f]/25 transition-colors cursor-pointer inline-flex items-center gap-1"
                    title="Check Wayback Machine Snapshots & Past Registration History"
                  >
                    <StudioIcon name="archive" class="w-3 h-3" />
                    <span>History</span>
                  </button>

                  <button
                    type="button"
                    onclick={(e) => {
                      e.stopPropagation();
                      onInspectDomain(item.domain);
                    }}
                    class="px-2.5 py-1.5 rounded-full text-xs font-display font-semibold bg-[#fffdf8] hover:bg-[#d9d6fc] text-[#19231f] border border-[#19231f]/25 transition-colors cursor-pointer"
                    title="Inspect RDAP & WHOIS Registration Date & DNS"
                  >
                    RDAP
                  </button>
                </div>
              </div>
            </div>
          {/each}
        </div>
      </div>

      <!-- 5-Column Sticky Appraisal Dossier Card -->
      <aside use:motionCard={{ delay: 0.1 }} class="lg:col-span-5 bento-card overflow-hidden lg:sticky lg:top-24">
        {#if activeItem}
          <div class="bg-[#d9d6fc] border-b border-[#19231f]/15 p-6">
            <div class="flex items-center justify-between gap-2">
              <span class="studio-label text-[#19231f]">Valuation Specimen</span>
              <div class="flex items-center gap-1.5">
                <a
                  href={`https://${activeItem.domain}`}
                  target="_blank"
                  rel="noopener noreferrer"
                  class="px-2.5 py-0.5 rounded-full text-xs font-display font-bold bg-[#fffdf8] hover:bg-[#dffc78] border border-[#19231f] text-[#19231f] inline-flex items-center gap-1"
                >
                  <span>Visit Site</span>
                  <StudioIcon name="external" class="w-3 h-3" />
                </a>
                <a
                  href={`https://web.archive.org/web/*/${activeItem.domain}`}
                  target="_blank"
                  rel="noopener noreferrer"
                  class="px-2.5 py-0.5 rounded-full text-xs font-display font-bold bg-[#19231f] text-[#dffc78] border border-[#19231f] inline-flex items-center gap-1"
                >
                  <StudioIcon name="archive" class="w-3 h-3" />
                  <span>Wayback</span>
                </a>
              </div>
            </div>

            <div class="mt-2 flex items-baseline justify-between gap-3 flex-wrap">
              <h4 class="text-2xl sm:text-3xl font-mono font-bold text-[#19231f] break-all">
                {activeItem.domain}
              </h4>
              <span
                class="px-3 py-1 rounded-full text-xs font-display font-bold bg-[#19231f] text-[#dffc78]"
              >
                {activeItem.valuation.tier}
              </span>
            </div>

            <!-- Dual Pricing Breakdown: Registrar Cost vs Aftermarket Flip -->
            <div class="mt-4 grid grid-cols-2 gap-2.5">
              <div class="rounded-2xl bg-[#fffdf8]/90 border border-[#19231f]/20 p-3.5">
                <div class="text-[10px] font-mono uppercase tracking-wider text-[#48534e]">
                  1st-Yr Registrar Cost
                </div>
                <div class="mt-0.5 text-lg font-display font-bold text-[#19231f]">
                  {activeItem.valuation.regFeeDisplay || '$12/yr Reg'}
                </div>
              </div>

              <div class="rounded-2xl bg-[#fffdf8]/90 border border-[#19231f]/20 p-3.5">
                <div class="text-[10px] font-mono uppercase tracking-wider text-[#48534e]">
                  {activeItem.available ? 'Est. Aftermarket Flip' : 'Est. Market Appraisal'}
                </div>
                <div class="mt-0.5 text-lg font-display font-bold text-[#5366e8]">
                  ${activeItem.valuation.estimatedMinUsd.toLocaleString()} – ${activeItem.valuation.estimatedMaxUsd.toLocaleString()}
                </div>
              </div>
            </div>
          </div>

          <!-- 4-Pillar Score Breakdown -->
          <div class="p-6 space-y-5 bg-[#fffdf8]">
            {#if !activeItem.available && (activeItem.registrar || activeItem.registeredAt)}
              <div class="rounded-2xl bg-[#ffc3a5]/45 border border-[#19231f]/25 p-3.5 text-xs font-mono space-y-1">
                <div class="font-bold text-[#19231f]">Authoritative WHOIS / RDAP Registry Record</div>
                {#if activeItem.registrar}
                  <div>Registrar: <strong>{activeItem.registrar}</strong></div>
                {/if}
                {#if activeItem.registeredAt}
                  <div>Registered Since: <strong>{activeItem.registeredAt}</strong></div>
                {/if}
              </div>
            {/if}

            <div class="space-y-3">
              <div class="studio-label">4-Factor Appraisal Matrix</div>

              <div class="space-y-1">
                <div class="flex justify-between text-xs font-medium">
                  <span>Brevity & Character Count ({activeItem.rootName.length} chars)</span>
                  <span class="font-mono font-bold">{activeItem.valuation.lengthScore}/35</span>
                </div>
                <div class="h-2 rounded-full bg-[#f4f1e9] overflow-hidden">
                  <div
                    class="h-full bg-[#19231f] rounded-full"
                    style="width: {(activeItem.valuation.lengthScore / 35) * 100}%"
                  ></div>
                </div>
              </div>

              <div class="space-y-1">
                <div class="flex justify-between text-xs font-medium">
                  <span>TLD Authority (.{activeItem.tld})</span>
                  <span class="font-mono font-bold">{activeItem.valuation.tldScore}/25</span>
                </div>
                <div class="h-2 rounded-full bg-[#f4f1e9] overflow-hidden">
                  <div
                    class="h-full bg-[#5366e8] rounded-full"
                    style="width: {(activeItem.valuation.tldScore / 25) * 100}%"
                  ></div>
                </div>
              </div>

              <div class="space-y-1">
                <div class="flex justify-between text-xs font-medium">
                  <span>Phonetic Pronounceability</span>
                  <span class="font-mono font-bold">{activeItem.valuation.phoneticScore}/25</span>
                </div>
                <div class="h-2 rounded-full bg-[#f4f1e9] overflow-hidden">
                  <div
                    class="h-full bg-[#19231f] rounded-full"
                    style="width: {(activeItem.valuation.phoneticScore / 25) * 100}%"
                  ></div>
                </div>
              </div>

              <div class="space-y-1">
                <div class="flex justify-between text-xs font-medium">
                  <span>Dictionary & Commercial Stem</span>
                  <span class="font-mono font-bold">{activeItem.valuation.keywordScore}/15</span>
                </div>
                <div class="h-2 rounded-full bg-[#f4f1e9] overflow-hidden">
                  <div
                    class="h-full bg-[#5366e8] rounded-full"
                    style="width: {(activeItem.valuation.keywordScore / 15) * 100}%"
                  ></div>
                </div>
              </div>
            </div>

            {#if !activeItem.available && onCheckTraffic}
              <button
                type="button"
                onclick={() => activeItem && onCheckTraffic(activeItem.domain)}
                class="w-full py-2.5 px-4 rounded-2xl text-xs font-display font-bold bg-[#19231f] hover:bg-[#5366e8] text-[#dffc78] border border-[#19231f] transition-all cursor-pointer inline-flex items-center justify-center gap-2 shadow-sm"
              >
                <StudioIcon name="chart" class="w-3.5 h-3.5" />
                <span>Check Registered Site Traffic & Global Rank →</span>
              </button>
            {/if}

            {#if onPrepareReport}
              <button
                type="button"
                onclick={() => activeItem && onPrepareReport(activeItem.domain)}
                class="w-full py-2.5 px-4 rounded-2xl text-xs font-display font-bold bg-[#dffc78] hover:bg-[#d9d6fc] text-[#19231f] border-[1.5px] border-[#19231f] transition-all cursor-pointer inline-flex items-center justify-center gap-2"
              >
                <StudioIcon name="sparkle" class="w-3.5 h-3.5" />
                <span>Prepare Full PDF Report ({activeItem.domain})</span>
              </button>
            {/if}

            <div class="grid grid-cols-2 gap-2.5 pt-1">
              <button
                type="button"
                onclick={() => activeItem && onCheckHistory(activeItem.domain)}
                class="studio-btn-primary py-2.5 px-3 text-xs text-center cursor-pointer inline-flex items-center justify-center gap-1.5"
              >
                <StudioIcon name="archive" class="w-3.5 h-3.5" />
                <span>Past History</span>
              </button>

              <button
                type="button"
                onclick={() => activeItem && onRunRecon(activeItem.domain)}
                class="studio-btn-ink py-2.5 px-3 text-xs text-center cursor-pointer inline-flex items-center justify-center gap-1.5"
              >
                <StudioIcon name="shield" class="w-3.5 h-3.5" />
                <span>Port & TLS Recon</span>
              </button>

              <button
                type="button"
                onclick={() => activeItem && onInspectDomain(activeItem.domain)}
                class="py-2.5 px-3 rounded-full text-xs font-display font-semibold bg-[#f4f1e9] hover:bg-[#d9d6fc] text-[#19231f] border border-[#19231f]/25 transition-colors cursor-pointer"
              >
                Full RDAP Dossier
              </button>

              <button
                type="button"
                onclick={() => activeItem && onToggleSave(activeItem.domain, activeItem.available)}
                class="py-2.5 px-3 rounded-full text-xs font-display font-semibold bg-[#f4f1e9] hover:bg-[#dffc78] text-[#19231f] border border-[#19231f]/25 transition-colors cursor-pointer inline-flex items-center justify-center gap-1.5"
              >
                <StudioIcon name={savedDomainsSet.has(activeItem.domain) ? 'bookmark-solid' : 'bookmark'} class="w-3.5 h-3.5" />
                <span>{savedDomainsSet.has(activeItem.domain) ? 'Saved in Vault' : 'Save to Vault'}</span>
              </button>
            </div>
          </div>
        {/if}
      </aside>
    </div>
  {/if}
</section>
