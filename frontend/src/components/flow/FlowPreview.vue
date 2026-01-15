<template>
  <div class="flow-preview-container">
    <VueFlow :id="id" :edges="graph?.edges || []" :nodes="graph?.nodes || []" :elementsSelectable="false"
             :selectNodesOnDrag="false"
             fit-view-on-init
             :min-zoom="0.01"
             :max-zoom="1"
             :panOnDrag="false" :zoomOnScroll="false" :zoomOnPinch="false" :zoomOnDoubleClick="false"
             :preventScrolling="false" :panOnScroll="false"
             :nodesDraggable="false" :edgesUpdatable="false" :nodesConnectable="false"
             class="pointer-events-none">
      <template #node-tinyNode="props">
        <TinyNode v-bind="props" :no-expire="true"/>
      </template>
      <template #edge-tinyEdge="props">
        <TinyEdge v-bind="props" :curvature="0.4" no-configure/>
      </template>
      <slot></slot>
    </VueFlow>
  </div>
</template>
<script setup>
import TinyEdge from './TinyEdge.vue';
import TinyNode from './TinyNode.vue';
import { VueFlow } from '@vue-flow/core'

const props = defineProps({
  graph: {
    type: Object,
    default: () => ({})
  },
  id: {
    type: String,
    default: ''
  },
})
</script>

<style scoped>
.flow-preview-container {
  width: 100%;
  height: 100%;
  position: relative;
  overflow: hidden;
}

.flow-preview-container :deep(.vue-flow) {
  width: 100%;
  height: 100%;
}
</style>
