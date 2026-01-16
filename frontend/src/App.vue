<script setup>
import {ref} from 'vue';

import ProjectList from "./components/ProjectList.vue";
import Project from "./components/Project.vue";

const ctx = ref(null)
const project = ref(null)

const selectProject = (p) => {
  project.value = p
}
const selectContext = (c) => {
  ctx.value = c
}
</script>
<template>
  <div class="h-screen w-screen overflow-hidden bg-gray-50 dark:bg-gray-900">
    <div class="w-full h-full">
      <Project v-if="project" @close="project = null" :ctx="ctx.name" :ns="ctx.ns" :name="project.name"></Project>
      <ProjectList :ctx="ctx" @selectProject="selectProject" @selectContext="selectContext" v-else></ProjectList>
    </div>
  </div>
</template>
