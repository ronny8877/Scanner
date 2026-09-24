<script lang="ts">
  import { onMount } from 'svelte';
  import AvailabilityScanner from './lib/components/AvailabilityScanner.svelte';
  import DomainInspector from './lib/components/DomainInspector.svelte';
  import SiteCrawler from './lib/components/SiteCrawler.svelte';
  import PortReconStudio from './lib/components/PortReconStudio.svelte';
  import WatchlistVault from './lib/components/WatchlistVault.svelte';
  import HistoryModal from './lib/components/HistoryModal.svelte';
  import JobQueueDrawer from './lib/components/JobQueueDrawer.svelte';
  import {
    checkBackendHealth,
    runDomainScan,
    runDomainInspect,
    runDomainHistory,
    runPortRecon,
    runSiteCrawl,
    runFullParallelSuite,
    fetchJobQueue,
    cancelRunningJob,
    fetchWatchlist,
    saveDomainToVault,
    removeDomainFromVault,
    recheckWatchlistParallel,
  } from './lib/api';
  import type {
    ScanReport,
    DomainInquiry,
    CrawlReport,
    HistoryReport,
    ReconReport,
    SavedDomain,
    Job,
  } from './lib/types';

  type Mode = 'scan' | 'inspect' | 'crawl' | 'recon' | 'vault';

  let activeMode = $state<Mode>('scan');
  let backendOnline = $state<boolean>(false);
  let copiedCli = $state<boolean>(false);

  // Mode 1: Bulk Availability & Dictionary Scan
  let scanReport = $state<ScanReport | null>(null);
  let scanLoading = $state<boolean>(false);

  // Mode 2: RDAP & DNS Dossier
  let inquiryData = $state<DomainInquiry | null>(null);
  let inspectLoading = $state<boolean>(false);

  // Mode 3: Site Cartography
  let crawlReport = $state<CrawlReport | null>(null);
  let crawlLoading = $state<boolean>(false);

  // Mode 4: Port, TLS & Security Surface Recon
  let reconReport = $state<ReconReport | null>(null);
  let reconLoading = $state<boolean>(false);

  // Mode 5: Saved Watchlist Vault
  let savedDomains = $state<SavedDomain[]>([]);
  let recheckingVault = $state<boolean>(false);

  // Past Registration History Modal
  let historyModalOpen = $state<boolean>(false);
  let historyLoading = $state<boolean>(false);
  let historyTarget = $state<string>('');
  let historyReport = $state<HistoryReport | null>(null);

  // Enterprise Job Queue
  let jobQueueOpen = $state<boolean>(false);
  let jobList = $state<Job[]>([]);

  let cliPreview = $state<string>('scanner scan veltrix nova --tlds com,ai,io,dev,co,app,xyz,sh --mutations');

  const savedDomainsSet = $derived(new Set(savedDomains.map((d) => d.domain.toLowerCase())));
  const runningJobsCount = $derived(jobList.filter((j) => j.status === 'RUNNING').length);
  const anyJobRunning = $derived(
    scanLoading || inspectLoading || crawlLoading || reconLoading || historyLoading || runningJobsCount > 0
  );

  onMount(async () => {
    backendOnline = await checkBackendHealth();
    savedDomains = await fetchWatchlist();
    await handleRunScan({
      keywords: ['veltrix', 'nova'],
      dictionaryPack: 'none',
      tlds: ['com', 'ai', 'io', 'dev', 'co', 'app', 'xyz', 'sh'],
      mutations: true,
      onlyAvailable: false,
      minScore: 0,
    });
    jobList = await fetchJobQueue();
  });

  async function refreshJobs() {
    jobList = await fetchJobQueue();
  }

  async function handleCancelJob(jobId?: string) {
    jobList = await cancelRunningJob(jobId);
    scanLoading = false;
    inspectLoading = false;
    crawlLoading = false;
    reconLoading = false;
    historyLoading = false;
  }

  async function handleRunScan(opts: {
    keywords: string[];
    dictionaryPack?: string;
    tlds: string[];
    mutations: boolean;
    onlyAvailable: boolean;
    minScore: number;
  }) {
    activeMode = 'scan';
    scanLoading = true;
    const dictFlag = opts.dictionaryPack && opts.dictionaryPack !== 'none' ? `--dict ${opts.dictionaryPack}` : '';
    const flags = [
      dictFlag,
      `--tlds ${opts.tlds.join(',')}`,
      opts.mutations ? '--mutations' : '--mutations=false',
      opts.onlyAvailable ? '--available' : '',
      opts.minScore > 0 ? `--min-score ${opts.minScore}` : '',
    ]
      .filter(Boolean)
      .join(' ');
    cliPreview = `scanner scan ${opts.keywords.join(' ')} ${flags}`.trim();

    const { report, liveBackend } = await runDomainScan(opts);
    scanReport = report;
    backendOnline = liveBackend;
    scanLoading = false;
    await refreshJobs();
  }

  async function handleInspectDomain(domain: string) {
    activeMode = 'inspect';
    inspectLoading = true;
    cliPreview = `scanner inspect ${domain}`;

    const { inquiry, liveBackend } = await runDomainInspect(domain);
    inquiryData = inquiry;
    backendOnline = liveBackend;
    inspectLoading = false;
    await refreshJobs();
  }

  async function handleCheckHistory(domain: string) {
    historyTarget = domain;
    historyModalOpen = true;
    historyLoading = true;
    cliPreview = `scanner history ${domain}`;

    const { history, liveBackend } = await runDomainHistory(domain);
    historyReport = history;
    backendOnline = liveBackend;
    historyLoading = false;
    await refreshJobs();
  }

  async function handleRunRecon(domain: string) {
    activeMode = 'recon';
    reconLoading = true;
    cliPreview = `scanner recon ${domain}`;

    const { recon, liveBackend } = await runPortRecon(domain);
    reconReport = recon;
    backendOnline = liveBackend;
    reconLoading = false;
    await refreshJobs();
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
    await refreshJobs();
  }

  async function handleRunCrawl(opts: { targetUrl: string; maxPages: number; maxDepth: number }) {
    activeMode = 'crawl';
    crawlLoading = true;
    cliPreview = `scanner crawl ${opts.targetUrl} --pages ${opts.maxPages} --depth ${opts.maxDepth}`;

    const { report, liveBackend } = await runSiteCrawl(opts);
    crawlReport = report;
    backendOnline = liveBackend;
    crawlLoading = false;
    await refreshJobs();
  }

  async function handleRunFullSuite(domain: string) {
    jobQueueOpen = false;
    reconLoading = true;
    inspectLoading = true;
    crawlLoading = true;
    cliPreview = `scanner recon ${domain} && scanner history ${domain} && scanner inspect ${domain}`;

    const { suite, liveBackend } = await runFullParallelSuite(domain);
    inquiryData = suite.inquiry;
    historyReport = suite.history;
    reconReport = suite.recon;
    crawlReport = suite.crawl;
    backendOnline = liveBackend;

    reconLoading = false;
    inspectLoading = false;
    crawlLoading = false;
    activeMode = 'recon';
    await refreshJobs();
  }

  async function handleToggleSave(domain: string, available: boolean) {
    const clean = domain.toLowerCase().trim();
    if (savedDomainsSet.has(clean)) {
      savedDomains = await removeDomainFromVault(clean);
    } else {
      savedDomains = await saveDomainToVault({
        domain: clean,
        available,
        notes: available ? 'High-value unclaimed domain candidate' : 'Active registered domain watch',
        tags: available ? ['Unclaimed Gem'] : ['Watch'],
      });
    }
  }

  async function handleAddVaultDomain(domain: string, notes: string, tag: string) {
    savedDomains = await saveDomainToVault({
      domain,
      notes,
      tags: [tag],
    });
  }

  async function handleRemoveVaultDomain(domain: string) {
    savedDomains = await removeDomainFromVault(domain);
  }

  async function handleRecheckVault() {
    recheckingVault = true;
    savedDomains = await recheckWatchlistParallel();
    recheckingVault = false;
    await refreshJobs();
  }

  function handleSelectJob(job: Job) {
    jobQueueOpen = false;
    if (job.type === 'scan') activeMode = 'scan';
    else if (job.type === 'inspect') activeMode = 'inspect';
    else if (job.type === 'recon' || job.type === 'parallel_suite') activeMode = 'recon';
    else if (job.type === 'crawl') activeMode = 'crawl';
    else if (job.type === 'history') {
      historyTarget = job.target;
      historyModalOpen = true;
    }
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
  <!-- RESPONSIVE FLOATING NAVIGATION (Clean 2-row card on mobile, floating pill on desktop) -->
  <div class="sticky top-3 z-40 px-3 sm:px-6">
    <header
      class="mx-auto max-w-6xl rounded-2xl lg:rounded-full bg-[#fffdf8]/95 backdrop-blur-md border-[1.5px] border-[#19231f] px-3 py-2.5 lg:py-2 shadow-[0_4px_0_#19231f,0_14px_30px_rgba(25,35,31,0.08)] flex flex-col lg:flex-row lg:items-center lg:justify-between gap-2.5"
    >
      <!-- Top Row on Mobile / Left+Right on Desktop -->
      <div class="flex items-center justify-between gap-2">
        <!-- Studio Mark -->
        <a
          href="#top"
          onclick={(e) => {
            e.preventDefault();
            activeMode = 'scan';
          }}
          class="flex items-center gap-2 pl-1.5 pr-3 py-1 rounded-full hover:bg-[#f4f1e9] transition-colors shrink-0"
        >
          <span
            class="w-7 h-7 rounded-full bg-[#19231f] text-[#dffc78] flex items-center justify-center font-mono text-sm font-bold shrink-0"
          >
            ◈
          </span>
          <span class="font-display font-bold text-sm tracking-tight text-[#19231f]">
            Scanner <span class="font-serif-editorial font-normal text-base">Studio</span>
          </span>
        </a>

        <!-- Mobile Right Actions (Queue + Cancel) -->
        <div class="flex lg:hidden items-center gap-1.5">
          {#if anyJobRunning}
            <button
              type="button"
              onclick={() => handleCancelJob()}
              class="px-2.5 py-1 rounded-full text-[11px] font-display font-bold bg-[#ffc3a5] text-[#19231f] border border-[#19231f] cursor-pointer"
            >
              ✕ Stop Job
            </button>
          {/if}

          <button
            type="button"
            onclick={async () => {
              await refreshJobs();
              jobQueueOpen = true;
            }}
            class="px-3 py-1 rounded-full text-xs font-display font-bold bg-[#d9d6fc] text-[#19231f] border border-[#19231f] cursor-pointer flex items-center gap-1"
          >
            <span>⚡ Queue ({jobList.length})</span>
            {#if runningJobsCount > 0}
              <span class="w-2 h-2 rounded-full bg-[#19231f] animate-ping"></span>
            {/if}
          </button>
        </div>
      </div>

      <!-- Mode Switcher Pills: Single-row horizontal scroll on mobile so it NEVER wraps into a tall oval -->
      <nav
        aria-label="Primary Studio Modes"
        class="w-full lg:w-auto flex items-center gap-1 overflow-x-auto whitespace-nowrap bg-[#f4f1e9] p-1 rounded-full border border-[#19231f]/15"
      >
        <button
          type="button"
          onclick={() => (activeMode = 'scan')}
          class="px-3 py-1.5 rounded-full text-xs font-display font-semibold transition-all cursor-pointer shrink-0 {activeMode ===
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
            if (!inquiryData) handleInspectDomain('supercoloring.com');
          }}
          class="px-3 py-1.5 rounded-full text-xs font-display font-semibold transition-all cursor-pointer shrink-0 {activeMode ===
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
          class="px-3 py-1.5 rounded-full text-xs font-display font-semibold transition-all cursor-pointer shrink-0 {activeMode ===
          'crawl'
            ? 'bg-[#19231f] text-[#fffdf8]'
            : 'text-[#48534e] hover:text-[#19231f]'}"
        >
          03. Site Tree
        </button>

        <button
          type="button"
          onclick={() => {
            activeMode = 'recon';
            if (!reconReport) handleRunRecon('supercoloring.com');
          }}
          class="px-3 py-1.5 rounded-full text-xs font-display font-semibold transition-all cursor-pointer shrink-0 {activeMode ===
          'recon'
            ? 'bg-[#19231f] text-[#fffdf8]'
            : 'text-[#48534e] hover:text-[#19231f]'}"
        >
          04. Ports & TLS
        </button>

        <button
          type="button"
          onclick={() => (activeMode = 'vault')}
          class="px-3 py-1.5 rounded-full text-xs font-display font-semibold transition-all cursor-pointer flex items-center gap-1.5 shrink-0 {activeMode ===
          'vault'
            ? 'bg-[#19231f] text-[#fffdf8]'
            : 'text-[#48534e] hover:text-[#19231f]'}"
        >
          <span>05. Saved Vault</span>
          <span
            class="px-1.5 py-0.2 rounded-full text-[10px] font-mono {activeMode === 'vault'
              ? 'bg-[#dffc78] text-[#19231f] font-bold'
              : 'bg-[#d9d6fc] text-[#19231f]'}"
          >
            {savedDomains.length}
          </span>
        </button>
      </nav>

      <!-- Desktop Right Actions: Queue + Cancel Job + CLI CTA -->
      <div class="hidden lg:flex items-center gap-2 shrink-0">
        {#if anyJobRunning}
          <button
            type="button"
            onclick={() => handleCancelJob()}
            class="px-3 py-1.5 rounded-full text-xs font-display font-bold bg-[#ffc3a5] hover:bg-[#ffad85] text-[#19231f] border border-[#19231f] transition-colors cursor-pointer"
            title="Cancel currently running background job"
          >
            ✕ Cancel Job
          </button>
        {/if}

        <button
          type="button"
          onclick={async () => {
            await refreshJobs();
            jobQueueOpen = true;
          }}
          class="px-3.5 py-1.5 rounded-full text-xs font-display font-bold bg-[#d9d6fc] hover:bg-[#c8c3fa] text-[#19231f] border border-[#19231f] transition-colors cursor-pointer flex items-center gap-1.5"
          title="Open Enterprise Job & Worker Queue"
        >
          <span>⚡ Queue ({jobList.length})</span>
          {#if runningJobsCount > 0}
            <span class="w-2 h-2 rounded-full bg-[#19231f] animate-ping"></span>
          {/if}
        </button>

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
  <div id="top" class="max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8 pt-8 sm:pt-10 pb-8">
    <div class="grid grid-cols-1 lg:grid-cols-12 gap-6 items-end">
      <!-- 7-Col Editorial Title & Context -->
      <div class="lg:col-span-7 space-y-3">
        <div class="inline-flex flex-wrap items-center gap-2 px-3 py-1 rounded-full bg-[#fffdf8] border border-[#19231f]/20 text-xs font-mono">
          <span
            class="w-2 h-2 rounded-full {backendOnline
              ? 'bg-[#19231f]'
              : 'bg-[#5366e8]'}"
          ></span>
          <span>
            {backendOnline
              ? 'Go Parallel Engine Online (:8080) · 24 TLDs · Dictionary Combinator · Wayback + CT + Port/TLS'
              : 'Standalone Studio · Run `./bin/scanner serve` for live sockets'}
          </span>
        </div>

        <h1 class="text-3xl sm:text-5xl lg:text-[3.35rem] font-display font-bold leading-[1.06] text-[#19231f]">
          Uncover <span class="font-serif-editorial font-normal underline decoration-[#dffc78] decoration-4 underline-offset-4">unclaimed</span> domains &amp; map the <span class="font-serif-editorial font-normal">living</span> web.
        </h1>

        <p class="text-sm sm:text-base text-[#48534e] max-w-2xl leading-relaxed">
          Editorial domain intelligence &amp; surface reconnaissance studio: 24-TLD dictionary search, calibrated
          registrar vs. aftermarket valuation, clickable Wayback Machine archives, 14-port TCP/TLS inspection, and live job cancellation.
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
          <span>Workers: Cancellable Go Pool</span>
          <span>Vault: {savedDomains.length} saved</span>
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
        onCancelJob={() => handleCancelJob()}
        onInspectDomain={handleInspectDomain}
        onCheckHistory={handleCheckHistory}
        onRunRecon={handleRunRecon}
        onCrawlDomain={handleCrawlDomain}
        onToggleSave={handleToggleSave}
        {savedDomainsSet}
      />
    {:else if activeMode === 'inspect'}
      <DomainInspector
        inquiry={inquiryData}
        loading={inspectLoading}
        onInspect={handleInspectDomain}
        onCancelJob={() => handleCancelJob()}
        onCheckHistory={handleCheckHistory}
        onRunRecon={handleRunRecon}
        onCrawlDomain={handleCrawlDomain}
        onToggleSave={handleToggleSave}
        {savedDomainsSet}
      />
    {:else if activeMode === 'crawl'}
      <SiteCrawler
        report={crawlReport}
        loading={crawlLoading}
        onRunCrawl={handleRunCrawl}
        onInspectDomain={handleInspectDomain}
      />
    {:else if activeMode === 'recon'}
      <PortReconStudio
        report={reconReport}
        loading={reconLoading}
        onRunRecon={handleRunRecon}
        onRunFullSuite={handleRunFullSuite}
        onCancelJob={() => handleCancelJob()}
        onCheckHistory={handleCheckHistory}
        onSaveDomain={handleToggleSave}
        {savedDomainsSet}
      />
    {:else if activeMode === 'vault'}
      <WatchlistVault
        items={savedDomains}
        rechecking={recheckingVault}
        onAddDomain={handleAddVaultDomain}
        onRemoveDomain={handleRemoveVaultDomain}
        onRecheckAll={handleRecheckVault}
        onInspectDomain={handleInspectDomain}
        onCheckHistory={handleCheckHistory}
        onRunRecon={handleRunRecon}
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
        <div class="md:col-span-3 bento-card p-5 space-y-2.5 bg-[#fffdf8]">
          <div class="flex items-center justify-between">
            <span class="studio-label">01 · Dictionary Scan</span>
            <span class="px-2 py-0.5 rounded-full text-[11px] font-display font-bold bg-[#dffc78] border border-[#19231f]">
              24 TLDs
            </span>
          </div>
          <p class="text-xs text-[#48534e] leading-relaxed">
            Scan curated dictionary packs across 24 TLDs with calibrated registrar &amp; flip pricing.
          </p>
          <pre class="rounded-xl bg-[#f4f1e9] border border-[#19231f]/15 p-3 text-xs font-mono text-[#19231f] overflow-x-auto">./bin/scanner veltrix --available</pre>
        </div>

        <div class="md:col-span-3 bento-card p-5 space-y-2.5 bg-[#fffdf8]">
          <div class="flex items-center justify-between">
            <span class="studio-label">02 · Past History</span>
            <span class="px-2 py-0.5 rounded-full text-[11px] font-display font-bold bg-[#ffc3a5] border border-[#19231f]">
              Wayback + RDAP + CT
            </span>
          </div>
          <p class="text-xs text-[#48534e] leading-relaxed">
            Inspect Wayback yearly snapshots, RDAP original creation date, and historical TLS logs.
          </p>
          <pre class="rounded-xl bg-[#f4f1e9] border border-[#19231f]/15 p-3 text-xs font-mono text-[#19231f] overflow-x-auto">./bin/scanner history supercoloring.com</pre>
        </div>

        <div class="md:col-span-3 bento-card p-5 space-y-2.5 bg-[#fffdf8]">
          <div class="flex items-center justify-between">
            <span class="studio-label">03 · Port & TLS Recon</span>
            <span class="px-2 py-0.5 rounded-full text-[11px] font-display font-bold bg-[#d9d6fc] border border-[#19231f]">
              28 Workers
            </span>
          </div>
          <p class="text-xs text-[#48534e] leading-relaxed">
            Probe 14 TCP ports, HTTP header ledger, `robots.txt`/`security.txt`, and CNAME takeover risk.
          </p>
          <pre class="rounded-xl bg-[#f4f1e9] border border-[#19231f]/15 p-3 text-xs font-mono text-[#19231f] overflow-x-auto">./bin/scanner recon supercoloring.com</pre>
        </div>

        <div class="md:col-span-3 bento-card p-5 space-y-2.5 bg-[#fffdf8]">
          <div class="flex items-center justify-between">
            <span class="studio-label">04 · Site Tree</span>
            <span class="px-2 py-0.5 rounded-full text-[11px] font-display font-bold bg-[#dffc78] border border-[#19231f]">
              Crawler
            </span>
          </div>
          <p class="text-xs text-[#48534e] leading-relaxed">
            Crawl internal link graphs and render a box-drawing hierarchy in terminal.
          </p>
          <pre class="rounded-xl bg-[#f4f1e9] border border-[#19231f]/15 p-3 text-xs font-mono text-[#19231f] overflow-x-auto">./bin/scanner crawl svelte.dev</pre>
        </div>
      </div>
    </section>
  </main>

  <!-- PAST REGISTRATION HISTORY DOSSIER MODAL -->
  <HistoryModal
    open={historyModalOpen}
    loading={historyLoading}
    targetDomain={historyTarget}
    report={historyReport}
    onClose={() => (historyModalOpen = false)}
    onCancelJob={() => handleCancelJob()}
    onInspectRDAP={handleInspectDomain}
    onRunRecon={handleRunRecon}
  />

  <!-- ENTERPRISE JOB & WORKER QUEUE DRAWER -->
  <JobQueueDrawer
    open={jobQueueOpen}
    jobs={jobList}
    onClose={() => (jobQueueOpen = false)}
    onSelectJob={handleSelectJob}
    onDispatchParallelSuite={handleRunFullSuite}
    onCancelJob={handleCancelJob}
  />
</div>
