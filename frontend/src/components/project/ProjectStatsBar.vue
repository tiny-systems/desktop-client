<script setup>
import { CubeIcon, DocumentTextIcon, RectangleGroupIcon } from '@heroicons/vue/24/outline'

const props = defineProps({
  stats: {
    type: Object,
    default: () => ({
      widgetsCount: 0,
      flowsCount: 0,
      nodesCount: 0
    })
  },
  activeTab: String,
})

const emit = defineEmits(['selectTab'])
</script>

<template>
  <div class="flex items-center space-x-4 px-4 py-2 bg-gray-50 dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700">
    <button
      @click="emit('selectTab', 'widgets')"
      :class="[
        'flex items-center space-x-2 px-3 py-1.5 rounded-lg text-sm transition-colors',
        activeTab === 'widgets'
          ? 'bg-sky-100 dark:bg-sky-900 text-sky-700 dark:text-sky-300'
          : 'text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700'
      ]"
    >
      <RectangleGroupIcon class="w-4 h-4" />
      <span>Widgets</span>
      <span class="font-semibold">{{ stats.widgetsCount }}</span>
    </button>
    <button
      @click="emit('selectTab', 'flows')"
      :class="[
        'flex items-center space-x-2 px-3 py-1.5 rounded-lg text-sm transition-colors',
        activeTab === 'flows'
          ? 'bg-sky-100 dark:bg-sky-900 text-sky-700 dark:text-sky-300'
          : 'text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700'
      ]"
    >
      <CubeIcon class="w-4 h-4" />
      <span>Flows/nodes</span>
      <span class="font-semibold">{{ stats.flowsCount }}/{{ stats.nodesCount }}</span>
    </button>
    <button
      @click="emit('selectTab', 'description')"
      :class="[
        'flex items-center space-x-2 px-3 py-1.5 rounded-lg text-sm transition-colors',
        activeTab === 'description'
          ? 'bg-sky-100 dark:bg-sky-900 text-sky-700 dark:text-sky-300'
          : 'text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700'
      ]"
    >
      <DocumentTextIcon class="w-4 h-4" />
      <span>Description</span>
    </button>
  </div>
</template>
