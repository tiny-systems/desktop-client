<script setup>
import ProjectStatusBar from "./ProjectStatusBar.vue";
import {onMounted, ref} from "vue";

const emit = defineEmits(['close'])
const props = defineProps({
  ctx: String,
  ns: String,
  name: String
})

const GoApp = window.go.main.App;

const statusMessage = ref('Initializing...');
const statusClass = ref('');
const isLoading = ref(false)
const project = ref(null);


const close =  () => {
  emit('close')
}

onMounted(() => {
  if (GoApp) {
    loadProject(props.ctx, props.ns, props.name)
    return
  }
  statusMessage.value = "Wails Go runtime not ready.";
  statusClass.value = 'error';
});


const loadProject = async (context, ns, name)=>{

  statusMessage.value = 'Attempting to read TinyProject...';
  isLoading.value = true
  project.value = null

  try {
    const fetchedProject = await GoApp.GetProject(context, ns, name);
    if (!fetchedProject) {
      statusMessage.value = 'Project not found';
      statusClass.value = ''
      return
    }

    project.value = fetchedProject;
    statusClass.value = 'success'
    statusMessage.value = 'Project data retrieved'

  } catch (error) {
    statusMessage.value = `Error loading project: ${error}`;
    statusClass.value = 'error';

    console.error('Error in getProject:', error);
  } finally {
    isLoading.value = false;
  }
}

</script>
<template>
  <div class="p-4 relative">
    {{name}}
    <div class="fixed top-4 right-4">
      <a href="#" class="text-2xl" @click.prevent="close">&times</a>
    </div>
    <ProjectStatusBar></ProjectStatusBar>
  </div>
</template>
