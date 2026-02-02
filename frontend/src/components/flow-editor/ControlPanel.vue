<script setup>
import { ref } from 'vue'
import { Menu, MenuButton, MenuItem, MenuItems } from '@headlessui/vue'
import { EllipsisVerticalIcon } from '@heroicons/vue/24/solid'
import { TrashIcon } from '@heroicons/vue/24/outline'
import FlowSwitcher from './FlowSwitcher.vue'

const GoApp = window.go?.main?.App

const props = defineProps({
  flowName: String,
  flowResourceName: String,
  projectName: String,
  projectResourceName: String,
  contextName: String,
  namespace: String,
  loading: Boolean,
})

const emit = defineEmits(['close', 'error', 'switch-flow', 'new-flow'])

const showUndeployDialog = ref(false)
const undeploying = ref(false)

const openUndeployDialog = () => {
  showUndeployDialog.value = true
}

const closeUndeployDialog = () => {
  showUndeployDialog.value = false
}

const undeployFlow = async () => {
  if (!GoApp || !props.flowResourceName) return

  undeploying.value = true
  try {
    await GoApp.UndeployFlow(props.contextName, props.namespace, props.flowResourceName)
    closeUndeployDialog()
    emit('close') // Navigate back to project after undeploy
  } catch (err) {
    emit('error', `Failed to undeploy flow: ${err}`)
  } finally {
    undeploying.value = false
  }
}

const handleFlowSwitch = (flowResourceName) => {
  emit('switch-flow', flowResourceName)
}
</script>

<template>
  <div class="lg:items-center lg:justify-center p-1 lg:p-3 border-b border-gray-200 dark:border-gray-700">
    <div class="flex justify-between align-middle">
      <div class="flex justify-left align-middle items-center">
        <!-- Back button -->
        <button
          type="button"
          @click="emit('close')"
          title="Back to project"
          class="text-sky-600 border border-sky-600 hover:bg-sky-600 hover:text-white focus:ring-4 focus:outline-none focus:ring-sky-300 font-medium rounded-lg text-sm p-2.5 text-center inline-flex items-center mr-2 dark:border-sky-500 dark:text-sky-500 dark:hover:text-white dark:focus:ring-sky-800 dark:hover:bg-sky-600"
        >
          <svg
            aria-hidden="true"
            class="w-5 h-5 rotate-180"
            fill="currentColor"
            viewBox="0 0 20 20"
            xmlns="http://www.w3.org/2000/svg"
          >
            <path
              fill-rule="evenodd"
              d="M10.293 3.293a1 1 0 011.414 0l6 6a1 1 0 010 1.414l-6 6a1 1 0 01-1.414-1.414L14.586 11H3a1 1 0 110-2h11.586l-4.293-4.293a1 1 0 010-1.414z"
              clip-rule="evenodd"
            />
          </svg>
          <span class="sr-only">Back to project</span>
        </button>

        <!-- Project / Flow header -->
        <div class="inline-flex items-center">
          <span v-if="projectName" class="text-lg font-medium text-gray-500 dark:text-gray-400 px-1">
            {{ projectName }}
          </span>
          <span v-if="projectName" class="text-gray-400 dark:text-gray-500 px-1">/</span>
          <FlowSwitcher
            :current-flow-name="flowName || flowResourceName"
            :current-flow-resource-name="flowResourceName"
            :project-name="projectResourceName"
            :context-name="contextName"
            :namespace="namespace"
            @switch="handleFlowSwitch"
            @new-flow="emit('new-flow')"
          />
        </div>

        <!-- Saving indicator -->
        <div v-if="loading" class="inline-flex items-center">
          <span class="text-sky-600 ml-3 text-sm">saving...</span>
        </div>
      </div>

      <!-- Three-dots menu -->
      <Menu as="div" class="relative">
        <MenuButton class="p-2 rounded-full text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-800 focus:outline-none">
          <EllipsisVerticalIcon class="w-5 h-5" />
        </MenuButton>
        <transition
          enter-active-class="transition ease-out duration-100"
          enter-from-class="transform opacity-0 scale-95"
          enter-to-class="transform opacity-100 scale-100"
          leave-active-class="transition ease-in duration-75"
          leave-from-class="transform opacity-100 scale-100"
          leave-to-class="transform opacity-0 scale-95"
        >
          <MenuItems class="absolute right-0 z-50 mt-2 w-48 origin-top-right rounded-md bg-white dark:bg-gray-900 shadow-lg ring-1 ring-black ring-opacity-5 dark:ring-gray-700 focus:outline-none">
            <div class="py-1">
              <MenuItem v-slot="{ active }">
                <button
                  @click="openUndeployDialog"
                  :class="[
                    active ? 'bg-red-50 dark:bg-red-900/20 text-red-700 dark:text-red-400' : 'text-red-600 dark:text-red-400',
                    'w-full flex items-center px-4 py-2 text-sm'
                  ]"
                >
                  <TrashIcon class="mr-3 h-4 w-4" />
                  Undeploy Flow
                </button>
              </MenuItem>
            </div>
          </MenuItems>
        </transition>
      </Menu>
    </div>
  </div>

  <!-- Undeploy Confirmation Dialog -->
  <teleport to="body">
    <div
      v-if="showUndeployDialog"
      class="fixed inset-0 z-50 flex items-center justify-center"
      @keydown.enter.prevent="undeployFlow"
      @keydown.escape="closeUndeployDialog"
    >
      <div class="fixed inset-0 bg-black/50 backdrop-blur-sm" @click="closeUndeployDialog"></div>
      <div class="relative bg-white dark:bg-gray-900 rounded-lg shadow-xl w-full max-w-md p-6 mx-4">
        <h3 class="text-lg font-medium text-gray-900 dark:text-gray-100 mb-2">
          Undeploy Flow
        </h3>
        <p class="text-sm text-gray-600 dark:text-gray-400 mb-4">
          Are you sure you want to undeploy "<strong class="text-gray-800 dark:text-gray-200">{{ flowName || flowResourceName }}</strong>"?
          This will delete all nodes in this flow. This action cannot be undone.
        </p>
        <div class="flex justify-end gap-2">
          <button
            @click="closeUndeployDialog"
            :disabled="undeploying"
            class="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-gray-100 dark:bg-gray-800 rounded-md hover:bg-gray-200 dark:hover:bg-gray-700 disabled:opacity-50"
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
  </teleport>
</template>
