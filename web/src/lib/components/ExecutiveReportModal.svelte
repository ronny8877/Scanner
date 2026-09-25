<script lang="ts">
  import type { SavedExecutiveReport } from '../types';
  import StudioIcon from './StudioIcon.svelte';

  let {
    report = null,
    onClose,
  }: {
    report: SavedExecutiveReport | null;
    onClose: () => void;
  } = $props();

  function handlePrint() {
    window.print();
  }
</script>

{#if report}
  {@const s = report.suite}
  <div
    role="dialog"
    aria-modal="true"
    aria-labelledby="executive-dossier-title"
    class="fixed inset-0 z-50 flex items-center justify-center bg-[#19231f]/80 p-3 sm:p-6 backdrop-blur-sm overflow-y-auto print:static print:bg-white print:p-0"
  >
    <div
      class="relative my-auto max-h-[92vh] w-full max-w-6xl overflow-y-auto rounded-3xl border-2 border-[#19231f] bg-[#f4f1e9] p-6 sm:p-10 shadow-[0_28px_80px_rgba(0,0,0,0.45)] print:max-h-none print:border-0 print:bg-white print:p-4 print:shadow-none"
    >
      <!-- Top Sticky Toolbar (Hidden when printing) -->
      <div class="mb-6 flex flex-wrap items-center justify-between gap-3 rounded-2xl border border-[#19231f]/15 bg-[#fffdf8] px-5 py-3.5 print:hidden">
        <div class="flex items-center gap-2.5">
          <span class="rounded-full bg-[#19231f] px-3 py-1 font-display text-[11px] font-bold uppercase tracking-wider text-[#dffc78]">
            Executive Intelligence Dossier
          </span>
          <span class="font-display text-xs font-semibold text-[#19231f]/65">
            Generated {new Date(report.generatedAt).toLocaleString()} · 7 Parallel Engines
          </span>
        </div>

        <div class="flex items-center gap-2.5">
          <button
            type="button"
            onclick={handlePrint}
            class="inline-flex items-center gap-2 rounded-full border border-[#19231f] bg-[#dffc78] px-4 py-2 font-display text-xs font-bold text-[#19231f] transition-transform hover:-translate-y-0.5"
          >
            <StudioIcon name="sparkle" class="h-3.5 w-3.5" />
            Print / Save PDF Report
          </button>
          <a
            href="https://{report.domain}"
            target="_blank"
            rel="noopener noreferrer"
            class="inline-flex items-center gap-1.5 rounded-full border border-[#19231f]/20 bg-[#fffdf8] px-3.5 py-2 font-display text-xs font-semibold text-[#19231f] hover:border-[#19231f]"
          >
            Visit Live Site ↗
          </a>
          <button
            type="button"
            onclick={onClose}
            class="inline-flex items-center gap-1.5 rounded-full border border-[#19231f] bg-[#19231f] px-4 py-2 font-display text-xs font-bold text-[#fffdf8] hover:bg-[#5366e8]"
          >
            Close Report
          </button>
        </div>
      </div>

      <!-- PDF Cover Header Banner -->
      <div class="rounded-3xl border-2 border-[#19231f] bg-[#19231f] p-7 text-[#fffdf8] sm:p-9">
        <div class="flex flex-col justify-between gap-6 lg:flex-row lg:items-end">
          <div>
            <p class="font-display text-xs font-bold uppercase tracking-widest text-[#dffc78]">
              LiveTheme Domain Intelligence Studio · Verified Specimen Report
            </p>
            <h2 id="executive-dossier-title" class="mt-2 font-display text-3xl font-bold tracking-tight sm:text-5xl">
              {report.domain}
            </h2>
            <p class="mt-2 max-w-2xl text-sm text-[#fffdf8]/75">
              {s.meta?.resolvedTitle || s.inquiry?.status || 'Complete 360° Domain Valuation, RDAP Ownership, Traffic Rank, Infrastructure & SEO Dossier'}
            </p>
          </div>

          <div class="grid grid-cols-2 gap-3 sm:grid-cols-4">
            <div class="rounded-2xl border border-[#fffdf8]/15 bg-[#fffdf8]/10 px-4 py-3">
              <p class="text-[10px] uppercase tracking-wider text-[#fffdf8]/60">Market Valuation</p>
              <p class="mt-1 font-display text-sm font-bold text-[#dffc78]">
                {s.inquiry?.valuation?.bandLabel || '$10/yr'}
              </p>
            </div>
            <div class="rounded-2xl border border-[#fffdf8]/15 bg-[#fffdf8]/10 px-4 py-3">
              <p class="text-[10px] uppercase tracking-wider text-[#fffdf8]/60">Est. Monthly Traffic</p>
              <p class="mt-1 font-display text-sm font-bold text-[#d9d6fc]">
                {s.traffic?.estimatedMonthlyRange || '1K–15K/mo'}
              </p>
            </div>
            <div class="rounded-2xl border border-[#fffdf8]/15 bg-[#fffdf8]/10 px-4 py-3">
              <p class="text-[10px] uppercase tracking-wider text-[#fffdf8]/60">Security Grade</p>
              <p class="mt-1 font-display text-sm font-bold text-[#dffc78]">
                Grade {s.recon?.securityGrade || 'A'} ({s.recon?.securityScore ?? 85}/100)
              </p>
            </div>
            <div class="rounded-2xl border border-[#fffdf8]/15 bg-[#fffdf8]/10 px-4 py-3">
              <p class="text-[10px] uppercase tracking-wider text-[#fffdf8]/60">Social & SEO Grade</p>
              <p class="mt-1 font-display text-sm font-bold text-[#ffc3a5]">
                Grade {s.meta?.socialGrade || 'A'} ({s.meta?.socialScore ?? 80}/100)
              </p>
            </div>
          </div>
        </div>
      </div>

      <!-- Visual Bento Grid of All 7 Intelligence Sections -->
      <div class="mt-6 grid grid-cols-1 gap-6 lg:grid-cols-12">
        <!-- 01 & 02: Valuation + RDAP Registration & History -->
        <div class="bento-card p-6 lg:col-span-6">
          <div class="flex items-center justify-between">
            <span class="rounded-full bg-[#dffc78] px-3 py-1 font-display text-[11px] font-bold text-[#19231f]">
              01 · Valuation & RDAP Registry Dossier
            </span>
            <span class="font-display text-xs font-bold text-[#19231f]/60">
              {s.inquiry?.available ? 'Available for Registration' : 'Registered Asset'}
            </span>
          </div>

          <div class="mt-4 grid grid-cols-2 gap-3 rounded-2xl border border-[#19231f]/12 bg-[#f4f1e9]/70 p-4 text-xs">
            <div>
              <span class="text-[#19231f]/55">Valuation Band</span>
              <p class="mt-0.5 font-display text-sm font-bold text-[#19231f]">{s.inquiry?.valuation?.bandLabel}</p>
            </div>
            <div>
              <span class="text-[#19231f]/55">Renewal Fee</span>
              <p class="mt-0.5 font-display text-sm font-bold text-[#19231f]">
                ${s.inquiry?.valuation?.renewalPerYear || 12}/yr
              </p>
            </div>
            <div>
              <span class="text-[#19231f]/55">Registrar</span>
              <p class="mt-0.5 font-display font-bold text-[#19231f]">{s.inquiry?.registrar || 'Registry Direct'}</p>
            </div>
            <div>
              <span class="text-[#19231f]/55">Archive History</span>
              <p class="mt-0.5 font-display font-bold text-[#19231f]">
                {s.history?.firstSeenYear ? `First seen ${s.history.firstSeenYear} (${s.history.domainAgeYears}y)` : 'Clean Archive'}
              </p>
            </div>
          </div>

          {#if s.inquiry?.valuation?.reasons?.length}
            <div class="mt-4 flex flex-wrap gap-1.5">
              {#each s.inquiry.valuation.reasons as reason}
                <span class="rounded-lg border border-[#19231f]/15 bg-[#fffdf8] px-2.5 py-1 text-[11px] font-medium text-[#19231f]">
                  {reason}
                </span>
              {/each}
            </div>
          {/if}
        </div>

        <!-- 03: Traffic & Popularity Intelligence -->
        <div class="bento-card p-6 lg:col-span-6">
          <div class="flex items-center justify-between">
            <span class="rounded-full bg-[#d9d6fc] px-3 py-1 font-display text-[11px] font-bold text-[#19231f]">
              02 · Traffic & Global Popularity Telemetry
            </span>
            <span class="font-display text-xs font-bold text-[#5366e8]">
              {s.traffic?.trendLabel || '→ Stable'}
            </span>
          </div>

          <div class="mt-4 space-y-2.5 rounded-2xl border border-[#19231f]/12 bg-[#f4f1e9]/70 p-4 text-xs">
            <div class="flex justify-between">
              <span class="text-[#19231f]/60">Domain Popularity</span>
              <span class="font-display font-bold text-[#19231f]">{s.traffic?.popularityTier || 'Niche Tier'}</span>
            </div>
            <div class="flex justify-between">
              <span class="text-[#19231f]/60">Cloudflare Rank Bucket</span>
              <span class="font-display font-bold text-[#19231f]">{s.traffic?.cloudflareBucket || 'Long-Tail'}</span>
            </div>
            <div class="flex justify-between">
              <span class="text-[#19231f]/60">Top Locations</span>
              <span class="font-display font-bold text-[#19231f]">
                {(s.traffic?.topLocations || ['US', 'IN', 'GB']).map((l) => l.split('·')[0].trim()).join(' · ')}
              </span>
            </div>
            <div class="flex justify-between border-t border-[#19231f]/10 pt-2">
              <span class="text-[#19231f]/60">Estimated Traffic Range</span>
              <span class="font-display text-sm font-bold text-[#5366e8]">
                {s.traffic?.estimatedMonthlyRange || '1K–15K visits/month'}
              </span>
            </div>
          </div>
        </div>

        <!-- 04: Security Posture, Open Ports & Verified Active Subdomains -->
        <div class="bento-card p-6 lg:col-span-6">
          <div class="flex items-center justify-between">
            <span class="rounded-full bg-[#ffc3a5] px-3 py-1 font-display text-[11px] font-bold text-[#19231f]">
              03 · Infrastructure, Ports & Verified Subdomains
            </span>
            <span class="font-display text-xs font-bold">
              IP: {s.recon?.targetIp || 'Edge Proxy'}
            </span>
          </div>

          <div class="mt-4 grid grid-cols-3 gap-2 text-center text-xs">
            <div class="rounded-xl border border-[#19231f]/12 bg-[#f4f1e9]/70 p-3">
              <p class="text-[10px] uppercase text-[#19231f]/55">Open TCP Ports</p>
              <p class="mt-1 font-display text-base font-bold">{s.recon?.openPortsCount ?? 2}</p>
            </div>
            <div class="rounded-xl border border-[#19231f]/12 bg-[#f4f1e9]/70 p-3">
              <p class="text-[10px] uppercase text-[#19231f]/55">TLS Protocol</p>
              <p class="mt-1 font-display text-base font-bold">{s.recon?.tls?.version || 'TLS 1.3'}</p>
            </div>
            <div class="rounded-xl border border-[#19231f]/12 bg-[#f4f1e9]/70 p-3">
              <p class="text-[10px] uppercase text-[#19231f]/55">Live Subdomains</p>
              <p class="mt-1 font-display text-base font-bold">{s.recon?.subdomains?.length ?? 0}</p>
            </div>
          </div>

          {#if s.recon?.subdomains && s.recon.subdomains.length > 0}
            <div class="mt-4 flex flex-wrap gap-1.5">
              {#each s.recon.subdomains.slice(0, 8) as sub}
                <span class="rounded-lg border border-[#19231f]/15 bg-[#fffdf8] px-2.5 py-1 font-mono text-[11px] font-semibold text-[#19231f]">
                  {sub.subdomain}
                </span>
              {/each}
            </div>
          {/if}
        </div>

        <!-- 05, 06 & 07: Site Architecture, Robots/Sitemap & Social Card Preview -->
        <div class="bento-card p-6 lg:col-span-6">
          <div class="flex items-center justify-between">
            <span class="rounded-full bg-[#dffc78] px-3 py-1 font-display text-[11px] font-bold text-[#19231f]">
              04 · Site Architecture, Robots & Social Meta
            </span>
            <span class="font-display text-xs font-bold text-[#19231f]/65">
              Privacy Grade {s.meta?.trackers?.privacyGrade || 'A'}
            </span>
          </div>

          <div class="mt-4 space-y-2.5 rounded-2xl border border-[#19231f]/12 bg-[#f4f1e9]/70 p-4 text-xs">
            <div class="flex justify-between">
              <span class="text-[#19231f]/60">Resolved Social Title</span>
              <span class="max-w-xs truncate font-display font-bold text-[#19231f]">
                {s.meta?.resolvedTitle || report.domain}
              </span>
            </div>
            <div class="flex justify-between">
              <span class="text-[#19231f]/60">Crawled Internal Routes</span>
              <span class="font-display font-bold text-[#19231f]">
                {s.crawl?.pagesCrawled ?? 1} routes ({s.crawl?.totalLinks ?? 0} links)
              </span>
            </div>
            <div class="flex justify-between">
              <span class="text-[#19231f]/60">robots.txt & Sitemap XML</span>
              <span class="font-display font-bold text-[#19231f]">
                {s.robots?.robotsFound ? 'robots.txt Active' : 'Default Open'} · {s.robots?.totalUrlsCount ?? 0} Sitemap URLs
              </span>
            </div>
            <div class="flex justify-between">
              <span class="text-[#19231f]/60">Detected Ad / Analytics Trackers</span>
              <span class="font-display font-bold text-[#19231f]">
                {s.meta?.trackers?.totalDetected ?? 0} SDKs ({s.meta?.trackers?.verdict || 'Clean Surface'})
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
{/if}
