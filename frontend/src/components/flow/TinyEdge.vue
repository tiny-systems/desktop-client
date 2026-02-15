<script setup>
import { BaseEdge, getBezierPath, EdgeLabelRenderer } from '@vue-flow/core'
import { PencilSquareIcon } from '@heroicons/vue/24/outline'
import { computed } from 'vue'
import { useFlowStore } from '../../stores/flow'

defineOptions({
  inheritAttrs: false
})

const props = defineProps({
  disabled: {
    type: Boolean,
  },
  noConfigure: {
    type: Boolean
  },
  selected: {
    type: Boolean
  },
  editorMode: {
    type: Boolean
  },
  id: {
    type: String,
    required: true,
  },
  sourceX: {
    type: Number,
    required: true,
  },
  sourceY: {
    type: Number,
    required: true,
  },
  targetX: {
    type: Number,
    required: true,
  },
  targetY: {
    type: Number,
    required: true,
  },
  sourcePosition: {
    type: String,
    required: true,
  },
  targetPosition: {
    type: String,
    required: true,
  },
  data: {
    type: Object,
    required: false,
  },
  markerEnd: {
    type: String,
    required: false,
  },
  curvature: {
    type: Number,
    required: false
  },
  style: {
    type: Object,
    required: false,
  },
})

const flowStore = useFlowStore()

// Handle edge selection
const handleEdgeSelect = () => {
  flowStore.select(props.id)
}

const path = computed(() => getBezierPath({
  sourceX: props.sourceX,
  sourceY: props.sourceY,
  targetX: props.targetX,
  targetY: props.targetY,
  sourcePosition: props.sourcePosition,
  targetPosition: props.targetPosition,
  curvature: props.curvature,
}))

// Compute edge style based on validation and trace state - matching platform
const edgeStyle = computed(() => {
  const baseStyle = { ...(props.style || {}) }

  // Validation error - red
  if (props.data?.valid === false) {
    baseStyle.stroke = '#fca5a5'
    return baseStyle
  }

  // Trace runtime error - red
  if (props.data?.trace?.error) {
    baseStyle.stroke = '#fca5a5'
    return baseStyle
  }

  // Was part of trace execution (has sequence) - blue
  if (props.data?.trace?.sequence !== undefined && props.data.trace.sequence >= 0) {
    baseStyle.stroke = '#00bfff'
    return baseStyle
  }

  // Default - no explicit stroke, let CSS theme handle it
  return baseStyle
})

// Icon class based on state - matches platform implementation exactly
const iconClass = computed(() => {
  // Invalid edge - red
  if (props.data?.valid === false) {
    return 'fill-red-500 stroke-red-200 dark:fill-red-700 dark:stroke-red-300 dark:opacity-70'
  }
  // Selected edge - sky blue
  if (props.selected) {
    return 'fill-sky-500 stroke-sky-200 dark:fill-sky-700 dark:stroke-sky-300 dark:opacity-70'
  }
  // Default - light gray to match platform
  return 'fill-gray-200 text-gray-400 dark:text-gray-300 dark:fill-gray-900 dark:opacity-40'
})

// Format milliseconds to readable time
const msToTime = (ms) => {
  if (ms < 1) return '<1ms'
  if (ms < 1000) return `${Math.round(ms)}ms`
  return `${(ms / 1000).toFixed(2)}s`
}
</script>

<template>
  <BaseEdge
    :id="id"
    :style="edgeStyle"
    :path="path[0]"
    :marker-end="markerEnd"
  />

  <!-- Edit button and trace info at edge midpoint -->
  <EdgeLabelRenderer>
    <div
      v-if="!data?.blocked"
      :style="{
        pointerEvents: 'all',
        position: 'absolute',
        color: 'black',
        textAlign: 'center',
        transform: `translate(-50%, -50%) translate(${path[1]}px, ${path[2]}px)`,
      }"
      class="nodrag nopan"
    >
      <!-- Edit button - shows when noConfigure is not set -->
      <button
        v-if="!noConfigure"
        v-tooltip="data?.valid === false && data?.error ? data.error : 'Configure'"
        @click="handleEdgeSelect"
      >
        <PencilSquareIcon
          :class="['w-5 h-5', iconClass]"
        />
      </button>

      <!-- Trace latency display -->
      <div
        v-if="data?.trace && data.trace.sequence !== undefined && data.trace.sequence > 0"
        class="text-xs w-full text-center text-sky-500"
        :title="'Span# ' + data.trace.sequence"
      >
        {{ msToTime(data.trace.latency || 0) }}
      </div>
    </div>
  </EdgeLabelRenderer>
</template>
