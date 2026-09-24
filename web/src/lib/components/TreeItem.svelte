<script lang="ts">
  import type { SiteNode } from '../types';
  import Self from './TreeItem.svelte';

  interface Props {
    node: SiteNode;
    depth?: number;
    selectedPath: string;
    onSelect: (node: SiteNode) => void;
  }

  let { node, depth = 0, selectedPath, onSelect }: Props = $props();
  let expanded = $state(true);

  const hasChildren = $derived(Boolean(node.children && node.children.length > 0));
  const isSelected = $derived(selectedPath === node.fullPath);

  function statusBadgeClass(code?: number) {
    if (!code) return 'bg-[#f4f1e9] text-[#48534e] border-[#19231f]/15';
    if (code >= 200 && code < 300) return 'bg-[#dffc78] text-[#19231f] border-[#19231f] font-semibold';
    if (code >= 300 && code < 400) return 'bg-[#d9d6fc] text-[#19231f] border-[#19231f]/40';
    return 'bg-[#ffc3a5] text-[#19231f] border-[#19231f]/40';
  }
</script>

<div class="select-none">
  <div
    class="group flex items-center justify-between gap-2 rounded-xl px-3 py-2 transition-colors cursor-pointer border {isSelected
      ? 'bg-[#19231f] text-[#fffdf8] border-[#19231f]'
      : 'border-transparent hover:bg-[#f4f1e9] text-[#19231f]'}"
    style="margin-left: {depth * 18}px"
    role="button"
    tabindex="0"
    onclick={() => onSelect(node)}
    onkeydown={(e) => e.key === 'Enter' && onSelect(node)}
  >
    <div class="flex items-center gap-2 min-w-0">
      {#if hasChildren}
        <button
          type="button"
          class="w-5 h-5 flex items-center justify-center rounded-md border text-xs font-mono cursor-pointer {isSelected
            ? 'border-[#fffdf8]/30 text-[#dffc78] hover:bg-[#fffdf8]/10'
            : 'border-[#19231f]/20 text-[#19231f] hover:bg-[#dffc78]'}"
          onclick={(e) => {
            e.stopPropagation();
            expanded = !expanded;
          }}
          aria-label="Toggle child routes"
        >
          {expanded ? '−' : '+'}
        </button>
      {:else}
        <span class="w-5 h-5 flex items-center justify-center text-xs font-mono opacity-45">├</span>
      {/if}

      <span class="font-mono text-sm font-semibold truncate">
        {depth === 0 ? `◈ ${node.segment}` : `/${node.segment}`}
      </span>

      {#if node.title}
        <span
          class="text-xs truncate hidden sm:inline {isSelected
            ? 'text-[#fffdf8]/75'
            : 'text-[#48534e]'}"
        >
          — {node.title}
        </span>
      {/if}
    </div>

    <div class="flex items-center gap-2 shrink-0">
      {#if node.latencyMs}
        <span
          class="text-[11px] font-mono {isSelected
            ? 'text-[#dffc78]'
            : 'text-[#6d7873]'}"
        >
          {node.latencyMs}ms
        </span>
      {/if}
      <span class="text-[11px] font-mono px-2 py-0.5 rounded-full border {statusBadgeClass(node.statusCode)}">
        {node.statusCode ? node.statusCode : 'route'}
      </span>
    </div>
  </div>

  {#if hasChildren && expanded}
    <div class="mt-0.5 space-y-0.5 border-l-[1.5px] border-[#19231f]/15 ml-4">
      {#each node.children ?? [] as child (child.fullPath)}
        <Self node={child} depth={depth + 1} {selectedPath} {onSelect} />
      {/each}
    </div>
  {/if}
</div>
