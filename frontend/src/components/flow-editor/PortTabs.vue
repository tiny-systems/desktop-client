<script setup>
import { computed } from 'vue'

const props = defineProps({
  handles: {
    type: Array,
    default: () => []
  },
  selectedHandleId: {
    type: String,
    default: null
  }
})

const emit = defineEmits(['select'])

// Filter handles - exclude those starting with underscore
const visibleHandles = computed(() => {
  return props.handles.filter(h => h.id && !h.id.startsWith('_'))
})

// Group handles by position (rotated_position: 0=top, 1=right, 2=bottom, 3=left)
const handlesByPosition = computed(() => {
  const groups = { top: [], right: [], bottom: [], left: [] }

  for (const handle of visibleHandles.value) {
    const pos = handle.rotated_position ?? 0
    if (pos === 0) groups.top.push(handle)
    else if (pos === 1) groups.right.push(handle)
    else if (pos === 2) groups.bottom.push(handle)
    else if (pos === 3) groups.left.push(handle)
  }

  return groups
})

const hasAnyHandles = computed(() => visibleHandles.value.length > 0)

const selectHandle = (handleId) => {
  emit('select', handleId)
}

const isSelected = (handleId) => {
  return props.selectedHandleId === handleId
}
</script>

<template>
  <div v-if="hasAnyHandles" class="port-tabs-container relative flex flex-col h-full">
    <!-- Top tabs row (always present for spacing) -->
    <div class="flex justify-center flex-shrink-0 tab-row-horizontal">
      <button
        v-for="handle in handlesByPosition.top"
        :key="handle.id"
        @click="selectHandle(handle.id)"
        :class="[
          'px-4 py-2 text-sm font-medium rounded-t-lg border-2 border-b-0 transition-colors',
          isSelected(handle.id)
            ? 'bg-sky-500 text-white border-sky-500'
            : 'bg-gray-100 dark:bg-gray-800 text-gray-700 dark:text-gray-200 border-gray-300 dark:border-gray-600 hover:bg-sky-100 dark:hover:bg-sky-900'
        ]"
        :title="handle.label || handle.id"
      >
        {{ handle.label || handle.id }}
      </button>
    </div>

    <!-- Middle row: Left tabs, content area, Right tabs -->
    <div class="flex flex-1 min-h-0">
      <!-- Left tabs column (always present for spacing) -->
      <div class="flex flex-col justify-center flex-shrink-0 tab-col-vertical">
        <button
          v-for="handle in handlesByPosition.left"
          :key="handle.id"
          @click="selectHandle(handle.id)"
          :class="[
            'px-3 py-4 text-sm font-medium rounded-l-lg border-2 border-r-0 transition-colors vertical-tab-left',
            isSelected(handle.id)
              ? 'bg-sky-500 text-white border-sky-500'
              : 'bg-gray-100 dark:bg-gray-800 text-gray-700 dark:text-gray-200 border-gray-300 dark:border-gray-600 hover:bg-sky-100 dark:hover:bg-sky-900'
          ]"
          :title="handle.label || handle.id"
        >
          <span class="vertical-text-left">{{ handle.label || handle.id }}</span>
        </button>
      </div>

      <!-- Center content slot -->
      <div class="flex-1 min-w-0 min-h-0">
        <slot></slot>
      </div>

      <!-- Right tabs column (always present for spacing) -->
      <div class="flex flex-col justify-center flex-shrink-0 tab-col-vertical">
        <button
          v-for="handle in handlesByPosition.right"
          :key="handle.id"
          @click="selectHandle(handle.id)"
          :class="[
            'px-3 py-4 text-sm font-medium rounded-r-lg border-2 border-l-0 transition-colors vertical-tab-right',
            isSelected(handle.id)
              ? 'bg-sky-500 text-white border-sky-500'
              : 'bg-gray-100 dark:bg-gray-800 text-gray-700 dark:text-gray-200 border-gray-300 dark:border-gray-600 hover:bg-sky-100 dark:hover:bg-sky-900'
          ]"
          :title="handle.label || handle.id"
        >
          <span class="vertical-text-right">{{ handle.label || handle.id }}</span>
        </button>
      </div>
    </div>

    <!-- Bottom tabs row (always present for spacing) -->
    <div class="flex justify-center flex-shrink-0 tab-row-horizontal">
      <button
        v-for="handle in handlesByPosition.bottom"
        :key="handle.id"
        @click="selectHandle(handle.id)"
        :class="[
          'px-4 py-2 text-sm font-medium rounded-b-lg border-2 border-t-0 transition-colors',
          isSelected(handle.id)
            ? 'bg-sky-500 text-white border-sky-500'
            : 'bg-gray-100 dark:bg-gray-800 text-gray-700 dark:text-gray-200 border-gray-300 dark:border-gray-600 hover:bg-sky-100 dark:hover:bg-sky-900'
        ]"
        :title="handle.label || handle.id"
      >
        {{ handle.label || handle.id }}
      </button>
    </div>
  </div>
  <div v-else class="h-full">
    <slot></slot>
  </div>
</template>

<style scoped>
.port-tabs-container {
  padding: 0;
}

/* Reserve space for tabs even when empty */
.tab-row-horizontal {
  min-height: 36px;
}

.tab-col-vertical {
  min-width: 32px;
}

.vertical-text-left {
  writing-mode: vertical-rl;
  transform: rotate(180deg);
  white-space: nowrap;
}

.vertical-text-right {
  writing-mode: vertical-rl;
  white-space: nowrap;
}

.vertical-tab-left,
.vertical-tab-right {
  min-height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>
