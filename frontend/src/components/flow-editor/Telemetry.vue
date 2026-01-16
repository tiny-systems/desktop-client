<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useFlowStore } from '../../stores/flow'
import { ChevronUpIcon, ChevronDownIcon } from '@heroicons/vue/24/outline'

const props = defineProps({
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
const selectedTraceId = ref(null)

// Extract trace data from edges in the flow
const traces = computed(() => {
  const traceMap = new Map()

  // Collect trace data from edges
  for (const element of flowStore.elements) {
    if (element.type === 'tinyEdge' && element.data?.trace) {
      const trace = element.data.trace
      if (trace.traceID) {
        if (!traceMap.has(trace.traceID)) {
          traceMap.set(trace.traceID, {
            id: trace.traceID,
            start: trace.start || Date.now(),
            end: trace.end || Date.now(),
            duration: trace.latency || 0,
            spans: 1,
            errors: trace.error ? 1 : 0,
            length: trace.dataSize || 0
          })
        } else {
          const existing = traceMap.get(trace.traceID)
          existing.spans++
          existing.duration = Math.max(existing.duration, trace.latency || 0)
          if (trace.error) existing.errors++
          existing.length += trace.dataSize || 0
        }
      }
    }
  }

  // Convert to array and sort by start time descending
  return Array.from(traceMap.values()).sort((a, b) => b.start - a.start).slice(0, 50)
})

const hasData = computed(() => traces.value.length > 0)

// Format duration from nanoseconds to readable string
const formatDuration = (ns) => {
  if (typeof ns !== 'number') return '0ms'
  const ms = ns / 1000000
  if (ms < 1) return '<1ms'
  if (ms < 1000) return `${Math.round(ms)}ms`
  return `${(ms / 1000).toFixed(2)}s`
}

// Format time from timestamp
const formatTime = (timestamp) => {
  if (!timestamp) return '-'
  const date = new Date(timestamp)
  const opts = {
    year: 'numeric',
    month: 'numeric',
    day: 'numeric',
    hour: 'numeric',
    minute: 'numeric',
    second: 'numeric',
    hour12: false
  }
  return new Intl.DateTimeFormat('default', opts).format(date)
}

// Format relative time
const formatRelativeTime = (timestamp) => {
  if (!timestamp) return '-'
  const now = Date.now()
  const diff = now - timestamp

  if (diff < 1000) return 'just now'
  if (diff < 60000) return `${Math.floor(diff / 1000)}s ago`
  if (diff < 3600000) return `${Math.floor(diff / 60000)}m ago`
  if (diff < 86400000) return `${Math.floor(diff / 3600000)}h ago`
  return `${Math.floor(diff / 86400000)}d ago`
}

const selectTrace = (trace) => {
  selectedTraceId.value = trace.id
  emit('trace', trace.id)
}

const toggleCollapsed = () => {
  collapsed.value = !collapsed.value
}
</script>

<template>
  <div
    :class="[
      'w-full border-t border-gray-200 dark:border-gray-700 text-sm relative bg-white dark:bg-gray-900',
      hasData && !collapsed ? 'min-h-48 h-1/4' : ''
    ]"
  >
    <!-- Header with collapse toggle -->
    <div
      class="flex items-center justify-between px-3 py-1.5 bg-gray-50 dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700 cursor-pointer"
      @click="toggleCollapsed"
    >
      <span class="text-xs font-medium text-gray-600 dark:text-gray-400">
        Telemetry
        <span v-if="traces.length > 0" class="ml-1 text-gray-400 dark:text-gray-500">
          ({{ traces.length }} traces)
        </span>
      </span>
      <button class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300">
        <ChevronDownIcon v-if="collapsed" class="w-4 h-4" />
        <ChevronUpIcon v-else class="w-4 h-4" />
      </button>
    </div>

    <!-- Content -->
    <div v-if="!collapsed" class="h-full overflow-hidden">
      <div class="h-full dark:text-gray-300">
        <!-- No data message -->
        <div v-if="!hasData" class="text-center p-4 text-xs font-mono text-gray-500 dark:text-gray-400">
          No telemetry data available. Traces will appear here when flows are executed.
        </div>

        <!-- Trace list -->
        <div v-else class="flex flex-col h-full max-h-48 overflow-y-auto bg-gray-50/50 dark:bg-gray-800/50 font-mono">
          <table class="w-full">
            <thead class="sticky top-0 bg-gray-100 dark:bg-gray-700 text-xs">
              <tr>
                <th class="px-2 py-1 text-left font-medium text-gray-600 dark:text-gray-300">ID</th>
                <th class="px-2 py-1 text-left font-medium text-gray-600 dark:text-gray-300">Start</th>
                <th class="px-2 py-1 text-left font-medium text-gray-600 dark:text-gray-300">End</th>
                <th class="px-2 py-1 text-left font-medium text-gray-600 dark:text-gray-300">Duration</th>
                <th class="px-2 py-1 text-left font-medium text-gray-600 dark:text-gray-300">Spans</th>
                <th class="px-2 py-1 text-left font-medium text-gray-600 dark:text-gray-300">Errors</th>
                <th class="px-2 py-1 text-left font-medium text-gray-600 dark:text-gray-300">Size</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="trace in traces"
                :key="trace.id"
                :class="[
                  'text-left text-xs hover:bg-gray-200 dark:hover:bg-gray-700 cursor-pointer transition-colors',
                  selectedTraceId === trace.id ? 'bg-gray-300 dark:bg-gray-600' : ''
                ]"
                @click="selectTrace(trace)"
              >
                <td class="px-2 py-1">
                  <span class="text-gray-700 dark:text-gray-300 truncate max-w-24 block" :title="trace.id">
                    {{ trace.id.slice(0, 8) }}...
                  </span>
                </td>
                <td class="px-2 py-1 text-gray-600 dark:text-gray-400">
                  {{ formatTime(trace.start) }}
                </td>
                <td class="px-2 py-1 text-gray-600 dark:text-gray-400">
                  {{ formatRelativeTime(trace.end) }}
                </td>
                <td class="px-2 py-1 text-gray-700 dark:text-gray-300">
                  {{ formatDuration(trace.duration) }}
                </td>
                <td class="px-2 py-1">
                  <span class="px-1.5 py-0.5 bg-sky-200 dark:bg-sky-700 text-sky-800 dark:text-sky-100 rounded text-xs">
                    {{ trace.spans }}
                  </span>
                </td>
                <td class="px-2 py-1">
                  <span
                    v-if="trace.errors > 0"
                    class="px-1.5 py-0.5 bg-red-200 dark:bg-red-700 text-red-800 dark:text-red-100 rounded text-xs"
                  >
                    {{ trace.errors }}
                  </span>
                  <span v-else class="text-gray-400">-</span>
                </td>
                <td class="px-2 py-1 text-gray-600 dark:text-gray-400">
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
