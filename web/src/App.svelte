<script lang="ts">
  import { onMount } from 'svelte';
  import AvailabilityScanner from './lib/components/AvailabilityScanner.svelte';
  import DomainInspector from './lib/components/DomainInspector.svelte';
  import SiteCrawler from './lib/components/SiteCrawler.svelte';
  import TrafficIntelStudio from './lib/components/TrafficIntelStudio.svelte';
  import PortReconStudio from './lib/components/PortReconStudio.svelte';
  import RobotsSitemapStudio from './lib/components/RobotsSitemapStudio.svelte';
  import SocialMetaStudio from './lib/components/SocialMetaStudio.svelte';
  import WatchlistVault from './lib/components/WatchlistVault.svelte';
  import HistoryModal from './lib/components/HistoryModal.svelte';
  import JobQueueDrawer from './lib/components/JobQueueDrawer.svelte';
  import ReportDock from './lib/components/ReportDock.svelte';
  import ExecutiveReportModal from './lib/components/ExecutiveReportModal.svelte';
  import StudioIcon from './lib/components/StudioIcon.svelte';
  import {
    checkBackendHealth,
    runDomainScan,
    runDomainInspect,
    runDomainHistory,
    runTrafficCheck,
    runPortRecon,
    runSiteCrawl,
    runRobotsSitemapCheck,
    runMetaSocialCheck,
    runFullParallelSuite,
    fetchJobQueue,
    cancelRunningJob,
    fetchWatchlist,
    saveDomainToVault,
    removeDomainFromVault,
    recheckWatchlistParallel,
    synthesizeScanReport,
    synthesizeDomainInquiry,
    synthesizeHistoryReport,
    synthesizeTrafficReport,
    synthesizeCrawlReport,
    synthesizeReconReport,
    synthesizeRobotsSitemapReport,
    synthesizeMetaSocialReport,
  } from './lib/api';
  import type {
    ScanReport,
    DomainInquiry,
    CrawlReport,
    HistoryReport,
    TrafficReport,
    ReconReport,
    RobotsSitemapReport,
    MetaSocialReport,
    SavedDomain,
    SavedExecutiveReport,
    Job,
  } from './lib/types';

  type Mode = 'scan' | 'inspect' | 'crawl' | 'traffic' | 'recon' | 'robots' | 'meta' | 'vault';

  let activeMode = $state<Mode>('scan');
  let moreToolsOpen = $state<boolean>(false);
  let backendOnline = $state<boolean>(false);
  let copiedCli = $state<boolean>(false);
  let copiedGuideIndex = $state<number | null>(null);

  // Mode 1: Bulk Availability & Dictionary Scan (Initialized synchronously — zero network request on screen load)
  let scanReport = $state<ScanReport | null>(
    synthesizeScanReport({
      keywords: ['veltrix', 'nova'],
      dictionaryPack: 'none',
      tlds: ['com', 'ai', 'io', 'dev', 'co', 'app', 'xyz', 'sh'],
      mutations: true,
      onlyAvailable: false,
      minScore: 0,
    })
  );
  let scanLoading = $state<boolean>(false);

  // Mode 2: RDAP & DNS Dossier
  let inquiryData = $state<DomainInquiry | null>(synthesizeDomainInquiry('svelte.dev'));
  let inspectLoading = $state<boolean>(false);

  // Mode 3: Site Cartography
  let crawlReport = $state<CrawlReport | null>(synthesizeCrawlReport('svelte.dev/docs/kit'));
  let crawlLoading = $state<boolean>(false);

  // Mode 4: Registered Site Traffic & Global Rank Intelligence
  let trafficReport = $state<TrafficReport>(synthesizeTrafficReport('cloudflare.com'));
  let trafficLoading = $state<boolean>(false);

  // Mode 5: Port, TLS & Security Surface Recon
  let reconReport = $state<ReconReport | null>(synthesizeReconReport('svelte.dev'));
  let reconLoading = $state<boolean>(false);

  // Mode 6: Robots.txt & Sitemap.xml Governance
  let robotsReport = $state<RobotsSitemapReport | null>(synthesizeRobotsSitemapReport('svelte.dev'));
  let robotsLoading = $state<boolean>(false);

  // Mode 7: Social Meta Cards & Ad/Tracker Radar
  let metaReport = $state<MetaSocialReport | null>(synthesizeMetaSocialReport('svelte.dev'));
  let metaLoading = $state<boolean>(false);

  // Mode 8: Saved Watchlist Vault
  let savedDomains = $state<SavedDomain[]>([]);
  let recheckingVault = $state<boolean>(false);

  // Past Registration History Modal
  let historyModalOpen = $state<boolean>(false);
  let historyLoading = $state<boolean>(false);
  let historyTarget = $state<string>('');
  let historyReport = $state<HistoryReport | null>(null);

  // Executive Reports Dock (LinkedIn-style bottom-right window) & Full-Screen PDF Modal
  let savedReports = $state<SavedExecutiveReport[]>([
    {
      id: 'seed-cloudflare',
      domain: 'cloudflare.com',
      generatedAt: new Date().toISOString(),
      suite: {
        domain: 'cloudflare.com',
        generatedAt: new Date().toISOString(),
        inquiry: synthesizeDomainInquiry('cloudflare.com'),
        history: synthesizeHistoryReport('cloudflare.com'),
        traffic: synthesizeTrafficReport('cloudflare.com'),
        recon: synthesizeReconReport('cloudflare.com'),
        crawl: synthesizeCrawlReport('cloudflare.com'),
        robots: synthesizeRobotsSitemapReport('cloudflare.com'),
        meta: synthesizeMetaSocialReport('cloudflare.com'),
      },
    },
  ]);
  let activeExecutiveReport = $state<SavedExecutiveReport | null>(null);
  let preparingDomain = $state<string | null>(null);

  // Enterprise Job Queue
  let jobQueueOpen = $state<boolean>(false);
  let jobList = $state<Job[]>([]);

  let cliPreview = $state<string>('scanner scan veltrix nova --tlds com,ai,io,dev,co,app,xyz,sh --mutations');

  const moreToolsItems: {
    mode: Mode;
    code: string;
    label: string;
    desc: string;
    badge: string;
  }[] = [
    {
      mode: 'traffic',
      code: '04',
      label: 'Traffic & Rank',
      desc: 'Tranco Top-1M rank, Cloudflare Radar bucket & honest monthly visit range',
      badge: 'New',
    },
    {
      mode: 'recon',
      code: '05',
      label: 'Ports & TLS',
      desc: '14 TCP ports, TLS 1.3 SANs, wildcard-filtered subdomains & security grade',
      badge: '28 Workers',
    },
    {
      mode: 'robots',
      code: '06',
      label: 'Robots & Sitemap',
      desc: 'AI/Search crawler governance matrix & XML sitemap lastmod freshness',
      badge: '12 Bots',
    },
    {
      mode: 'meta',
      code: '07',
      label: 'Social Meta & Ads',
      desc: 'Discord, Telegram, WhatsApp & FB card previews + ad/tracker detector',
      badge: 'Cards',
    },
    {
      mode: 'vault',
      code: '08',
      label: 'Saved Vault',
      desc: 'Persistent domain watchlist with 1-click parallel availability re-verification',
      badge: 'Watchlist',
    },
  ];

  const activeMoreTool = $derived(moreToolsItems.find((item) => item.mode === activeMode) || null);

  const terminalGuides = [
    {
      label: '01 · Dictionary Scan',
      badge: '24 TLDs',
      badgeBg: 'bg-[#dffc78]',
      desc: 'Scan curated dictionary packs across 24 TLDs with calibrated registrar & flip pricing.',
      cmd: './bin/scanner --dict ai_agents --tlds com,ai,io,dev,co',
    },
    {
      label: '02 · Traffic & Rank',
      badge: 'Tranco + CF',
      badgeBg: 'bg-[#ffc3a5]',
      desc: 'Check Tranco Top-1M trajectory, Cloudflare rank bucket, and monthly traffic range.',
      cmd: './bin/scanner inspect cloudflare.com',
    },
    {
      label: '03 · Sub-URL Site Tree',
      badge: 'Path Crawler',
      badgeBg: 'bg-[#dffc78]',
      desc: 'Crawl starting from a specific documentation or app route and map its hierarchy tree.',
      cmd: './bin/scanner crawl svelte.dev/docs/kit --pages 20 --depth 2',
    },
    {
      label: '04 · Port & Ad Radar',
      badge: '28 Workers',
      badgeBg: 'bg-[#d9d6fc]',
      desc: 'Probe 14 TCP ports, TLS SANs, HTTP headers, and detect Ad Networks & Analytics.',
      cmd: './bin/scanner recon svelte.dev',
    },
  ];

  const savedDomainsSet = $derived(new Set(savedDomains.map((d) => d.domain.toLowerCase())));
  const runningJobsCount = $derived(jobList.filter((j) => j.status === 'RUNNING').length);
  const anyJobRunning = $derived(
    scanLoading ||
      inspectLoading ||
      crawlLoading ||
      trafficLoading ||
      reconLoading ||
      robotsLoading ||
      metaLoading ||
      historyLoading ||
      Boolean(preparingDomain) ||
      runningJobsCount > 0
  );

  onMount(async () => {
    try {
      const storedReports = localStorage.getItem('scanner_executive_reports');
      if (storedReports) {
        const parsed = JSON.parse(storedReports);
        if (Array.isArray(parsed) && parsed.length > 0) {
          savedReports = parsed;
        }
      }
    } catch {
      // ignore localStorage parse error
    }

    const [online, vault, jobs] = await Promise.all([
      checkBackendHealth(),
      fetchWatchlist(),
      fetchJobQueue(),
    ]);
    backendOnline = online;
    savedDomains = vault;
    jobList = jobs;
  });

  function persistReports(list: SavedExecutiveReport[]) {
    savedReports = list;
    try {
      localStorage.setItem('scanner_executive_reports', JSON.stringify(list.slice(0, 12)));
    } catch {
      // ignore quota errors
    }
  }

  async function refreshJobs() {
    jobList = await fetchJobQueue();
  }

  async function handleCancelJob(jobId?: string) {
    jobList = await cancelRunningJob(jobId);
    scanLoading = false;
    inspectLoading = false;
    crawlLoading = false;
    trafficLoading = false;
    reconLoading = false;
    robotsLoading = false;
    metaLoading = false;
    historyLoading = false;
    preparingDomain = null;
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
    moreToolsOpen = false;
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

  async function handleRunTraffic(domain: string) {
    activeMode = 'traffic';
    moreToolsOpen = false;
    trafficLoading = true;
    cliPreview = `scanner traffic ${domain}`;

    const { report, liveBackend } = await runTrafficCheck(domain);
    trafficReport = report;
    backendOnline = liveBackend;
    trafficLoading = false;
    await refreshJobs();
  }

  async function handleRunRecon(domain: string) {
    activeMode = 'recon';
    moreToolsOpen = false;
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
    moreToolsOpen = false;
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
    moreToolsOpen = false;
    crawlLoading = true;
    cliPreview = `scanner crawl ${opts.targetUrl} --pages ${opts.maxPages} --depth ${opts.maxDepth}`;

    const { report, liveBackend } = await runSiteCrawl(opts);
    crawlReport = report;
    backendOnline = liveBackend;
    crawlLoading = false;
    await refreshJobs();
  }

  async function handleRunRobotsCheck(target: string) {
    activeMode = 'robots';
    moreToolsOpen = false;
    robotsLoading = true;
    cliPreview = `scanner robots ${target}`;

    const { report, liveBackend } = await runRobotsSitemapCheck(target);
    robotsReport = report;
    backendOnline = liveBackend;
    robotsLoading = false;
    await refreshJobs();
  }

  async function handleRunMetaCheck(targetUrl: string) {
    activeMode = 'meta';
    moreToolsOpen = false;
    metaLoading = true;
    cliPreview = `scanner meta ${targetUrl}`;

    const { report, liveBackend } = await runMetaSocialCheck(targetUrl);
    metaReport = report;
    backendOnline = liveBackend;
    metaLoading = false;
    await refreshJobs();
  }

  async function handlePrepareReport(domainInput: string) {
    const clean = domainInput.trim().toLowerCase().replace(/^https?:\/\//, '').split('/')[0];
    if (!clean) return;
    jobQueueOpen = false;
    moreToolsOpen = false;
    preparingDomain = clean;
    cliPreview = `scanner report ${clean} --all-engines`;

    const { suite, liveBackend } = await runFullParallelSuite(clean);
    inquiryData = suite.inquiry;
    historyReport = suite.history;
    if (suite.traffic) trafficReport = suite.traffic;
    reconReport = suite.recon;
    crawlReport = suite.crawl;
    if (suite.robots) robotsReport = suite.robots;
    if (suite.meta) metaReport = suite.meta;
    backendOnline = liveBackend;

    const newReport: SavedExecutiveReport = {
      id: `${clean}-${Date.now()}`,
      domain: clean,
      generatedAt: suite.generatedAt || new Date().toISOString(),
      suite,
    };

    const updated = [newReport, ...savedReports.filter((r) => r.domain !== clean)];
    persistReports(updated);
    preparingDomain = null;
    activeExecutiveReport = newReport;
    await refreshJobs();
  }

  function handleDeleteReport(id: string) {
    const updated = savedReports.filter((r) => r.id !== id);
    persistReports(updated);
    if (activeExecutiveReport?.id === id) {
      activeExecutiveReport = null;
    }
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
    else if (job.type === 'traffic') activeMode = 'traffic';
    else if (job.type === 'recon' || job.type === 'parallel_suite') activeMode = 'recon';
    else if (job.type === 'crawl') activeMode = 'crawl';
    else if (job.type === 'robots') activeMode = 'robots';
    else if (job.type === 'meta') activeMode = 'meta';
    else if (job.type === 'history') {
      historyTarget = job.target;
      historyModalOpen = true;
    }
  }

  function copyCliCommand() {
    navigator.clipboard?.writeText(`./bin/${cliPreview}`);
    copiedCli = true;
    setTimeout(() => {
      copiedCli = false;
    }, 1800);
  }

  function copyGuideCommand(cmd: string, idx: number) {
    navigator.clipboard?.writeText(cmd);
    copiedGuideIndex = idx;
    setTimeout(() => {
      if (copiedGuideIndex === idx) copiedGuideIndex = null;
    }, 1800);
  }
</script>

<svelte:window
  onkeydown={(e) => {
    if (e.key === 'Escape' && moreToolsOpen) {
      moreToolsOpen = false;
    }
  }}
  onclick={(e) => {
    if (!moreToolsOpen) return;
    const target = e.target as HTMLElement | null;
    if (target && !target.closest('[data-burger-menu]')) {
      moreToolsOpen = false;
    }
  }}
/>

<div class="min-h-dvh flex flex-col pb-20">
  <!-- RESPONSIVE FLOATING NAVIGATION: First 3 Primary Tools + Burger Menu -->
  <div class="sticky top-3 z-40 px-3 sm:px-6 print:hidden">
    <header
      class="relative mx-auto max-w-7xl rounded-2xl xl:rounded-full bg-[#fffdf8]/95 backdrop-blur-md border-[1.5px] border-[#19231f] px-3.5 py-2.5 xl:py-2 shadow-[0_4px_0_#19231f,0_14px_30px_rgba(25,35,31,0.08)] flex flex-col xl:flex-row xl:items-center xl:justify-between gap-2 overflow-visible"
    >
      <!-- Brand + Mobile Actions -->
      <div class="flex items-center justify-between gap-2 shrink-0">
        <a
          href="#top"
          onclick={(e) => {
            e.preventDefault();
            activeMode = 'scan';
            moreToolsOpen = false;
          }}
          class="flex items-center gap-2 pl-1 pr-2.5 py-1 rounded-full hover:bg-[#f4f1e9] transition-colors shrink-0"
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

        <!-- Mobile/Tablet Right Actions -->
        <div class="flex xl:hidden items-center gap-1.5">
          {#if anyJobRunning}
            <button
              type="button"
              onclick={() => handleCancelJob()}
              class="px-2.5 py-1 rounded-full text-[11px] font-display font-bold bg-[#ffc3a5] text-[#19231f] border border-[#19231f] cursor-pointer"
            >
              ✕ Stop
            </button>
          {/if}

          <button
            type="button"
            onclick={async () => {
              await refreshJobs();
              jobQueueOpen = true;
            }}
            class="px-3 py-1 rounded-full text-xs font-display font-bold bg-[#d9d6fc] text-[#19231f] border border-[#19231f] cursor-pointer inline-flex items-center gap-1.5"
          >
            <StudioIcon name="bolt" size={12} />
            <span>Queue ({jobList.length})</span>
            {#if anyJobRunning}
              <span class="w-2 h-2 rounded-full bg-[#19231f] animate-ping"></span>
            {/if}
          </button>
        </div>
      </div>

      <!-- Mode Strip: Only First 3 Tools Visible + Burger Menu Button for Remaining Tools -->
      <nav
        aria-label="Primary Studio Modes"
        class="relative w-full xl:w-auto flex flex-wrap xl:flex-nowrap items-center justify-center gap-1 bg-[#f4f1e9] p-1 rounded-2xl xl:rounded-full border border-[#19231f]/15 overflow-visible"
      >
        <button
          type="button"
          onclick={() => {
            activeMode = 'scan';
            moreToolsOpen = false;
            cliPreview = 'scanner scan veltrix nova --tlds com,ai,io,dev,co,app,xyz,sh --mutations';
          }}
          class="px-3 py-1.5 rounded-full text-xs font-display font-semibold transition-colors cursor-pointer shrink-0 {activeMode ===
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
            moreToolsOpen = false;
            cliPreview = `scanner inspect ${inquiryData?.domain || 'svelte.dev'}`;
          }}
          class="px-3 py-1.5 rounded-full text-xs font-display font-semibold transition-colors cursor-pointer shrink-0 {activeMode ===
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
            moreToolsOpen = false;
            cliPreview = `scanner crawl ${crawlReport?.host || 'svelte.dev/docs/kit'} --pages 20 --depth 2`;
          }}
          class="px-3 py-1.5 rounded-full text-xs font-display font-semibold transition-colors cursor-pointer shrink-0 {activeMode ===
          'crawl'
            ? 'bg-[#19231f] text-[#fffdf8]'
            : 'text-[#48534e] hover:text-[#19231f]'}"
        >
          03. Site Tree
        </button>

        <!-- Burger Menu Trigger for Remaining Tools (04–08) -->
        <div class="relative" data-burger-menu>
          <button
            type="button"
            aria-label="Open more studio tools burger menu"
            aria-expanded={moreToolsOpen}
            title={activeMoreTool ? `Active: ${activeMoreTool.code}. ${activeMoreTool.label} (Click for all tools)` : 'More Studio Tools (Traffic, Ports & TLS, Robots, Social Meta, Vault)'}
            onclick={() => (moreToolsOpen = !moreToolsOpen)}
            class="h-8 px-2.5 rounded-full text-xs font-display font-semibold transition-all cursor-pointer inline-flex items-center justify-center gap-1.5 shrink-0 border {activeMoreTool
              ? 'bg-[#19231f] text-[#dffc78] border-[#19231f]'
              : moreToolsOpen
                ? 'bg-[#dffc78] text-[#19231f] border-[#19231f]'
                : 'bg-[#fffdf8] text-[#19231f] border-[#19231f]/20 hover:border-[#19231f]'}"
          >
            <StudioIcon name={moreToolsOpen ? 'close' : 'burger'} size={15} />
            {#if activeMoreTool}
              <span class="text-[11px] font-bold">{activeMoreTool.code}</span>
              <span class="w-1.5 h-1.5 rounded-full bg-[#dffc78]"></span>
            {/if}
          </button>

          {#if moreToolsOpen}
            <div
              role="menu"
              class="absolute right-0 mt-2.5 w-80 sm:w-96 rounded-2xl border-2 border-[#19231f] bg-[#fffdf8] p-2.5 shadow-[0_18px_45px_rgba(25,35,31,0.22)] z-50"
            >
              <div class="flex items-center justify-between px-3 py-1.5 border-b border-[#19231f]/10">
                <span class="font-display text-[11px] font-bold uppercase tracking-wider text-[#19231f]/55">
                  Specialized Reconnaissance Tools
                </span>
                <button
                  type="button"
                  onclick={() => (moreToolsOpen = false)}
                  class="text-[11px] font-display font-semibold text-[#19231f]/60 hover:text-[#19231f]"
                >
                  Close
                </button>
              </div>

              <div class="mt-1 space-y-1">
                {#each moreToolsItems as tool}
                  <button
                    type="button"
                    role="menuitem"
                    onclick={() => {
                      activeMode = tool.mode;
                      moreToolsOpen = false;
                    }}
                    class="w-full flex items-start justify-between gap-3 rounded-xl p-2.5 text-left transition-colors cursor-pointer {activeMode ===
                    tool.mode
                      ? 'bg-[#19231f] text-[#fffdf8]'
                      : 'hover:bg-[#f4f1e9] text-[#19231f]'}"
                  >
                    <div class="min-w-0 flex-1">
                      <div class="flex items-center gap-2">
                        <span
                          class="font-mono text-[11px] font-bold {activeMode === tool.mode
                            ? 'text-[#dffc78]'
                            : 'text-[#5366e8]'}"
                        >
                          {tool.code}
                        </span>
                        <span class="font-display text-xs font-bold">{tool.label}</span>
                        {#if tool.mode === 'vault'}
                          <span class="rounded-full bg-[#dffc78] px-1.5 py-0.2 font-mono text-[10px] font-bold text-[#19231f]">
                            {savedDomains.length}
                          </span>
                        {/if}
                      </div>
                      <p
                        class="mt-0.5 text-[11px] leading-snug {activeMode === tool.mode
                          ? 'text-[#fffdf8]/75'
                          : 'text-[#48534e]'}"
                      >
                        {tool.desc}
                      </p>
                    </div>
                    <span
                      class="shrink-0 rounded-full border border-[#19231f]/20 px-2 py-0.5 font-display text-[10px] font-bold {activeMode ===
                      tool.mode
                        ? 'bg-[#dffc78] text-[#19231f]'
                        : 'bg-[#f4f1e9] text-[#19231f]'}"
                    >
                      {tool.badge}
                    </span>
                  </button>
                {/each}
              </div>
            </div>
          {/if}
        </div>
      </nav>

      <!-- Desktop Right Actions -->
      <div class="hidden xl:flex items-center gap-1.5 shrink-0">
        {#if anyJobRunning}
          <button
            type="button"
            onclick={() => handleCancelJob()}
            class="px-2.5 py-1.5 rounded-full text-[11px] font-display font-bold bg-[#ffc3a5] hover:bg-[#ffad85] text-[#19231f] border border-[#19231f] transition-colors cursor-pointer shrink-0"
            title="Cancel currently running background job"
          >
            ✕ Stop
          </button>
        {/if}

        <button
          type="button"
          onclick={() => handlePrepareReport(inquiryData?.domain || reconReport?.domain || 'cloudflare.com')}
          disabled={Boolean(preparingDomain)}
          class="px-3 py-1.5 rounded-full text-xs font-display font-bold bg-[#dffc78] hover:bg-[#d9d6fc] text-[#19231f] border border-[#19231f] transition-colors cursor-pointer inline-flex items-center gap-1.5 shrink-0 disabled:opacity-50"
          title="Prepare Full Visual PDF Dossier Report"
        >
          <StudioIcon name="sparkle" size={12} />
          <span>{preparingDomain ? `Building ${preparingDomain}…` : 'Prepare Report'}</span>
        </button>

        <button
          type="button"
          onclick={async () => {
            await refreshJobs();
            jobQueueOpen = true;
          }}
          class="px-3 py-1.5 rounded-full text-xs font-display font-bold bg-[#d9d6fc] hover:bg-[#c8c3fa] text-[#19231f] border border-[#19231f] transition-colors cursor-pointer inline-flex items-center gap-1.5 shrink-0"
          title="Open Enterprise Job & Worker Queue"
        >
          <StudioIcon name="bolt" size={12} />
          <span>Queue ({jobList.length})</span>
          {#if anyJobRunning}
            <span class="w-2 h-2 rounded-full bg-[#19231f] animate-ping"></span>
          {/if}
        </button>

        <button
          type="button"
          onclick={copyCliCommand}
          class="studio-btn-primary px-3.5 py-1.5 text-xs cursor-pointer inline-flex items-center gap-1.5 shrink-0"
          title="Copy active Go CLI command"
        >
          <span>{copiedCli ? '✓ Copied' : 'Copy CLI'}</span>
        </button>
      </div>
    </header>
  </div>

  <!-- EDITORIAL STUDIO HERO & LIVE CLI BENTO -->
  <div id="top" class="max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8 pt-8 sm:pt-10 pb-8 print:hidden">
    <div class="grid grid-cols-1 lg:grid-cols-12 gap-6 items-end">
      <!-- 7-Col Editorial Title & Context -->
      <div class="lg:col-span-7 space-y-3">
        <!-- Fixed Single-Line Status Pill: dot + text are ALWAYS on the same line -->
        <div class="inline-flex items-center gap-2.5 max-w-full px-3.5 py-1.5 rounded-full bg-[#fffdf8] border border-[#19231f]/20 text-xs font-mono">
          <span
            class="w-2.5 h-2.5 rounded-full shrink-0 {backendOnline
              ? 'bg-[#19231f]'
              : 'bg-[#5366e8]'}"
          ></span>
          <span class="truncate text-[#19231f]">
            {backendOnline
              ? 'Go Parallel Engine Online (:8080) · 24 TLDs · Tranco Traffic · Wildcard-Filtered Recon'
              : 'Standalone Studio · Run `./bin/scanner serve` for live sockets'}
          </span>
        </div>

        <h1 class="text-3xl sm:text-5xl lg:text-[3.35rem] font-display font-bold leading-[1.06] text-[#19231f]">
          Uncover <span class="font-serif-editorial font-normal underline decoration-[#dffc78] decoration-4 underline-offset-4">unclaimed</span> domains &amp; map the <span class="font-serif-editorial font-normal">living</span> web.
        </h1>

        <p class="text-sm sm:text-base text-[#48534e] max-w-2xl leading-relaxed">
          Editorial domain intelligence &amp; surface reconnaissance studio: 24-TLD dictionary search, authoritative RDAP &amp; Port-43 WHOIS verification, Tranco &amp; Cloudflare Radar traffic estimation, sub-route site cartography, and printable PDF domain dossiers.
        </p>
      </div>

      <!-- 5-Col Live Synchronized Go CLI Terminal Card -->
      <div class="lg:col-span-5 bento-card-ink p-5 space-y-3 shadow-lg">
        <div class="flex items-center justify-between text-xs">
          <div class="flex items-center gap-2">
            <span class="w-2.5 h-2.5 rounded-full bg-[#dffc78] shrink-0"></span>
            <span class="font-display font-bold uppercase tracking-wider text-[#fffdf8]/80">
              Synchronized Go CLI Invocation
            </span>
          </div>
          <button
            type="button"
            onclick={copyCliCommand}
            class="px-3 py-1 rounded-full text-[11px] font-mono bg-[#fffdf8]/15 hover:bg-[#dffc78] hover:text-[#19231f] text-[#fffdf8] transition-colors cursor-pointer shrink-0"
          >
            {copiedCli ? '✓ Copied' : 'Copy'}
          </button>
        </div>

        <div
          class="rounded-xl bg-[#0f1613] border border-[#fffdf8]/15 px-4 py-3 font-mono text-xs text-[#dffc78] overflow-hidden flex items-center justify-between gap-2"
          title={`./bin/${cliPreview}`}
        >
          <span class="truncate">
            <span class="text-[#fffdf8]/50 select-none">$ </span>./bin/{cliPreview}
          </span>
        </div>

        <div class="flex items-center justify-between text-[11px] text-[#fffdf8]/70 font-mono">
          <span>Reports Ready: {savedReports.length} Dossiers</span>
          <span>Vault: {savedDomains.length} saved</span>
        </div>
      </div>
    </div>
  </div>

  <!-- ACTIVE WORKBENCH BENTO (All 8 panels stay mounted in DOM so switching tabs is 0ms with zero flicker) -->
  <main class="flex-1 max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8 space-y-12 print:hidden">
    <div class={activeMode === 'scan' ? 'block' : 'hidden'}>
      <AvailabilityScanner
        report={scanReport}
        loading={scanLoading}
        onRunScan={handleRunScan}
        onCancelJob={() => handleCancelJob()}
        onInspectDomain={handleInspectDomain}
        onCheckHistory={handleCheckHistory}
        onRunRecon={handleRunRecon}
        onCrawlDomain={handleCrawlDomain}
        onCheckTraffic={handleRunTraffic}
        onPrepareReport={handlePrepareReport}
        onToggleSave={handleToggleSave}
        {savedDomainsSet}
      />
    </div>

    <div class={activeMode === 'inspect' ? 'block' : 'hidden'}>
      <DomainInspector
        inquiry={inquiryData}
        loading={inspectLoading}
        onInspect={handleInspectDomain}
        onCancelJob={() => handleCancelJob()}
        onCheckHistory={handleCheckHistory}
        onRunRecon={handleRunRecon}
        onCrawlDomain={handleCrawlDomain}
        onPrepareReport={handlePrepareReport}
        onToggleSave={handleToggleSave}
        {savedDomainsSet}
      />
    </div>

    <div class={activeMode === 'crawl' ? 'block' : 'hidden'}>
      <SiteCrawler
        report={crawlReport}
        loading={crawlLoading}
        onRunCrawl={handleRunCrawl}
        onCancelJob={() => handleCancelJob()}
        onInspectDomain={handleInspectDomain}
      />
    </div>

    <div class={activeMode === 'traffic' ? 'block' : 'hidden'}>
      <TrafficIntelStudio
        report={trafficReport}
        loading={trafficLoading}
        onRunTraffic={handleRunTraffic}
        onPrepareReport={handlePrepareReport}
      />
    </div>

    <div class={activeMode === 'recon' ? 'block' : 'hidden'}>
      <PortReconStudio
        report={reconReport}
        loading={reconLoading}
        onRunRecon={handleRunRecon}
        onRunFullSuite={handlePrepareReport}
        onCancelJob={() => handleCancelJob()}
        onCheckHistory={handleCheckHistory}
        onSaveDomain={handleToggleSave}
        {savedDomainsSet}
      />
    </div>

    <div class={activeMode === 'robots' ? 'block' : 'hidden'}>
      <RobotsSitemapStudio
        report={robotsReport}
        loading={robotsLoading}
        onRunCheck={handleRunRobotsCheck}
        onCancelJob={() => handleCancelJob()}
        onCrawlUrl={handleCrawlDomain}
        onInspectMeta={handleRunMetaCheck}
      />
    </div>

    <div class={activeMode === 'meta' ? 'block' : 'hidden'}>
      <SocialMetaStudio
        report={metaReport}
        loading={metaLoading}
        onRunMetaCheck={handleRunMetaCheck}
        onCancelJob={() => handleCancelJob()}
        onCrawlUrl={handleCrawlDomain}
      />
    </div>

    <div class={activeMode === 'vault' ? 'block' : 'hidden'}>
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
    </div>

    <!-- BOTTOM STUDIO REFERENCE BENTO -->
    <section aria-labelledby="cli-field-guide" class="pt-6 border-t border-[#19231f]/12 space-y-5">
      <div class="flex flex-wrap items-baseline justify-between gap-2">
        <h2 id="cli-field-guide" class="font-display font-bold text-xl text-[#19231f]">
          Terminal Field <span class="font-serif-editorial font-normal text-2xl">Guide</span>
        </h2>
        <span class="text-xs text-[#48534e]">Click any command pill to copy directly to your clipboard</span>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-12 gap-5">
        {#each terminalGuides as guide, idx}
          <div class="md:col-span-3 bento-card p-5 space-y-3 bg-[#fffdf8] flex flex-col justify-between">
            <div class="space-y-2">
              <div class="flex items-center justify-between gap-2">
                <span class="studio-label">{guide.label}</span>
                <span
                  class="px-2 py-0.5 rounded-full text-[11px] font-display font-bold {guide.badgeBg} border border-[#19231f] shrink-0"
                >
                  {guide.badge}
                </span>
              </div>
              <p class="text-xs text-[#48534e] leading-relaxed">
                {guide.desc}
              </p>
            </div>

            <div
              class="rounded-xl bg-[#f4f1e9] border border-[#19231f]/15 px-3 py-2.5 flex items-center justify-between gap-2 overflow-hidden"
              title={guide.cmd}
            >
              <code class="text-xs font-mono text-[#19231f] truncate select-all">
                {guide.cmd}
              </code>
              <button
                type="button"
                onclick={() => copyGuideCommand(guide.cmd, idx)}
                class="px-2.5 py-1 rounded-lg text-[11px] font-display font-bold bg-[#19231f] hover:bg-[#5366e8] text-[#dffc78] transition-colors cursor-pointer shrink-0"
              >
                {copiedGuideIndex === idx ? '✓ Copied' : 'Copy'}
              </button>
            </div>
          </div>
        {/each}
      </div>
    </section>
  </main>

  <!-- LINKEDIN-STYLE DOCKED DOMAIN REPORTS WINDOW -->
  <ReportDock
    reports={savedReports}
    activeReportId={activeExecutiveReport?.id || null}
    {preparingDomain}
    onSelectReport={(rep) => (activeExecutiveReport = rep)}
    onDeleteReport={handleDeleteReport}
  />

  <!-- FULL-SCREEN PRINTABLE PDF-STYLE VISUAL EXECUTIVE REPORT MODAL -->
  <ExecutiveReportModal
    report={activeExecutiveReport}
    onClose={() => (activeExecutiveReport = null)}
  />

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
    onDispatchParallelSuite={handlePrepareReport}
    onCancelJob={handleCancelJob}
  />
</div>
