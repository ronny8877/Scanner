<script lang="ts">
  import type { TrafficReport } from '../types';
  import { validateDomainOrUrl } from '../validation';
  import StudioIcon from './StudioIcon.svelte';
  import ValidationBanner from './ValidationBanner.svelte';

  let {
    report,
    loading,
    onRunTraffic,
    onPrepareReport,
  }: {
    report: TrafficReport;
    loading: boolean;
    onRunTraffic: (domain: string) => void;
    onPrepareReport?: (domain: string) => void;
  } = $props();

  let targetInput = $state('cloudflare.com');
  let validationError = $state<string | null>(null);
  let validationSuggestion = $state<string | null>(null);

  const presets = ['cloudflare.com', 'svelte.dev', 'vercel.com', 'stripe.com', 'linear.app', 'outbid.lol'];

  function handleSubmit(e: Event) {
    e.preventDefault();
    triggerCheck(targetInput);
  }

  function triggerCheck(raw: string) {
    const check = validateDomainOrUrl(raw);
    if (!check.valid) {
      validationError = check.error || 'Please enter a valid domain name.';
      validationSuggestion = check.suggestion || null;
      return;
    }
    validationError = null;
    validationSuggestion = null;
    targetInput = check.normalizedDomain;
    onRunTraffic(check.normalizedDomain);
  }

  function handlePrepare() {
    const check = validateDomainOrUrl(targetInput || report.domain);
    if (!check.valid) {
      validationError = check.error || 'Please enter a valid domain name.';
      validationSuggestion = check.suggestion || null;
      return;
    }
    validationError = null;
    validationSuggestion = null;
    if (onPrepareReport) {
      onPrepareReport(check.normalizedDomain);
    }
  }
</script>

<section aria-labelledby="traffic-heading" class="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
  <div class="bento-card p-6 sm:p-8">
    <div class="flex flex-col justify-between gap-4 lg:flex-row lg:items-end">
      <div>
        <span class="pill-tag bg-[#dffc78] text-[#19231f]">
          <span class="h-1.5 w-1.5 rounded-full bg-[#19231f]"></span>
          04 · Registered Site Traffic & Global Rank Intelligence
        </span>
        <h2 id="traffic-heading" class="mt-3 font-display text-2xl font-bold tracking-tight sm:text-3xl">
          Tranco Top-1M, <span class="font-serif-accent font-normal">Cloudflare Radar</span> & Traffic Range Estimator
        </h2>
        <p class="mt-1 max-w-2xl text-sm text-[#19231f]/70">
          Combines live Tranco 30-day research rankings, Cloudflare Radar DNS resolver buckets, XML Sitemap index inventory, and Certificate Transparency velocity into honest monthly visit ranges.
        </p>
      </div>

      <form onsubmit={handleSubmit} class="flex w-full max-w-2xl flex-wrap gap-2 sm:flex-nowrap">
        <label for="traffic-domain-input" class="sr-only">Domain for Traffic & Popularity Check</label>
        <input
          id="traffic-domain-input"
          type="text"
          bind:value={targetInput}
          oninput={() => {
            if (validationError) {
              validationError = null;
              validationSuggestion = null;
            }
          }}
          placeholder="cloudflare.com"
          class="w-full rounded-xl border border-[#19231f]/25 bg-[#f4f1e9]/70 px-4 py-2.5 font-display text-sm font-medium text-[#19231f] focus:border-[#5366e8] focus:bg-[#fffdf8] focus:outline-none"
        />
        <button
          type="submit"
          disabled={loading}
          class="inline-flex shrink-0 items-center gap-2 rounded-xl border border-[#19231f] bg-[#19231f] px-5 py-2.5 font-display text-sm font-semibold text-[#fffdf8] transition-all hover:bg-[#5366e8] disabled:opacity-50"
        >
          {#if loading}
            <span class="h-2 w-2 animate-ping rounded-full bg-[#dffc78]"></span>
            Checking Rank...
          {:else}
            Check Traffic & Rank
          {/if}
        </button>
        {#if onPrepareReport}
          <button
            type="button"
            onclick={handlePrepare}
            disabled={loading}
            class="inline-flex shrink-0 items-center gap-1.5 rounded-xl border border-[#19231f] bg-[#dffc78] px-4 py-2.5 font-display text-xs font-bold text-[#19231f] transition-transform hover:-translate-y-0.5 disabled:opacity-50"
          >
            <StudioIcon name="sparkle" class="h-3.5 w-3.5" />
            Prepare Report
          </button>
        {/if}
      </form>
    </div>

    <ValidationBanner
      error={validationError}
      suggestion={validationSuggestion}
      onApplySuggestion={(fixed) => {
        targetInput = fixed;
        triggerCheck(fixed);
      }}
      onDismiss={() => {
        validationError = null;
        validationSuggestion = null;
      }}
    />

    <div class="mt-4 flex flex-wrap items-center gap-2 border-t border-[#19231f]/10 pt-4">
      <span class="font-display text-xs font-semibold uppercase tracking-wider text-[#19231f]/45">Quick Presets:</span>
      {#each presets as item}
        <button
          type="button"
          onclick={() => {
            targetInput = item;
            triggerCheck(item);
          }}
          class="rounded-full border border-[#19231f]/15 bg-[#f4f1e9] px-3 py-1 font-display text-xs font-medium text-[#19231f] transition-colors hover:border-[#19231f] hover:bg-[#dffc78]"
        >
          {item}
        </button>
      {/each}
    </div>
  </div>

  <!-- Executive 5-Line Traffic Summary Card + KPI Bento -->
  <div class="mt-6 grid grid-cols-1 gap-6 lg:grid-cols-12">
    <!-- Left 5 cols: Editorial Specimen Traffic Card (High-Contrast Dark Ink on Surface) -->
    <div class="bento-card flex flex-col justify-between p-6 text-[#19231f] sm:p-7 lg:col-span-5">
      <div>
        <div class="flex items-center justify-between gap-2">
          <span class="rounded-full border border-[#19231f] bg-[#dffc78] px-3 py-1 font-display text-[11px] font-bold uppercase tracking-wider text-[#19231f]">
            {report.isRegistered ? 'Registered Site Telemetry' : 'Unregistered Domain'}
          </span>
          <span class="rounded-full border border-[#19231f]/15 bg-[#f4f1e9] px-2.5 py-0.5 font-mono text-xs font-semibold text-[#19231f]">
            {report.durationMs}ms probe
          </span>
        </div>

        <h3 class="mt-4 font-display text-2xl font-bold tracking-tight text-[#19231f] sm:text-3xl">
          {report.domain}
        </h3>
        <p class="mt-1 text-xs text-[#48534e]">
          Category / Edge: <span class="font-bold text-[#19231f]">{report.edgeNetwork}</span>
        </p>

        <div class="mt-5 space-y-3 rounded-2xl border-[1.5px] border-[#19231f]/20 bg-[#f4f1e9] p-5">
          <div class="flex items-center justify-between gap-3 border-b border-[#19231f]/12 pb-2.5">
            <span class="text-xs font-semibold text-[#48534e]">Domain Popularity</span>
            <span class="font-display text-sm font-bold text-[#19231f]">{report.popularityTier}</span>
          </div>

          <div class="flex items-center justify-between gap-3 border-b border-[#19231f]/12 pb-2.5">
            <span class="text-xs font-semibold text-[#48534e]">Cloudflare Rank</span>
            <span class="rounded-lg border border-[#19231f]/25 bg-[#d9d6fc] px-2.5 py-0.5 font-display text-xs font-bold text-[#19231f]">
              {report.cloudflareBucket}
            </span>
          </div>

          <div class="flex items-center justify-between gap-3 border-b border-[#19231f]/12 pb-2.5">
            <span class="text-xs font-semibold text-[#48534e]">Top Locations</span>
            <span class="font-mono text-xs font-bold text-[#19231f]">
              {report.topLocations.map((loc) => loc.split('·')[0].trim()).join(' · ')}
            </span>
          </div>

          <div class="flex items-center justify-between gap-3 border-b border-[#19231f]/12 pb-2.5">
            <span class="text-xs font-semibold text-[#48534e]">Popularity Trend</span>
            <span
              class="rounded-full border border-[#19231f] px-2.5 py-0.5 font-display text-xs font-bold {report.trendDirection === 'RISING'
                ? 'bg-[#dffc78] text-[#19231f]'
                : report.trendDirection === 'COOLING'
                  ? 'bg-[#ffc3a5] text-[#19231f]'
                  : 'bg-[#d9d6fc] text-[#19231f]'}"
            >
              {report.trendLabel}
            </span>
          </div>

          <div class="flex items-center justify-between gap-3 pt-1">
            <span class="text-xs font-semibold text-[#48534e]">Est. Traffic Range</span>
            <span class="rounded-lg border border-[#19231f] bg-[#dffc78] px-2.5 py-1 font-display text-sm font-bold text-[#19231f]">
              {report.estimatedMonthlyRange}
            </span>
          </div>
        </div>
      </div>

      <div class="mt-5 flex flex-wrap items-center justify-between gap-2 border-t border-[#19231f]/15 pt-4 text-xs text-[#48534e]">
        <span>Daily Band: <strong class="text-[#19231f]">{report.estimatedDailyRange}</strong></span>
        <span>Confidence: <strong class="text-[#19231f]">{report.confidenceLevel}</strong></span>
      </div>
    </div>

    <!-- Right 7 cols: Multi-Signal Telemetry Breakdown & 30-Day Tranco Trajectory -->
    <div class="flex flex-col gap-6 lg:col-span-7">
      <div class="bento-card p-6">
        <div class="flex flex-wrap items-center justify-between gap-2">
          <div>
            <p class="font-display text-xs font-bold uppercase tracking-wider text-[#19231f]/50">
              Multi-Signal Traffic Calibration
            </p>
            <h3 class="mt-0.5 font-display text-lg font-bold text-[#19231f]">
              Why We Report Honest Ranges Instead of Fake Exact Counts
            </h3>
          </div>
          <span class="rounded-full border border-[#19231f]/15 bg-[#f4f1e9] px-3 py-1 font-display text-xs font-bold">
            {report.signals.length} Verified Signals
          </span>
        </div>

        <p class="mt-2 text-xs leading-relaxed text-[#19231f]/70">
          {report.methodologyNote}
        </p>

        <div class="mt-4 grid grid-cols-1 gap-3 sm:grid-cols-2">
          {#each report.signals as sig}
            <div class="rounded-xl border border-[#19231f]/12 bg-[#f4f1e9]/65 p-3.5">
              <div class="flex items-center justify-between gap-2">
                <span class="font-display text-xs font-bold text-[#19231f]">{sig.source}</span>
                <span class="rounded bg-[#19231f] px-2 py-0.5 font-display text-[10px] font-bold text-[#dffc78]">
                  {sig.weight}
                </span>
              </div>
              <p class="mt-1.5 font-display text-base font-bold text-[#5366e8]">{sig.value}</p>
              <p class="mt-1 text-[11px] leading-relaxed text-[#19231f]/65">{sig.description}</p>
            </div>
          {/each}
        </div>
      </div>

      <!-- 30-Day Tranco Rank Trajectory & Regional Distribution -->
      <div class="grid grid-cols-1 gap-6 sm:grid-cols-2">
        <div class="bento-card p-5">
          <div class="flex items-center justify-between">
            <h4 class="font-display text-xs font-bold uppercase tracking-wider text-[#19231f]/55">
              30-Day Tranco Rank History
            </h4>
            {#if report.isRanked}
              <span class="rounded-full bg-[#dffc78] px-2.5 py-0.5 font-display text-[11px] font-bold text-[#19231f]">
                #{report.trancoRank.toLocaleString()}
              </span>
            {/if}
          </div>

          {#if report.rankHistory && report.rankHistory.length > 0}
            <div class="mt-4 space-y-2">
              {#each report.rankHistory.slice(-6) as pt}
                <div class="flex items-center justify-between rounded-lg border border-[#19231f]/10 bg-[#f4f1e9]/60 px-3 py-1.5 text-xs">
                  <span class="font-mono text-[11px] text-[#19231f]/65">{pt.date}</span>
                  <span class="font-display font-bold text-[#19231f]">Rank #{pt.rank.toLocaleString()}</span>
                </div>
              {/each}
            </div>
          {:else}
            <p class="mt-4 rounded-xl border border-[#19231f]/10 bg-[#f4f1e9]/50 p-3.5 text-xs text-[#19231f]/65">
              Domain is outside the Tranco Top 1M global list; traffic range is calibrated from Sitemap URLs ({report.sitemapPagesCount}), CT Subdomains ({report.subdomainCount}), and Wayback history ({report.waybackYearsCount} yrs).
            </p>
          {/if}
        </div>

        <div class="bento-card p-5">
          <h4 class="font-display text-xs font-bold uppercase tracking-wider text-[#19231f]/55">
            Primary Audience Geographies
          </h4>
          <div class="mt-4 space-y-2.5">
            {#each report.topLocations as loc, idx}
              <div class="flex items-center justify-between rounded-xl border border-[#19231f]/10 bg-[#f4f1e9]/60 px-3.5 py-2">
                <div class="flex items-center gap-2">
                  <span class="flex h-5 w-5 items-center justify-center rounded-full bg-[#19231f] font-display text-[10px] font-bold text-[#fffdf8]">
                    {idx + 1}
                  </span>
                  <span class="font-display text-xs font-bold text-[#19231f]">{loc}</span>
                </div>
                <span class="text-[11px] font-medium text-[#19231f]/55">
                  {idx === 0 ? 'Primary Share' : idx === 1 ? 'Secondary' : 'Active Region'}
                </span>
              </div>
            {/each}
          </div>
        </div>
      </div>
    </div>
  </div>
</section>
