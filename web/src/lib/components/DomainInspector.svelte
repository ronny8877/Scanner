<script lang="ts">
  import type { DomainInquiry } from '../types';

  interface Props {
    inquiry: DomainInquiry | null;
    loading: boolean;
    onInspect: (domain: string) => void;
    onCrawlDomain: (domain: string) => void;
  }

  let { inquiry, loading, onInspect, onCrawlDomain }: Props = $props();
  let domainInput = $state('svelte.dev');

  const quickExamples = ['svelte.dev', 'golang.org', 'linear.app', 'veltrixhq.ai'];

  $effect(() => {
    if (inquiry?.domain) {
      domainInput = inquiry.domain;
    }
  });

  function handleSubmit(e: Event) {
    e.preventDefault();
    if (domainInput.trim()) {
      onInspect(domainInput.trim());
    }
  }

  function formatDate(iso?: string) {
    if (!iso) return 'Unpublished by registry';
    try {
      const d = new Date(iso);
      return d.toLocaleDateString('en-US', {
        year: 'numeric',
        month: 'short',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
      });
    } catch {
      return iso;
    }
  }
</script>

<section aria-labelledby="rdap-dossier-heading" class="space-y-6">
  <!-- Top Inquiry Bar Card -->
  <form
    onsubmit={handleSubmit}
    class="bento-card p-6 sm:p-7 flex flex-col lg:flex-row lg:items-end justify-between gap-5"
  >
    <div class="flex-1 space-y-3">
      <div class="flex items-center justify-between">
        <h2 id="rdap-dossier-heading" class="studio-label">
          02 // Authoritative RDAP Registration & DNS Dossier
        </h2>
        <span class="text-xs font-mono text-[#48534e]">ICANN Registry JSON Protocol + Live Resolver</span>
      </div>

      <div class="flex flex-col sm:flex-row gap-3">
        <input
          id="inspect-domain"
          type="text"
          bind:value={domainInput}
          placeholder="Enter domain name (e.g. svelte.dev, golang.org)..."
          class="flex-1 rounded-2xl bg-[#f4f1e9] border-[1.5px] border-[#19231f]/20 focus:border-[#19231f] px-4 py-3.5 text-base font-mono text-[#19231f] placeholder-[#6d7873]"
        />
        <button
          type="submit"
          disabled={loading}
          class="studio-btn-primary px-7 py-3.5 text-sm cursor-pointer disabled:opacity-50 shrink-0"
        >
          {loading ? 'Querying Registry…' : 'Inspect Dossier →'}
        </button>
      </div>

      <div class="flex flex-wrap items-center gap-2 pt-0.5">
        <span class="text-xs text-[#6d7873]">Inspect real domains:</span>
        {#each quickExamples as ex}
          <button
            type="button"
            onclick={() => {
              domainInput = ex;
              onInspect(ex);
            }}
            class="px-3 py-1 rounded-full text-xs font-mono bg-[#f4f1e9] hover:bg-[#d9d6fc] text-[#19231f] border border-[#19231f]/15 transition-colors cursor-pointer"
          >
            {ex}
          </button>
        {/each}
      </div>
    </div>
  </form>

  {#if inquiry}
    <!-- Asymmetric Dossier Header Bento: 8-Col Identity Certificate + 4-Col Valuation Stamp -->
    <div class="grid grid-cols-1 lg:grid-cols-12 gap-6 items-stretch">
      <div
        class="lg:col-span-8 bento-card p-6 sm:p-7 flex flex-col justify-between gap-4 {inquiry.available
          ? 'bg-[#dffc78]/35'
          : 'bg-[#fffdf8]'}"
      >
        <div class="flex flex-wrap items-center justify-between gap-3">
          <span class="studio-label">Registry Certificate</span>
          <span class="text-xs font-mono text-[#48534e]">
            Resolved in {inquiry.checkLatencyMs}ms · {inquiry.registryHandle || 'Authoritative RDAP'}
          </span>
        </div>

        <div class="flex flex-wrap items-center gap-4 my-1">
          <h3 class="text-4xl sm:text-5xl font-display font-bold tracking-tight text-[#19231f] break-all">
            {inquiry.domain}
          </h3>

          {#if inquiry.available}
            <span
              class="px-4 py-1.5 rounded-full text-xs font-display font-bold uppercase bg-[#dffc78] text-[#19231f] border-[1.5px] border-[#19231f]"
            >
              ★ Unclaimed & Available
            </span>
          {:else}
            <span
              class="px-4 py-1.5 rounded-full text-xs font-display font-bold uppercase bg-[#ffc3a5] text-[#19231f] border-[1.5px] border-[#19231f]"
            >
              ● Registered Record
            </span>
          {/if}
        </div>

        <div class="pt-4 border-t border-[#19231f]/10 flex flex-wrap items-center justify-between gap-3 text-xs">
          {#if !inquiry.available && inquiry.domainAge}
            <p class="text-[#19231f] text-sm">
              Continuous registration tenure of
              <span class="font-serif-editorial text-xl px-1">{inquiry.domainAge}</span>
              via <strong>{inquiry.registrar || 'Registry Direct'}</strong>.
            </p>
          {:else if inquiry.available}
            <p class="text-[#19231f] text-sm">
              No active NS delegation or registry lock detected — open for immediate registration.
            </p>
          {:else}
            <p class="text-[#48534e]">Active DNS delegation verified across global nameservers.</p>
          {/if}

          {#if !inquiry.available}
            <button
              type="button"
              onclick={() => onCrawlDomain(inquiry.domain)}
              class="studio-btn-ink px-4 py-2 text-xs cursor-pointer"
            >
              Map Site Architecture →
            </button>
          {/if}
        </div>
      </div>

      <!-- 4-Column Aftermarket Stamp Card -->
      <div class="lg:col-span-4 bento-card p-6 sm:p-7 flex flex-col justify-between bg-[#d9d6fc]/55">
        <div class="flex items-center justify-between">
          <span class="studio-label text-[#19231f]">Valuation Appraisal</span>
          <span class="px-2.5 py-0.5 rounded-full text-xs font-mono bg-[#fffdf8] border border-[#19231f]">
            {inquiry.valuation.score}/100
          </span>
        </div>

        <div class="my-3">
          <div class="font-serif-editorial text-2xl text-[#19231f]">{inquiry.valuation.tier}</div>
          <div class="text-3xl font-display font-bold text-[#19231f] mt-0.5">
            {inquiry.valuation.estimatedDisplay}
          </div>
        </div>

        <ul class="space-y-1 text-xs text-[#19231f] pt-3 border-t border-[#19231f]/15">
          {#each inquiry.valuation.highlights as h}
            <li class="flex items-center gap-1.5">
              <span class="font-bold">·</span>
              <span>{h}</span>
            </li>
          {/each}
        </ul>
      </div>
    </div>

    <!-- Main Asymmetric Split: 6-Col Registration Chronology + 6-Col DNS Blueprint -->
    <div class="grid grid-cols-1 lg:grid-cols-12 gap-6">
      <!-- Registration Timeline & EPP Statuses -->
      <div class="lg:col-span-6 bento-card p-6 sm:p-7 space-y-5">
        <div class="flex items-center justify-between border-b border-[#19231f]/10 pb-3.5">
          <h4 class="font-display font-bold text-base text-[#19231f]">
            Registration Chronology & Custody
          </h4>
          <span class="studio-label">WHOIS / RDAP</span>
        </div>

        {#if inquiry.available}
          <div class="rounded-2xl bg-[#dffc78]/45 border border-[#19231f] p-5 space-y-2">
            <div class="font-display font-bold text-base text-[#19231f]">
              Unoccupied in Authoritative Registry
            </div>
            <p class="text-xs text-[#19231f]/80 leading-relaxed">
              This domain has zero active `NS` or `A` records and returned HTTP 404 from the ICANN RDAP gateway,
              confirming it is empty and claimable.
            </p>
          </div>
        {:else}
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div class="rounded-2xl bg-[#f4f1e9] border border-[#19231f]/12 p-4">
              <div class="studio-label">Registered On (Created)</div>
              <div class="mt-1.5 font-mono text-sm font-semibold text-[#19231f]">
                {formatDate(inquiry.registeredAt)}
              </div>
              {#if inquiry.domainAge}
                <div class="mt-2 inline-block px-2.5 py-0.5 rounded-full bg-[#fffdf8] border border-[#19231f]/20 text-xs font-mono">
                  Tenure: {inquiry.domainAge}
                </div>
              {/if}
            </div>

            <div class="rounded-2xl bg-[#f4f1e9] border border-[#19231f]/12 p-4">
              <div class="studio-label">Registry Expiration</div>
              <div class="mt-1.5 font-mono text-sm font-semibold text-[#19231f]">
                {formatDate(inquiry.expiresAt)}
              </div>
              {#if inquiry.daysToExpiry}
                <div class="mt-2 inline-block px-2.5 py-0.5 rounded-full bg-[#ffc3a5] border border-[#19231f]/30 text-xs font-mono font-semibold">
                  {inquiry.daysToExpiry} days remaining
                </div>
              {/if}
            </div>

            <div class="rounded-2xl bg-[#f4f1e9] border border-[#19231f]/12 p-4">
              <div class="studio-label">Sponsoring Registrar</div>
              <div class="mt-1.5 font-display text-sm font-bold text-[#19231f]">
                {inquiry.registrar || 'Registry Direct / Protected'}
              </div>
            </div>

            <div class="rounded-2xl bg-[#f4f1e9] border border-[#19231f]/12 p-4">
              <div class="studio-label">Last Registry Update</div>
              <div class="mt-1.5 font-mono text-sm text-[#19231f]">
                {formatDate(inquiry.updatedAt)}
              </div>
            </div>
          </div>

          {#if inquiry.statusFlags && inquiry.statusFlags.length > 0}
            <div class="pt-2">
              <div class="studio-label mb-2">EPP Security & Transfer Locks</div>
              <div class="flex flex-wrap gap-2">
                {#each inquiry.statusFlags as flag}
                  <span class="px-3 py-1 rounded-full bg-[#f4f1e9] border border-[#19231f]/20 text-xs font-mono text-[#19231f]">
                    {flag}
                  </span>
                {/each}
              </div>
            </div>
          {/if}
        {/if}
      </div>

      <!-- DNS Infrastructure Blueprint Card -->
      <div class="lg:col-span-6 bento-card p-6 sm:p-7 space-y-5">
        <div class="flex items-center justify-between border-b border-[#19231f]/10 pb-3.5">
          <h4 class="font-display font-bold text-base text-[#19231f]">
            DNS Zone Blueprint & Routing
          </h4>
          <span class="studio-label">Live Records</span>
        </div>

        <div class="space-y-3.5 font-mono text-xs">
          <div class="rounded-2xl bg-[#f4f1e9] border border-[#19231f]/12 p-4">
            <div class="studio-label mb-2">Authoritative Nameservers (NS)</div>
            {#if inquiry.nameservers && inquiry.nameservers.length > 0}
              <div class="flex flex-wrap gap-2">
                {#each inquiry.nameservers as ns}
                  <span class="px-3 py-1 rounded-full bg-[#fffdf8] border border-[#19231f]/20 text-[#5366e8] font-medium">
                    {ns}
                  </span>
                {/each}
              </div>
            {:else}
              <span class="text-[#6d7873]">No NS delegation</span>
            {/if}
          </div>

          <div class="grid grid-cols-1 sm:grid-cols-2 gap-3.5">
            <div class="rounded-2xl bg-[#f4f1e9] border border-[#19231f]/12 p-4">
              <div class="studio-label mb-1.5">IPv4 (A)</div>
              {#if inquiry.dns?.a && inquiry.dns.a.length > 0}
                {#each inquiry.dns.a as ip}
                  <div class="text-[#19231f] font-semibold">{ip}</div>
                {/each}
              {:else}
                <span class="text-[#6d7873]">None</span>
              {/if}
            </div>

            <div class="rounded-2xl bg-[#f4f1e9] border border-[#19231f]/12 p-4">
              <div class="studio-label mb-1.5">IPv6 (AAAA)</div>
              {#if inquiry.dns?.aaaa && inquiry.dns.aaaa.length > 0}
                {#each inquiry.dns.aaaa as ip6}
                  <div class="text-[#19231f] truncate">{ip6}</div>
                {/each}
              {:else}
                <span class="text-[#6d7873]">None</span>
              {/if}
            </div>
          </div>

          <div class="rounded-2xl bg-[#f4f1e9] border border-[#19231f]/12 p-4">
            <div class="studio-label mb-1.5">Mail Exchangers (MX)</div>
            {#if inquiry.dns?.mx && inquiry.dns.mx.length > 0}
              <div class="space-y-1">
                {#each inquiry.dns.mx as mx}
                  <div class="text-[#19231f]">{mx}</div>
                {/each}
              </div>
            {:else}
              <span class="text-[#6d7873]">No MX records</span>
            {/if}
          </div>

          {#if inquiry.dns?.txt && inquiry.dns.txt.length > 0}
            <div class="rounded-2xl bg-[#f4f1e9] border border-[#19231f]/12 p-4">
              <div class="studio-label mb-1.5">TXT Verification & SPF Records</div>
              <div class="space-y-1.5">
                {#each inquiry.dns.txt as txt}
                  <div class="bg-[#fffdf8] px-3 py-1.5 rounded-lg border border-[#19231f]/10 text-[#19231f] break-all">
                    {txt}
                  </div>
                {/each}
              </div>
            </div>
          {/if}
        </div>
      </div>
    </div>
  {/if}
</section>
