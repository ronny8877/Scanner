<script lang="ts">
  import type { RobotsSitemapReport } from '../types';
  import LoadingProgressBanner from './LoadingProgressBanner.svelte';
  import { motionCard } from '../motion';

  interface Props {
    report: RobotsSitemapReport | null;
    loading: boolean;
    onRunCheck: (target: string) => void;
    onCancelJob?: () => void;
    onCrawlUrl?: (url: string) => void;
    onInspectMeta?: (url: string) => void;
  }

  let { report, loading, onRunCheck, onCancelJob, onCrawlUrl, onInspectMeta }: Props = $props();

  let targetInput = $state('supercoloring.com');
  let sitemapSearch = $state('');
  let showRawRobots = $state(false);

  const sampleTargets = ['supercoloring.com', 'bemee.in', 'svelte.dev', 'github.com', 'cloudflare.com'];

  $effect(() => {
    if (report?.host) {
      targetInput = report.host;
    }
  });

  const filteredEntries = $derived(
    (report?.entries ?? []).filter((e) => {
      if (!sitemapSearch.trim()) return true;
      const q = sitemapSearch.toLowerCase();
      return e.loc.toLowerCase().includes(q) || (e.lastMod || '').toLowerCase().includes(q);
    })
  );

  const blockedBotsCount = $derived(
    (report?.botMatrix ?? []).filter((b) => b.status === 'BLOCKED').length
  );
  const partialBotsCount = $derived(
    (report?.botMatrix ?? []).filter((b) => b.status === 'PARTIAL').length
  );

  function handleSubmit(e: Event) {
    e.preventDefault();
    if (targetInput.trim()) {
      onRunCheck(targetInput.trim());
    }
  }
</script>

<section aria-labelledby="robots-sitemap-heading" class="space-y-6">
  <!-- Command Header Card -->
  <form use:motionCard onsubmit={handleSubmit} class="bento-card p-6 sm:p-7 space-y-4">
    <div class="flex flex-wrap items-center justify-between gap-2">
      <div>
        <h2 id="robots-sitemap-heading" class="studio-label">
          05 // Robots.txt Crawler Governance & Sitemap.xml Update Inspector
        </h2>
        <p class="text-sm text-[#48534e] mt-1">
          Inspect which AI/LLM &amp; Search crawlers are blocked, audit disallowed route rules, and verify XML Sitemap update timestamps (`&lt;lastmod&gt;`).
        </p>
      </div>
      <span class="px-3 py-1 rounded-full text-xs font-mono bg-[#dffc78] border border-[#19231f] font-bold">
        12-Bot Matrix + XML Parser
      </span>
    </div>

    <div class="flex flex-col sm:flex-row gap-3">
      <input
        type="text"
        bind:value={targetInput}
        placeholder="Enter domain or URL (e.g. supercoloring.com, bemee.in, svelte.dev)..."
        class="flex-1 rounded-2xl bg-[#f4f1e9] border-[1.5px] border-[#19231f]/20 focus:border-[#19231f] px-4 py-3.5 text-base font-mono text-[#19231f]"
      />
      <button
        type="submit"
        disabled={loading}
        class="studio-btn-primary px-6 py-3.5 text-sm cursor-pointer disabled:opacity-50 shrink-0"
      >
        {loading ? 'Auditing Robots & Sitemaps…' : 'Inspect Robots & Sitemap →'}
      </button>
    </div>

    <div class="flex flex-wrap items-center gap-2">
      <span class="text-xs text-[#6d7873]">Quick test domain:</span>
      {#each sampleTargets as st}
        <button
          type="button"
          onclick={() => {
            targetInput = st;
            onRunCheck(st);
          }}
          class="px-3 py-1 rounded-full text-xs font-mono bg-[#f4f1e9] hover:bg-[#dffc78] text-[#19231f] border border-[#19231f]/15 transition-colors cursor-pointer"
        >
          {st}
        </button>
      {/each}
    </div>
  </form>

  {#if loading}
    <LoadingProgressBanner
      title="Robots.txt & Sitemap.xml Governance Audit"
      target={targetInput}
      workers={6}
      steps={[
        'Fetching /robots.txt & Parsing User-agent Disallow/Allow Directives',
        'Evaluating 12 Search, SEO & AI Crawler Permissions (GPTBot, ClaudeBot, Googlebot)',
        'Parsing XML Sitemap Index & Extracting <lastmod> Update Timestamps',
      ]}
      onCancel={onCancelJob}
    />
  {:else if report}
    <!-- 4-Card Telemetry Summary Row -->
    <div class="grid grid-cols-1 md:grid-cols-12 gap-5">
      <!-- Robots.txt Status -->
      <div class="md:col-span-3 bento-card p-5 bg-[#fffdf8] space-y-2">
        <div class="flex items-center justify-between">
          <span class="studio-label">robots.txt Policy</span>
          <span
            class="px-2.5 py-0.5 rounded-full text-xs font-mono font-bold border {report.robotsFound
              ? 'bg-[#dffc78] text-[#19231f] border-[#19231f]'
              : 'bg-[#ffc3a5] text-[#19231f] border-[#19231f]'}"
          >
            {report.robotsFound ? `HTTP ${report.robotsStatus}` : 'MISSING'}
          </span>
        </div>
        <div class="text-2xl font-display font-bold text-[#19231f]">
          {report.totalDisallowCount} Blocked · {report.totalAllowCount} Allowed
        </div>
        <div class="flex items-center justify-between text-xs font-mono text-[#48534e] pt-1">
          <span>{(report.robotsSizeBytes / 1024).toFixed(1)} KB file</span>
          <a
            href={report.robotsUrl}
            target="_blank"
            rel="noopener noreferrer"
            class="text-[#5366e8] font-bold hover:underline"
          >
            Open robots.txt ↗
          </a>
        </div>
      </div>

      <!-- Crawler Matrix Summary -->
      <div class="md:col-span-3 bento-card p-5 bg-[#fffdf8] space-y-2">
        <div class="flex items-center justify-between">
          <span class="studio-label">Bot Access Matrix</span>
          <span class="px-2.5 py-0.5 rounded-full text-xs font-mono bg-[#d9d6fc] border border-[#19231f]/30 text-[#19231f] font-bold">
            12 Bots Tested
          </span>
        </div>
        <div class="text-2xl font-display font-bold text-[#19231f]">
          {blockedBotsCount} Blocked · {partialBotsCount} Restricted
        </div>
        <p class="text-xs text-[#48534e]">
          {12 - blockedBotsCount - partialBotsCount} crawlers have unrestricted full-site access.
        </p>
      </div>

      <!-- Sitemap Discovery -->
      <div class="md:col-span-3 bento-card p-5 bg-[#fffdf8] space-y-2">
        <div class="flex items-center justify-between">
          <span class="studio-label">Sitemap.xml Index</span>
          <span
            class="px-2.5 py-0.5 rounded-full text-xs font-mono font-bold border {report.sitemapFound
              ? 'bg-[#dffc78] text-[#19231f] border-[#19231f]'
              : 'bg-[#ffc3a5] text-[#19231f] border-[#19231f]'}"
          >
            {report.sitemapFound ? (report.isSitemapIndex ? 'INDEX MAP' : 'URLSET') : 'NONE'}
          </span>
        </div>
        <div class="text-2xl font-display font-bold text-[#19231f]">
          {report.totalUrlsCount.toLocaleString()} Indexed URLs
        </div>
        <div class="flex items-center justify-between text-xs font-mono text-[#48534e] pt-1">
          <span>{report.childSitemaps?.length || 0} child sitemaps</span>
          {#if report.sitemapUrl}
            <a
              href={report.sitemapUrl}
              target="_blank"
              rel="noopener noreferrer"
              class="text-[#5366e8] font-bold hover:underline"
            >
              Open XML ↗
            </a>
          {/if}
        </div>
      </div>

      <!-- Sitemap Freshness / LastMod Dates -->
      <div class="md:col-span-3 bento-card-ink p-5 space-y-2">
        <div class="flex items-center justify-between">
          <span class="text-[11px] font-display uppercase tracking-wider text-[#dffc78]">
            Sitemap Update Cadence
          </span>
          <span class="text-xs font-mono text-[#fffdf8]/70">{report.durationMs}ms</span>
        </div>
        <div class="text-xl font-mono font-bold text-[#fffdf8]">
          Latest: {report.newestLastMod || 'No <lastmod>'}
        </div>
        <div class="text-xs font-mono text-[#fffdf8]/80 flex items-center justify-between pt-1">
          <span>7d: <strong class="text-[#dffc78]">{report.updatedLast7Days}</strong></span>
          <span>30d: <strong class="text-[#dffc78]">{report.updatedLast30Days}</strong></span>
          <span>1yr: <strong class="text-[#dffc78]">{report.updatedLastYear}</strong></span>
        </div>
      </div>
    </div>

    <!-- Main Asymmetric Grid: 6-Col Bot Permission Matrix & Disallow Rules + 6-Col Sitemap Update Ledger -->
    <div class="grid grid-cols-1 lg:grid-cols-12 gap-6 items-start">
      <!-- Left 6-Col: Crawler Permission Matrix & Blocked Routes -->
      <div class="lg:col-span-6 space-y-6">
        <!-- 12-Bot Permission Matrix -->
        <div class="bento-card p-6 space-y-4 bg-[#fffdf8]">
          <div class="flex flex-wrap items-center justify-between gap-2 border-b border-[#19231f]/10 pb-3">
            <div>
              <h3 class="font-display font-bold text-lg text-[#19231f]">
                Crawler & AI Bot Permission Matrix
              </h3>
              <p class="text-xs text-[#48534e]">
                Evaluated against explicit <code>User-agent</code> blocks and wildcard <code>*</code> fallback rules
              </p>
            </div>
          </div>

          <div class="grid grid-cols-1 sm:grid-cols-2 gap-2.5">
            {#each report.botMatrix as bot}
              <div class="rounded-2xl bg-[#f4f1e9] border border-[#19231f]/15 p-3.5 space-y-1.5">
                <div class="flex items-center justify-between gap-2">
                  <span class="font-display font-bold text-xs text-[#19231f]">{bot.botName}</span>
                  {#if bot.status === 'BLOCKED'}
                    <span class="px-2 py-0.5 rounded-full text-[10px] font-mono font-bold bg-[#ffc3a5] border border-[#19231f] text-[#19231f]">
                      ✕ BLOCKED
                    </span>
                  {:else if bot.status === 'PARTIAL'}
                    <span class="px-2 py-0.5 rounded-full text-[10px] font-mono font-bold bg-[#d9d6fc] border border-[#19231f]/40 text-[#19231f]">
                      ◐ RESTRICTED
                    </span>
                  {:else}
                    <span class="px-2 py-0.5 rounded-full text-[10px] font-mono font-bold bg-[#dffc78] border border-[#19231f] text-[#19231f]">
                      ✓ ALLOWED
                    </span>
                  {/if}
                </div>
                <div class="text-[11px] font-mono text-[#48534e] truncate" title={bot.matchedRule}>
                  {bot.matchedRule}
                </div>
              </div>
            {/each}
          </div>
        </div>

        <!-- Blocked & Allowed Routes per User-Agent Group -->
        <div class="bento-card p-6 space-y-4 bg-[#fffdf8]">
          <div class="flex items-center justify-between border-b border-[#19231f]/10 pb-3">
            <div>
              <h3 class="font-display font-bold text-lg text-[#19231f]">
                Blocked & Allowed Routes by User-Agent ({report.agentGroups?.length || 0} groups)
              </h3>
              <p class="text-xs text-[#48534e]">All route patterns declared in <code>robots.txt</code></p>
            </div>
            {#if report.rawRobotsPreview}
              <button
                type="button"
                onclick={() => (showRawRobots = !showRawRobots)}
                class="px-3 py-1 rounded-full text-xs font-mono bg-[#f4f1e9] hover:bg-[#dffc78] text-[#19231f] border border-[#19231f]/20 cursor-pointer"
              >
                {showRawRobots ? 'Hide Raw robots.txt' : 'View Raw File'}
              </button>
            {/if}
          </div>

          {#if showRawRobots && report.rawRobotsPreview}
            <pre class="rounded-2xl bg-[#19231f] text-[#dffc78] p-4 text-xs font-mono max-h-64 scroll-panel overflow-x-auto">{report.rawRobotsPreview}</pre>
          {/if}

          {#if report.agentGroups && report.agentGroups.length > 0}
            <div class="space-y-3 max-h-[460px] scroll-panel pr-1">
              {#each report.agentGroups as group}
                <div class="rounded-2xl bg-[#f4f1e9] border border-[#19231f]/15 p-4 space-y-2.5">
                  <div class="flex items-center justify-between">
                    <span class="font-mono text-xs font-bold px-2.5 py-1 rounded-lg bg-[#19231f] text-[#dffc78]">
                      User-agent: {group.userAgent}
                    </span>
                    {#if group.crawlDelay}
                      <span class="text-xs font-mono text-[#5366e8] font-bold">
                        Crawl-delay: {group.crawlDelay}s
                      </span>
                    {/if}
                  </div>

                  {#if group.disallow && group.disallow.length > 0}
                    <div class="space-y-1">
                      <div class="text-[10px] font-mono uppercase text-[#48534e]">
                        Blocked Routes (Disallow: {group.disallow.length})
                      </div>
                      <div class="flex flex-wrap gap-1.5">
                        {#each group.disallow as dPath}
                          <span class="px-2.5 py-0.5 rounded-lg bg-[#ffc3a5]/70 border border-[#19231f]/25 text-xs font-mono text-[#19231f]">
                            ✕ {dPath}
                          </span>
                        {/each}
                      </div>
                    </div>
                  {/if}

                  {#if group.allow && group.allow.length > 0}
                    <div class="space-y-1 pt-1">
                      <div class="text-[10px] font-mono uppercase text-[#48534e]">
                        Allowed Exceptions (Allow: {group.allow.length})
                      </div>
                      <div class="flex flex-wrap gap-1.5">
                        {#each group.allow as aPath}
                          <span class="px-2.5 py-0.5 rounded-lg bg-[#dffc78] border border-[#19231f]/25 text-xs font-mono text-[#19231f]">
                            ✓ {aPath}
                          </span>
                        {/each}
                      </div>
                    </div>
                  {/if}
                </div>
              {/each}
            </div>
          {:else}
            <div class="rounded-2xl bg-[#f4f1e9] p-4 text-xs text-[#48534e]">
              No explicit <code>Disallow</code> or <code>Allow</code> blocks found in <code>robots.txt</code>.
            </div>
          {/if}
        </div>
      </div>

      <!-- Right 6-Col: Sitemap.xml Child Maps & Indexed URLs with Update Dates -->
      <div class="lg:col-span-6 space-y-6">
        {#if report.childSitemaps && report.childSitemaps.length > 0}
          <div class="bento-card p-6 space-y-3 bg-[#fffdf8]">
            <div class="flex items-center justify-between border-b border-[#19231f]/10 pb-3">
              <h3 class="font-display font-bold text-base text-[#19231f]">
                Sitemap Index Partition Maps ({report.childSitemaps.length})
              </h3>
              <span class="studio-label">&lt;sitemapindex&gt;</span>
            </div>
            <div class="space-y-2 max-h-52 scroll-panel pr-1">
              {#each report.childSitemaps as sm}
                <div class="rounded-xl bg-[#f4f1e9] border border-[#19231f]/12 p-3 flex items-center justify-between gap-2 text-xs font-mono">
                  <a
                    href={sm.loc}
                    target="_blank"
                    rel="noopener noreferrer"
                    class="text-[#5366e8] font-semibold truncate hover:underline"
                  >
                    {sm.path} ↗
                  </a>
                  <div class="flex items-center gap-2 shrink-0">
                    {#if sm.lastMod}
                      <span class="px-2 py-0.5 rounded-full bg-[#dffc78] border border-[#19231f] text-[#19231f] font-bold">
                        {sm.lastMod}
                      </span>
                    {/if}
                    {#if sm.ageLabel}
                      <span class="text-[#48534e]">{sm.ageLabel}</span>
                    {/if}
                  </div>
                </div>
              {/each}
            </div>
          </div>
        {/if}

        <!-- Sitemap URLs & Update Timestamps Table -->
        <div class="bento-card overflow-hidden bg-[#fffdf8]">
          <div class="p-6 border-b border-[#19231f]/10 space-y-3">
            <div class="flex flex-wrap items-center justify-between gap-2">
              <div>
                <h3 class="font-display font-bold text-lg text-[#19231f]">
                  Sitemap URLs & Last Modified Dates ({filteredEntries.length})
                </h3>
                <p class="text-xs text-[#48534e]">
                  Click any route to inspect its Meta tags or start a Site Tree crawl from that path
                </p>
              </div>
              {#if report.oldestLastMod && report.newestLastMod}
                <span class="px-3 py-1 rounded-full text-xs font-mono bg-[#f4f1e9] border border-[#19231f]/20 text-[#19231f]">
                  Span: {report.oldestLastMod} → {report.newestLastMod}
                </span>
              {/if}
            </div>

            <input
              type="text"
              bind:value={sitemapSearch}
              placeholder="Filter sitemap routes or dates (e.g. 2026, /blog, /docs)..."
              class="w-full rounded-xl bg-[#f4f1e9] border border-[#19231f]/20 px-3.5 py-2 text-xs font-mono text-[#19231f]"
            />
          </div>

          {#if filteredEntries.length > 0}
            <div class="max-h-[520px] scroll-panel divide-y divide-[#19231f]/10">
              {#each filteredEntries as entry}
                <div class="px-5 py-3.5 hover:bg-[#f4f1e9]/70 transition-colors flex flex-col sm:flex-row sm:items-center justify-between gap-2">
                  <div class="min-w-0 space-y-0.5">
                    <a
                      href={entry.loc}
                      target="_blank"
                      rel="noopener noreferrer"
                      class="font-mono text-xs font-bold text-[#5366e8] hover:underline block truncate"
                      title={entry.loc}
                    >
                      {entry.path || entry.loc} ↗
                    </a>
                    <div class="flex flex-wrap items-center gap-2 text-[11px] font-mono text-[#48534e]">
                      {#if entry.changeFreq}
                        <span>freq: {entry.changeFreq}</span>
                      {/if}
                      {#if entry.priority}
                        <span>priority: {entry.priority}</span>
                      {/if}
                    </div>
                  </div>

                  <div class="flex items-center gap-2 shrink-0">
                    {#if entry.lastMod}
                      <span class="px-2.5 py-0.5 rounded-full text-xs font-mono font-bold bg-[#dffc78] border border-[#19231f] text-[#19231f]">
                        {entry.lastMod}
                      </span>
                      {#if entry.ageLabel}
                        <span class="text-[11px] font-mono text-[#48534e]">{entry.ageLabel}</span>
                      {/if}
                    {:else}
                      <span class="text-[11px] font-mono text-[#6d7873]">No &lt;lastmod&gt;</span>
                    {/if}

                    {#if onCrawlUrl}
                      <button
                        type="button"
                        onclick={() => onCrawlUrl?.(entry.loc)}
                        class="px-2.5 py-1 rounded-full text-[11px] font-display font-semibold bg-[#f4f1e9] hover:bg-[#dffc78] text-[#19231f] border border-[#19231f]/20 cursor-pointer"
                        title="Crawl site tree starting from this URL"
                      >
                        Crawl Here
                      </button>
                    {/if}

                    {#if onInspectMeta}
                      <button
                        type="button"
                        onclick={() => onInspectMeta?.(entry.loc)}
                        class="px-2.5 py-1 rounded-full text-[11px] font-display font-semibold bg-[#d9d6fc] hover:bg-[#c6c1fa] text-[#19231f] border border-[#19231f]/25 cursor-pointer"
                        title="Preview Social Meta cards for this URL"
                      >
                        Meta Card
                      </button>
                    {/if}
                  </div>
                </div>
              {/each}
            </div>
          {:else}
            <div class="p-8 text-center text-xs text-[#48534e]">
              No XML sitemap entries found for this filter.
            </div>
          {/if}
        </div>
      </div>
    </div>
  {/if}
</section>
