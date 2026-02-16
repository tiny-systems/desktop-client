<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { ChevronUpIcon, ChevronDownIcon, ExclamationTriangleIcon, ArrowPathIcon } from '@heroicons/vue/24/outline'
import { EventsOn } from '../../../wailsjs/runtime/runtime'
import { GetTraces } from '../../../wailsjs/go/main/App'
import { useFlowStore } from '../../stores/flow'

const props = defineProps({
  ctx: {
    type: String,
    required: true
  },
  ns: {
    type: String,
    required: true
  },
  flowName: {
    type: String,
    required: true
  },
  projectName: {
    type: String,
    required: true
  }
})

const emit = defineEmits(['trace'])
const flowStore = useFlowStore()

const collapsed = ref(false)
const telemetryError = ref(null)
const loading = ref(false)
const initialLoadDone = ref(false)
const traces = ref([])

// Store the callback reference so we can remove only our listener
let errorEventCallback = null
let refreshTimeout = null

// Computed metrics from traces
const metrics = computed(() => {
  if (traces.value.length === 0) {
    return { total: 0, errors: 0, avgDuration: 0 }
  }

  const total = traces.value.length
  const errors = traces.value.reduce((sum, t) => sum + (t.errors || 0), 0)
  const totalDuration = traces.value.reduce((sum, t) => sum + (t.duration || 0), 0)
  const avgDuration = total > 0 ? totalDuration / total : 0

  return { total, errors, avgDuration }
})

// Selected trace ID from flow store
const selectedTraceId = computed(() => flowStore.trace)

// Load traces from the backend
const loadTraces = async () => {
  if (!props.ctx || !props.ns || !props.projectName || !props.flowName) {
    return
  }

  loading.value = true
  try {
    const response = await GetTraces(
      props.ctx,
      props.ns,
      props.projectName,
      '', // empty - fetch all project traces, not filtered by flow
      0, // start - 0 means use default (last 15 minutes)
      0, // end - 0 means use default (now)
      0  // offset
    )

    if (response && response.traces) {
      traces.value = response.traces.map(t => ({
        id: t.id,
        spans: t.spans,
        errors: t.errors,
        data: t.data,
        length: t.length,
        duration: t.duration,
        start: t.start,
        end: t.end
      }))
    } else {
      traces.value = []
    }
    telemetryError.value = null
  } catch (err) {
    console.error('Failed to load traces:', err)
    telemetryError.value = `Failed to load traces: ${err}`
  } finally {
    loading.value = false
    initialLoadDone.value = true
  }
}

// Debounced reload to avoid too many API calls
const scheduleReload = () => {
  if (refreshTimeout) {
    clearTimeout(refreshTimeout)
  }
  refreshTimeout = setTimeout(() => {
    loadTraces()
  }, 500)
}

// Listen for flow events to trigger trace reload
onMounted(() => {
  errorEventCallback = (event) => {
    if (event?.type === 'ERROR' && event?.id) {
      telemetryError.value = event.id
    }
    // Schedule reload when we get stats updates (indicates flow activity)
    if (event?.type === 'STATS' || event?.type === 'MODIFIED' || event?.type === 'ADDED') {
      scheduleReload()
    }
  }
  EventsOn('flowNodeUpdate', errorEventCallback)

  // Initial load
  loadTraces()
})

onUnmounted(() => {
  errorEventCallback = null
  if (refreshTimeout) {
    clearTimeout(refreshTimeout)
  }
})

// Reload when flow changes
watch(
  () => [props.ctx, props.ns, props.projectName, props.flowName],
  () => {
    initialLoadDone.value = false
    traces.value = []
    flowStore.clearTrace()
    loadTraces()
  }
)

const clearError = () => {
  telemetryError.value = null
}

const hasData = ref(false)
watch(traces, (newTraces) => {
  hasData.value = newTraces.length > 0
}, { immediate: true })

// Format duration from nanoseconds to readable string
const formatDuration = (ns) => {
  if (typeof ns !== 'number') return '0ms'
  const ms = ns / 1000000
  if (ms < 1) return '<1ms'
  if (ms < 1000) return `${Math.round(ms)}ms`
  return `${(ms / 1000).toFixed(2)}s`
}

// Format time from microseconds timestamp (proto uses microseconds)
const formatTime = (timestamp) => {
  if (!timestamp) return '-'
  // Timestamp is in microseconds, convert to milliseconds
  const ms = timestamp / 1000
  const date = new Date(ms)
  const opts = {
    hour: 'numeric',
    minute: 'numeric',
    second: 'numeric',
    hour12: false
  }
  return new Intl.DateTimeFormat('default', opts).format(date)
}

// Format relative time from microseconds
const formatRelativeTime = (timestamp) => {
  if (!timestamp) return '-'
  // Timestamp is in microseconds, convert to milliseconds
  const ms = timestamp / 1000
  const now = Date.now()
  const diff = now - ms

  if (diff < 1000) return 'just now'
  if (diff < 60000) return `${Math.floor(diff / 1000)}s ago`
  if (diff < 3600000) return `${Math.floor(diff / 60000)}m ago`
  if (diff < 86400000) return `${Math.floor(diff / 3600000)}h ago`
  return `${Math.floor(diff / 86400000)}d ago`
}

const selectTrace = (trace) => {
  flowStore.setTrace(trace.id)
  emit('trace', trace.id)
}

const toggleCollapsed = () => {
  collapsed.value = !collapsed.value
}

const refresh = () => {
  loadTraces()
}
</script>

<template>
  <div
    :class="[
      'w-full border-t border-gray-200 dark:border-gray-700 text-sm relative bg-gray-50 dark:bg-gray-900',
      !collapsed ? 'min-h-64 h-1/4' : ''
    ]"
  >
    <!-- Header with collapse toggle -->
    <div
      class="flex items-center justify-between px-3 py-1.5 bg-gray-100 dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700 cursor-pointer"
      @click="toggleCollapsed"
    >
      <div class="flex items-center gap-4">
        <span class="text-xs font-medium text-gray-600 dark:text-gray-400">
          Telemetry
        </span>
        <!-- Metrics summary -->
        <div v-if="hasData && !collapsed" class="flex items-center gap-3 text-xs">
          <span class="px-2 py-0.5 bg-sky-100 dark:bg-sky-900/50 text-sky-700 dark:text-sky-300 rounded">
            {{ metrics.total }} traces
          </span>
          <span v-if="metrics.errors > 0" class="px-2 py-0.5 bg-red-100 dark:bg-red-900/50 text-red-700 dark:text-red-300 rounded">
            {{ metrics.errors }} errors
          </span>
          <span class="text-gray-500">
            avg: {{ formatDuration(metrics.avgDuration) }}
          </span>
        </div>
        <span v-else-if="traces.length > 0" class="text-gray-500 text-xs">
          ({{ traces.length }})
        </span>
      </div>
      <div class="flex items-center gap-2">
        <button class="text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300">
          <ChevronDownIcon v-if="collapsed" class="w-4 h-4" />
          <ChevronUpIcon v-else class="w-4 h-4" />
        </button>
        <button
          @click.stop="refresh"
          class="text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300 p-0.5"
          :class="{ 'animate-spin': loading }"
          title="Refresh traces"
        >
          <ArrowPathIcon class="w-4 h-4" />
        </button>
      </div>
    </div>

    <!-- Content -->
    <div v-if="!collapsed" class="h-full overflow-hidden relative">
      <!-- Error Overlay -->
      <div
        v-if="telemetryError"
        class="absolute inset-0 z-10 flex items-center justify-center bg-gray-100/60 dark:bg-gray-900/60 backdrop-blur-sm"
      >
        <div class="flex flex-col items-center gap-2 p-4 max-w-md text-center">
          <ExclamationTriangleIcon class="w-8 h-8 text-amber-500" />
          <p class="text-sm text-gray-900 dark:text-white font-medium">{{ telemetryError }}</p>
          <button
            @click="clearError"
            class="mt-2 px-3 py-1 text-xs text-gray-700 dark:text-white bg-gray-200 dark:bg-gray-700 hover:bg-gray-300 dark:hover:bg-gray-600 rounded transition-colors"
          >
            Dismiss
          </button>
        </div>
      </div>

      <div class="h-full text-gray-700 dark:text-gray-300">
        <!-- Loading state - only show before initial load completes -->
        <div v-if="!initialLoadDone && loading" class="text-center p-4 text-xs font-mono text-gray-500">
          Loading traces...
        </div>

        <!-- No data message - only show after initial load -->
        <div v-else-if="initialLoadDone && !hasData" class="text-center p-4 text-xs font-mono text-gray-500">
          No telemetry data available. Traces will appear here when flows are executed.
        </div>

        <!-- Trace list -->
        <div v-else class="flex flex-col h-full max-h-48 overflow-y-auto bg-gray-100/50 dark:bg-gray-800/50 font-mono">
          <table class="w-full">
            <thead class="sticky top-0 bg-gray-200 dark:bg-gray-700 text-xs">
              <tr>
                <th class="px-2 py-1 text-left font-medium text-gray-700 dark:text-gray-300">Trace ID</th>
                <th class="px-2 py-1 text-left font-medium text-gray-700 dark:text-gray-300">Time</th>
                <th class="px-2 py-1 text-left font-medium text-gray-700 dark:text-gray-300">Age</th>
                <th class="px-2 py-1 text-left font-medium text-gray-700 dark:text-gray-300">Duration</th>
                <th class="px-2 py-1 text-left font-medium text-gray-700 dark:text-gray-300">Spans</th>
                <th class="px-2 py-1 text-left font-medium text-gray-700 dark:text-gray-300">Errors</th>
                <th class="px-2 py-1 text-left font-medium text-gray-700 dark:text-gray-300">Size</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="trace in traces"
                :key="trace.id"
                :class="[
                  'text-left text-xs hover:bg-gray-200 dark:hover:bg-gray-700 cursor-pointer transition-colors',
                  selectedTraceId === trace.id ? 'bg-sky-100 dark:bg-sky-900/50' : ''
                ]"
                @click="selectTrace(trace)"
              >
                <td class="px-2 py-1">
                  <span :class="[
                    'font-mono',
                    selectedTraceId === trace.id ? 'text-sky-600 dark:text-sky-300' : 'text-gray-700 dark:text-gray-300'
                  ]">
                    {{ trace.id.substring(0, 16) }}...
                  </span>
                </td>
                <td class="px-2 py-1 text-gray-500 dark:text-gray-400">
                  {{ formatTime(trace.start) }}
                </td>
                <td class="px-2 py-1 text-gray-500 dark:text-gray-400">
                  {{ formatRelativeTime(trace.end) }}
                </td>
                <td class="px-2 py-1 text-gray-700 dark:text-gray-300">
                  {{ formatDuration(trace.duration) }}
                </td>
                <td class="px-2 py-1">
                  <span class="px-1.5 py-0.5 bg-sky-500 dark:bg-sky-700 text-white dark:text-sky-100 rounded text-xs">
                    {{ trace.spans }}
                  </span>
                </td>
                <td class="px-2 py-1">
                  <span
                    v-if="trace.errors > 0"
                    class="px-1.5 py-0.5 bg-red-500 dark:bg-red-700 text-white dark:text-red-100 rounded text-xs"
                  >
                    {{ trace.errors }}
                  </span>
                  <span v-else class="text-gray-400 dark:text-gray-500">-</span>
                </td>
                <td class="px-2 py-1 text-gray-500 dark:text-gray-400">
                  {{ trace.length }} B
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </div>
</template>
