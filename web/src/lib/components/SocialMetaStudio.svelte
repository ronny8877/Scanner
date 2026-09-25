<script lang="ts">
  import type { MetaSocialReport } from '../types';
  import LoadingProgressBanner from './LoadingProgressBanner.svelte';
  import TrackerPostureCard from './TrackerPostureCard.svelte';
  import { motionCard } from '../motion';

  interface Props {
    report: MetaSocialReport | null;
    loading: boolean;
    onRunMetaCheck: (targetUrl: string) => void;
    onCancelJob?: () => void;
    onCrawlUrl?: (url: string) => void;
  }

  let { report, loading, onRunMetaCheck, onCancelJob, onCrawlUrl }: Props = $props();

  let targetInput = $state('svelte.dev');
  let activePlatform = $state<'all' | 'discord' | 'telegram' | 'whatsapp' | 'facebook'>('all');

  const sampleUrls = ['svelte.dev', 'github.com/sveltejs/svelte', 'golang.org', 'cloudflare.com'];

  $effect(() => {
    if (report?.targetUrl) {
      targetInput = report.targetUrl.replace(/^https?:\/\//, '');
    }
  });

  function handleSubmit(e: Event) {
    e.preventDefault();
    if (targetInput.trim()) {
      onRunMetaCheck(targetInput.trim());
    }
  }
</script>

<section aria-labelledby="social-meta-heading" class="space-y-6">
  <!-- Command Bar -->
  <form use:motionCard onsubmit={handleSubmit} class="bento-card p-6 sm:p-7 space-y-4">
    <div class="flex flex-wrap items-center justify-between gap-2">
      <div>
        <h2 id="social-meta-heading" class="studio-label">
          06 // Social Meta Tag Inspector, Platform Card Previews & Tracker Telemetry
        </h2>
        <p class="text-sm text-[#48534e] mt-1">
          Preview how any domain or sub-route renders on <strong>Discord, Telegram, WhatsApp &amp; Facebook</strong>, audit OpenGraph tags, and inspect Ad Networks &amp; Analytics trackers.
        </p>
      </div>
      <span class="px-3 py-1 rounded-full text-xs font-mono bg-[#d9d6fc] border border-[#19231f] font-bold">
        Discord · Telegram · WhatsApp · FB + Ad/Tracker Radar
      </span>
    </div>

    <div class="flex flex-col sm:flex-row gap-3">
      <input
        type="text"
        bind:value={targetInput}
        placeholder="Enter any URL or sub-path (e.g. svelte.dev, github.com/sveltejs/svelte, golang.org)..."
        class="flex-1 rounded-2xl bg-[#f4f1e9] border-[1.5px] border-[#19231f]/20 focus:border-[#19231f] px-4 py-3.5 text-base font-mono text-[#19231f]"
      />
      <button
        type="submit"
        disabled={loading}
        class="studio-btn-primary px-6 py-3.5 text-sm cursor-pointer disabled:opacity-50 shrink-0"
      >
        {loading ? 'Fetching Meta & Trackers…' : 'Preview Social Cards & Trackers →'}
      </button>
    </div>

    <div class="flex flex-wrap items-center gap-2">
      <span class="text-xs text-[#6d7873]">Sample URLs:</span>
      {#each sampleUrls as u}
        <button
          type="button"
          onclick={() => {
            targetInput = u;
            onRunMetaCheck(u);
          }}
          class="px-3 py-1 rounded-full text-xs font-mono bg-[#f4f1e9] hover:bg-[#dffc78] text-[#19231f] border border-[#19231f]/15 transition-colors cursor-pointer"
        >
          {u}
        </button>
      {/each}
    </div>
  </form>

  {#if loading}
    <LoadingProgressBanner
      title="Social OpenGraph & Ad/Tracker Inspection"
      target={targetInput}
      workers={4}
      steps={[
        'Fetching HTML Document Head & Resolving Canonical / OpenGraph Assets',
        'Synthesizing Discord, Telegram, WhatsApp & Facebook Share Previews',
        'Scanning Page Scripts for Ad Networks, Analytics Engines & Retargeting Pixels',
      ]}
      onCancel={onCancelJob}
    />
  {:else if report}
    <!-- Platform Filter Switcher + Grade Summary -->
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div class="flex flex-wrap items-center gap-1.5 bg-[#fffdf8] p-1.5 rounded-full border border-[#19231f]/20">
        <button
          type="button"
          onclick={() => (activePlatform = 'all')}
          class="px-3.5 py-1.5 rounded-full text-xs font-display font-bold cursor-pointer transition-colors {activePlatform ===
          'all'
            ? 'bg-[#19231f] text-[#dffc78]'
            : 'text-[#48534e] hover:text-[#19231f]'}"
        >
          All 4 Platforms
        </button>
        <button
          type="button"
          onclick={() => (activePlatform = 'discord')}
          class="px-3.5 py-1.5 rounded-full text-xs font-display font-bold cursor-pointer transition-colors {activePlatform ===
          'discord'
            ? 'bg-[#5865F2] text-white'
            : 'text-[#48534e] hover:text-[#19231f]'}"
        >
          Discord
        </button>
        <button
          type="button"
          onclick={() => (activePlatform = 'telegram')}
          class="px-3.5 py-1.5 rounded-full text-xs font-display font-bold cursor-pointer transition-colors {activePlatform ===
          'telegram'
            ? 'bg-[#24A1DE] text-white'
            : 'text-[#48534e] hover:text-[#19231f]'}"
        >
          Telegram
        </button>
        <button
          type="button"
          onclick={() => (activePlatform = 'whatsapp')}
          class="px-3.5 py-1.5 rounded-full text-xs font-display font-bold cursor-pointer transition-colors {activePlatform ===
          'whatsapp'
            ? 'bg-[#25D366] text-[#19231f]'
            : 'text-[#48534e] hover:text-[#19231f]'}"
        >
          WhatsApp
        </button>
        <button
          type="button"
          onclick={() => (activePlatform = 'facebook')}
          class="px-3.5 py-1.5 rounded-full text-xs font-display font-bold cursor-pointer transition-colors {activePlatform ===
          'facebook'
            ? 'bg-[#1877F2] text-white'
            : 'text-[#48534e] hover:text-[#19231f]'}"
        >
          Facebook / LinkedIn
        </button>
      </div>

      <div class="flex items-center gap-2">
        <a
          href={report.finalUrl}
          target="_blank"
          rel="noopener noreferrer"
          class="px-3.5 py-1.5 rounded-full text-xs font-display font-bold bg-[#dffc78] text-[#19231f] border border-[#19231f]"
        >
          ↗ Open Live Page
        </a>
        {#if onCrawlUrl}
          <button
            type="button"
            onclick={() => onCrawlUrl?.(report.finalUrl)}
            class="px-3.5 py-1.5 rounded-full text-xs font-display font-bold bg-[#fffdf8] hover:bg-[#d9d6fc] text-[#19231f] border border-[#19231f] cursor-pointer"
          >
            Crawl Site Tree from URL →
          </button>
        {/if}
      </div>
    </div>

    <!-- Main Asymmetric Grid: 7-Col Live Platform Card Simulators + 5-Col Ad/Tracker Posture & Meta Audit -->
    <div class="grid grid-cols-1 lg:grid-cols-12 gap-6 items-start">
      <!-- 7-Col Social Platform Embed Simulators -->
      <div class="lg:col-span-7 space-y-6">
        <!-- 1. DISCORD EMBED PREVIEW -->
        {#if activePlatform === 'all' || activePlatform === 'discord'}
          <div class="bento-card overflow-hidden bg-[#313338] text-white border-[1.5px] border-[#19231f]">
            <div class="px-5 py-3 bg-[#1e1f22] border-b border-white/10 flex items-center justify-between">
              <div class="flex items-center gap-2">
                <span class="w-2.5 h-2.5 rounded-full bg-[#5865F2]"></span>
                <span class="font-display font-bold text-xs uppercase tracking-wider text-white/90">
                  Discord Chat Embed Preview
                </span>
              </div>
              <span class="text-[11px] font-mono text-white/60">
                theme-color: {report.resolvedThemeColor}
              </span>
            </div>

            <div class="p-5 space-y-2">
              <div class="text-xs font-mono text-[#00a8fc] hover:underline break-all">
                {report.finalUrl}
              </div>

              <!-- Discord Embed Container with Left Accent Border -->
              <div
                class="rounded-md bg-[#2b2d31] p-4 max-w-lg space-y-2 border-l-4"
                style="border-left-color: {report.resolvedThemeColor || '#5865F2'}"
              >
                <div class="text-xs text-[#b5bac1] font-medium">
                  {report.resolvedSiteName}
                </div>
                <div class="font-display font-bold text-base text-[#00a8fc] hover:underline leading-snug">
                  {report.resolvedTitle}
                </div>
                <p class="text-xs text-[#dbdee1] leading-relaxed line-clamp-3">
                  {report.resolvedDescription}
                </p>

                {#if report.resolvedImage}
                  <div class="pt-2">
                    <img
                      src={report.resolvedImage}
                      alt={report.resolvedTitle}
                      class="rounded-md max-h-60 w-full object-cover border border-white/10 bg-[#1e1f22]"
                      onerror={(e) => ((e.currentTarget as HTMLElement).style.display = 'none')}
                    />
                  </div>
                {:else}
                  <div class="rounded-md bg-[#1e1f22] border border-dashed border-white/20 p-4 text-center text-xs text-white/50 font-mono">
                    No og:image / twitter:image tag detected — Discord renders text-only embed
                  </div>
                {/if}
              </div>
            </div>
          </div>
        {/if}

        <!-- 2. TELEGRAM LINK PREVIEW -->
        {#if activePlatform === 'all' || activePlatform === 'telegram'}
          <div class="bento-card overflow-hidden bg-[#17212b] text-white border-[1.5px] border-[#19231f]">
            <div class="px-5 py-3 bg-[#0e1621] border-b border-white/10 flex items-center justify-between">
              <div class="flex items-center gap-2">
                <span class="w-2.5 h-2.5 rounded-full bg-[#24A1DE]"></span>
                <span class="font-display font-bold text-xs uppercase tracking-wider text-white/90">
                  Telegram Instant Link Preview
                </span>
              </div>
              <span class="text-[11px] font-mono text-white/60">Telegram Bot / Client</span>
            </div>

            <div class="p-5">
              <div class="rounded-2xl bg-[#182533] p-4 max-w-lg space-y-2 shadow-md border border-white/10">
                <div class="text-xs font-mono text-[#6ab3f3] break-all">
                  {report.finalUrl}
                </div>
                <div class="border-l-2 border-[#6ab3f3] pl-3 space-y-1.5">
                  <div class="text-xs font-bold text-[#6ab3f3]">{report.resolvedSiteName}</div>
                  <div class="font-display font-bold text-sm text-white leading-snug">
                    {report.resolvedTitle}
                  </div>
                  <p class="text-xs text-white/80 leading-relaxed line-clamp-3">
                    {report.resolvedDescription}
                  </p>
                  {#if report.resolvedImage}
                    <img
                      src={report.resolvedImage}
                      alt={report.resolvedTitle}
                      class="rounded-lg max-h-56 w-full object-cover mt-2 bg-[#0e1621]"
                      onerror={(e) => ((e.currentTarget as HTMLElement).style.display = 'none')}
                    />
                  {/if}
                </div>
              </div>
            </div>
          </div>
        {/if}

        <!-- 3. WHATSAPP CHAT LINK BUBBLE -->
        {#if activePlatform === 'all' || activePlatform === 'whatsapp'}
          <div class="bento-card overflow-hidden bg-[#efeae2] border-[1.5px] border-[#19231f]">
            <div class="px-5 py-3 bg-[#075e54] text-white flex items-center justify-between">
              <div class="flex items-center gap-2">
                <span class="w-2.5 h-2.5 rounded-full bg-[#25D366]"></span>
                <span class="font-display font-bold text-xs uppercase tracking-wider">
                  WhatsApp Message Link Preview
                </span>
              </div>
              <span class="text-[11px] font-mono text-white/80">Mobile & Desktop Share</span>
            </div>

            <div class="p-5">
              <div class="rounded-2xl bg-[#d9fdd3] p-2.5 max-w-md shadow border border-[#19231f]/15 space-y-2">
                <!-- Inner Link Box -->
                <div class="rounded-xl bg-[#f0f2f5] overflow-hidden border border-black/5">
                  {#if report.resolvedImage}
                    <img
                      src={report.resolvedImage}
                      alt={report.resolvedTitle}
                      class="max-h-48 w-full object-cover bg-white"
                      onerror={(e) => ((e.currentTarget as HTMLElement).style.display = 'none')}
                    />
                  {/if}
                  <div class="p-3 space-y-1">
                    <div class="font-display font-bold text-xs text-[#111b21] line-clamp-2">
                      {report.resolvedTitle}
                    </div>
                    <p class="text-[11px] text-[#667781] line-clamp-2">
                      {report.resolvedDescription}
                    </p>
                    <div class="text-[10px] font-mono text-[#8696a0] pt-0.5">
                      {report.host}
                    </div>
                  </div>
                </div>
                <div class="px-1 text-xs font-mono text-[#027eb5] underline break-all">
                  {report.finalUrl}
                </div>
              </div>
            </div>
          </div>
        {/if}

        <!-- 4. FACEBOOK / LINKEDIN FEED CARD -->
        {#if activePlatform === 'all' || activePlatform === 'facebook'}
          <div class="bento-card overflow-hidden bg-[#fffdf8] border-[1.5px] border-[#19231f]">
            <div class="px-5 py-3 bg-[#f4f1e9] border-b border-[#19231f]/15 flex items-center justify-between">
              <div class="flex items-center gap-2">
                <span class="w-2.5 h-2.5 rounded-full bg-[#1877F2]"></span>
                <span class="font-display font-bold text-xs uppercase tracking-wider text-[#19231f]">
                  Facebook / LinkedIn / X Large Card Preview
                </span>
              </div>
              <span class="text-[11px] font-mono text-[#48534e]">
                {report.twitterCard || 'summary_large_image'}
              </span>
            </div>

            <div class="p-5">
              <div class="rounded-xl overflow-hidden border border-[#19231f]/20 bg-[#f0f2f5] max-w-lg">
                {#if report.resolvedImage}
                  <img
                    src={report.resolvedImage}
                    alt={report.resolvedTitle}
                    class="max-h-64 w-full object-cover bg-[#19231f]"
                    onerror={(e) => ((e.currentTarget as HTMLElement).style.display = 'none')}
                  />
                {:else}
                  <div class="h-36 bg-[#e4e6eb] flex items-center justify-center text-xs font-mono text-[#65676b]">
                    No og:image Banner Configured (1200×630 recommended)
                  </div>
                {/if}
                <div class="p-3.5 space-y-1 bg-[#f0f2f5]">
                  <div class="text-[11px] font-mono uppercase tracking-wider text-[#65676b]">
                    {report.host}
                  </div>
                  <div class="font-display font-bold text-sm text-[#050505] line-clamp-1">
                    {report.resolvedTitle}
                  </div>
                  <p class="text-xs text-[#65676b] line-clamp-2">
                    {report.resolvedDescription}
                  </p>
                </div>
              </div>
            </div>
          </div>
        {/if}
      </div>

      <!-- 5-Col Right Column: Ad Networks / Analytics Tracker Card + Social Audit Checklist + Raw Meta Ledger -->
      <div class="lg:col-span-5 space-y-6">
        <!-- Ad Networks, Analytics & Tracker Telemetry Card -->
        {#if report.trackers}
          <TrackerPostureCard trackers={report.trackers} />
        {/if}

        <!-- Social & SEO Meta Readiness Grade Card -->
        <div class="bento-card overflow-hidden bg-[#fffdf8]">
          <div class="bg-[#19231f] text-[#fffdf8] p-5 flex items-center justify-between">
            <div>
              <span class="text-[11px] font-display uppercase tracking-widest text-[#dffc78]">
                Social Card Readiness Audit
              </span>
              <h3 class="text-lg font-display font-bold mt-0.5 text-[#fffdf8]">
                OpenGraph &amp; Twitter Tag Score: {report.socialScore}/100
              </h3>
            </div>
            <div class="w-12 h-12 rounded-2xl bg-[#dffc78] text-[#19231f] font-display font-bold text-xl flex items-center justify-center">
              {report.socialGrade}
            </div>
          </div>

          <div class="p-5 space-y-2.5">
            {#each report.auditChecks as chk}
              <div class="rounded-xl bg-[#f4f1e9] border border-[#19231f]/12 p-3 flex items-start justify-between gap-2 text-xs">
                <div>
                  <div class="font-display font-bold text-[#19231f]">{chk.label}</div>
                  <div class="text-[#48534e] font-mono text-[11px] break-all mt-0.5">{chk.details}</div>
                </div>
                <span
                  class="px-2 py-0.5 rounded-full text-[10px] font-mono font-bold border shrink-0 {chk.status ===
                  'PASS'
                    ? 'bg-[#dffc78] text-[#19231f] border-[#19231f]'
                    : chk.status === 'WARN'
                      ? 'bg-[#d9d6fc] text-[#19231f] border-[#19231f]/40'
                      : 'bg-[#ffc3a5] text-[#19231f] border-[#19231f]'}"
                >
                  {chk.status}
                </span>
              </div>
            {/each}
          </div>
        </div>

        <!-- Extracted HTML Meta Tags Ledger -->
        <div class="bento-card p-5 space-y-3 bg-[#fffdf8]">
          <div class="flex items-center justify-between border-b border-[#19231f]/10 pb-2.5">
            <h4 class="font-display font-bold text-sm text-[#19231f]">
              Extracted HTML &lt;head&gt; Metadata
            </h4>
            <span class="studio-label">Raw Values</span>
          </div>

          <div class="space-y-2 text-xs font-mono max-h-80 scroll-panel pr-1">
            {#each [
              ['title', report.title],
              ['description', report.description],
              ['canonical', report.canonicalUrl],
              ['og:title', report.ogTitle],
              ['og:description', report.ogDescription],
              ['og:image', report.ogImage],
              ['og:site_name', report.ogSiteName],
              ['twitter:card', report.twitterCard],
              ['twitter:image', report.twitterImage],
              ['theme-color', report.themeColor],
              ['robots', report.robotsMeta],
              ['generator', report.generator],
            ].filter(([, v]) => Boolean(v)) as [k, val]}
              <div class="rounded-xl bg-[#f4f1e9] border border-[#19231f]/10 p-2.5">
                <div class="text-[10px] font-bold text-[#5366e8] uppercase">{k}</div>
                <div class="text-[#19231f] break-all mt-0.5">{val}</div>
              </div>
            {/each}
          </div>
        </div>
      </div>
    </div>
  {/if}
</section>
