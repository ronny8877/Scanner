<script lang="ts">
  import type { Job } from '../types';
  import { motionCard } from '../motion';

  interface Props {
    open: boolean;
    jobs: Job[];
    onClose: () => void;
    onSelectJob: (job: Job) => void;
    onDispatchParallelSuite: (target: string) => void;
  }

  let { open, jobs, onClose, onSelectJob, onDispatchParallelSuite }: Props = $props();
  let parallelTarget = $state('svelte.dev');

  function submitSuite(e: Event) {
    e.preventDefault();
    if (parallelTarget.trim()) {
      onDispatchParallelSuite(parallelTarget.trim());
    }
  }
</script>

{#if open}
  <div
    class="fixed inset-0 z-50 bg-[#19231f]/50 backdrop-blur-sm flex justify-end"
    role="dialog"
    aria-modal="true"
    aria-labelledby="job-queue-title"
  >
    <div
      use:motionCard={{ y: 0 }}
      class="w-full max-w-xl bg-[#f4f1e9] border-l-2 border-[#19231f] h-full flex flex-col shadow-2xl"
    >
      <!-- Drawer Header -->
      <div class="bg-[#19231f] text-[#fffdf8] px-6 py-5 flex items-center justify-between">
        <div>
          <div class="flex items-center gap-2">
            <span class="w-2.5 h-2.5 rounded-full bg-[#dffc78]"></span>
            <h2 id="job-queue-title" class="font-display font-bold text-base uppercase tracking-wider text-[#dffc78]">
              Enterprise Job & Worker Queue
            </h2>
          </div>
          <p class="text-xs text-[#fffdf8]/75 mt-0.5">
            Concurrent Go goroutine dispatcher & execution history ({jobs.length} tracked jobs)
          </p>
        </div>

        <button
          type="button"
          onclick={onClose}
          class="px-3.5 py-1.5 rounded-full text-xs font-mono bg-[#fffdf8]/15 hover:bg-[#dffc78] hover:text-[#19231f] text-[#fffdf8] transition-colors cursor-pointer"
        >
          Close
        </button>
      </div>

      <!-- Dispatch Multi-Pipeline Parallel Suite Form -->
      <form onsubmit={submitSuite} class="p-5 bg-[#fffdf8] border-b border-[#19231f]/15 space-y-3">
        <div class="flex items-center justify-between">
          <span class="studio-label">Dispatch 4-Engine Parallel Suite (32 Workers)</span>
          <span class="text-[11px] font-mono text-[#5366e8]">RDAP + Wayback/CT + Ports/TLS + Crawl</span>
        </div>

        <div class="flex gap-2">
          <input
            type="text"
            bind:value={parallelTarget}
            placeholder="Target domain (e.g. svelte.dev)..."
            class="flex-1 rounded-full bg-[#f4f1e9] border border-[#19231f]/25 px-4 py-2 text-xs font-mono text-[#19231f]"
          />
          <button type="submit" class="studio-btn-primary px-4 py-2 text-xs cursor-pointer shrink-0">
            ⚡ Dispatch Suite
          </button>
        </div>
      </form>

      <!-- Jobs List -->
      <div class="flex-1 p-5 space-y-3 scroll-panel">
        {#if jobs.length === 0}
          <div class="bento-card p-6 text-center text-xs text-[#48534e]">
            No jobs recorded yet in this session. Trigger any scan to populate the worker queue.
          </div>
        {:else}
          {#each jobs as job (job.id)}
            <div
              class="bento-card bento-card-interactive p-4 bg-[#fffdf8] space-y-2.5 cursor-pointer"
              role="button"
              tabindex="0"
              onclick={() => onSelectJob(job)}
              onkeydown={(e) => e.key === 'Enter' && onSelectJob(job)}
            >
              <div class="flex items-center justify-between gap-2">
                <div class="flex items-center gap-2">
                  <span class="font-mono text-xs font-bold px-2 py-0.5 rounded bg-[#19231f] text-[#dffc78]">
                    {job.id}
                  </span>
                  <span class="font-display font-bold text-sm text-[#19231f]">{job.title}</span>
                </div>

                {#if job.status === 'RUNNING'}
                  <span class="px-2.5 py-0.5 rounded-full text-[11px] font-mono font-bold bg-[#dffc78] border border-[#19231f] text-[#19231f] animate-pulse">
                    ● RUNNING ({job.progress}%)
                  </span>
                {:else if job.status === 'COMPLETED'}
                  <span class="px-2.5 py-0.5 rounded-full text-[11px] font-mono font-semibold bg-[#d9d6fc] border border-[#19231f]/30 text-[#19231f]">
                    ✓ COMPLETED ({job.durationMs}ms)
                  </span>
                {:else}
                  <span class="px-2.5 py-0.5 rounded-full text-[11px] font-mono bg-[#ffc3a5] border border-[#19231f]/30 text-[#19231f]">
                    {job.status}
                  </span>
                {/if}
              </div>

              <div class="flex items-center justify-between text-xs font-mono">
                <span class="text-[#5366e8] font-semibold truncate">Target: {job.target}</span>
                <span class="text-[#48534e]">{job.workers} parallel workers</span>
              </div>

              {#if job.status === 'RUNNING'}
                <div class="h-2 rounded-full bg-[#f4f1e9] border border-[#19231f]/15 overflow-hidden">
                  <div class="h-full bg-[#19231f] rounded-full" style="width: {job.progress}%"></div>
                </div>
                <div class="text-[11px] font-mono text-[#48534e]">{job.phase}</div>
              {:else if job.resultSummary}
                <div class="rounded-xl bg-[#f4f1e9] border border-[#19231f]/10 px-3 py-1.5 text-xs font-mono text-[#19231f] flex items-center justify-between">
                  <span>{job.resultSummary}</span>
                  <span class="text-[#5366e8] font-bold">Open →</span>
                </div>
              {/if}
            </div>
          {/each}
        {/if}
      </div>
    </div>
  </div>
{/if}
