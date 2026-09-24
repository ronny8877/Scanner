<script lang="ts">
  import type { HistoryReport } from '../types';
  import { motionCard } from '../motion';
  import LoadingProgressBanner from './LoadingProgressBanner.svelte';

  interface Props {
    open: boolean;
    loading: boolean;
    targetDomain: string;
    report: HistoryReport | null;
    onClose: () => void;
    onInspectRDAP: (domain: string) => void;
    onRunRecon: (domain: string) => void;
  }

  let { open, loading, targetDomain, report, onClose, onInspectRDAP, onRunRecon }: Props = $props();
</script>

{#if open}
  <div
    class="fixed inset-0 z-50 bg-[#19231f]/55 backdrop-blur-sm flex items-center justify-center p-4 sm:p-6"
    role="dialog"
    aria-modal="true"
    aria-labelledby="history-modal-title"
  >
    <div
      use:motionCard={{ y: 18 }}
      class="w-full max-w-3xl bento-card bg-[#fffdf8] border-2 border-[#19231f] shadow-[0_12px_0_#19231f] overflow-hidden max-h-[90dvh] flex flex-col"
    >
      <!-- Header -->
      <div class="bg-[#19231f] text-[#fffdf8] px-6 py-4 flex items-center justify-between gap-4">
        <div class="flex items-center gap-2.5">
          <span class="w-2.5 h-2.5 rounded-full bg-[#dffc78]"></span>
          <span id="history-modal-title" class="font-display font-bold text-sm uppercase tracking-wider text-[#dffc78]">
            Past Registration & Archive History Dossier (Wayback + CT Logs)
          </span>
        </div>

        <button
          type="button"
          onclick={onClose}
          class="px-3 py-1 rounded-full text-xs font-mono bg-[#fffdf8]/15 hover:bg-[#dffc78] hover:text-[#19231f] text-[#fffdf8] transition-colors cursor-pointer"
        >
          Close [Esc]
        </button>
      </div>

      <!-- Body -->
      <div class="p-6 space-y-5 scroll-panel">
        {#if loading}
          <LoadingProgressBanner
            title="Scanning Historical Archives"
            target={targetDomain}
            workers={4}
            steps={[
              'Querying Wayback Machine CDX Index',
              'Scanning crt.sh Certificate Transparency Logs',
              'Correlating Past Subdomains & Timestamps',
            ]}
          />
        {:else if report}
          <!-- Verdict Card -->
          <div
            class="rounded-2xl border-[1.5px] border-[#19231f] p-5 flex flex-col sm:flex-row sm:items-center justify-between gap-4 {report.previouslyRegistered
              ? 'bg-[#ffc3a5]/45'
              : 'bg-[#dffc78]/45'}"
          >
            <div class="space-y-1">
              <div class="flex items-center gap-3 flex-wrap">
                <h3 class="text-2xl sm:text-3xl font-display font-bold text-[#19231f] font-mono">
                  {report.domain}
                </h3>
                {#if report.previouslyRegistered}
                  <span class="px-3 py-1 rounded-full text-xs font-display font-bold uppercase bg-[#ffc3a5] border border-[#19231f] text-[#19231f]">
                    ● Registered in the Past ({report.firstSeenYear}–{report.lastSeenYear})
                  </span>
                {:else}
                  <span class="px-3 py-1 rounded-full text-xs font-display font-bold uppercase bg-[#dffc78] border border-[#19231f] text-[#19231f]">
                    ★ Clean Virgin History
                  </span>
                {/if}
              </div>
              <p class="text-xs text-[#19231f]/85 leading-relaxed">
                {report.summaryNote}
              </p>
            </div>

            <div class="text-right shrink-0 font-mono text-xs text-[#48534e]">
              Checked in {report.checkLatencyMs}ms
            </div>
          </div>

          {#if report.previouslyRegistered}
            <!-- Historical Metrics Grid -->
            <div class="grid grid-cols-2 sm:grid-cols-4 gap-3.5">
              <div class="rounded-2xl bg-[#f4f1e9] border border-[#19231f]/15 p-4">
                <div class="studio-label">First Seen</div>
                <div class="mt-1 font-mono text-base font-bold text-[#19231f]">
                  {report.firstSeenAt || 'N/A'}
                </div>
              </div>

              <div class="rounded-2xl bg-[#f4f1e9] border border-[#19231f]/15 p-4">
                <div class="studio-label">Last Seen</div>
                <div class="mt-1 font-mono text-base font-bold text-[#19231f]">
                  {report.lastSeenAt || 'N/A'}
                </div>
              </div>

              <div class="rounded-2xl bg-[#f4f1e9] border border-[#19231f]/15 p-4">
                <div class="studio-label">Wayback Captures</div>
                <div class="mt-1 font-display text-xl font-bold text-[#5366e8]">
                  {report.waybackSnapshots}
                </div>
              </div>

              <div class="rounded-2xl bg-[#f4f1e9] border border-[#19231f]/15 p-4">
                <div class="studio-label">Past TLS Certs</div>
                <div class="mt-1 font-display text-xl font-bold text-[#19231f]">
                  {report.certCount}
                </div>
              </div>
            </div>

            <!-- Active Archive Years Timeline -->
            {#if report.activeYears && report.activeYears.length > 0}
              <div class="rounded-2xl bg-[#f4f1e9] border border-[#19231f]/15 p-4 space-y-2">
                <div class="studio-label">Years With Confirmed Web Hosting Activity</div>
                <div class="flex flex-wrap gap-1.5">
                  {#each report.activeYears as yr}
                    <span class="px-3 py-1 rounded-full bg-[#fffdf8] border border-[#19231f] text-xs font-mono font-bold text-[#19231f]">
                      {yr}
                    </span>
                  {/each}
                </div>
              </div>
            {/if}

            <!-- Milestones & Past Subdomains -->
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
              {#if report.milestones && report.milestones.length > 0}
                <div class="rounded-2xl bg-[#f4f1e9] border border-[#19231f]/15 p-4 space-y-2.5">
                  <div class="studio-label">Historical Timeline Milestones</div>
                  <div class="space-y-2">
                    {#each report.milestones as ms}
                      <div class="rounded-xl bg-[#fffdf8] border border-[#19231f]/10 p-2.5 text-xs">
                        <div class="flex items-center justify-between font-mono text-[11px] text-[#5366e8] font-semibold">
                          <span>{ms.date}</span>
                          <span>{ms.source}</span>
                        </div>
                        <div class="mt-0.5 text-[#19231f]">{ms.event}</div>
                      </div>
                    {/each}
                  </div>
                </div>
              {/if}

              <div class="rounded-2xl bg-[#f4f1e9] border border-[#19231f]/15 p-4 space-y-3">
                <div>
                  <div class="studio-label mb-1.5">Past Subdomains Found in CT Logs</div>
                  {#if report.pastSubdomains && report.pastSubdomains.length > 0}
                    <div class="flex flex-wrap gap-1.5">
                      {#each report.pastSubdomains as sub}
                        <span class="px-2.5 py-0.5 rounded-full bg-[#fffdf8] border border-[#19231f]/20 text-xs font-mono text-[#19231f]">
                          {sub}
                        </span>
                      {/each}
                    </div>
                  {:else}
                    <span class="text-xs text-[#6d7873]">No historical subdomains recorded</span>
                  {/if}
                </div>

                {#if report.pastIssuers && report.pastIssuers.length > 0}
                  <div class="pt-2 border-t border-[#19231f]/10">
                    <div class="studio-label mb-1.5">Historical Certificate Authorities</div>
                    <div class="flex flex-wrap gap-1.5">
                      {#each report.pastIssuers as iss}
                        <span class="px-2.5 py-0.5 rounded-full bg-[#d9d6fc] border border-[#19231f]/20 text-xs font-mono text-[#19231f]">
                          {iss}
                        </span>
                      {/each}
                    </div>
                  </div>
                {/if}
              </div>
            </div>
          {/if}

          <!-- Footer Actions -->
          <div class="pt-3 border-t border-[#19231f]/12 flex flex-wrap items-center justify-between gap-3">
            <div class="flex flex-wrap gap-2">
              <button
                type="button"
                onclick={() => {
                  onClose();
                  onInspectRDAP(report.domain);
                }}
                class="studio-btn-primary px-4 py-2 text-xs cursor-pointer"
              >
                Open Live RDAP Dossier →
              </button>
              <button
                type="button"
                onclick={() => {
                  onClose();
                  onRunRecon(report.domain);
                }}
                class="studio-btn-ink px-4 py-2 text-xs cursor-pointer"
              >
                Scan TCP Ports & TLS →
              </button>
            </div>

            <button
              type="button"
              onclick={onClose}
              class="px-4 py-2 rounded-full text-xs font-display font-semibold bg-[#f4f1e9] hover:bg-[#19231f] hover:text-[#fffdf8] border border-[#19231f]/20 transition-colors cursor-pointer"
            >
              Done
            </button>
          </div>
        {/if}
      </div>
    </div>
  </div>
{/if}
