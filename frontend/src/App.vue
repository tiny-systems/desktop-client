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
  <div class="h-screen max-w-7xl mx-auto">
    <Project v-if="project" @close="project = null" :ctx="ctx.name" :ns="ctx.ns" :name="project.name"></Project>
    <ProjectList :ctx="ctx" @selectProject="selectProject" @selectContext="selectContext" v-else></ProjectList>
  </div>
</template>
