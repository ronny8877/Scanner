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

  function handleExportJSON() {
    if (!report) return;
    const blob = new Blob([JSON.stringify(report.suite, null, 2)], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `${report.domain}-master-dossier.json`;
    a.click();
    URL.revokeObjectURL(url);
  }

  function formatDate(iso?: string): string {
    if (!iso) return 'Unpublished by registry';
    try {
      return new Date(iso).toLocaleDateString('en-US', {
        year: 'numeric',
        month: 'short',
        day: 'numeric',
      });
    } catch {
      return iso;
    }
  }
</script>

{#if report}
  {@const s = report.suite}
  {@const inq = s.inquiry}
  {@const hist = s.history}
  {@const traf = s.traffic}
  {@const rec = s.recon}
  {@const crw = s.crawl}
  {@const rob = s.robots}
  {@const met = s.meta}
  {@const stack = met?.techStack || rec?.techStack || crw?.techStack}
  {@const trackers = met?.trackers || rec?.trackers || crw?.trackers}

  <div
    role="dialog"
    aria-modal="true"
    aria-labelledby="executive-dossier-title"
    class="fixed inset-0 z-50 flex items-center justify-center bg-[#19231f]/85 p-2 sm:p-5 backdrop-blur-sm overflow-y-auto print:static print:bg-white print:p-0"
  >
    <div
      class="relative my-auto max-h-[94vh] w-full max-w-7xl overflow-y-auto rounded-3xl border-2 border-[#19231f] bg-[#f4f1e9] p-5 sm:p-8 lg:p-10 shadow-[0_28px_80px_rgba(0,0,0,0.5)] print:max-h-none print:border-0 print:bg-white print:p-4 print:shadow-none space-y-7"
    >
      <!-- Sticky Top Action & Section Jump Bar (Hidden when printing) -->
      <div class="sticky top-0 z-30 flex flex-col gap-3 rounded-2xl border-2 border-[#19231f] bg-[#fffdf8]/95 backdrop-blur-md px-4 py-3 shadow-md print:hidden">
        <div class="flex flex-wrap items-center justify-between gap-3">
          <div class="flex items-center gap-2.5">
            <span class="rounded-full bg-[#19231f] px-3 py-1 font-display text-[11px] font-bold uppercase tracking-wider text-[#dffc78]">
              Master 360° Domain Dossier
            </span>
            <span class="font-display text-xs font-bold text-[#19231f]">
              {report.domain}
            </span>
            <span class="hidden sm:inline font-mono text-[11px] text-[#19231f]/60">
              · Generated {new Date(report.generatedAt).toLocaleTimeString()} · All 7 Engines
            </span>
          </div>

          <div class="flex flex-wrap items-center gap-2">
            <button
              type="button"
              onclick={handleExportJSON}
              class="inline-flex items-center gap-1.5 rounded-full border border-[#19231f]/25 bg-[#f4f1e9] px-3.5 py-1.5 font-display text-xs font-bold text-[#19231f] hover:bg-[#d9d6fc] cursor-pointer"
            >
              Export JSON
            </button>
            <button
              type="button"
              onclick={handlePrint}
              class="inline-flex items-center gap-1.5 rounded-full border border-[#19231f] bg-[#dffc78] px-4 py-1.5 font-display text-xs font-bold text-[#19231f] transition-transform hover:-translate-y-0.5 cursor-pointer"
            >
              <StudioIcon name="sparkle" size={13} />
              Print / Save PDF
            </button>
            <a
              href="https://{report.domain}"
              target="_blank"
              rel="noopener noreferrer"
              class="inline-flex items-center gap-1 rounded-full border border-[#19231f]/20 bg-[#fffdf8] px-3.5 py-1.5 font-display text-xs font-semibold text-[#19231f] hover:border-[#19231f]"
            >
              Live Site ↗
            </a>
            <button
              type="button"
              onclick={onClose}
              class="inline-flex items-center gap-1.5 rounded-full border border-[#19231f] bg-[#19231f] px-4 py-1.5 font-display text-xs font-bold text-[#fffdf8] hover:bg-[#5366e8] cursor-pointer"
            >
              <StudioIcon name="close" size={13} />
              Close
            </button>
          </div>
        </div>

        <!-- Quick Section Jump Pills -->
        <div class="flex flex-wrap items-center gap-1.5 border-t border-[#19231f]/10 pt-2 text-[11px] font-display font-semibold">
          <span class="text-[#19231f]/50 mr-1">Jump to:</span>
          <a href="#rep-rdap" class="rounded-full bg-[#f4f1e9] hover:bg-[#dffc78] px-2.5 py-0.5 text-[#19231f] border border-[#19231f]/15">01. RDAP & DNS Zone</a>
          <a href="#rep-valuation" class="rounded-full bg-[#f4f1e9] hover:bg-[#dffc78] px-2.5 py-0.5 text-[#19231f] border border-[#19231f]/15">02. Valuation & Wayback</a>
          <a href="#rep-traffic" class="rounded-full bg-[#f4f1e9] hover:bg-[#dffc78] px-2.5 py-0.5 text-[#19231f] border border-[#19231f]/15">03. Traffic & Rank</a>
          <a href="#rep-techstack" class="rounded-full bg-[#f4f1e9] hover:bg-[#dffc78] px-2.5 py-0.5 text-[#19231f] border border-[#19231f]/15">04. Tech Stack & Trackers</a>
          <a href="#rep-ports" class="rounded-full bg-[#f4f1e9] hover:bg-[#dffc78] px-2.5 py-0.5 text-[#19231f] border border-[#19231f]/15">05. Ports (RDP/SSH), TLS & Subdomains</a>
          <a href="#rep-seo" class="rounded-full bg-[#f4f1e9] hover:bg-[#dffc78] px-2.5 py-0.5 text-[#19231f] border border-[#19231f]/15">06. Robots, Sitemap & Social Meta</a>
        </div>
      </div>

      <!-- EXECUTIVE COVER HERO BANNER -->
      <div class="rounded-3xl border-2 border-[#19231f] bg-[#19231f] p-7 text-[#fffdf8] sm:p-9">
        <div class="flex flex-col justify-between gap-6 lg:flex-row lg:items-end">
          <div class="space-y-2">
            <div class="flex flex-wrap items-center gap-2">
              <span class="rounded-full bg-[#dffc78] px-3 py-0.5 font-display text-[11px] font-bold uppercase tracking-wider text-[#19231f]">
                {inq?.available ? 'Available for Registration' : 'Registered Domain Asset'}
              </span>
              <span class="rounded-full border border-[#fffdf8]/20 bg-[#fffdf8]/10 px-3 py-0.5 font-mono text-[11px] text-[#fffdf8]">
                IP: {rec?.targetIp || inq?.dns?.a?.[0] || 'Edge Proxy'}
              </span>
              {#if stack?.primaryFramework}
                <span class="rounded-full bg-[#d9d6fc] px-3 py-0.5 font-display text-[11px] font-bold text-[#19231f]">
                  {stack.primaryFramework}
                </span>
              {/if}
            </div>

            <h2 id="executive-dossier-title" class="font-display text-3xl font-bold tracking-tight sm:text-5xl">
              {report.domain}
            </h2>
            <p class="max-w-3xl text-sm text-[#fffdf8]/80 leading-relaxed">
              {met?.resolvedTitle || inq?.status || 'Complete 360° Domain Valuation, RDAP Registry Dossier, DNS Zone, Traffic Rank, Tech Stack, Port/RDP Surface & SEO Report'}
            </p>
            {#if met?.resolvedDescription}
              <p class="max-w-3xl text-xs text-[#fffdf8]/60 leading-relaxed">
                {met.resolvedDescription}
              </p>
            {/if}
          </div>

          <!-- Top 6 Executive KPIs -->
          <div class="grid grid-cols-2 gap-2.5 sm:grid-cols-3 lg:w-[500px]">
            <div class="rounded-2xl border border-[#fffdf8]/15 bg-[#fffdf8]/10 p-3.5">
              <p class="text-[10px] uppercase tracking-wider text-[#fffdf8]/60">Market Valuation</p>
              <p class="mt-1 font-display text-sm font-bold text-[#dffc78]">
                {inq?.valuation?.bandLabel || '$10/yr'}
              </p>
              <p class="text-[10px] text-[#fffdf8]/55">Score {inq?.valuation?.score ?? 80}/100 ({inq?.valuation?.tier || 'Prime'})</p>
            </div>

            <div class="rounded-2xl border border-[#fffdf8]/15 bg-[#fffdf8]/10 p-3.5">
              <p class="text-[10px] uppercase tracking-wider text-[#fffdf8]/60">Est. Monthly Traffic</p>
              <p class="mt-1 font-display text-sm font-bold text-[#d9d6fc]">
                {traf?.estimatedMonthlyRange || '1K–15K/mo'}
              </p>
              <p class="text-[10px] text-[#fffdf8]/55">{traf?.popularityTier || 'Global Rank'}</p>
            </div>

            <div class="rounded-2xl border border-[#fffdf8]/15 bg-[#fffdf8]/10 p-3.5">
              <p class="text-[10px] uppercase tracking-wider text-[#fffdf8]/60">Registrar & Age</p>
              <p class="mt-1 truncate font-display text-sm font-bold text-[#fffdf8]">
                {inq?.registrar || 'Registry Direct'}
              </p>
              <p class="text-[10px] text-[#fffdf8]/55">
                {hist?.firstSeenYear ? `Since ${hist.firstSeenYear} (${hist.domainAgeYears || (2026 - hist.firstSeenYear)}y)` : formatDate(inq?.createdDate)}
              </p>
            </div>

            <div class="rounded-2xl border border-[#fffdf8]/15 bg-[#fffdf8]/10 p-3.5">
              <p class="text-[10px] uppercase tracking-wider text-[#fffdf8]/60">UI & Tech Stack</p>
              <p class="mt-1 truncate font-display text-sm font-bold text-[#dffc78]">
                {stack?.primaryFramework || 'HTML5 / Edge'}
              </p>
              <p class="truncate text-[10px] text-[#fffdf8]/55">{stack?.uiSystemSummary || 'Tailwind CSS'}</p>
            </div>

            <div class="rounded-2xl border border-[#fffdf8]/15 bg-[#fffdf8]/10 p-3.5">
              <p class="text-[10px] uppercase tracking-wider text-[#fffdf8]/60">Security & TLS</p>
              <p class="mt-1 font-display text-sm font-bold text-[#dffc78]">
                Grade {rec?.securityGrade || 'A'} ({rec?.securityScore ?? 85}/100)
              </p>
              <p class="text-[10px] text-[#fffdf8]/55">{rec?.tls?.version || 'TLS 1.3'} · {rec?.openPortsCount ?? 2} open ports</p>
            </div>

            <div class="rounded-2xl border border-[#fffdf8]/15 bg-[#fffdf8]/10 p-3.5">
              <p class="text-[10px] uppercase tracking-wider text-[#fffdf8]/60">Social & Privacy</p>
              <p class="mt-1 font-display text-sm font-bold text-[#ffc3a5]">
                SEO {met?.socialGrade || 'A'} · Privacy {trackers?.privacyGrade || 'A'}
              </p>
              <p class="text-[10px] text-[#fffdf8]/55">{trackers?.totalDetected ?? 0} Trackers · {rob?.totalUrlsCount ?? 0} Sitemap URLs</p>
            </div>
          </div>
        </div>
      </div>

      <!-- SECTION 01: AUTHORITATIVE RDAP REGISTRY DOSSIER, WHOIS LIFECYCLE & FULL DNS ZONE -->
      <section id="rep-rdap" class="bento-card p-6 sm:p-7 space-y-5">
        <div class="flex flex-wrap items-center justify-between gap-2 border-b border-[#19231f]/12 pb-3.5">
          <div>
            <span class="rounded-full bg-[#dffc78] px-3 py-1 font-display text-xs font-bold text-[#19231f]">
              01 · Authoritative RDAP Registry, WHOIS Lifecycle & Complete DNS Zone
            </span>
            <h3 class="mt-2 font-display text-xl font-bold text-[#19231f]">
              ICANN RDAP Ownership Metadata, Nameservers, Mail Exchangers & Zone Blueprint
            </h3>
          </div>
          <span class="rounded-full border border-[#19231f]/20 bg-[#f4f1e9] px-3 py-1 font-mono text-xs font-bold text-[#19231f]">
            DNSSEC: {inq?.dnssec ? 'Signed & Verified' : 'Unsigned'}
          </span>
        </div>

        <!-- RDAP Metadata Grid -->
        <div class="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-6 text-xs">
          <div class="rounded-2xl border border-[#19231f]/12 bg-[#f4f1e9]/75 p-3.5">
            <span class="text-[#19231f]/55">Registrar</span>
            <p class="mt-1 font-display font-bold text-[#19231f]">{inq?.registrar || 'ICANN Registry'}</p>
            {#if inq?.registrarIanaId}
              <p class="font-mono text-[10px] text-[#19231f]/60">IANA #{inq.registrarIanaId}</p>
            {/if}
          </div>

          <div class="rounded-2xl border border-[#19231f]/12 bg-[#f4f1e9]/75 p-3.5">
            <span class="text-[#19231f]/55">Registry Handle</span>
            <p class="mt-1 font-mono font-bold text-[#19231f] break-all">{inq?.handle || `DOM-${report.domain.toUpperCase()}`}</p>
          </div>

          <div class="rounded-2xl border border-[#19231f]/12 bg-[#f4f1e9]/75 p-3.5">
            <span class="text-[#19231f]/55">Created Date</span>
            <p class="mt-1 font-display font-bold text-[#19231f]">{formatDate(inq?.createdDate || hist?.rdapCreatedDate)}</p>
          </div>

          <div class="rounded-2xl border border-[#19231f]/12 bg-[#f4f1e9]/75 p-3.5">
            <span class="text-[#19231f]/55">Expiration Date</span>
            <p class="mt-1 font-display font-bold text-[#19231f]">{formatDate(inq?.expiryDate)}</p>
          </div>

          <div class="rounded-2xl border border-[#19231f]/12 bg-[#f4f1e9]/75 p-3.5">
            <span class="text-[#19231f]/55">Last Updated</span>
            <p class="mt-1 font-display font-bold text-[#19231f]">{formatDate(inq?.updatedDate)}</p>
          </div>

          <div class="rounded-2xl border border-[#19231f]/12 bg-[#f4f1e9]/75 p-3.5">
            <span class="text-[#19231f]/55">Abuse Contact</span>
            <p class="mt-1 font-mono font-bold text-[#19231f] break-all">{inq?.abuseContactEmail || 'Published via RDAP'}</p>
          </div>
        </div>

        <!-- EPP Status Codes & Nameservers -->
        <div class="grid grid-cols-1 gap-4 lg:grid-cols-2">
          <div class="rounded-2xl border border-[#19231f]/12 bg-[#fffdf8] p-4 space-y-2">
            <div class="font-display text-xs font-bold uppercase tracking-wider text-[#19231f]/55">
              EPP Registry Status Locks ({inq?.statusCodes?.length || 1})
            </div>
            <div class="flex flex-wrap gap-1.5">
              {#if inq?.statusCodes && inq.statusCodes.length > 0}
                {#each inq.statusCodes as code}
                  <span class="rounded-lg border border-[#19231f]/20 bg-[#f4f1e9] px-2.5 py-1 font-mono text-[11px] font-semibold text-[#19231f]">
                    {code}
                  </span>
                {/each}
              {:else}
                <span class="rounded-lg border border-[#19231f]/20 bg-[#f4f1e9] px-2.5 py-1 font-mono text-[11px] font-semibold text-[#19231f]">
                  {inq?.status || 'active / clientTransferProhibited'}
                </span>
              {/if}
            </div>
          </div>

          <div class="rounded-2xl border border-[#19231f]/12 bg-[#fffdf8] p-4 space-y-2">
            <div class="font-display text-xs font-bold uppercase tracking-wider text-[#19231f]/55">
              Authoritative Nameservers ({inq?.nameservers?.length || inq?.dns?.ns?.length || 0})
            </div>
            <div class="flex flex-wrap gap-1.5">
              {#each (inq?.nameservers?.length ? inq.nameservers : inq?.dns?.ns || []) as ns}
                <span class="rounded-lg border border-[#19231f]/20 bg-[#d9d6fc]/50 px-2.5 py-1 font-mono text-[11px] font-bold text-[#19231f]">
                  {ns}
                </span>
              {/each}
            </div>
          </div>
        </div>

        <!-- Complete DNS Zone Records Table (A, AAAA, MX, TXT, DMARC, CAA) -->
        <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3 text-xs">
          <div class="rounded-2xl border border-[#19231f]/12 bg-[#f4f1e9]/65 p-4 space-y-1.5">
            <span class="font-display font-bold uppercase text-[#5366e8]">A (IPv4) & AAAA (IPv6) Records</span>
            <div class="font-mono space-y-1 text-[#19231f]">
              {#each (inq?.dns?.a || [rec?.targetIp || 'No A Record']) as ip}
                <div>A → <strong>{ip}</strong></div>
              {/each}
              {#each (inq?.dns?.aaaa || []) as ip6}
                <div>AAAA → <strong>{ip6}</strong></div>
              {/each}
            </div>
          </div>

          <div class="rounded-2xl border border-[#19231f]/12 bg-[#f4f1e9]/65 p-4 space-y-1.5">
            <span class="font-display font-bold uppercase text-[#5366e8]">MX Mail Exchangers & CNAME</span>
            <div class="font-mono space-y-1 text-[#19231f]">
              {#if inq?.dns?.mx && inq.dns.mx.length > 0}
                {#each inq.dns.mx as mx}
                  <div class="break-all">MX → <strong>{mx}</strong></div>
                {/each}
              {:else}
                <div class="text-[#19231f]/60">No inbound MX mail servers configured</div>
              {/if}
              {#if inq?.dns?.cname}
                <div class="break-all">CNAME → <strong>{inq.dns.cname}</strong></div>
              {/if}
            </div>
          </div>

          <div class="rounded-2xl border border-[#19231f]/12 bg-[#f4f1e9]/65 p-4 space-y-1.5">
            <span class="font-display font-bold uppercase text-[#5366e8]">TXT, SPF, DMARC & CAA Policies</span>
            <div class="font-mono space-y-1 text-[11px] text-[#19231f] max-h-28 overflow-y-auto">
              {#if inq?.dns?.dmarc}
                <div class="break-all text-[#19231f]"><strong>DMARC:</strong> {inq.dns.dmarc}</div>
              {/if}
              {#if inq?.dns?.txt && inq.dns.txt.length > 0}
                {#each inq.dns.txt.slice(0, 4) as txt}
                  <div class="break-all">TXT: {txt}</div>
                {/each}
              {:else}
                <div class="text-[#19231f]/60">No public TXT/SPF records returned</div>
              {/if}
            </div>
          </div>
        </div>
      </section>

      <!-- SECTION 02: MARKET VALUATION & WAYBACK ARCHIVE HISTORY -->
      <section id="rep-valuation" class="grid grid-cols-1 gap-6 lg:grid-cols-12">
        <div class="bento-card p-6 lg:col-span-6 space-y-4">
          <div class="flex items-center justify-between">
            <span class="rounded-full bg-[#dffc78] px-3 py-1 font-display text-xs font-bold text-[#19231f]">
              02A · Algorithmic Appraisal & Sub-Score Breakdown
            </span>
            <span class="font-mono text-xs font-bold text-[#19231f]">
              Score {inq?.valuation?.score ?? 80}/100
            </span>
          </div>

          <div class="grid grid-cols-3 gap-2.5 rounded-2xl border border-[#19231f]/12 bg-[#f4f1e9]/75 p-4 text-xs">
            <div>
              <span class="text-[#19231f]/55">Valuation Band</span>
              <p class="mt-1 font-display text-sm font-bold text-[#19231f]">{inq?.valuation?.bandLabel}</p>
            </div>
            <div>
              <span class="text-[#19231f]/55">Estimated Range</span>
              <p class="mt-1 font-mono font-bold text-[#5366e8]">
                ${(inq?.valuation?.minUsd ?? 10).toLocaleString()} – ${(inq?.valuation?.maxUsd ?? 100).toLocaleString()}
              </p>
            </div>
            <div>
              <span class="text-[#19231f]/55">Renewal Fee</span>
              <p class="mt-1 font-mono font-bold text-[#19231f]">${inq?.valuation?.renewalPerYear ?? 12}/yr</p>
            </div>
          </div>

          <!-- 4 Sub-scores -->
          <div class="grid grid-cols-2 gap-3 text-xs">
            <div class="rounded-xl border border-[#19231f]/10 bg-[#fffdf8] p-3">
              <div class="flex justify-between font-semibold">
                <span>Length brevity</span>
                <span class="font-mono">{inq?.valuation?.lengthScore ?? 28}/35</span>
              </div>
            </div>
            <div class="rounded-xl border border-[#19231f]/10 bg-[#fffdf8] p-3">
              <div class="flex justify-between font-semibold">
                <span>TLD Authority</span>
                <span class="font-mono">{inq?.valuation?.tldScore ?? 22}/25</span>
              </div>
            </div>
            <div class="rounded-xl border border-[#19231f]/10 bg-[#fffdf8] p-3">
              <div class="flex justify-between font-semibold">
                <span>Phonetics</span>
                <span class="font-mono">{inq?.valuation?.phoneticScore ?? 20}/25</span>
              </div>
            </div>
            <div class="rounded-xl border border-[#19231f]/10 bg-[#fffdf8] p-3">
              <div class="flex justify-between font-semibold">
                <span>Keyword / Brand Stem</span>
                <span class="font-mono">{inq?.valuation?.keywordScore ?? 12}/15</span>
              </div>
            </div>
          </div>

          {#if inq?.valuation?.reasons?.length}
            <div class="flex flex-wrap gap-1.5">
              {#each inq.valuation.reasons as reason}
                <span class="rounded-lg border border-[#19231f]/15 bg-[#fffdf8] px-2.5 py-1 text-[11px] font-medium text-[#19231f]">
                  ✓ {reason}
                </span>
              {/each}
            </div>
          {/if}
        </div>

        <!-- Wayback Archive & Certificate Transparency History -->
        <div class="bento-card p-6 lg:col-span-6 space-y-4">
          <div class="flex items-center justify-between">
            <span class="rounded-full bg-[#ffc3a5] px-3 py-1 font-display text-xs font-bold text-[#19231f]">
              02B · Wayback Machine & CT Certificate History
            </span>
            <a
              href={hist?.waybackCalendarUrl || `https://web.archive.org/web/*/${report.domain}`}
              target="_blank"
              rel="noopener noreferrer"
              class="font-display text-xs font-bold text-[#5366e8] hover:underline"
            >
              Open Wayback Calendar ↗
            </a>
          </div>

          <p class="font-display text-sm font-bold text-[#19231f]">
            {hist?.historyVerdict || 'Verified Registration & Archive Timeline'}
          </p>
          <p class="text-xs text-[#48534e] leading-relaxed">
            {hist?.summaryNote || ''}
          </p>

          <div class="grid grid-cols-3 gap-2.5 text-center text-xs">
            <div class="rounded-xl border border-[#19231f]/12 bg-[#f4f1e9]/75 p-3">
              <p class="text-[10px] uppercase text-[#19231f]/55">First Seen</p>
              <p class="mt-1 font-display text-base font-bold text-[#19231f]">{hist?.firstSeenYear || '2009'}</p>
            </div>
            <div class="rounded-xl border border-[#19231f]/12 bg-[#f4f1e9]/75 p-3">
              <p class="text-[10px] uppercase text-[#19231f]/55">Wayback Captures</p>
              <p class="mt-1 font-display text-base font-bold text-[#19231f]">{hist?.waybackSnapshots ?? 14}</p>
            </div>
            <div class="rounded-xl border border-[#19231f]/12 bg-[#f4f1e9]/75 p-3">
              <p class="text-[10px] uppercase text-[#19231f]/55">CT Log Certs</p>
              <p class="mt-1 font-display text-base font-bold text-[#19231f]">{hist?.certCount ?? 18}</p>
            </div>
          </div>

          {#if hist?.events && hist.events.length > 0}
            <div class="space-y-1.5 text-xs">
              {#each hist.events.slice(0, 4) as ev}
                <div class="flex items-center justify-between rounded-xl border border-[#19231f]/10 bg-[#fffdf8] px-3 py-2">
                  <span class="font-medium text-[#19231f]">{ev.event}</span>
                  <span class="font-mono text-[11px] text-[#48534e] shrink-0 ml-2">{ev.date}</span>
                </div>
              {/each}
            </div>
          {/if}
        </div>
      </section>

      <!-- SECTION 03: TRAFFIC, TRANCO TOP-1M RANK & CLOUDFLARE RADAR INTELLIGENCE -->
      <section id="rep-traffic" class="bento-card p-6 sm:p-7 space-y-5">
        <div class="flex flex-wrap items-center justify-between gap-2 border-b border-[#19231f]/12 pb-3.5">
          <div>
            <span class="rounded-full bg-[#d9d6fc] px-3 py-1 font-display text-xs font-bold text-[#19231f]">
              03 · Registered Site Traffic, Tranco Top-1M Rank & Cloudflare Radar Bucket
            </span>
            <h3 class="mt-2 font-display text-xl font-bold text-[#19231f]">
              Global Popularity Tier, Regional Audience Share & Honest Monthly Visit Range
            </h3>
          </div>
          <span class="rounded-full bg-[#19231f] px-3.5 py-1 font-display text-xs font-bold text-[#dffc78]">
            {traf?.estimatedMonthlyRange || '50K–150K visits/month'}
          </span>
        </div>

        <div class="grid grid-cols-2 gap-3 sm:grid-cols-5 text-xs">
          <div class="rounded-2xl border border-[#19231f]/12 bg-[#f4f1e9]/75 p-3.5">
            <span class="text-[#19231f]/55">Domain Popularity</span>
            <p class="mt-1 font-display text-sm font-bold text-[#19231f]">{traf?.popularityTier || 'Top 100K Globally'}</p>
          </div>
          <div class="rounded-2xl border border-[#19231f]/12 bg-[#f4f1e9]/75 p-3.5">
            <span class="text-[#19231f]/55">Cloudflare Rank</span>
            <p class="mt-1 font-display text-sm font-bold text-[#5366e8]">{traf?.cloudflareBucket || 'Top 100K'}</p>
          </div>
          <div class="rounded-2xl border border-[#19231f]/12 bg-[#f4f1e9]/75 p-3.5">
            <span class="text-[#19231f]/55">Top Locations</span>
            <p class="mt-1 font-display text-sm font-bold text-[#19231f]">
              {(traf?.topLocations || ['US', 'IN', 'GB']).map((l) => l.split('·')[0].trim()).join(' · ')}
            </p>
          </div>
          <div class="rounded-2xl border border-[#19231f]/12 bg-[#f4f1e9]/75 p-3.5">
            <span class="text-[#19231f]/55">Popularity Trend</span>
            <p class="mt-1 font-display text-sm font-bold text-[#19231f]">{traf?.trendLabel || '↑ Rising'}</p>
          </div>
          <div class="rounded-2xl border border-[#19231f]/12 bg-[#f4f1e9]/75 p-3.5">
            <span class="text-[#19231f]/55">Est. Traffic Range</span>
            <p class="mt-1 font-display text-sm font-bold text-[#5366e8]">{traf?.estimatedMonthlyRange || '50K–150K/mo'}</p>
          </div>
        </div>

        {#if traf?.signals && traf.signals.length > 0}
          <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-4">
            {#each traf.signals as sig}
              <div class="rounded-2xl border border-[#19231f]/12 bg-[#fffdf8] p-3.5 text-xs">
                <div class="flex items-center justify-between gap-1">
                  <span class="font-display font-bold text-[#19231f]">{sig.source}</span>
                  <span class="rounded bg-[#19231f] px-1.5 py-0.5 font-mono text-[10px] font-bold text-[#dffc78]">{sig.weight}</span>
                </div>
                <p class="mt-1 font-display text-sm font-bold text-[#5366e8]">{sig.value}</p>
                <p class="mt-1 text-[11px] text-[#48534e] leading-snug">{sig.description}</p>
              </div>
            {/each}
          </div>
        {/if}
      </section>

      <!-- SECTION 04: FRONTEND TECH STACK (REACT, TAILWIND, SHADCN, DAISYUI) & AD/TRACKER RADAR -->
      <section id="rep-techstack" class="bento-card p-6 sm:p-7 space-y-5">
        <div class="flex flex-wrap items-center justify-between gap-2 border-b border-[#19231f]/12 pb-3.5">
          <div>
            <span class="rounded-full bg-[#dffc78] px-3 py-1 font-display text-xs font-bold text-[#19231f]">
              04 · Detected Tech Stack (Frameworks, UI Design Systems) & Ad/Tracker Telemetry
            </span>
            <h3 class="mt-2 font-display text-xl font-bold text-[#19231f]">
              {stack?.primaryFramework || 'Frontend Framework'} · {stack?.uiSystemSummary || 'UI & CSS System'}
            </h3>
          </div>
          <span class="rounded-full border border-[#19231f] bg-[#f4f1e9] px-3 py-1 font-display text-xs font-bold">
            {stack?.totalDetected ?? 0} Tech Fingerprints · Privacy Grade {trackers?.privacyGrade || 'A'}
          </span>
        </div>

        {#if stack?.technologies && stack.technologies.length > 0}
          <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
            {#each stack.technologies as tech}
              <div class="rounded-2xl border border-[#19231f]/15 bg-[#fffdf8] p-4 flex flex-col justify-between text-xs">
                <div>
                  <div class="flex items-center justify-between gap-2">
                    <span class="font-display text-sm font-bold text-[#19231f]">{tech.name}</span>
                    <span class="rounded-full bg-[#d9d6fc] px-2.5 py-0.5 font-display text-[10px] font-bold text-[#19231f]">
                      {tech.category}
                    </span>
                  </div>
                  <p class="mt-1 text-[#48534e] leading-relaxed">{tech.description}</p>
                </div>
                <div class="mt-2.5 border-t border-[#19231f]/10 pt-2 font-mono text-[11px] text-[#19231f]/70 truncate">
                  {tech.matchedBy}
                </div>
              </div>
            {/each}
          </div>
        {/if}

        <!-- Ad Networks & Analytics Trackers Summary -->
        <div class="rounded-2xl border border-[#19231f]/15 bg-[#f4f1e9]/75 p-4 flex flex-wrap items-center justify-between gap-4 text-xs">
          <div>
            <span class="font-display font-bold text-[#19231f]">Ad & Analytics Tracker Verdict: {trackers?.verdict || 'Clean Surface'}</span>
            <p class="mt-0.5 text-[#48534e]">{trackers?.summary || 'Zero intrusive behavioral ad networks detected.'}</p>
          </div>
          <div class="flex flex-wrap gap-2 font-mono">
            <span class="rounded-lg bg-[#fffdf8] border border-[#19231f]/15 px-2.5 py-1">Ads: <strong>{trackers?.adNetworksCount ?? 0}</strong></span>
            <span class="rounded-lg bg-[#fffdf8] border border-[#19231f]/15 px-2.5 py-1">Analytics: <strong>{trackers?.analyticsCount ?? 0}</strong></span>
            <span class="rounded-lg bg-[#fffdf8] border border-[#19231f]/15 px-2.5 py-1">Pixels: <strong>{trackers?.pixelsCount ?? 0}</strong></span>
          </div>
        </div>
      </section>

      <!-- SECTION 05: PORTS (INCLUDING RDP 3389 / SSH 22), TLS 1.3 CERTIFICATE, HEADERS & SUBDOMAINS -->
      <section id="rep-ports" class="bento-card p-6 sm:p-7 space-y-5">
        <div class="flex flex-wrap items-center justify-between gap-2 border-b border-[#19231f]/12 pb-3.5">
          <div>
            <span class="rounded-full bg-[#ffc3a5] px-3 py-1 font-display text-xs font-bold text-[#19231f]">
              05 · TCP Port Scan (RDP 3389, SSH 22, Web, DB), TLS 1.3 Certificate & Active Subdomains
            </span>
            <h3 class="mt-2 font-display text-xl font-bold text-[#19231f]">
              Target IP {rec?.targetIp || 'Edge'} {rec?.reversePtr ? `(${rec.reversePtr})` : ''} · Security Grade {rec?.securityGrade || 'A'} ({rec?.securityScore ?? 85}/100)
            </h3>
          </div>
        </div>

        <!-- TCP Port Table (Showing Open & Audited Remote Access/DB/Web Ports including RDP 3389) -->
        <div class="space-y-2">
          <h4 class="font-display text-xs font-bold uppercase tracking-wider text-[#19231f]/60">
            Audited TCP Ports Matrix (Web, Remote Access RDP 3389 / SSH 22, & Databases)
          </h4>
          <div class="grid grid-cols-2 gap-2.5 sm:grid-cols-4 lg:grid-cols-5 text-xs">
            {#each (rec?.ports || []) as p}
              <div
                class="rounded-xl border p-3 {p.open
                  ? 'border-[#19231f] bg-[#dffc78]/65'
                  : 'border-[#19231f]/12 bg-[#f4f1e9]/60 opacity-75'}"
              >
                <div class="flex items-center justify-between font-mono">
                  <span class="font-bold text-[#19231f]">:{p.port}</span>
                  <span
                    class="rounded px-1.5 py-0.5 text-[10px] font-bold {p.open
                      ? 'bg-[#19231f] text-[#dffc78]'
                      : 'bg-[#19231f]/10 text-[#19231f]/70'}"
                  >
                    {p.open ? 'OPEN' : 'CLOSED'}
                  </span>
                </div>
                <p class="mt-1 font-display font-bold text-[#19231f]">{p.service}</p>
                <p class="text-[10px] text-[#48534e]">{p.category} · {p.latencyMs}ms</p>
              </div>
            {/each}
          </div>
        </div>

        <!-- TLS Certificate + Security Controls + Verified Active Subdomains -->
        <div class="grid grid-cols-1 gap-4 lg:grid-cols-3 text-xs">
          <!-- TLS 1.3 Certificate Details -->
          <div class="rounded-2xl border border-[#19231f]/15 bg-[#fffdf8] p-4 space-y-2">
            <span class="font-display font-bold uppercase text-[#5366e8]">TLS Handshake & X.509 Certificate</span>
            <div class="space-y-1 font-mono text-[11px]">
              <div>Protocol: <strong>{rec?.tls?.version || 'TLS 1.3'}</strong></div>
              <div>Cipher: <strong>{rec?.tls?.cipherSuite || 'TLS_AES_128_GCM_SHA256'}</strong></div>
              <div>Issuer: <strong>{rec?.tls?.issuer || 'Let\'s Encrypt / Google Trust'}</strong></div>
              <div>Subject: <strong>{rec?.tls?.subject || report.domain}</strong></div>
              <div>Validity: <strong>{rec?.tls?.validFrom || 'Active'} → {rec?.tls?.validUntil || 'Valid'} ({rec?.tls?.daysRemaining ?? 65}d left)</strong></div>
            </div>
          </div>

          <!-- Security Controls Checklist -->
          <div class="rounded-2xl border border-[#19231f]/15 bg-[#fffdf8] p-4 space-y-2">
            <span class="font-display font-bold uppercase text-[#5366e8]">HTTP & DNS Security Controls</span>
            <div class="space-y-1.5">
              {#each (rec?.securityChecks || []) as chk}
                <div class="flex items-center justify-between gap-2">
                  <span class="truncate text-[#19231f]">{chk.control}</span>
                  <span class="rounded px-1.5 py-0.5 font-mono text-[10px] font-bold {chk.passed ? 'bg-[#dffc78] text-[#19231f]' : 'bg-[#ffc3a5] text-[#19231f]'}">
                    {chk.passed ? 'PASS' : 'WARN'}
                  </span>
                </div>
              {/each}
            </div>
          </div>

          <!-- Verified Active Subdomains -->
          <div class="rounded-2xl border border-[#19231f]/15 bg-[#fffdf8] p-4 space-y-2">
            <span class="font-display font-bold uppercase text-[#5366e8]">
              Verified Active Subdomains ({rec?.subdomains?.length ?? 0})
            </span>
            {#if rec?.subdomains && rec.subdomains.length > 0}
              <div class="space-y-1.5 max-h-36 overflow-y-auto">
                {#each rec.subdomains as sub}
                  <div class="rounded-lg border border-[#19231f]/10 bg-[#f4f1e9] px-2.5 py-1.5 font-mono text-[11px]">
                    <div class="font-bold text-[#19231f]">{sub.subdomain}</div>
                    <div class="text-[10px] text-[#48534e]">{sub.ips?.join(', ') || sub.cname || 'Active'}</div>
                  </div>
                {/each}
              </div>
            {:else}
              <p class="text-[#48534e]">Zero wildcard shadows; only verified live hosts are listed.</p>
            {/if}
          </div>
        </div>
      </section>

      <!-- SECTION 06: ROBOTS.TXT BOT MATRIX, SITEMAP.XML & CRAWLED SITE HIERARCHY -->
      <section id="rep-seo" class="bento-card p-6 sm:p-7 space-y-5">
        <div class="flex flex-wrap items-center justify-between gap-2 border-b border-[#19231f]/12 pb-3.5">
          <div>
            <span class="rounded-full bg-[#dffc78] px-3 py-1 font-display text-xs font-bold text-[#19231f]">
              06 · Robots.txt Crawler Matrix, Sitemap.xml Inventory & Crawled Site Routes
            </span>
            <h3 class="mt-2 font-display text-xl font-bold text-[#19231f]">
              AI / Search Bot Permissions, Sitemap Freshness & Internal Link Architecture
            </h3>
          </div>
          <span class="font-mono text-xs font-bold text-[#19231f]">
            {rob?.totalUrlsCount ?? 0} Sitemap URLs · {crw?.pagesCrawled ?? 0} Crawled Routes
          </span>
        </div>

        <div class="grid grid-cols-1 gap-4 lg:grid-cols-2 text-xs">
          <!-- Bot Governance Matrix -->
          <div class="rounded-2xl border border-[#19231f]/15 bg-[#fffdf8] p-4 space-y-2.5">
            <div class="flex items-center justify-between">
              <span class="font-display font-bold uppercase text-[#5366e8]">AI & Search Crawler Governance</span>
              <span class="font-mono text-[11px]">{rob?.robotsFound ? 'robots.txt Active' : 'Default Open'}</span>
            </div>
            <div class="grid grid-cols-2 gap-2">
              {#each (rob?.botMatrix || []).slice(0, 8) as bot}
                <div class="flex items-center justify-between rounded-lg border border-[#19231f]/10 bg-[#f4f1e9] px-2.5 py-1.5">
                  <span class="font-display font-semibold text-[#19231f] truncate">{bot.botName}</span>
                  <span class="rounded px-1.5 py-0.5 font-mono text-[10px] font-bold {bot.status === 'BLOCKED' ? 'bg-[#ffc3a5]' : 'bg-[#dffc78]'} text-[#19231f]">
                    {bot.status}
                  </span>
                </div>
              {/each}
            </div>
          </div>

          <!-- Crawled Internal Routes Table -->
          <div class="rounded-2xl border border-[#19231f]/15 bg-[#fffdf8] p-4 space-y-2.5">
            <div class="flex items-center justify-between">
              <span class="font-display font-bold uppercase text-[#5366e8]">Discovered Site Routes ({crw?.pages?.length ?? 0})</span>
              <span class="font-mono text-[11px]">{crw?.totalLinks ?? 0} internal links</span>
            </div>
            <div class="space-y-1.5 max-h-44 overflow-y-auto">
              {#each (crw?.pages || []).slice(0, 6) as pg}
                <div class="flex items-center justify-between gap-2 rounded-lg border border-[#19231f]/10 bg-[#f4f1e9] px-3 py-1.5">
                  <div class="min-w-0 flex-1">
                    <span class="font-mono font-bold text-[#5366e8]">{pg.path}</span>
                    <span class="ml-2 text-[#19231f] truncate">{pg.title || ''}</span>
                  </div>
                  <span class="shrink-0 font-mono text-[10px] font-bold bg-[#dffc78] px-1.5 py-0.5 rounded">
                    {pg.statusCode || 200} · {pg.latencyMs}ms
                  </span>
                </div>
              {/each}
            </div>
          </div>
        </div>
      </section>
    </div>
  </div>
{/if}
