<script lang="ts">
  import type { SavedDomain } from '../types';
  import { motionCard } from '../motion';
  import LoadingProgressBanner from './LoadingProgressBanner.svelte';

  interface Props {
    items: SavedDomain[];
    rechecking: boolean;
    onAddDomain: (domain: string, notes: string, tag: string) => void;
    onRemoveDomain: (domain: string) => void;
    onRecheckAll: () => void;
    onInspectDomain: (domain: string) => void;
    onCheckHistory: (domain: string) => void;
    onRunRecon: (domain: string) => void;
  }

  let {
    items,
    rechecking,
    onAddDomain,
    onRemoveDomain,
    onRecheckAll,
    onInspectDomain,
    onCheckHistory,
    onRunRecon,
  }: Props = $props();

  let newDomain = $state('');
  let newNotes = $state('');
  let newTag = $state('Shortlist');
  let filterMode = $state<'all' | 'available' | 'registered'>('all');

  const filteredItems = $derived(
    items.filter((i) => {
      if (filterMode === 'available') return i.available;
      if (filterMode === 'registered') return !i.available;
      return true;
    })
  );

  function handleAdd(e: Event) {
    e.preventDefault();
    if (!newDomain.trim()) return;
    onAddDomain(newDomain.trim(), newNotes.trim(), newTag);
    newDomain = '';
    newNotes = '';
  }
</script>

<section aria-labelledby="watchlist-vault-heading" class="space-y-6">
  <!-- Top Asymmetric Bento: 8-Col Add to Vault Form + 4-Col Parallel Re-Check Action Card -->
  <div class="grid grid-cols-1 lg:grid-cols-12 gap-6 items-stretch">
    <form
      use:motionCard
      onsubmit={handleAdd}
      class="lg:col-span-8 bento-card p-6 sm:p-7 flex flex-col justify-between gap-4"
    >
      <div class="flex items-center justify-between">
        <h2 id="watchlist-vault-heading" class="studio-label">
          05 // Persistent Saved Domains Vault & Watchlist
        </h2>
        <span class="text-xs font-mono text-[#48534e]">Synced to data/watchlist.json</span>
      </div>

      <div class="grid grid-cols-1 sm:grid-cols-12 gap-3">
        <input
          type="text"
          bind:value={newDomain}
          placeholder="Domain to save (e.g. veltrixhq.ai)..."
          class="sm:col-span-5 rounded-2xl bg-[#f4f1e9] border-[1.5px] border-[#19231f]/20 focus:border-[#19231f] px-4 py-3 text-sm font-mono text-[#19231f] placeholder-[#6d7873]"
        />
        <input
          type="text"
          bind:value={newNotes}
          placeholder="Editorial note or project idea..."
          class="sm:col-span-4 rounded-2xl bg-[#f4f1e9] border-[1.5px] border-[#19231f]/20 focus:border-[#19231f] px-4 py-3 text-sm text-[#19231f] placeholder-[#6d7873]"
        />
        <button
          type="submit"
          class="sm:col-span-3 studio-btn-primary py-3 px-4 text-xs cursor-pointer"
        >
          ★ Save to Vault
        </button>
      </div>

      <div class="flex flex-wrap items-center justify-between gap-2 pt-2 border-t border-[#19231f]/10 text-xs">
        <div class="flex items-center gap-2">
          <span class="text-[#6d7873]">Filter vault:</span>
          <button
            type="button"
            onclick={() => (filterMode = 'all')}
            class="px-3 py-1 rounded-full font-display font-semibold border cursor-pointer {filterMode === 'all'
              ? 'bg-[#19231f] text-[#fffdf8] border-[#19231f]'
              : 'bg-[#f4f1e9] text-[#19231f] border-[#19231f]/15'}"
          >
            All ({items.length})
          </button>
          <button
            type="button"
            onclick={() => (filterMode = 'available')}
            class="px-3 py-1 rounded-full font-display font-semibold border cursor-pointer {filterMode === 'available'
              ? 'bg-[#dffc78] text-[#19231f] border-[#19231f]'
              : 'bg-[#f4f1e9] text-[#19231f] border-[#19231f]/15'}"
          >
            Unclaimed Gems ({items.filter((i) => i.available).length})
          </button>
          <button
            type="button"
            onclick={() => (filterMode = 'registered')}
            class="px-3 py-1 rounded-full font-display font-semibold border cursor-pointer {filterMode === 'registered'
              ? 'bg-[#ffc3a5] text-[#19231f] border-[#19231f]'
              : 'bg-[#f4f1e9] text-[#19231f] border-[#19231f]/15'}"
          >
            Registered Watch ({items.filter((i) => !i.available).length})
          </button>
        </div>
      </div>
    </form>

    <!-- 4-Column Parallel Batch Re-Check Card -->
    <div
      use:motionCard={{ delay: 0.05 }}
      class="lg:col-span-4 bento-card p-6 sm:p-7 flex flex-col justify-between bg-[#d9d6fc]/55"
    >
      <div>
        <span class="studio-label text-[#19231f]">Batch Registry Verification</span>
        <h3 class="mt-2 text-2xl font-display font-bold text-[#19231f]">
          Re-Verify All {items.length} Saved Domains in <span class="font-serif-editorial font-normal text-3xl">parallel</span>
        </h3>
        <p class="mt-1 text-xs text-[#19231f]/80 leading-relaxed">
          Concurrently checks live RDAP registration status and Wayback/CT past history for every saved domain.
        </p>
      </div>

      <button
        type="button"
        disabled={rechecking || items.length === 0}
        onclick={onRecheckAll}
        class="mt-4 studio-btn-primary w-full py-3 px-4 text-xs cursor-pointer disabled:opacity-50"
      >
        {rechecking ? 'Re-Verifying All Domains in Parallel…' : '⚡ Re-Verify All Saved Domains Now'}
      </button>
    </div>
  </div>

  {#if rechecking}
    <LoadingProgressBanner
      title="Parallel Watchlist Re-Verification"
      target={`${items.length} Saved Domains`}
      workers={Math.max(4, items.length * 2)}
      steps={[
        'Resolving Live DNS NS/A Delegation',
        'Querying Authoritative RDAP Status',
        'Checking Wayback & CT Past Registration Traces',
      ]}
    />
  {/if}

  <!-- Saved Domains Bento Grid -->
  <div class="grid grid-cols-1 md:grid-cols-2 gap-5">
    {#each filteredItems as item (item.domain)}
      <div
        use:motionCard
        class="bento-card bento-card-interactive p-6 flex flex-col justify-between gap-4 bg-[#fffdf8]"
      >
        <div class="space-y-2.5">
          <div class="flex items-start justify-between gap-3">
            <div class="flex items-center gap-2.5 flex-wrap">
              <h3 class="text-2xl font-display font-bold font-mono text-[#19231f]">
                {item.domain}
              </h3>

              {#if item.available}
                <span class="px-2.5 py-0.5 rounded-full text-[11px] font-display font-bold uppercase bg-[#dffc78] border border-[#19231f] text-[#19231f]">
                  ● Unclaimed
                </span>
              {:else}
                <span class="px-2.5 py-0.5 rounded-full text-[11px] font-display font-semibold uppercase bg-[#ffc3a5] border border-[#19231f] text-[#19231f]">
                  Registered
                </span>
              {/if}

              {#if item.previouslyRegistered === true}
                <span class="px-2.5 py-0.5 rounded-full text-[11px] font-mono bg-[#d9d6fc] border border-[#19231f]/40 text-[#19231f]">
                  Past Reg ({item.firstSeenYear || 'Archive'})
                </span>
              {:else if item.previouslyRegistered === false}
                <span class="px-2.5 py-0.5 rounded-full text-[11px] font-mono bg-[#dffc78]/60 border border-[#19231f]/40 text-[#19231f]">
                  Virgin History
                </span>
              {/if}
            </div>

            <button
              type="button"
              onclick={() => onRemoveDomain(item.domain)}
              class="text-xs font-mono text-[#6d7873] hover:text-[#19231f] px-2 py-0.5 rounded-full hover:bg-[#f4f1e9] cursor-pointer"
              title="Remove from Vault"
            >
              ✕ Remove
            </button>
          </div>

          {#if item.notes}
            <p class="text-xs text-[#48534e] bg-[#f4f1e9] rounded-xl px-3.5 py-2 border border-[#19231f]/10">
              “{item.notes}”
            </p>
          {/if}

          <div class="flex items-center justify-between text-xs pt-1">
            <span class="font-mono text-[#48534e]">
              Valuation: <strong class="text-[#19231f]">{item.valuation.estimatedDisplay}</strong> ({item.valuation.score}/100)
            </span>
            <span class="font-serif-editorial text-base text-[#19231f]">{item.valuation.tier}</span>
          </div>
        </div>

        <div class="pt-3 border-t border-[#19231f]/10 flex flex-wrap items-center justify-between gap-2">
          <div class="flex flex-wrap gap-1.5">
            <button
              type="button"
              onclick={() => onCheckHistory(item.domain)}
              class="px-3 py-1.5 rounded-full text-xs font-display font-semibold bg-[#f4f1e9] hover:bg-[#ffc3a5] text-[#19231f] border border-[#19231f]/25 transition-colors cursor-pointer"
            >
              ⏳ Past History
            </button>
            <button
              type="button"
              onclick={() => onInspectDomain(item.domain)}
              class="px-3 py-1.5 rounded-full text-xs font-display font-semibold bg-[#f4f1e9] hover:bg-[#d9d6fc] text-[#19231f] border border-[#19231f]/25 transition-colors cursor-pointer"
            >
              RDAP Dossier
            </button>
            <button
              type="button"
              onclick={() => onRunRecon(item.domain)}
              class="px-3 py-1.5 rounded-full text-xs font-display font-semibold bg-[#f4f1e9] hover:bg-[#dffc78] text-[#19231f] border border-[#19231f]/25 transition-colors cursor-pointer"
            >
              Ports & TLS
            </button>
          </div>
        </div>
      </div>
    {/each}
  </div>
</section>
