<script lang="ts" setup>
defineOptions({
  inheritAttrs: false
})

import { ref, watch } from 'vue'
import type { CSSProperties } from 'vue'
import { Handle, Position } from '@vue-flow/core'

interface HandleProps {
  id?: string
  position?: Number
  virtual?: Boolean
  rotated_position: Number
  type?: string
  label?: string
  error?: string
}

interface NodeData {
  handles: HandleProps[]
  blocked: boolean
  label?: string
}

const emits = defineEmits(['updateNodeInternals'])

const calculateHandlerClass = (h: HandleProps, nodeData: NodeData) => {
  if (h.error) {
    return 'bg-red-300 dark:border-gray-400 dark:bg-red-800'
  }
  if (nodeData?.blocked) {
    return 'bg-gray-100 dark:border-gray-400 dark:bg-gray-600'
  }
  return 'bg-gray-300 dark:border-gray-400 dark:bg-gray-700'
}

const calculateHandlerStyle = (h: HandleProps, nodeData: NodeData) => {
  let style: CSSProperties = {
    height: '15px',
    width: '15px',
  }

  let idx = 0, i = 0
  const sameSideHandles = (nodeData?.handles || []).filter(a => a.rotated_position == h.rotated_position && !a.virtual)
  // find current handle position
  for (let ha of sameSideHandles) {
    if (h.id == ha.id) {
      idx = i
    }
    i++
  }
  const total = sameSideHandles.length || 1
  if (h.rotated_position == 1) { // right
    style.right = '0px';
    style.paddingLeft = '13px';
    style.top = (idx + 1) * 100 / (total + 1) + '%'
    style.lineHeight = '14px'
  } else if (h.rotated_position == 3) { // left
    style.left = '0px';
    style.paddingRight = '13px';
    style.direction = 'rtl';
    style.top = (idx + 1) * 100 / (total + 1) + '%'
    style.lineHeight = '13px'
  } else if (h.rotated_position == 2) { // bottom
    style.writingMode = 'vertical-rl'
    style.bottom = '0px'
    style.paddingTop = '13px'
    style.paddingLeft = '4px'
    style.lineHeight = '13px'
    style.left = (idx + 1) * 100 / (total + 1) + '%'
  } else { // top
    style.left = (idx + 1) * 100 / (total + 1) + '%'
    style.writingMode = 'vertical-lr'
    style.direction = 'rtl'
    style.paddingBottom = '13px'
    style.paddingRight = '4px'
    style.lineHeight = '14px'
  }
  return style
}

const posIntToStr = (n: Number) => {
  switch (n) {
    case 0:
      return Position.Top
    case 1:
      return Position.Right
    case 2:
      return Position.Bottom
    case 3:
      return Position.Left
  }
}

const calculateBoxStyle = (data: any): CSSProperties => {
  const leftRightElementCount = Math.max((data.handles || []).filter((a: HandleProps) => a.rotated_position == 1 && !a.virtual).length, (data.handles || []).filter((a: HandleProps) => a.rotated_position == 3 && !a.virtual).length)
  const topBottomElementCount = Math.max((data.handles || []).filter((a: HandleProps) => a.rotated_position == 0 && !a.virtual).length, (data.handles || []).filter((a: HandleProps) => a.rotated_position == 2 && !a.virtual).length)
  return {
    minWidth: ((topBottomElementCount || 1) * 30 + 70) + 'px',
    textAlign: 'center',
    padding: '14px',
    minHeight: ((leftRightElementCount || 1) * 30 + 50) + 'px',
  }
}

const props = defineProps(['data', 'selected', 'id', 'noExpire'])
watch(props, (a) => {
  emits('updateNodeInternals')
}, {
  deep: true
})

const calcBoxClass = () => {
  if (props.data?.blocked) {
    if (props.selected) {
      return 'bg-gray-300 border border-indigo-200 dark:bg-gray-800'
    }
    return 'bg-gray-100 border border-gray-300 dark:bg-gray-700'
  }

  if (props.selected) {
    return 'bg-sky-500 border border-gray-700 dark:border-gray-600 dark:bg-gray-900 text-white'
  }
  return 'bg-sky-100 border border-solid dark:border-gray-600 dark:bg-gray-800 border-sky-300'
}
</script>

<template>
  <div :style="calculateBoxStyle(props.data || {})" :class="['dark:text-gray-300 rounded', calcBoxClass()]">
    {{ props.data?.label || props.id }}
  </div>
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
        v-if="(h.type === 'source' && (h.rotated_position === 1 || h.rotated_position === 2)) ||
              (h.type === 'target' && (h.rotated_position === 3 || h.rotated_position === 0))"
      >
        {{ h.label }}
      </template>
      <template
        v-else-if="(h.type === 'target' && (h.rotated_position === 1 || h.rotated_position === 2)) ||
                   h.type === 'source'"
      >
        {{ h.label }}
      </template>
    </Handle>
  </template>
</template>
