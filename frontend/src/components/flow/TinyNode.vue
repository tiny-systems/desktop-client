<script lang="ts" setup>
import type { CSSProperties } from 'vue'
import { ref, watch, onMounted, onUnmounted } from 'vue'
import { Handle, Position } from '@vue-flow/core'
import { AdjustmentsHorizontalIcon, GlobeAltIcon } from '@heroicons/vue/24/outline'

defineOptions({
  inheritAttrs: false,
})

interface HandleProps {
  id?: string
  type?: 'source' | 'target'
  label?: string
  position?: number
  virtual?: boolean
  rotated_position: number
  error?: string
}

interface NodeData {
  handles: HandleProps[]
  blocked: boolean
  label?: string
  error?: boolean
  emitter?: boolean
  emitting?: boolean
  emit?: boolean
  comment?: string
  dashboard?: string
  shared_with_flows?: string
  stats?: unknown
  last_status_update?: number
  trace?: {
    error?: boolean
    sequence?: number
    port?: string
    latency?: number
  }
}

const props = defineProps<{
  data: NodeData
  selected?: boolean
  id: string
  noExpire?: boolean
  editorMode?: boolean
}>()

const emit = defineEmits(['updateNodeInternals'])

// Format milliseconds to readable time
function msToTime(ms: number): string {
  if (!ms && ms !== 0) return ''
  if (ms < 1) return '<1ms'
  if (ms < 1000) return `${Math.round(ms)}ms`
  return `${(ms / 1000).toFixed(2)}s`
}

const loading = ref(false)
let interval: ReturnType<typeof setInterval> | null = null

onMounted(() => {
  if (!props.noExpire) {
    interval = setInterval(() => {
      if (!props.data) {
        return
      }
      const n = new Date((props.data.last_status_update || 0) * 1000).getTime()
      if (!n) {
        loading.value = true
        return
      }

      if (Date.now().valueOf() - n > 10 * 60 * 1000) {
        loading.value = true
      } else {
        loading.value = false
      }
    }, 2000)
  }
})

onUnmounted(() => {
  if (interval) {
    clearInterval(interval)
  }
})

watch(() => props.data, () => {
  emit('updateNodeInternals')
}, { deep: true })

function calculateHandlerClass(h: HandleProps, nodeData: NodeData): string {
  if (h.error) {
    return 'bg-red-300 dark:border-gray-400 dark:bg-red-800'
  }
  if (nodeData?.blocked) {
    return 'bg-gray-100 dark:border-gray-400 dark:bg-gray-600'
  }
  return 'bg-gray-300 dark:border-gray-400 dark:bg-gray-700'
}

function calculateHandlerStyle(h: HandleProps, nodeData: NodeData): CSSProperties {
  const style: CSSProperties = {
    height: '15px',
    width: '15px',
  }

  let idx = 0
  let i = 0
  const handles = nodeData?.handles || []
  const sameSideHandles = handles.filter(
    a => a.rotated_position === h.rotated_position && !a.virtual
  )

  for (const ha of sameSideHandles) {
    if (h.id === ha.id) {
      idx = i
    }
    i++
  }

  const total = sameSideHandles.length || 1

  if (h.rotated_position === 1) {
    // right
    style.right = '0px'
    style.paddingLeft = '13px'
    style.top = `${((idx + 1) * 100) / (total + 1)}%`
    style.lineHeight = '14px'
  } else if (h.rotated_position === 3) {
    // left
    style.left = '0px'
    style.paddingRight = '13px'
    style.direction = 'rtl'
    style.top = `${((idx + 1) * 100) / (total + 1)}%`
    style.lineHeight = '13px'
  } else if (h.rotated_position === 2) {
    // bottom
    style.writingMode = 'vertical-rl'
    style.bottom = '0px'
    style.paddingTop = '13px'
    style.paddingLeft = '4px'
    style.lineHeight = '13px'
    style.left = `${((idx + 1) * 100) / (total + 1)}%`
  } else {
    // top (0)
    style.left = `${((idx + 1) * 100) / (total + 1)}%`
    style.writingMode = 'vertical-lr'
    style.direction = 'rtl'
    style.paddingBottom = '13px'
    style.paddingRight = '4px'
    style.lineHeight = '14px'
  }

  return style
}

function posIntToStr(n: number): Position {
  switch (n) {
    case 0:
      return Position.Top
    case 1:
      return Position.Right
    case 2:
      return Position.Bottom
    case 3:
      return Position.Left
    default:
      return Position.Top
  }
}

function calculateBoxStyle(data: NodeData): CSSProperties {
  const handles = data?.handles || []
  const leftRightElementCount = Math.max(
    handles.filter(a => a.rotated_position === 1 && !a.virtual).length,
    handles.filter(a => a.rotated_position === 3 && !a.virtual).length
  )
  const topBottomElementCount = Math.max(
    handles.filter(a => a.rotated_position === 0 && !a.virtual).length,
    handles.filter(a => a.rotated_position === 2 && !a.virtual).length
  )
  return {
    minWidth: `${(topBottomElementCount || 1) * 30 + 70}px`,
    textAlign: 'center',
    padding: '14px',
    minHeight: `${(leftRightElementCount || 1) * 30 + 50}px`,
  }
}

function calcBoxClass(): string {
  if (props.data?.trace?.error) {
    if (props.selected) {
      return 'bg-red-200 dark:bg-red-950 border border-gray-400'
    }
    return 'bg-red-100 dark:bg-red-800 border border-gray-300'
  }
  if (props.data?.blocked) {
    if (props.selected) {
      return 'bg-gray-300 border border-indigo-200 dark:bg-gray-800 blocked'
    }
    return 'bg-gray-100 border border-gray-300 dark:bg-gray-700 blocked'
  }

  if (props.selected) {
    return 'bg-sky-500 border border-gray-700 dark:border-gray-600 dark:bg-gray-900 text-white'
  }
  return 'bg-sky-100 border border-solid dark:border-gray-600 dark:bg-gray-800 border-sky-300'
}
</script>

<template>
  <div :style="calculateBoxStyle(props.data)" :class="['dark:text-gray-300', calcBoxClass()]">
    {{ props.data?.label || props.id }}
    <span
      v-if="props.data?.emitter"
      class="dot"
      :class="props.data?.emitting ? 'emitting' : props.data?.emit ? 'emit' : 'stop'"
    />
    <div v-if="props.data?.stats" />
    <div
      v-if="props.data?.trace && props.data.trace.sequence !== undefined && props.data.trace.sequence >= 0"
      class="text-xs w-full text-center text-sky-500"
      :title="'Span# ' + props.data.trace.sequence"
    >
      {{ props.data.trace.port }}: {{ msToTime(props.data.trace.latency || 0) }}
    </div>
    <div class="pt-5 space-x-0.5">
      <span v-if="props.data?.error" class="bg-red-500 text-xs rounded text-white p-1">error</span>
      <span
        v-if="props.data?.shared_with_flows"
        class="bg-emerald-400 text-xs rounded text-white p-0.5 inline-block"
        title="Shared node"
      >
        <GlobeAltIcon class="w-4 h-4 inline-block" />
      </span>
      <span
        v-if="props.data?.dashboard === 'true'"
        class="bg-fuchsia-400 text-xs rounded text-white p-0.5 inline-block"
        title="Added to dashboard"
      >
        <AdjustmentsHorizontalIcon class="w-4 h-4 inline-block" />
      </span>
      <span class="text-xs">{{ props.data?.comment }}</span>
    </div>
  </div>
  <template v-if="!props.noExpire">
    <div
      v-if="(props.data?.handles || []).length === 0 || loading"
      class="absolute inset-0 bg-gray-900/50 rounded flex items-center justify-center"
    >
      <div class="text-center">
        <div class="animate-spin rounded-full h-4 w-4 border-2 border-white border-t-transparent mx-auto"></div>
        <span class="text-white text-xs mt-1 block">Loading</span>
      </div>
    </div>
  </template>
  <template v-for="h in props.data?.handles || []" :key="h.id">
    <Handle
      v-if="h.id !== '_settings' && h.id !== '_control'"
      :id="h.id"
      :type="h.type || 'source'"
      :position="posIntToStr(h.rotated_position)"
      :style="calculateHandlerStyle(h, props.data)"
      :class="calculateHandlerClass(h, props.data)"
      :title="h.error || ''"
    >
      <template
        v-if="
          (h.type === 'source' && (h.rotated_position === 1 || h.rotated_position === 2)) ||
          (h.type === 'target' && (h.rotated_position === 3 || h.rotated_position === 0))
        "
      >
        &rarr;{{ h.label }}
      </template>
      <template
        v-else-if="
          (h.type === 'target' && (h.rotated_position === 1 || h.rotated_position === 2)) ||
          h.type === 'source'
        "
      >
        &larr;{{ h.label }}
      </template>
    </Handle>
  </template>
</template>

<style scoped>
.dot {
  height: 10px;
  width: 10px;
  border-radius: 50%;
  display: inline-block;
  margin-left: 5px;
}
.dot.emitting {
  background-color: #22c55e;
  animation: pulse 1s infinite;
}
.dot.emit {
  background-color: #22c55e;
}
.dot.stop {
  background-color: #9ca3af;
}
@keyframes pulse {
  0%,
  100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}
</style>
