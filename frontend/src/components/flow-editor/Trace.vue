<script setup>
import { ref, computed, watch } from 'vue'
import VueJsonPretty from 'vue-json-pretty'
import 'vue-json-pretty/lib/styles.css'
import { GetTraceByID } from '../../../wailsjs/go/main/App'

const props = defineProps({
  trace: {
    type: String,
    required: true
  },
  ctx: {
    type: String,
    required: true
  },
  ns: {
    type: String,
    required: true
  },
  projectName: {
    type: String,
    required: true
  }
})

const emit = defineEmits(['close'])

const expanded = ref(false)
const loading = ref(false)
const traceData = ref(null)
const error = ref('')
const copied = ref(false)

const formattedTraceData = computed(() => {
  if (!traceData.value?.spans) return {}

  return {
    traceId: traceData.value.traceId,
    spans: traceData.value.spans.map((span) => ({
      name: span.name,
      spanId: span.span_id,
      duration: span.end_time_unix_nano && span.start_time_unix_nano
        ? `${((span.end_time_unix_nano - span.start_time_unix_nano) / 1000000).toFixed(2)}ms`
        : null,
      attributes: span.attributes?.reduce((acc, attr) => {
        acc[attr.key] = attr.value
        return acc
      }, {}) || {},
      events: span.events?.map((e) => ({
        name: e.name,
        attributes: e.attributes?.reduce((acc, attr) => {
          acc[attr.key] = attr.value
          return acc
        }, {}) || {}
      })) || []
    }))
  }
})

async function copyTrace() {
  if (!formattedTraceData.value) return

  try {
    await navigator.clipboard.writeText(JSON.stringify(formattedTraceData.value, null, 2))
    copied.value = true
    setTimeout(() => {
      copied.value = false
    }, 2000)
  } catch (e) {
    console.error('Failed to copy trace:', e)
  }
}

async function fetchTraceData() {
  if (!props.trace || !props.projectName) return

  loading.value = true
  error.value = ''

  try {
    const response = await GetTraceByID(props.ctx, props.ns, props.projectName, props.trace)
    traceData.value = response
  } catch (e) {
    error.value = e.message || 'Failed to load trace'
    console.error('Failed to fetch trace:', e)
  } finally {
    loading.value = false
  }
}

function toggleExpand() {
  expanded.value = !expanded.value
  if (expanded.value && !traceData.value) {
    fetchTraceData()
  }
}

function resetTrace() {
  expanded.value = false
  traceData.value = null
  emit('close')
}

// Reset state when trace changes
watch(() => props.trace, () => {
  traceData.value = null
  error.value = ''
  if (expanded.value) {
    fetchTraceData()
  }
})
</script>

<template>
  <div class="absolute z-50 top-2 left-2 text-white text-sm" v-if="props.trace">
    <!-- Collapsed header -->
    <div class="border border-gray-700 rounded bg-sky-600 p-1 px-2 flex items-center gap-2">
      <a href="#" @click.prevent="toggleExpand" class="hover:underline font-mono text-xs">
        trace#{{ props.trace.substring(0, 16) }}...
      </a>
      <span v-if="loading" class="text-xs opacity-70">loading...</span>
      <a href="#" @click.prevent="resetTrace" class="px-1 hover:bg-sky-500 rounded">&times;</a>
    </div>

    <!-- Expanded content -->
    <div v-if="expanded" class="mt-1 bg-gray-900 border border-gray-700 rounded shadow-lg max-w-3xl max-h-[32rem] overflow-auto">
      <div class="sticky top-0 z-10 bg-gray-800 px-3 py-2 border-b border-gray-700 flex justify-between items-center gap-4">
        <span class="text-gray-300 font-mono text-xs truncate">{{ traceData?.traceId || props.trace }}</span>
        <div class="flex items-center gap-3 flex-shrink-0">
          <span class="text-gray-500 text-xs whitespace-nowrap">{{ traceData?.spans?.length || 0 }} spans</span>
          <button
            @click="copyTrace"
            class="text-gray-400 hover:text-gray-200 text-xs px-1.5 py-0.5 rounded border border-gray-600 hover:bg-gray-700 transition-colors"
            :title="copied ? 'Copied!' : 'Copy trace JSON'"
          >
            {{ copied ? 'Copied!' : 'Copy' }}
          </button>
        </div>
      </div>
      <div class="p-2" v-if="traceData">
        <VueJsonPretty :data="formattedTraceData" theme="dark" :deep="2" />
      </div>
      <div v-else-if="error" class="p-3 text-red-400 text-xs">
        {{ error }}
      </div>
      <div v-else class="p-3 text-gray-500 text-xs">
        No trace data available
      </div>
    </div>
  </div>
</template>

<style scoped>
:deep(.vjs-tree) {
  font-size: 10px !important;
  line-height: 1.4 !important;
}
</style>
