<script lang="ts">
  import { onMount } from 'svelte';
  import AvailabilityScanner from './lib/components/AvailabilityScanner.svelte';
  import DomainInspector from './lib/components/DomainInspector.svelte';
  import SiteCrawler from './lib/components/SiteCrawler.svelte';
  import { checkBackendHealth, runDomainScan, runDomainInspect, runSiteCrawl } from './lib/api';
  import type { ScanReport, DomainInquiry, CrawlReport } from './lib/types';

  type Mode = 'scan' | 'inspect' | 'crawl';

  let activeMode = $state<Mode>('scan');
  let backendOnline = $state<boolean>(false);
  let copiedCli = $state<boolean>(false);

  let scanReport = $state<ScanReport | null>(null);
  let scanLoading = $state<boolean>(false);

  let inquiryData = $state<DomainInquiry | null>(null);
  let inspectLoading = $state<boolean>(false);

  let crawlReport = $state<CrawlReport | null>(null);
  let crawlLoading = $state<boolean>(false);

  let cliPreview = $state<string>('scanner scan veltrix nova --tlds com,ai,io,dev,co,app --mutations');

  onMount(async () => {
    backendOnline = await checkBackendHealth();
    await handleRunScan({
      keywords: ['veltrix', 'nova'],
      tlds: ['com', 'ai', 'io', 'dev', 'co', 'app'],
      mutations: true,
      onlyAvailable: false,
      minScore: 0,
    });
  });

  async function handleRunScan(opts: {
    keywords: string[];
    tlds: string[];
    mutations: boolean;
    onlyAvailable: boolean;
    minScore: number;
  }) {
    activeMode = 'scan';
    scanLoading = true;
    const flags = [
      `--tlds ${opts.tlds.join(',')}`,
      opts.mutations ? '--mutations' : '--mutations=false',
      opts.onlyAvailable ? '--available' : '',
      opts.minScore > 0 ? `--min-score ${opts.minScore}` : '',
    ]
      .filter(Boolean)
      .join(' ');
    cliPreview = `scanner scan ${opts.keywords.join(' ')} ${flags}`;

    const { report, liveBackend } = await runDomainScan(opts);
    scanReport = report;
    backendOnline = liveBackend;
    scanLoading = false;
  }

  async function handleInspectDomain(domain: string) {
    activeMode = 'inspect';
    inspectLoading = true;
    cliPreview = `scanner inspect ${domain}`;

    const { inquiry, liveBackend } = await runDomainInspect(domain);
    inquiryData = inquiry;
    backendOnline = liveBackend;
    inspectLoading = false;
  }

  async function handleCrawlDomain(target: string) {
    activeMode = 'crawl';
    crawlLoading = true;
    cliPreview = `scanner crawl ${target} --pages 20 --depth 2`;

    const { report, liveBackend } = await runSiteCrawl({
      targetUrl: target,
      maxPages: 20,
      maxDepth: 2,
    });
    crawlReport = report;
    backendOnline = liveBackend;
    crawlLoading = false;
  }

  async function handleRunCrawl(opts: { targetUrl: string; maxPages: number; maxDepth: number }) {
    activeMode = 'crawl';
    crawlLoading = true;
    cliPreview = `scanner crawl ${opts.targetUrl} --pages ${opts.maxPages} --depth ${opts.maxDepth}`;

    const { report, liveBackend } = await runSiteCrawl(opts);
    crawlReport = report;
    backendOnline = liveBackend;
    crawlLoading = false;
  }

  function copyCliCommand() {
    navigator.clipboard?.writeText(cliPreview);
    copiedCli = true;
    setTimeout(() => {
      copiedCli = false;
    }, 1800);
  }
</script>

<div class="min-h-dvh flex flex-col pb-16">
  <!-- FLOATING PILL NAVIGATION -->
  <div class="sticky top-4 z-40 px-4 sm:px-6">
    <header
      class="mx-auto max-w-5xl rounded-full bg-[#fffdf8]/95 backdrop-blur-md border-[1.5px] border-[#19231f] px-3 py-2 shadow-[0_4px_0_#19231f,0_14px_30px_rgba(25,35,31,0.08)] flex items-center justify-between gap-2"
    >
      <!-- Studio Mark -->
      <a
        href="#top"
        onclick={(e) => {
          e.preventDefault();
          activeMode = 'scan';
        }}
        class="flex items-center gap-2.5 pl-2 pr-3 py-1 rounded-full hover:bg-[#f4f1e9] transition-colors"
      >
        <span
          class="w-7 h-7 rounded-full bg-[#19231f] text-[#dffc78] flex items-center justify-center font-mono text-sm font-bold"
        >
          ◈
        </span>
        <span class="font-display font-bold text-sm tracking-tight text-[#19231f]">
          Scanner <span class="font-serif-editorial font-normal text-base">Studio</span>
        </span>
      </a>

      <!-- Mode Switcher Pills -->
      <nav aria-label="Primary Studio Modes" class="flex items-center gap-1 bg-[#f4f1e9] p-1 rounded-full border border-[#19231f]/15">
        <button
          type="button"
          onclick={() => (activeMode = 'scan')}
          class="px-3.5 py-1.5 rounded-full text-xs font-display font-semibold transition-all cursor-pointer {activeMode ===
          'scan'
            ? 'bg-[#19231f] text-[#fffdf8]'
            : 'text-[#48534e] hover:text-[#19231f]'}"
        >
          01. Find & Value
        </button>

        <button
          type="button"
          onclick={() => {
            activeMode = 'inspect';
            if (!inquiryData) handleInspectDomain('svelte.dev');
          }}
          class="px-3.5 py-1.5 rounded-full text-xs font-display font-semibold transition-all cursor-pointer {activeMode ===
          'inspect'
            ? 'bg-[#19231f] text-[#fffdf8]'
            : 'text-[#48534e] hover:text-[#19231f]'}"
        >
          02. RDAP Dossier
        </button>

        <button
          type="button"
          onclick={() => {
            activeMode = 'crawl';
            if (!crawlReport) handleCrawlDomain('svelte.dev');
          }}
          class="px-3.5 py-1.5 rounded-full text-xs font-display font-semibold transition-all cursor-pointer {activeMode ===
          'crawl'
            ? 'bg-[#19231f] text-[#fffdf8]'
            : 'text-[#48534e] hover:text-[#19231f]'}"
        >
          03. Site Cartography
        </button>
      </nav>

      <!-- Prominent Studio Editor / CLI Action CTA -->
      <div class="flex items-center gap-2">
        <button
          type="button"
          onclick={copyCliCommand}
          class="studio-btn-primary px-4 py-1.5 text-xs cursor-pointer flex items-center gap-1.5"
          title="Copy active Go CLI command"
        >
          <span>{copiedCli ? '✓ CLI Copied' : 'Copy CLI Cmd'}</span>
        </button>
      </div>
    </header>
  </div>

  <!-- EDITORIAL STUDIO HERO & LIVE CLI BENTO -->
  <div id="top" class="max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8 pt-10 pb-8">
    <div class="grid grid-cols-1 lg:grid-cols-12 gap-6 items-end">
      <!-- 7-Col Editorial Title & Context -->
      <div class="lg:col-span-7 space-y-3">
        <div class="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-[#fffdf8] border border-[#19231f]/20 text-xs font-mono">
          <span
            class="w-2 h-2 rounded-full {backendOnline
              ? 'bg-[#19231f]'
              : 'bg-[#5366e8]'}"
          ></span>
          <span>
            {backendOnline
              ? 'Go Engine Connected (:8080) · Live DNS & RDAP'
              : 'Standalone Studio · Start `./bin/scanner serve` for live sockets'}
          </span>
        </div>

        <h1 class="text-4xl sm:text-5xl lg:text-[3.35rem] font-display font-bold leading-[1.06] text-[#19231f]">
          Uncover <span class="font-serif-editorial font-normal underline decoration-[#dffc78] decoration-4 underline-offset-4">unclaimed</span> domains &amp; map the <span class="font-serif-editorial font-normal">living</span> web.
        </h1>

        <p class="text-base text-[#48534e] max-w-2xl leading-relaxed">
          An editorial domain intelligence workbench pairing a concurrent Go CLI with interactive valuation scoring,
          authoritative ICANN RDAP registration chronology, and live site-tree cartography.
        </p>
      </div>

      <!-- 5-Col Live Synchronized Go CLI Terminal Card -->
      <div class="lg:col-span-5 bento-card-ink p-5 space-y-3 shadow-lg">
        <div class="flex items-center justify-between text-xs">
          <div class="flex items-center gap-2">
            <span class="w-2.5 h-2.5 rounded-full bg-[#dffc78]"></span>
            <span class="font-display font-bold uppercase tracking-wider text-[#fffdf8]/80">
              Synchronized Go CLI Invocation
            </span>
          </div>
          <button
            type="button"
            onclick={copyCliCommand}
            class="px-2.5 py-0.5 rounded-full text-[11px] font-mono bg-[#fffdf8]/10 hover:bg-[#dffc78] hover:text-[#19231f] text-[#fffdf8] transition-colors cursor-pointer"
          >
            {copiedCli ? 'Copied!' : 'Copy'}
          </button>
        </div>

        <div class="rounded-xl bg-[#0f1613] border border-[#fffdf8]/15 px-4 py-3 font-mono text-xs text-[#dffc78] overflow-x-auto">
          <span class="text-[#fffdf8]/50 select-none">$ </span>./bin/{cliPreview}
        </div>

        <div class="flex items-center justify-between text-[11px] text-[#fffdf8]/70 font-mono">
          <span>Binary: ./bin/scanner</span>
          <span>API: http://localhost:8080</span>
        </div>
      </div>
    </div>
  </div>

  <!-- ACTIVE WORKBENCH BENTO -->
  <main class="flex-1 max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8 space-y-12">
    {#if activeMode === 'scan'}
      <AvailabilityScanner
        report={scanReport}
        loading={scanLoading}
        onRunScan={handleRunScan}
        onInspectDomain={handleInspectDomain}
        onCrawlDomain={handleCrawlDomain}
      />
    {:else if activeMode === 'inspect'}
      <DomainInspector
        inquiry={inquiryData}
        loading={inspectLoading}
        onInspect={handleInspectDomain}
        onCrawlDomain={handleCrawlDomain}
      />
    {:else if activeMode === 'crawl'}
      <SiteCrawler
        report={crawlReport}
        loading={crawlLoading}
        onRunCrawl={handleRunCrawl}
        onInspectDomain={handleInspectDomain}
      />
    {/if}

    <!-- BOTTOM STUDIO REFERENCE BENTO: Real CLI Guides & Capabilities -->
    <section aria-labelledby="cli-field-guide" class="pt-6 border-t border-[#19231f]/12 space-y-5">
      <div class="flex flex-wrap items-baseline justify-between gap-2">
        <h2 id="cli-field-guide" class="font-display font-bold text-xl text-[#19231f]">
          Terminal Field <span class="font-serif-editorial font-normal text-2xl">Guide</span>
        </h2>
        <span class="text-xs text-[#48534e]">Every UI action maps 1:1 to a local Go subcommand</span>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-12 gap-5">
        <div class="md:col-span-5 bento-card p-5 space-y-2.5 bg-[#fffdf8]">
          <div class="flex items-center justify-between">
            <span class="studio-label">Default Mode · Bulk Discovery</span>
            <span class="px-2 py-0.5 rounded-full text-[11px] font-display font-bold bg-[#dffc78] border border-[#19231f]">
              Valuation Engine
            </span>
          </div>
          <p class="text-xs text-[#48534e] leading-relaxed">
            Scan seed words across multiple TLDs with automatic affix mutations and filter strictly for unclaimed domains.
          </p>
          <pre class="rounded-xl bg-[#f4f1e9] border border-[#19231f]/15 p-3 text-xs font-mono text-[#19231f] overflow-x-auto">./bin/scanner veltrix nova --available --min-score 75</pre>
        </div>

        <div class="md:col-span-4 bento-card p-5 space-y-2.5 bg-[#fffdf8]">
          <div class="flex items-center justify-between">
            <span class="studio-label">Mode 02 · RDAP Custody</span>
            <span class="px-2 py-0.5 rounded-full text-[11px] font-display font-bold bg-[#d9d6fc] border border-[#19231f]">
              WHOIS + DNS
            </span>
          </div>
          <p class="text-xs text-[#48534e] leading-relaxed">
            Query creation timestamps, domain age, registrar custody, and live NS/MX/TXT records.
          </p>
          <pre class="rounded-xl bg-[#f4f1e9] border border-[#19231f]/15 p-3 text-xs font-mono text-[#19231f] overflow-x-auto">./bin/scanner inspect svelte.dev</pre>
        </div>

        <div class="md:col-span-3 bento-card p-5 space-y-2.5 bg-[#fffdf8]">
          <div class="flex items-center justify-between">
            <span class="studio-label">Mode 03 · Site Tree</span>
            <span class="px-2 py-0.5 rounded-full text-[11px] font-display font-bold bg-[#ffc3a5] border border-[#19231f]">
              Crawler
            </span>
          </div>
          <p class="text-xs text-[#48534e] leading-relaxed">
            Crawl internal routes and print a box-drawing site hierarchy in your terminal.
          </p>
          <pre class="rounded-xl bg-[#f4f1e9] border border-[#19231f]/15 p-3 text-xs font-mono text-[#19231f] overflow-x-auto">./bin/scanner crawl svelte.dev -p 20</pre>
        </div>
      </div>
    </section>
  </main>
</div>
