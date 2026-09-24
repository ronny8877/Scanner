<script lang="ts">
  import type { CrawlReport, SiteNode, PageInfo } from '../types';
  import TreeItem from './TreeItem.svelte';
  import LoadingProgressBanner from './LoadingProgressBanner.svelte';
  import { motionCard } from '../motion';

  interface Props {
    report: CrawlReport | null;
    loading: boolean;
    onRunCrawl: (opts: { targetUrl: string; maxPages: number; maxDepth: number }) => void;
    onInspectDomain: (domain: string) => void;
  }

  let { report, loading, onRunCrawl, onInspectDomain }: Props = $props();

  let targetUrl = $state('svelte.dev');
  let maxPages = $state(20);
  let maxDepth = $state(2);
  let selectedPath = $state('/');

  const sampleSites = ['svelte.dev', 'golang.org', 'example.com'];

  $effect(() => {
    if (report?.host) {
      targetUrl = report.host;
      selectedPath = '/';
    }
  });

  const selectedPage = $derived<PageInfo | undefined>(
    report?.pages.find((p) => p.path === selectedPath) || report?.pages[0]
  );

  function submitCrawl(e: Event) {
    e.preventDefault();
    if (targetUrl.trim()) {
      onRunCrawl({ targetUrl: targetUrl.trim(), maxPages, maxDepth });
    }
  }

  function handleNodeSelect(node: SiteNode) {
    selectedPath = node.fullPath;
  }
</script>

<section aria-labelledby="site-cartography-heading" class="space-y-6">
  <!-- Top Crawler Command Card -->
  <form
    use:motionCard
    onsubmit={submitCrawl}
    class="bento-card p-6 sm:p-7 flex flex-col lg:flex-row lg:items-end justify-between gap-5"
  >
    <div class="flex-1 space-y-3">
      <div class="flex items-center justify-between">
        <h2 id="site-cartography-heading" class="studio-label">
          03 // Live Site Cartography & Structural Tree Mapper
        </h2>
        <span class="text-xs font-mono text-[#48534e]">Concurrent HTML Parser & Link Graph</span>
      </div>

      <div class="flex flex-col sm:flex-row gap-3">
        <input
          id="crawl-url"
          type="text"
          bind:value={targetUrl}
          placeholder="Enter live domain or URL (e.g. svelte.dev, golang.org)..."
          class="flex-1 rounded-2xl bg-[#f4f1e9] border-[1.5px] border-[#19231f]/20 focus:border-[#19231f] px-4 py-3.5 text-base font-mono text-[#19231f] placeholder-[#6d7873]"
        />

        <div class="flex items-center gap-2">
          <select
            aria-label="Maximum pages to crawl"
            bind:value={maxPages}
            class="rounded-2xl bg-[#f4f1e9] border-[1.5px] border-[#19231f]/20 px-3.5 py-3.5 text-xs font-mono font-semibold text-[#19231f]"
          >
            <option value={12}>12 pages</option>
            <option value={20}>20 pages</option>
            <option value={35}>35 pages</option>
          </select>

          <select
            aria-label="Crawl depth"
            bind:value={maxDepth}
            class="rounded-2xl bg-[#f4f1e9] border-[1.5px] border-[#19231f]/20 px-3.5 py-3.5 text-xs font-mono font-semibold text-[#19231f]"
          >
            <option value={1}>Depth 1</option>
            <option value={2}>Depth 2</option>
            <option value={3}>Depth 3</option>
          </select>

          <button
            type="submit"
            disabled={loading}
            class="studio-btn-primary px-6 py-3.5 text-sm cursor-pointer disabled:opacity-50 shrink-0"
          >
            {loading ? 'Mapping Tree…' : 'Crawl Structure →'}
          </button>
        </div>
      </div>

      <div class="flex flex-wrap items-center gap-2 pt-0.5">
        <span class="text-xs text-[#6d7873]">Map sample architecture:</span>
        {#each sampleSites as site}
          <button
            type="button"
            onclick={() => {
              targetUrl = site;
              onRunCrawl({ targetUrl: site, maxPages, maxDepth });
            }}
            class="px-3 py-1 rounded-full text-xs font-mono bg-[#f4f1e9] hover:bg-[#dffc78] text-[#19231f] border border-[#19231f]/15 transition-colors cursor-pointer"
          >
            {site}
          </button>
        {/each}
      </div>
    </div>
  </form>

  {#if loading}
    <LoadingProgressBanner
      title="Concurrent Site Structure Crawl"
      target={targetUrl}
      workers={8}
      steps={[
        `Fetching Root Document & Discovering Internal Anchor Graph`,
        `Crawling Up to ${maxPages} Pages Across Depth ${maxDepth}`,
        'Constructing Hierarchical URL Segment Tree',
      ]}
    />
  {:else if report}
    <!-- Asymmetric Bento: 7-Col Interactive Hierarchy Tree + 5-Col Route Specimen & Telemetry -->
    <div class="grid grid-cols-1 lg:grid-cols-12 gap-6 items-start">
      <!-- 7-Col Interactive Site Tree -->
      <div class="lg:col-span-7 bento-card p-6 sm:p-7 space-y-5">
        <div class="flex flex-wrap items-center justify-between gap-3 border-b border-[#19231f]/10 pb-4">
          <div>
            <span class="studio-label">Hierarchical Route Map</span>
            <h3 class="text-2xl font-display font-bold text-[#19231f] mt-0.5">
              {report.host}
              <span class="font-serif-editorial font-normal text-xl text-[#48534e]">
                ({report.pagesCrawled} crawled nodes · {report.totalLinks} internal links)
              </span>
            </h3>
          </div>

          <button
            type="button"
            onclick={() => onInspectDomain(report.host)}
            class="px-3.5 py-1.5 rounded-full text-xs font-display font-semibold bg-[#f4f1e9] hover:bg-[#d9d6fc] text-[#19231f] border border-[#19231f]/25 transition-colors cursor-pointer"
          >
            Inspect RDAP Dossier →
          </button>
        </div>

        <div class="scroll-panel max-h-[520px] pr-2">
          {#if report.tree}
            <TreeItem node={report.tree} {selectedPath} onSelect={handleNodeSelect} />
          {/if}
        </div>
      </div>

      <!-- 5-Col Selected Route Specimen + Tech Fingerprint -->
      <aside class="lg:col-span-5 space-y-6 lg:sticky lg:top-24">
        <div class="bento-card overflow-hidden">
          <div class="bg-[#19231f] text-[#fffdf8] p-6">
            <div class="flex items-center justify-between">
              <span class="text-[11px] font-display uppercase tracking-widest text-[#dffc78]">
                Selected Route Specimen
              </span>
              {#if selectedPage}
                <span class="px-2.5 py-0.5 rounded-full text-xs font-mono bg-[#dffc78] text-[#19231f] font-bold">
                  HTTP {selectedPage.statusCode || 200}
                </span>
              {/if}
            </div>
            <h4 class="mt-2 text-2xl font-mono font-bold text-[#fffdf8] break-all">
              {selectedPath}
            </h4>
          </div>

          <div class="p-6 space-y-4 bg-[#fffdf8]">
            {#if selectedPage}
              <div class="rounded-2xl bg-[#f4f1e9] border border-[#19231f]/12 p-4">
                <div class="studio-label">Document Title</div>
                <div class="mt-1 font-display font-semibold text-sm text-[#19231f]">
                  {selectedPage.title || '(Untitled route)'}
                </div>
              </div>

              {#if selectedPage.h1}
                <div class="rounded-2xl bg-[#f4f1e9] border border-[#19231f]/12 p-4">
                  <div class="studio-label">Primary Heading (H1)</div>
                  <div class="mt-1 font-serif-editorial text-xl text-[#19231f]">
                    “{selectedPage.h1}”
                  </div>
                </div>
              {/if}

              {#if selectedPage.description}
                <div class="rounded-2xl bg-[#f4f1e9] border border-[#19231f]/12 p-4">
                  <div class="studio-label">Meta Description</div>
                  <p class="mt-1 text-xs text-[#48534e] leading-relaxed">
                    {selectedPage.description}
                  </p>
                </div>
              {/if}

              <div class="grid grid-cols-3 gap-3 font-mono text-center">
                <div class="rounded-2xl bg-[#f4f1e9] border border-[#19231f]/12 p-3">
                  <div class="text-[10px] text-[#48534e] uppercase">Latency</div>
                  <div class="text-sm font-bold text-[#19231f] mt-0.5">{selectedPage.latencyMs}ms</div>
                </div>
                <div class="rounded-2xl bg-[#f4f1e9] border border-[#19231f]/12 p-3">
                  <div class="text-[10px] text-[#48534e] uppercase">Internal</div>
                  <div class="text-sm font-bold text-[#5366e8] mt-0.5">{selectedPage.internalLinks}</div>
                </div>
                <div class="rounded-2xl bg-[#f4f1e9] border border-[#19231f]/12 p-3">
                  <div class="text-[10px] text-[#48534e] uppercase">External</div>
                  <div class="text-sm font-bold text-[#19231f] mt-0.5">{selectedPage.externalLinks}</div>
                </div>
              </div>
            {:else}
              <p class="text-xs text-[#48534e]">
                Discovered route branch in site link topology. Click any crawled node to inspect its document headers.
              </p>
            {/if}

            {#if report.techHeaders && report.techHeaders.length > 0}
              <div class="pt-3 border-t border-[#19231f]/10">
                <div class="studio-label mb-2">Detected Edge & Server Headers</div>
                <div class="flex flex-wrap gap-1.5">
                  {#each report.techHeaders as th}
                    <span class="px-2.5 py-1 rounded-full bg-[#d9d6fc] border border-[#19231f]/20 text-xs font-mono text-[#19231f]">
                      {th}
                    </span>
                  {/each}
                </div>
              </div>
            {/if}
          </div>
        </div>
      </aside>
    </div>

    <!-- Crawled Page Ledger Table -->
    <div class="bento-card overflow-hidden">
      <div class="px-6 py-4 border-b border-[#19231f]/10 flex items-center justify-between">
        <h4 class="font-display font-bold text-base text-[#19231f]">
          Crawled Route Inventory ({report.pages.length} documents)
        </h4>
        <span class="text-xs font-mono text-[#48534e]">Total crawl duration: {report.durationMs}ms</span>
      </div>

      <div class="overflow-x-auto">
        <table class="w-full text-left text-xs">
          <thead class="bg-[#f4f1e9] border-b border-[#19231f]/10 font-mono uppercase text-[#48534e]">
            <tr>
              <th class="px-6 py-3">Route Path</th>
              <th class="px-4 py-3">HTTP</th>
              <th class="px-4 py-3">Response</th>
              <th class="px-4 py-3">Link Graph</th>
              <th class="px-6 py-3">Document Title</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-[#19231f]/10">
            {#each report.pages as pg (pg.path)}
              <tr
                class="cursor-pointer transition-colors {selectedPath === pg.path
                  ? 'bg-[#dffc78]/45'
                  : 'hover:bg-[#f4f1e9]'}"
                onclick={() => (selectedPath = pg.path)}
              >
                <td class="px-6 py-3.5 font-mono font-semibold text-[#5366e8]">{pg.path}</td>
                <td class="px-4 py-3.5 font-mono">
                  <span class="px-2.5 py-0.5 rounded-full bg-[#dffc78] border border-[#19231f] text-[#19231f] font-bold">
                    {pg.statusCode || 200}
                  </span>
                </td>
                <td class="px-4 py-3.5 font-mono text-[#48534e]">{pg.latencyMs}ms</td>
                <td class="px-4 py-3.5 font-mono text-[#19231f]">
                  {pg.internalLinks} int · {pg.externalLinks} ext
                </td>
                <td class="px-6 py-3.5 font-medium text-[#19231f] truncate max-w-md">{pg.title || '—'}</td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    </div>
  {/if}
</section>
