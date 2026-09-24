<script lang="ts">
  import type { ReconReport } from '../types';
  import { motionCard } from '../motion';
  import LoadingProgressBanner from './LoadingProgressBanner.svelte';

  interface Props {
    report: ReconReport | null;
    loading: boolean;
    onRunRecon: (domain: string) => void;
    onRunFullSuite: (domain: string) => void;
    onCheckHistory: (domain: string) => void;
    onSaveDomain: (domain: string, available: boolean) => void;
    savedDomainsSet: Set<string>;
  }

  let {
    report,
    loading,
    onRunRecon,
    onRunFullSuite,
    onCheckHistory,
    onSaveDomain,
    savedDomainsSet,
  }: Props = $props();

  let targetInput = $state('svelte.dev');

  const quickTargets = ['svelte.dev', 'golang.org', 'cloudflare.com', 'example.com'];

  $effect(() => {
    if (report?.domain) {
      targetInput = report.domain;
    }
  });

  function handleSubmit(e: Event) {
    e.preventDefault();
    if (targetInput.trim()) {
      onRunRecon(targetInput.trim());
    }
  }
</script>

<section aria-labelledby="port-recon-heading" class="space-y-6">
  <!-- Command Header Card -->
  <form
    use:motionCard
    onsubmit={handleSubmit}
    class="bento-card p-6 sm:p-7 flex flex-col lg:flex-row lg:items-end justify-between gap-5"
  >
    <div class="flex-1 space-y-3">
      <div class="flex flex-wrap items-center justify-between gap-2">
        <h2 id="port-recon-heading" class="studio-label">
          04 // Parallel Port Scanner, TLS Handshake & Security Surface Audit
        </h2>
        <span class="text-xs font-mono text-[#48534e]">
          28 Concurrent Goroutines · 14 TCP Ports + TLS 1.3 + Subdomains + DMARC/SPF
        </span>
      </div>

      <div class="flex flex-col sm:flex-row gap-3">
        <input
          type="text"
          bind:value={targetInput}
          placeholder="Enter domain or host (e.g. svelte.dev, golang.org)..."
          class="flex-1 rounded-2xl bg-[#f4f1e9] border-[1.5px] border-[#19231f]/20 focus:border-[#19231f] px-4 py-3.5 text-base font-mono text-[#19231f] placeholder-[#6d7873]"
        />

        <div class="flex flex-wrap gap-2">
          <button
            type="submit"
            disabled={loading}
            class="studio-btn-primary px-6 py-3.5 text-sm cursor-pointer disabled:opacity-50"
          >
            {loading ? 'Probing Surface…' : 'Scan Ports & TLS →'}
          </button>

          <button
            type="button"
            disabled={loading}
            onclick={() => targetInput.trim() && onRunFullSuite(targetInput.trim())}
            class="studio-btn-ink px-5 py-3.5 text-xs cursor-pointer disabled:opacity-50"
            title="Run RDAP + Past History + Port Scan + Site Crawl all in parallel"
          >
            ⚡ Run All 4 Engines in Parallel
          </button>
        </div>
      </div>

      <div class="flex flex-wrap items-center gap-2 pt-0.5">
        <span class="text-xs text-[#6d7873]">Quick infrastructure targets:</span>
        {#each quickTargets as t}
          <button
            type="button"
            onclick={() => {
              targetInput = t;
              onRunRecon(t);
            }}
            class="px-3 py-1 rounded-full text-xs font-mono bg-[#f4f1e9] hover:bg-[#dffc78] text-[#19231f] border border-[#19231f]/15 transition-colors cursor-pointer"
          >
            {t}
          </button>
        {/each}
      </div>
    </div>
  </form>

  {#if loading}
    <LoadingProgressBanner
      title="Parallel Port, TLS & Subdomain Recon"
      target={targetInput}
      workers={28}
      steps={[
        'Dialing 14 TCP Ports in Parallel (HTTP, SSH, DBs)',
        'Inspecting Port 443 TLS Handshake & SAN Certs',
        'Enumerating Subdomains & Auditing SPF/DMARC/HSTS',
      ]}
    />
  {:else if report}
    <!-- Top Asymmetric Summary Bento: 8-Col Host Surface + 4-Col Security Grade Stamp -->
    <div class="grid grid-cols-1 lg:grid-cols-12 gap-6 items-stretch">
      <div use:motionCard={{ delay: 0.04 }} class="lg:col-span-8 bento-card p-6 sm:p-7 flex flex-col justify-between gap-4">
        <div class="flex flex-wrap items-center justify-between gap-2">
          <span class="studio-label">Infrastructure Surface Summary</span>
          <span class="text-xs font-mono text-[#48534e]">
            Primary IP: <strong class="text-[#19231f]">{report.targetIp || 'Resolved'}</strong> · {report.durationMs}ms
          </span>
        </div>

        <div class="flex flex-wrap items-center justify-between gap-4 my-1">
          <div>
            <h3 class="text-3xl sm:text-4xl font-display font-bold text-[#19231f] font-mono">
              {report.domain}
            </h3>
            <p class="text-xs text-[#48534e] mt-1">
              <strong class="text-[#19231f]">{report.openPortsCount} open TCP ports</strong> detected out of {report.portsScanned} scanned ·
              <strong class="text-[#19231f]">{report.subdomains.length} active subdomains</strong> resolved
            </p>
          </div>

          <div class="flex flex-wrap items-center gap-2">
            <button
              type="button"
              onclick={() => onCheckHistory(report.domain)}
              class="px-3.5 py-2 rounded-full text-xs font-display font-semibold bg-[#f4f1e9] hover:bg-[#ffc3a5] text-[#19231f] border border-[#19231f] transition-colors cursor-pointer"
            >
              ⏳ Check Past Registration History
            </button>

            <button
              type="button"
              onclick={() => onSaveDomain(report.domain, false)}
              class="px-3.5 py-2 rounded-full text-xs font-display font-semibold border border-[#19231f] transition-colors cursor-pointer {savedDomainsSet.has(
                report.domain
              )
                ? 'bg-[#dffc78] text-[#19231f]'
                : 'bg-[#fffdf8] hover:bg-[#dffc78] text-[#19231f]'}"
            >
              {savedDomainsSet.has(report.domain) ? '★ Saved in Vault' : '☆ Save Domain'}
            </button>
          </div>
        </div>

        <!-- TLS Quick Strip -->
        {#if report.tls.supported}
          <div class="rounded-2xl bg-[#f4f1e9] border border-[#19231f]/15 px-4 py-3 flex flex-wrap items-center justify-between gap-3 text-xs font-mono">
            <div class="flex items-center gap-2">
              <span class="px-2.5 py-0.5 rounded-full bg-[#dffc78] border border-[#19231f] font-bold text-[#19231f]">
                {report.tls.version}
              </span>
              <span class="text-[#19231f] font-semibold">{report.tls.issuer}</span>
            </div>
            <div class="text-[#48534e]">
              Valid {report.tls.validFrom} → {report.tls.validUntil}
              <strong class="text-[#19231f]">({report.tls.daysRemaining}d left)</strong>
            </div>
          </div>
        {/if}
      </div>

      <!-- 4-Column Security Posture Grade Card -->
      <div
        use:motionCard={{ delay: 0.08 }}
        class="lg:col-span-4 bento-card p-6 sm:p-7 flex flex-col justify-between {report.securityScore >= 75
          ? 'bg-[#dffc78]/45'
          : 'bg-[#ffc3a5]/45'}"
      >
        <div class="flex items-center justify-between">
          <span class="studio-label text-[#19231f]">Security Posture Grade</span>
          <span class="px-2.5 py-0.5 rounded-full text-xs font-mono bg-[#fffdf8] border border-[#19231f]">
            {report.securityScore}/100
          </span>
        </div>

        <div class="my-2 flex items-baseline gap-3">
          <span class="text-6xl font-display font-bold text-[#19231f]">{report.securityGrade}</span>
          <span class="font-serif-editorial text-2xl text-[#19231f]">
            {report.securityScore >= 75 ? 'hardened perimeter' : 'headers missing'}
          </span>
        </div>

        <div class="space-y-1.5 pt-3 border-t border-[#19231f]/15 text-xs">
          {#each report.securityChecks as chk}
            <div class="flex items-center justify-between gap-2">
              <span class="truncate text-[#19231f]">{chk.control}</span>
              <span
                class="font-mono font-bold px-2 py-0.5 rounded-full text-[10px] border {chk.passed
                  ? 'bg-[#fffdf8] text-[#19231f] border-[#19231f]'
                  : 'bg-[#ffc3a5] text-[#19231f] border-[#19231f]'}"
              >
                {chk.passed ? 'PASS' : 'WARN'}
              </span>
            </div>
          {/each}
        </div>
      </div>
    </div>

    <!-- Main Asymmetric Grid: 7-Col Parallel Port Matrix + 5-Col Subdomains & TLS Certificate -->
    <div class="grid grid-cols-1 lg:grid-cols-12 gap-6 items-start">
      <!-- 7-Col Parallel TCP Port Matrix -->
      <div use:motionCard={{ delay: 0.1 }} class="lg:col-span-7 bento-card overflow-hidden">
        <div class="px-6 py-4 border-b border-[#19231f]/10 flex items-center justify-between">
          <div>
            <h4 class="font-display font-bold text-base text-[#19231f]">
              Parallel TCP Port Matrix ({report.ports.length} ports probed)
            </h4>
            <p class="text-xs text-[#48534e]">Open listeners sorted to top with TCP SYN/ACK handshake latency</p>
          </div>
          <span class="px-3 py-1 rounded-full text-xs font-mono bg-[#dffc78] border border-[#19231f] font-bold">
            {report.openPortsCount} Open
          </span>
        </div>

        <div class="divide-y divide-[#19231f]/10">
          {#each report.ports as p (p.port)}
            <div
              class="px-6 py-3.5 flex items-center justify-between gap-4 {p.open
                ? 'bg-[#dffc78]/20'
                : 'opacity-75 hover:opacity-100'}"
            >
              <div class="flex items-center gap-3.5">
                <span
                  class="w-16 font-mono text-sm font-bold px-2.5 py-1 rounded-xl text-center border {p.open
                    ? 'bg-[#19231f] text-[#dffc78] border-[#19231f]'
                    : 'bg-[#f4f1e9] text-[#48534e] border-[#19231f]/15'}"
                >
                  :{p.port}
                </span>

                <div>
                  <div class="flex items-center gap-2">
                    <span class="font-display font-bold text-sm text-[#19231f]">{p.service}</span>
                    <span class="text-[11px] font-mono px-2 py-0.5 rounded-full bg-[#f4f1e9] border border-[#19231f]/15 text-[#48534e]">
                      {p.category}
                    </span>
                  </div>
                  {#if p.riskNote}
                    <p class="text-xs text-[#48534e]">{p.riskNote}</p>
                  {/if}
                </div>
              </div>

              <div class="flex items-center gap-3 shrink-0 font-mono text-xs">
                <span class="text-[#48534e]">{p.latencyMs}ms</span>
                {#if p.open}
                  <span class="px-3 py-1 rounded-full bg-[#dffc78] border border-[#19231f] text-[#19231f] font-bold">
                    ● OPEN
                  </span>
                {:else}
                  <span class="px-3 py-1 rounded-full bg-[#f4f1e9] border border-[#19231f]/15 text-[#6d7873]">
                    CLOSED
                  </span>
                {/if}
              </div>
            </div>
          {/each}
        </div>
      </div>

      <!-- 5-Col Subdomains & TLS Certificate Blueprint -->
      <div use:motionCard={{ delay: 0.14 }} class="lg:col-span-5 space-y-6">
        <div class="bento-card p-6 space-y-4">
          <div class="flex items-center justify-between border-b border-[#19231f]/10 pb-3">
            <h4 class="font-display font-bold text-base text-[#19231f]">
              Active Subdomains ({report.subdomains.length})
            </h4>
            <span class="studio-label">Parallel DNS Enumeration</span>
          </div>

          {#if report.subdomains.length > 0}
            <div class="space-y-2">
              {#each report.subdomains as sub}
                <div class="rounded-xl bg-[#f4f1e9] border border-[#19231f]/15 px-4 py-2.5 flex items-center justify-between gap-2 font-mono text-xs">
                  <span class="font-semibold text-[#5366e8]">{sub.subdomain}</span>
                  <span class="text-[#19231f]">{sub.ips.join(', ')}</span>
                </div>
              {/each}
            </div>
          {:else}
            <p class="text-xs text-[#48534e]">No additional common subdomains resolved.</p>
          {/if}
        </div>

        {#if report.tls.supported}
          <div class="bento-card p-6 space-y-3 bg-[#d9d6fc]/40">
            <div class="flex items-center justify-between border-b border-[#19231f]/15 pb-3">
              <h4 class="font-display font-bold text-base text-[#19231f]">TLS Certificate SANs</h4>
              <span class="px-2.5 py-0.5 rounded-full text-xs font-mono bg-[#fffdf8] border border-[#19231f]">
                {report.tls.cipherSuite}
              </span>
            </div>

            <div class="text-xs space-y-1.5 font-mono">
              <div><span class="text-[#48534e]">Subject:</span> <strong>{report.tls.subject}</strong></div>
              <div><span class="text-[#48534e]">Issuer:</span> <strong>{report.tls.issuer}</strong></div>
            </div>

            {#if report.tls.sans && report.tls.sans.length > 0}
              <div class="pt-2">
                <div class="studio-label mb-1.5">Subject Alternative Names</div>
                <div class="flex flex-wrap gap-1.5">
                  {#each report.tls.sans as san}
                    <span class="px-2.5 py-0.5 rounded-full bg-[#fffdf8] border border-[#19231f]/20 text-xs font-mono text-[#19231f]">
                      {san}
                    </span>
                  {/each}
                </div>
              </div>
            {/if}
          </div>
        {/if}
      </div>
    </div>
  {/if}
</section>
