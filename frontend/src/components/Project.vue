<script setup>
import ProjectPage from "./project/ProjectPage.vue";

const emit = defineEmits(['close'])
const props = defineProps({
  ctx: String,
  ns: String,
  name: String
})

const close = () => {
  emit('close')
}
</script>

<template>
  <ProjectPage
    :ctx="ctx"
    :ns="ns"
    :name="name"
    @close="close"
  />
</template>
