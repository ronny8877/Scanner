<script lang="ts">
  import StudioIcon from './StudioIcon.svelte';

  let {
    error = null,
    suggestion = null,
    onApplySuggestion,
    onDismiss,
  }: {
    error: string | null;
    suggestion?: string | null;
    onApplySuggestion?: (fixed: string) => void;
    onDismiss?: () => void;
  } = $props();
</script>

{#if error}
  <div
    role="alert"
    class="mt-3 flex flex-wrap items-center justify-between gap-3 rounded-2xl border-2 border-[#19231f] bg-[#ffc3a5]/55 px-4 py-3 text-xs text-[#19231f] shadow-[0_6px_18px_rgba(25,35,31,0.08)]"
  >
    <div class="flex items-center gap-2.5">
      <span class="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-[#19231f] text-[#ffc3a5]">
        <StudioIcon name="alert" class="h-3.5 w-3.5" />
      </span>
      <div>
        <span class="font-display font-bold uppercase tracking-wider">Domain Validation Guard · </span>
        <span class="font-medium">{error}</span>
      </div>
    </div>

    <div class="flex items-center gap-2">
      {#if suggestion && onApplySuggestion}
        <button
          type="button"
          onclick={() => onApplySuggestion(suggestion)}
          class="inline-flex items-center gap-1.5 rounded-full border border-[#19231f] bg-[#dffc78] px-3.5 py-1.5 font-display text-xs font-bold text-[#19231f] transition-transform hover:-translate-y-0.5"
        >
          <StudioIcon name="check" class="h-3.5 w-3.5" />
          Use <span class="underline">{suggestion}</span> instead
        </button>
      {/if}
      {#if onDismiss}
        <button
          type="button"
          onclick={onDismiss}
          aria-label="Dismiss validation notice"
          class="rounded-full border border-[#19231f]/20 bg-[#fffdf8] px-2.5 py-1 font-display text-[11px] font-semibold text-[#19231f]/70 hover:text-[#19231f]"
        >
          Dismiss
        </button>
      {/if}
    </div>
  </div>
{/if}
