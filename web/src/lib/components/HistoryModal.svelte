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
    onCancelJob: () => void;
    onInspectRDAP: (domain: string) => void;
    onRunRecon: (domain: string) => void;
  }

  let { open, loading, targetDomain, report, onClose, onCancelJob, onInspectRDAP, onRunRecon }: Props = $props();
</script>

{#if open}
  <div
    class="fixed inset-0 z-50 bg-[#19231f]/55 backdrop-blur-sm flex items-center justify-center p-3 sm:p-6"
    role="dialog"
    aria-modal="true"
    aria-labelledby="history-modal-title"
  >
    <div
      use:motionCard={{ y: 18 }}
      class="w-full max-w-4xl bento-card bg-[#fffdf8] border-2 border-[#19231f] shadow-[0_12px_0_#19231f] overflow-hidden max-h-[92dvh] flex flex-col"
    >
      <!-- Header -->
      <div class="bg-[#19231f] text-[#fffdf8] px-6 py-4 flex flex-wrap items-center justify-between gap-3">
        <div class="flex items-center gap-2.5">
          <span class="w-2.5 h-2.5 rounded-full bg-[#dffc78]"></span>
          <span id="history-modal-title" class="font-display font-bold text-sm uppercase tracking-wider text-[#dffc78]">
            Past Registration & Archive History (Wayback + RDAP + CT Logs)
          </span>
        </div>

        <div class="flex items-center gap-2">
          <a
            href={`https://${targetDomain}`}
            target="_blank"
            rel="noopener noreferrer"
            class="px-3 py-1 rounded-full text-xs font-display font-bold bg-[#dffc78] text-[#19231f] hover:bg-[#e6fe8e] transition-colors"
          >
            ↗ Open Live Site
          </a>
          <a
            href={`https://web.archive.org/web/*/${targetDomain}`}
            target="_blank"
            rel="noopener noreferrer"
            class="px-3 py-1 rounded-full text-xs font-display font-bold bg-[#d9d6fc] text-[#19231f] hover:opacity-90 transition-colors"
          >
            🏛️ Wayback Calendar ↗
          </a>
          <button
            type="button"
            onclick={onClose}
            class="px-3 py-1 rounded-full text-xs font-mono bg-[#fffdf8]/15 hover:bg-[#ffc3a5] hover:text-[#19231f] text-[#fffdf8] transition-colors cursor-pointer"
          >
            Close [Esc]
          </button>
        </div>
      </div>

      <!-- Body -->
      <div class="p-6 space-y-5 scroll-panel">
        {#if loading}
          <LoadingProgressBanner
            title="Correlating Past Registration & Wayback Snapshots"
            target={targetDomain}
            workers={6}
            onCancel={onCancelJob}
            steps={[
              'Querying Wayback Machine Yearly CDX & Availability API',
              'Checking Authoritative RDAP Creation Date & Registrar',
              'Scanning crt.sh Certificate Transparency Subdomains',
            ]}
          />
        {:else if report}
          <!-- Verdict Card -->
          <div
            class="rounded-2xl border-[1.5px] border-[#19231f] p-5 flex flex-col sm:flex-row sm:items-center justify-between gap-4 {report.previouslyRegistered
              ? 'bg-[#ffc3a5]/45'
              : 'bg-[#dffc78]/45'}"
          >
            <div class="space-y-1.5">
              <div class="flex items-center gap-3 flex-wrap">
                <h3 class="text-2xl sm:text-3xl font-display font-bold text-[#19231f] font-mono">
                  {report.domain}
                </h3>
                {#if report.previouslyRegistered}
                  <span class="px-3 py-1 rounded-full text-xs font-display font-bold uppercase bg-[#ffc3a5] border border-[#19231f] text-[#19231f]">
                    ● {report.historyVerdict}
                  </span>
                {:else}
                  <span class="px-3 py-1 rounded-full text-xs font-display font-bold uppercase bg-[#dffc78] border border-[#19231f] text-[#19231f]">
                    ★ Clean Virgin History
                  </span>
                {/if}
              </div>
              <p class="text-xs text-[#19231f]/90 leading-relaxed">
                {report.summaryNote}
              </p>
            </div>

            <div class="flex flex-col sm:items-end gap-1.5 shrink-0">
              <a
                href={report.waybackCalendarUrl || `https://web.archive.org/web/*/${report.domain}`}
                target="_blank"
                rel="noopener noreferrer"
                class="studio-btn-primary px-4 py-1.5 text-xs text-center"
              >
                🏛️ View Old Snapshots on Wayback ↗
              </a>
              <span class="font-mono text-[11px] text-[#48534e]">Resolved in {report.checkLatencyMs}ms</span>
            </div>
          </div>

          {#if report.previouslyRegistered}
            <!-- Historical Metrics Grid -->
            <div class="grid grid-cols-2 sm:grid-cols-4 gap-3.5">
              <div class="rounded-2xl bg-[#f4f1e9] border border-[#19231f]/15 p-4">
                <div class="studio-label">First Registered / Seen</div>
                <div class="mt-1 font-mono text-base font-bold text-[#19231f]">
                  {report.rdapCreatedDate || report.firstSeenAt || 'N/A'}
                </div>
              </div>

              <div class="rounded-2xl bg-[#f4f1e9] border border-[#19231f]/15 p-4">
                <div class="studio-label">Most Recent Activity</div>
                <div class="mt-1 font-mono text-base font-bold text-[#19231f]">
                  {report.lastSeenAt || 'Active Now'}
                </div>
              </div>

              <div class="rounded-2xl bg-[#f4f1e9] border border-[#19231f]/15 p-4">
                <div class="studio-label">Wayback Captures</div>
                <div class="mt-1 font-display text-xl font-bold text-[#5366e8]">
                  {report.waybackSnapshots}+
                </div>
              </div>

              <div class="rounded-2xl bg-[#f4f1e9] border border-[#19231f]/15 p-4">
                <div class="studio-label">Active Span</div>
                <div class="mt-1 font-display text-xl font-bold text-[#19231f]">
                  {report.totalSpanYears} {report.totalSpanYears === 1 ? 'Year' : 'Years'}
                </div>
              </div>
            </div>

            <!-- CLICKABLE WAYBACK MACHINE SNAPSHOT TIME MACHINE -->
            {#if report.snapshots && report.snapshots.length > 0}
              <div class="rounded-2xl bg-[#f4f1e9] border-[1.5px] border-[#19231f]/20 p-5 space-y-3">
                <div class="flex flex-wrap items-center justify-between gap-2">
                  <div>
                    <div class="studio-label text-[#19231f]">
                      🏛️ Click Any Year to Open Historical Snapshot on Wayback Machine
                    </div>
                    <p class="text-xs text-[#48534e]">
                      Direct links to archived snapshots of <strong>{report.domain}</strong> across its history
                    </p>
                  </div>
                  <a
                    href={`https://web.archive.org/web/*/${report.domain}`}
                    target="_blank"
                    rel="noopener noreferrer"
                    class="text-xs font-mono font-bold text-[#5366e8] hover:underline"
                  >
                    Open Full Interactive Timeline ↗
                  </a>
                </div>

                <div class="grid grid-cols-2 sm:grid-cols-4 md:grid-cols-5 gap-2.5">
                  {#each report.snapshots as snap (snap.year)}
                    <a
                      href={snap.archiveUrl}
                      target="_blank"
                      rel="noopener noreferrer"
                      class="group rounded-xl bg-[#fffdf8] hover:bg-[#dffc78] border border-[#19231f] p-3 transition-all flex flex-col justify-between gap-1 shadow-[0_2px_0_#19231f]"
                    >
                      <div class="flex items-center justify-between">
                        <span class="font-display font-bold text-base text-[#19231f]">{snap.year}</span>
                        <span class="text-xs font-mono text-[#5366e8] group-hover:text-[#19231f] font-bold">↗</span>
                      </div>
                      <span class="text-[11px] font-mono text-[#48534e] group-hover:text-[#19231f]">
                        {snap.date}
                      </span>
                    </a>
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
                      <div class="rounded-xl bg-[#fffdf8] border border-[#19231f]/10 p-3 text-xs space-y-1">
                        <div class="flex items-center justify-between font-mono text-[11px] text-[#5366e8] font-semibold">
                          <span>{ms.date}</span>
                          <span>{ms.source}</span>
                        </div>
                        <div class="text-[#19231f]">{ms.event}</div>
                        {#if ms.archiveUrl}
                          <a
                            href={ms.archiveUrl}
                            target="_blank"
                            rel="noopener noreferrer"
                            class="inline-block text-[11px] font-mono font-bold text-[#5366e8] hover:underline pt-0.5"
                          >
                            View Archived Page Snapshot ↗
                          </a>
                        {/if}
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
                        <a
                          href={`https://${sub}`}
                          target="_blank"
                          rel="noopener noreferrer"
                          class="px-2.5 py-0.5 rounded-full bg-[#fffdf8] hover:bg-[#dffc78] border border-[#19231f]/25 text-xs font-mono text-[#19231f] transition-colors"
                        >
                          {sub} ↗
                        </a>
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
