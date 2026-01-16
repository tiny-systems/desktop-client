<script setup>
import { ref, computed, watch, onMounted, nextTick } from 'vue'
import { VueFlow, useVueFlow, MarkerType } from '@vue-flow/core'
import { Background } from '@vue-flow/background'
import { Controls } from '@vue-flow/controls'
import { MiniMap } from '@vue-flow/minimap'
import { useFlowStore } from '../../stores/flow'
import TinyNode from '../flow/TinyNode.vue'
import TinyEdge from '../flow/TinyEdge.vue'
import { PlusIcon } from '@heroicons/vue/24/outline'
import { debounce } from 'lodash'

const emit = defineEmits(['error', 'add-node'])

const flowStore = useFlowStore()

const {
  onConnect,
  onNodesChange,
  onEdgesChange,
  onNodeDragStop,
  onPaneReady,
  fitView,
  setViewport,
  getViewport
} = useVueFlow()

// Track dragged nodes for batch position update
const draggedNodes = ref(new Map())

// Computed elements from store
const nodes = computed(() => flowStore.nodes)
const edges = computed(() => flowStore.edges)

// Default edge options - use very light gray to match preview
const defaultEdgeOptions = {
  type: 'tinyEdge',
  markerEnd: {
    type: MarkerType.ArrowClosed,
    width: 20,
    height: 20,
    color: '#d1d5db',
  },
  style: {
    stroke: '#d1d5db',
    strokeWidth: 1,
  },
}

// Handle new connections
onConnect(async (params) => {
  try {
    await flowStore.connectNodes(
      params.source,
      params.sourceHandle,
      params.target,
      params.targetHandle
    )
  } catch (err) {
    emit('error', `Failed to connect nodes: ${err}`)
  }
})

// Track node drag for position updates
onNodeDragStop((event) => {
  const { node } = event
  if (node) {
    draggedNodes.value.set(node.id, { x: node.position.x, y: node.position.y })
    debouncedSavePositions()
  }
})

// Debounced position save
const debouncedSavePositions = debounce(async () => {
  if (draggedNodes.value.size === 0) return

  const positions = {}
  draggedNodes.value.forEach((pos, id) => {
    positions[id] = pos
  })

  try {
    await flowStore.batchUpdateNodePositions(positions)
  } catch (err) {
    console.error('Failed to save positions:', err)
  }

  draggedNodes.value.clear()
}, 500)

// Save viewport when it changes
const debouncedSaveMeta = debounce(async () => {
  const viewport = getViewport()
  if (viewport) {
    flowStore.setMeta({
      x: viewport.x,
      y: viewport.y,
      zoom: viewport.zoom
    })
    await flowStore.saveMeta()
  }
}, 1000)

// Handle pane ready - restore viewport
onPaneReady(() => {
  nextTick(() => {
    if (flowStore.meta?.x !== undefined && flowStore.meta?.y !== undefined && flowStore.meta?.zoom !== undefined) {
      setViewport({
        x: flowStore.meta.x,
        y: flowStore.meta.y,
        zoom: flowStore.meta.zoom
      })
    } else {
      fitView({ padding: 0.2 })
    }
  })
})

// Handle node/edge selection
const handleNodeClick = (event) => {
  const nodeId = event.node?.id
  if (nodeId) {
    flowStore.select(nodeId)
  }
}

const handleEdgeClick = (event) => {
  const edgeId = event.edge?.id
  if (edgeId) {
    flowStore.select(edgeId)
  }
}

const handlePaneClick = () => {
  flowStore.selectElement(null)
}

// Handle move end (viewport change)
const handleMoveEnd = () => {
  debouncedSaveMeta()
}

// Handle double click on pane to add node
const handlePaneDoubleClick = (event) => {
  // Get the position in flow coordinates from the event
  // The event.event contains the native MouseEvent
  const mouseEvent = event.event
  if (!mouseEvent) return

  // Get canvas bounds and calculate position
  const bounds = mouseEvent.target.getBoundingClientRect()
  const viewport = getViewport()

  // Convert screen coordinates to flow coordinates
  const x = (mouseEvent.clientX - bounds.left - viewport.x) / viewport.zoom
  const y = (mouseEvent.clientY - bounds.top - viewport.y) / viewport.zoom

  emit('add-node', { x: Math.round(x), y: Math.round(y) })
}

// Handle click on floating plus button - add at center of viewport
const handleAddClick = () => {
  const viewport = getViewport()
  // Calculate center of visible area in flow coordinates
  // Assuming canvas is roughly 800x600 visible area
  const x = (400 - viewport.x) / viewport.zoom
  const y = (200 - viewport.y) / viewport.zoom
  emit('add-node', { x: Math.round(x), y: Math.round(y) })
}
</script>

<template>
  <div class="flow-canvas w-full h-full relative">
    <!-- Floating Add Component Button -->
    <button
      @click="handleAddClick"
      title="Add component"
      class="absolute top-2 right-2 z-10 w-6 h-6 flex items-center justify-center rounded-full border border-sky-500 text-sky-500 bg-white dark:bg-gray-900 hover:bg-sky-50 dark:hover:bg-sky-900/30 transition-colors"
    >
      <PlusIcon class="w-4 h-4" />
    </button>

    <VueFlow
      :nodes="nodes"
      :edges="edges"
      :default-edge-options="defaultEdgeOptions"
      :connection-mode="'strict'"
      :snap-to-grid="true"
      :snap-grid="[15, 15]"
      @node-click="handleNodeClick"
      @edge-click="handleEdgeClick"
      @pane-click="handlePaneClick"
      @move-end="handleMoveEnd"
      @dblclick="handlePaneDoubleClick"
      class="bg-gray-50 dark:bg-gray-950"
    >
      <!-- Custom node types -->
      <template #node-tinyNode="nodeProps">
        <TinyNode
          v-bind="nodeProps"
          :editor-mode="true"
        />
      </template>

      <!-- Custom edge types -->
      <template #edge-tinyEdge="edgeProps">
        <TinyEdge
          v-bind="edgeProps"
          :editor-mode="true"
        />
      </template>

      <!-- Background -->
      <Background
        :variant="'dots'"
        :gap="20"
        :size="1"
        class="dark:opacity-30"
      />

      <!-- Controls -->
      <Controls
        :show-zoom="true"
        :show-fit-view="true"
        :show-interactive="false"
        position="bottom-left"
        class="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg"
      />

      <!-- MiniMap -->
      <MiniMap
        :pannable="true"
        :zoomable="true"
        position="bottom-right"
        class="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg"
      />
    </VueFlow>
  </div>
</template>

<style>
/* Vue Flow styles */
@import '@vue-flow/core/dist/style.css';
@import '@vue-flow/core/dist/theme-default.css';
@import '@vue-flow/controls/dist/style.css';
@import '@vue-flow/minimap/dist/style.css';

/* Handle label styling - critical for proper label display */
.vue-flow__handle {
  font-size: small;
}

/* Node styling */
.vue-flow__node {
  color: #555;
}

.vue-flow__node.selected {
  background-color: #0284c7;
}

/* Edge styling - light gray to match preview */
.vue-flow__edge path,
.vue-flow__edge-path {
  stroke-width: 1px !important;
  stroke: #d1d5db !important;
}

/* Animated edge styles - dashed stroke animation */
.vue-flow__edge.animated path,
.vue-flow__edge.animated .vue-flow__edge-path {
  stroke-dasharray: 5 !important;
  animation: dashdraw 0.5s linear infinite !important;
}

@keyframes dashdraw {
  from {
    stroke-dashoffset: 10;
  }
  to {
    stroke-dashoffset: 0;
  }
}

/* Selected edge styling */
.vue-flow__edge.selected path,
.vue-flow__edge.selected .vue-flow__edge-path {
  stroke: #0ea5e9 !important;
  stroke-width: 2px !important;
}

/* Dark mode overrides */
.dark .vue-flow__minimap {
  background-color: #333;
}

.dark .vue-flow__minimap rect {
  fill: #222;
}

.dark .vue-flow__minimap-mask {
  fill: #111;
}

.dark .vue-flow__node {
  background: #222;
  border-color: #333;
  border-width: 1px;
  color: #aaa;
}

.dark .vue-flow__node.selected {
  background-color: #075985;
  color: #fff;
  border-color: #075985;
}

.dark .vue-flow__background {
  opacity: 0.3;
}

.dark .vue-flow__edge-path {
  stroke: #4b5563 !important;
}

.dark .vue-flow__edge button {
  color: #666;
}

.dark .vue-flow__controls {
  background: #1f2937;
  border-color: #374151;
}

.dark .vue-flow__controls-button {
  background: #1f2937;
  border-color: #374151;
  color: #9ca3af;
}

.dark .vue-flow__controls-button:hover {
  background: #374151;
}

.dark .vue-flow__controls-button:last-child {
  border: none;
}
</style>
