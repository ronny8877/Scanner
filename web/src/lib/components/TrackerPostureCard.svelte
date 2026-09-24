<script lang="ts">
  import type { TrackerTelemetry } from '../types';

  interface Props {
    trackers: TrackerTelemetry;
    title?: string;
  }

  let { trackers, title = 'Ad Networks, Analytics & Tracker Telemetry' }: Props = $props();

  function categoryBadgeStyle(cat: string): string {
    switch (cat) {
      case 'AD_NETWORK':
        return 'bg-[#ffc3a5] text-[#19231f] border-[#19231f]';
      case 'PIXEL_TRACKER':
        return 'bg-[#19231f] text-[#ffc3a5] border-[#19231f]';
      case 'ANALYTICS':
        return 'bg-[#d9d6fc] text-[#19231f] border-[#19231f]';
      default:
        return 'bg-[#dffc78] text-[#19231f] border-[#19231f]';
    }
  }

  function categoryLabel(cat: string): string {
    switch (cat) {
      case 'AD_NETWORK':
        return 'Ad Network';
      case 'PIXEL_TRACKER':
        return 'Retargeting Pixel';
      case 'ANALYTICS':
        return 'Analytics Engine';
      default:
        return 'SDK / Telemetry';
    }
  }
</script>

<div class="bento-card overflow-hidden">
  <!-- Dark Ink Posture Header (Matching Security Posture Grade Card) -->
  <div class="bg-[#19231f] text-[#fffdf8] p-6 flex items-center justify-between gap-4">
    <div>
      <span class="text-[11px] font-display uppercase tracking-widest text-[#dffc78]">
        {title}
      </span>
      <h4 class="text-xl font-display font-bold mt-1 text-[#fffdf8]">
        {trackers.verdict}
      </h4>
      <p class="text-xs text-[#fffdf8]/75 mt-1 leading-relaxed">
        {trackers.summary}
      </p>
    </div>

    <div
      class="w-16 h-16 rounded-2xl border-2 border-[#19231f] flex flex-col items-center justify-center shrink-0 {trackers.privacyGrade.startsWith(
        'A'
      )
        ? 'bg-[#dffc78] text-[#19231f]'
        : trackers.privacyGrade === 'B'
          ? 'bg-[#d9d6fc] text-[#19231f]'
          : 'bg-[#ffc3a5] text-[#19231f]'}"
    >
      <span class="text-[9px] font-mono uppercase font-bold tracking-wider">Posture</span>
      <span class="font-display font-bold text-2xl leading-none">{trackers.privacyGrade}</span>
    </div>
  </div>

  <!-- 4-Category Counter Grid -->
  <div class="p-5 bg-[#fffdf8] space-y-4">
    <div class="grid grid-cols-2 sm:grid-cols-4 gap-2.5 text-center font-mono">
      <div class="rounded-2xl bg-[#f4f1e9] border border-[#19231f]/15 p-3">
        <div class="text-[10px] text-[#48534e] uppercase">Ad Networks</div>
        <div class="text-lg font-display font-bold text-[#19231f] mt-0.5">
          {trackers.adNetworksCount}
        </div>
      </div>

      <div class="rounded-2xl bg-[#f4f1e9] border border-[#19231f]/15 p-3">
        <div class="text-[10px] text-[#48534e] uppercase">Analytics</div>
        <div class="text-lg font-display font-bold text-[#5366e8] mt-0.5">
          {trackers.analyticsCount}
        </div>
      </div>

      <div class="rounded-2xl bg-[#f4f1e9] border border-[#19231f]/15 p-3">
        <div class="text-[10px] text-[#48534e] uppercase">Ad Pixels</div>
        <div class="text-lg font-display font-bold text-[#19231f] mt-0.5">
          {trackers.pixelsCount}
        </div>
      </div>

      <div class="rounded-2xl bg-[#f4f1e9] border border-[#19231f]/15 p-3">
        <div class="text-[10px] text-[#48534e] uppercase">SDK / CMP</div>
        <div class="text-lg font-display font-bold text-[#19231f] mt-0.5">
          {trackers.telemetryCount}
        </div>
      </div>
    </div>

    <!-- Detected Trackers / Ad Networks List -->
    {#if trackers.detectedTrackers && trackers.detectedTrackers.length > 0}
      <div class="space-y-2.5 pt-1">
        <div class="studio-label">Detected Monetization, Analytics & Tracking Scripts ({trackers.detectedTrackers.length})</div>
        <div class="space-y-2 max-h-72 scroll-panel pr-1">
          {#each trackers.detectedTrackers as item}
            <div class="rounded-2xl bg-[#f4f1e9] border border-[#19231f]/12 p-3.5 space-y-1.5">
              <div class="flex flex-wrap items-center justify-between gap-2">
                <div class="flex items-center gap-2">
                  <span class="font-display font-bold text-sm text-[#19231f]">{item.name}</span>
                  <span class="text-[11px] font-mono text-[#48534e]">by {item.provider}</span>
                </div>
                <span
                  class="px-2.5 py-0.5 rounded-full text-[10px] font-display font-bold border {categoryBadgeStyle(
                    item.category
                  )}"
                >
                  {categoryLabel(item.category)}
                </span>
              </div>
              <p class="text-xs text-[#48534e]">{item.description}</p>
              <div class="text-[11px] font-mono text-[#5366e8] truncate">
                Matched signature: <code>{item.matchedRule}</code>
              </div>
            </div>
          {/each}
        </div>
      </div>
    {:else}
      <div class="rounded-2xl bg-[#dffc78]/35 border border-[#19231f]/20 p-4 flex items-center justify-between">
        <div>
          <div class="font-display font-bold text-xs text-[#19231f]">
            ✓ Zero Third-Party Ad Networks or Tracking Pixels Detected
          </div>
          <div class="text-xs text-[#48534e] mt-0.5">
            Scanned HTML scripts, inline tags, and network references — no AdSense, GTM, Meta Pixel, or behavioral trackers found.
          </div>
        </div>
        <span class="px-2.5 py-1 rounded-full text-xs font-mono font-bold bg-[#dffc78] border border-[#19231f] text-[#19231f] shrink-0">
          CLEAN
        </span>
      </div>
    {/if}
  </div>
</div>
