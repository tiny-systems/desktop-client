<script setup>
import { ref, nextTick, onMounted, onUnmounted } from 'vue'
import { VueFlow, useVueFlow, MarkerType } from '@vue-flow/core'
import { Background } from '@vue-flow/background'
import { Controls, ControlButton } from '@vue-flow/controls'
import { MiniMap } from '@vue-flow/minimap'
import { useFlowStore } from '../../stores/flow'
import { useLayout } from '../../composables/useLayout'
import TinyNode from '../flow/TinyNode.vue'
import TinyEdge from '../flow/TinyEdge.vue'
import { PlusIcon, ArrowPathIcon, Squares2X2Icon } from '@heroicons/vue/24/outline'
import { LockClosedIcon, LockOpenIcon } from '@heroicons/vue/24/solid'
import { debounce } from 'lodash'

const emit = defineEmits(['error', 'add-node', 'delete-node', 'delete-edge'])

const flowStore = useFlowStore()
flowStore.initReadOnly()
const { layout } = useLayout()

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

// Handle Delete key press - show confirmation dialog instead of deleting directly
const handleKeyDown = (event) => {
  if (event.key === 'Delete' || event.key === 'Backspace') {
    // Ignore if typing in an input field
    const target = event.target
    const tagName = target?.tagName?.toLowerCase()
    if (tagName === 'input' || tagName === 'textarea' || target?.isContentEditable) {
      return // Let the input handle the keypress
    }

    if (flowStore.readOnly) return

    // Prevent default behavior
    event.preventDefault()

    // Check if a node is selected
    const selectedNode = flowStore.selectedNode
    if (selectedNode) {
      emit('delete-node', selectedNode)
      return
    }

    // Check if an edge is selected
    const selectedEdge = flowStore.selectedEdge
    if (selectedEdge) {
      emit('delete-edge', selectedEdge)
      return
    }
  }
}

// Add keyboard listener on mount
onMounted(() => {
  window.addEventListener('keydown', handleKeyDown)
})

// Remove keyboard listener on unmount
onUnmounted(() => {
  window.removeEventListener('keydown', handleKeyDown)
})

// Default edge options - let CSS handle colors for proper dark mode support
const defaultEdgeOptions = {
  type: 'tinyEdge',
  markerEnd: {
    type: MarkerType.ArrowClosed,
    width: 20,
    height: 20,
  },
}

// Handle new connections
onConnect(async (params) => {
  if (flowStore.readOnly) return
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
  if (flowStore.readOnly) return
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
      // Use lower default zoom (maxZoom limits how much fitView zooms in)
      fitView({ padding: 0.3, maxZoom: 0.8 })
    }
  })
})

// Handle node/edge selection
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

// Handle rotate node
const handleRotate = async () => {
  if (!flowStore.selectedNode) return
  try {
    await flowStore.rotateNode(flowStore.selectedNode.id)
  } catch (err) {
    emit('error', `Failed to rotate node: ${err}`)
  }
}

// Handle auto-layout
const handleAutoLayout = async () => {
  if (flowStore.nodes.length === 0) return

  try {
    // Apply dagre layout algorithm
    const layoutedNodes = layout(flowStore.nodes, flowStore.edges, 'LR')

    // Collect position updates
    const positions = {}
    layoutedNodes.forEach(node => {
      positions[node.id] = { x: node.position.x, y: node.position.y }
    })

    // Batch update positions to backend
    await flowStore.batchUpdateNodePositions(positions)

    // Fit view after layout
    nextTick(() => {
      fitView({ padding: 0.3, maxZoom: 0.8 })
    })
  } catch (err) {
    emit('error', `Failed to auto-layout: ${err}`)
  }
}

</script>

<template>
  <div class="flow-canvas w-full h-full relative">
    <!-- Read-only toggle -->
    <button
      @click="flowStore.toggleReadOnly()"
      :title="flowStore.readOnly ? 'Switch to editing mode' : 'Switch to read-only mode'"
      :class="[
        flowStore.readOnly
          ? 'bg-amber-100 text-amber-700 border-amber-300 dark:bg-amber-900/40 dark:text-amber-400 dark:border-amber-700'
          : 'bg-green-100 text-green-700 border-green-300 dark:bg-green-900/40 dark:text-green-400 dark:border-green-700'
      ]"
      class="absolute top-2 right-12 z-10 flex items-center gap-1 px-2.5 py-1 rounded-full border text-[11px] font-semibold"
    >
      <LockClosedIcon v-if="flowStore.readOnly" class="w-3.5 h-3.5" />
      <LockOpenIcon v-else class="w-3.5 h-3.5" />
      {{ flowStore.readOnly ? 'READ ONLY' : 'EDITING' }}
    </button>

    <!-- Floating Add Component Button -->
    <button
      v-if="!flowStore.readOnly"
      @click="handleAddClick"
      title="Add component"
      class="absolute top-2 right-2 z-10 w-6 h-6 flex items-center justify-center rounded-full border border-sky-500 text-sky-500 bg-white dark:bg-gray-900 hover:bg-sky-50 dark:hover:bg-sky-900/30 transition-colors"
    >
      <PlusIcon class="w-4 h-4" />
    </button>

    <VueFlow
      v-model="flowStore.elements"
      :default-edge-options="defaultEdgeOptions"
      :connection-mode="'strict'"
      :min-zoom="0.5"
      :max-zoom="1"
      :delete-key-code="null"
      :nodes-draggable="!flowStore.readOnly"
      :nodes-connectable="!flowStore.readOnly"
      @edge-click="handleEdgeClick"
      @pane-click="handlePaneClick"
      @move-end="handleMoveEnd"
      @dblclick="handlePaneDoubleClick"
      class="bg-gray-50 dark:bg-black"
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
      >
        <template #top>
          <ControlButton
            v-if="flowStore.selectedNode && !flowStore.readOnly"
            title="Rotate node"
            @click="handleRotate"
          >
            <ArrowPathIcon class="w-4 h-4" />
          </ControlButton>
        </template>
      </Controls>

      <!-- MiniMap -->
      <MiniMap
        :pannable="true"
        :zoomable="true"
        position="bottom-right"
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

/* Handle styling */
.vue-flow__handle {
  font-size: small;
}

/* Node styling */
.vue-flow__node {
  background: transparent;
  border: none;
  border-radius: 10px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08), 0 1px 2px rgba(0, 0, 0, 0.06);
}

.vue-flow__node.selected {
  background: transparent;
  box-shadow: 0 4px 12px rgba(14, 165, 233, 0.15), 0 1px 3px rgba(0, 0, 0, 0.08);
}

/* Edge styling - match platform */
.vue-flow__edge path {
  stroke-width: 2px;
  stroke: #ccc;
}

/* Marker (arrow) fill color */
.vue-flow__marker {
  fill: #ccc;
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

/* Dark mode overrides using prefers-color-scheme */
@media (prefers-color-scheme: dark) {
  /* MiniMap styles */
  .vue-flow__minimap {
    background-color: #333 !important;
  }

  .vue-flow__minimap svg {
    background-color: #333 !important;
  }

  .vue-flow__minimap rect {
    fill: #222 !important;
  }

  .vue-flow__minimap-node {
    fill: #60a5fa !important;
    stroke: #3b82f6 !important;
    stroke-width: 1px !important;
  }

  .vue-flow__minimap-mask {
    fill: #111 !important;
    opacity: 0.6 !important;
  }

  /* Node styles */
  .vue-flow__node {
    background: transparent !important;
    border: none !important;
    box-shadow: 0 1px 4px rgba(0, 0, 0, 0.3);
  }

  .vue-flow__node.selected {
    background: transparent !important;
    box-shadow: 0 4px 12px rgba(14, 165, 233, 0.2), 0 1px 3px rgba(0, 0, 0, 0.3);
  }

  /* Background */
  .vue-flow__background {
    opacity: 0.3;
  }

  /* Edge styles - no !important so inline trace/error stroke colors can override */
  .vue-flow__edge-path,
  .vue-flow__edge path {
    stroke: #555;
  }

  .vue-flow__edge button {
    color: #666 !important;
  }

  /* Marker (arrow) fill color for dark mode */
  .vue-flow__marker {
    fill: #555 !important;
  }

  /* Controls styles */
  .vue-flow__controls {
    background-color: #333 !important;
    border-color: #444 !important;
  }

  .vue-flow__controls-button {
    background-color: #333 !important;
    color: #ccc !important;
    border-color: #444 !important;
    fill: #ccc !important;
  }

  .vue-flow__controls-button svg {
    fill: #ccc !important;
  }

  .vue-flow__controls-button:hover {
    background-color: #555 !important;
  }

  .vue-flow__controls-button:hover svg {
    fill: #fff !important;
  }
}
</style>
