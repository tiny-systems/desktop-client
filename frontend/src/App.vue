<script setup>
import {ref, onMounted} from 'vue';
import {GetBuildInfo} from "../wailsjs/go/main/App";

import ProjectList from "./components/ProjectList.vue";
import Project from "./components/Project.vue";

const ctx = ref(null)
const project = ref(null)
const buildInfo = ref(null)

const selectProject = (p) => {
  project.value = p
}
const selectContext = (c) => {
  ctx.value = c
}

onMounted(async () => {
  try {
    buildInfo.value = await GetBuildInfo()
  } catch (e) {
    console.error('Failed to get build info:', e)
  }
})
</script>
<template>
  <div class="h-screen w-screen overflow-hidden bg-gray-50 dark:bg-gray-900">
    <div class="w-full h-full">
      <Project v-if="project" @close="project = null" :ctx="ctx.name" :ns="ctx.ns" :name="project.name"></Project>
      <ProjectList :ctx="ctx" @selectProject="selectProject" @selectContext="selectContext" v-else></ProjectList>
    </div>
    <div v-if="buildInfo" class="fixed bottom-1 right-2 text-[10px] text-gray-400 dark:text-gray-600 pointer-events-none select-none">
      SDK {{ buildInfo.sdkVersion }} · Built {{ buildInfo.buildTime }}
    </div>
  </div>
</template>
