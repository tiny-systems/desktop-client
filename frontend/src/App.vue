<script setup>
import {ref, onMounted, onUnmounted} from 'vue';
import {GetBuildInfo, GetPendingDeepLink} from "../wailsjs/go/main/App";
import {EventsOn, EventsOff} from "../wailsjs/runtime/runtime";

import ProjectList from "./components/ProjectList.vue";
import Project from "./components/Project.vue";
import DeepLinkImportModal from "./components/DeepLinkImportModal.vue";

const ctx = ref(null)
const project = ref(null)
const buildInfo = ref(null)
const deepLinkData = ref(null) // { token, api }
const initialTab = ref('projects')

const selectProject = (p) => {
  project.value = p
}
const selectContext = (c) => {
  ctx.value = c
}

const parseDeepLink = (rawUrl) => {
  // Extract token and api from tinysystems://deploy?token=<token>&api=<encoded-api-base>
  try {
    const u = new URL(rawUrl)
    const token = u.searchParams.get('token')
    const api = u.searchParams.get('api')
    if (token && api) {
      return { token, api }
    }
    // Backwards compat: support old url= format
    const url = u.searchParams.get('url')
    if (url) {
      return { legacyUrl: url }
    }
    return null
  } catch {
    return null
  }
}

const handleDeepLink = (rawUrl) => {
  console.log('[DEEPLINK] handleDeepLink called with:', rawUrl)
  const data = parseDeepLink(rawUrl)
  console.log('[DEEPLINK] parsed data:', data)
  if (data) {
    deepLinkData.value = data
  }
}

onMounted(async () => {
  console.log('[DEEPLINK] App.vue onMounted')
  try {
    buildInfo.value = await GetBuildInfo()
  } catch (e) {
    console.error('Failed to get build info:', e)
  }

  EventsOn('deeplink:deploy', (url) => {
    console.log('[DEEPLINK] EventsOn deeplink:deploy received:', url)
    handleDeepLink(url)
  })

  // Check for deep link that arrived before frontend was ready (cold-start)
  try {
    console.log('[DEEPLINK] calling GetPendingDeepLink...')
    const pending = await GetPendingDeepLink()
    console.log('[DEEPLINK] GetPendingDeepLink returned:', pending)
    if (pending) handleDeepLink(pending)
  } catch (e) {
    console.error('Failed to check pending deep link:', e)
  }

  EventsOn('navigate:modules', () => {
    project.value = null
    initialTab.value = 'modules'
  })
})

onUnmounted(() => {
  EventsOff('deeplink:deploy')
  EventsOff('navigate:modules')
})
</script>
<template>
  <div class="h-screen w-screen overflow-hidden bg-gray-50 dark:bg-gray-900">
    <div class="w-full h-full">
      <Project v-if="project" @close="project = null" :ctx="ctx.name" :ns="ctx.ns" :name="project.name"></Project>
      <ProjectList :ctx="ctx" :initial-tab="initialTab" @selectProject="selectProject" @selectContext="selectContext" v-else></ProjectList>
    </div>
    <div v-if="buildInfo" class="fixed bottom-1 right-2 text-[10px] text-gray-400 dark:text-gray-600 pointer-events-none select-none">
      SDK {{ buildInfo.sdkVersion }} · Built {{ buildInfo.buildTime }}
    </div>

    <!-- Deep link import modal (overlays everything) -->
    <DeepLinkImportModal
      v-if="deepLinkData"
      :deep-link-data="deepLinkData"
      :ctx="ctx"
      @close="deepLinkData = null"
      @success="deepLinkData = null"
    />
  </div>
</template>
