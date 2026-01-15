<script setup>
import { BaseEdge, getBezierPath } from '@vue-flow/core'
import { computed } from 'vue'

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

const path = computed(() => getBezierPath(props))
</script>

<template>
  <BaseEdge :id="id" :style="style" :path="path[0]" :marker-end="markerEnd"/>
</template>
