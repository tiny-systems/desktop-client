<script setup>
import { ref, onMounted, nextTick } from 'vue'
import FlowPreview from '../flow/FlowPreview.vue'
import { EllipsisVerticalIcon, PencilIcon, TrashIcon, XMarkIcon } from '@heroicons/vue/24/outline'

const GoApp = window.go.main.App

const props = defineProps({
  ctx: String,
  ns: String,
  projectName: String,
  flow: {
    type: Object,
    required: true
  }
})

const emit = defineEmits(['error', 'undeploy', 'rename', 'open'])

const graph = ref(null)
const loading = ref(true)
const showMenu = ref(false)
const showRenameDialog = ref(false)
const showUndeployDialog = ref(false)
const newFlowName = ref('')
const renaming = ref(false)
const undeploying = ref(false)
const renameInputRef = ref(null)

onMounted(async () => {
  if (!GoApp) return
  try {
    const result = await GoApp.GetFlowGraph(props.ctx, props.ns, props.projectName, props.flow.resourceName)
    graph.value = result
  } catch (err) {
    console.error('Failed to load flow graph:', err)
  } finally {
    loading.value = false
  }
})

const openRenameDialog = () => {
  showMenu.value = false
  newFlowName.value = props.flow.name
  showRenameDialog.value = true
  nextTick(() => {
    renameInputRef.value?.focus()
    renameInputRef.value?.select()
  })
}

const closeRenameDialog = () => {
  showRenameDialog.value = false
  newFlowName.value = ''
}

const renameFlow = async () => {
  if (!GoApp || !newFlowName.value.trim() || newFlowName.value.trim() === props.flow.name) {
    closeRenameDialog()
    return
  }

  renaming.value = true
  try {
    await GoApp.RenameFlow(props.ctx, props.ns, props.flow.resourceName, newFlowName.value.trim())
    emit('rename', props.flow.resourceName, newFlowName.value.trim())
    closeRenameDialog()
  } catch (err) {
    emit('error', `Failed to rename flow: ${err}`)
  } finally {
    renaming.value = false
  }
}

const openUndeployDialog = () => {
  showMenu.value = false
  showUndeployDialog.value = true
}

const closeUndeployDialog = () => {
  showUndeployDialog.value = false
}

const undeployFlow = async () => {
  if (!GoApp) return

  undeploying.value = true
  try {
    await GoApp.UndeployFlow(props.ctx, props.ns, props.flow.resourceName)
    emit('undeploy', props.flow.resourceName)
    closeUndeployDialog()
  } catch (err) {
    emit('error', `Failed to undeploy flow: ${err}`)
  } finally {
    undeploying.value = false
  }
}
</script>

<template>
  <div
    class="flow-card bg-white dark:bg-gray-900 rounded-lg border border-gray-200 dark:border-gray-700 overflow-hidden hover:shadow-md transition-shadow cursor-pointer"
    @click="emit('open', flow)"
    tabindex="0"
    @keydown.enter="emit('open', flow)"
  >
    <div class="flow-preview h-40 bg-gray-50 dark:bg-gray-800 relative">
      <div v-if="loading" class="flex items-center justify-center h-full">
        <span class="text-xs text-gray-400">Loading...</span>
      </div>
      <FlowPreview v-else-if="graph" :graph="graph" :id="flow.resourceName" />
      <div v-else class="flex items-center justify-center h-full">
        <span class="text-xs text-gray-400">No preview</span>
      </div>

      <!-- Menu button -->
      <div class="absolute top-2 right-2">
        <button
          @click.stop="showMenu = !showMenu"
          class="p-1.5 bg-white/80 dark:bg-gray-800/80 hover:bg-white dark:hover:bg-gray-700 rounded-full shadow-sm"
          title="More options"
        >
          <EllipsisVerticalIcon class="w-4 h-4 text-gray-500 dark:text-gray-400" />
        </button>

        <!-- Backdrop to close menu -->
        <div
          v-if="showMenu"
          class="fixed inset-0 z-40"
          @click="showMenu = false"
        ></div>

        <!-- Dropdown menu -->
        <div
          v-if="showMenu"
          class="absolute right-0 top-8 z-50 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-600 rounded-md shadow-lg py-1 min-w-36"
        >
          <button
            @click="openRenameDialog"
            class="w-full px-4 py-2 text-left text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 flex items-center space-x-2"
          >
            <PencilIcon class="w-4 h-4" />
            <span>Rename</span>
          </button>
          <button
            @click="openUndeployDialog"
            class="w-full px-4 py-2 text-left text-sm text-red-600 dark:text-red-400 hover:bg-gray-100 dark:hover:bg-gray-700 flex items-center space-x-2"
          >
            <TrashIcon class="w-4 h-4" />
            <span>Undeploy</span>
          </button>
        </div>
      </div>
    </div>
    <div class="p-3">
      <h3 class="text-sm font-medium text-gray-900 dark:text-white truncate">
        {{ flow.name }}
      </h3>
      <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
        {{ flow.nodeCount }} nodes
      </p>
    </div>

    <!-- Rename Dialog -->
    <div
      v-if="showRenameDialog"
      class="fixed inset-0 z-50 overflow-y-auto"
      @keydown.enter.prevent="renameFlow"
      @keydown.escape="closeRenameDialog"
    >
      <div class="fixed inset-0 bg-black/40 backdrop-blur-md" @click="closeRenameDialog"></div>
      <div class="flex min-h-full items-center justify-center p-4">
        <div class="relative bg-white dark:bg-gray-900 rounded-lg shadow-xl w-full max-w-md p-6">
          <div class="flex items-center justify-between mb-4">
            <h3 class="text-lg font-medium text-gray-900 dark:text-gray-100">
              Rename Flow
            </h3>
            <button @click="closeRenameDialog" class="text-gray-400 hover:text-gray-500">
              <XMarkIcon class="w-5 h-5" />
            </button>
          </div>
          <div class="mb-4">
            <label for="flow-name" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              Flow Name
            </label>
            <input
              id="flow-name"
              ref="renameInputRef"
              v-model="newFlowName"
              type="text"
              class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-sky-500 focus:border-transparent"
            />
          </div>
          <div class="flex justify-end gap-2">
            <button
              @click="closeRenameDialog"
              class="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-md hover:bg-gray-50 dark:hover:bg-gray-700"
            >
              Cancel
            </button>
            <button
              @click="renameFlow"
              :disabled="!newFlowName.trim() || renaming"
              class="px-4 py-2 text-sm font-medium text-white bg-sky-600 rounded-md hover:bg-sky-700 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {{ renaming ? 'Renaming...' : 'Rename' }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Undeploy Confirmation Dialog -->
    <div
      v-if="showUndeployDialog"
      class="fixed inset-0 z-50 overflow-y-auto"
      @keydown.enter.prevent="undeployFlow"
      @keydown.escape="closeUndeployDialog"
    >
      <div class="fixed inset-0 bg-black/40 backdrop-blur-md" @click="closeUndeployDialog"></div>
      <div class="flex min-h-full items-center justify-center p-4">
        <div class="relative bg-white dark:bg-gray-900 rounded-lg shadow-xl w-full max-w-md p-6">
          <h3 class="text-lg font-medium text-gray-900 dark:text-gray-100 mb-4">
            Undeploy Flow?
          </h3>
          <p class="text-sm text-gray-500 dark:text-gray-400 mb-6">
            Are you sure you want to undeploy "<strong class="text-gray-700 dark:text-gray-300">{{ flow.name }}</strong>"?
            This will delete all {{ flow.nodeCount }} nodes in this flow. This action cannot be undone.
          </p>
          <div class="flex justify-end gap-2">
            <button
              @click="closeUndeployDialog"
              class="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-md hover:bg-gray-50 dark:hover:bg-gray-700"
            >
              Cancel
            </button>
            <button
              @click="undeployFlow"
              :disabled="undeploying"
              class="px-4 py-2 text-sm font-medium text-white bg-red-600 rounded-md hover:bg-red-700 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {{ undeploying ? 'Undeploying...' : 'Undeploy' }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
