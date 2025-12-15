<template>

  <header class="sticky top-0 bg-white z-50 shadow-sm">
    <div class="px-4 py-3">
      <ContextSelector @select="onSelect" :ctx="ctx"/>
    </div>
  </header>
  <main v-if="ctx">
    <div>
      <div class="p-4">
        <div class="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4" v-if="projects.length > 0">

          <div v-for="p in projects" :key="p.name" class="tile-card bg-white p-6 rounded-sm border border-gray-100 cursor-pointer" tabindex="0" @click="handleProjectSelection(p)">
            <div class="flex items-center space-x-4" >
              <div class="p-3 rounded-full bg-blue-100 text-blue-600">
                <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"></path>
                </svg>
              </div>
              <h3 class="text-xl font-light text-gray-800">
                {{p.title}}
              </h3>
            </div>

            <p class="mt-4 text-gray-600 text-sm font-light" v-if="p.description">
              {{p.description}}
            </p>
          </div>
        </div>
        <div v-else :class="[statusClass === 'error' ? 'text-red-600' : '', 'text-center p-4']">
          {{statusMessage}}
        </div>
      </div>
    </div>
  </main>


</template>
<script setup>
import {onMounted, ref} from 'vue';
import ContextSelector from "./ContextSelector.vue";
const props = defineProps({
  ctx: Object
})


const ctx = ref(props.ctx)

const onSelect = function (c) {
  ctx.value = c
  emit('select-context', c)
  if (!c) {
    return
  }
  loadProjects(c.name, c.ns)
}
const emit = defineEmits(['select-project', 'select-context'])

// Define the path to your Go backend functions.
// This assumes your main Go struct is named 'App' and is bound to the Wails runtime.
const GoApp = window.go.main.App;

const statusMessage = ref('Initializing...');
const statusClass = ref('');
const isLoading = ref(false)
const projects = ref([]);

const handleProjectSelection = (prj) => {
  statusMessage.value = ``;
  emit('select-project', prj)
};

const loadProjects = async (name, ns)=>{

  statusMessage.value = 'Attempting to read TinyProjects...';
  isLoading.value = true
  projects.value = []
  try {
    const fetchedProjects = await GoApp.GetProjects(name, ns);
    if (!fetchedProjects) {
      statusMessage.value = 'No projects found in ' + ns + ' namespace';
      statusClass.value = ''
      return
    }

    if (fetchedProjects.length === 0) {
      statusMessage.value = 'No projects found in ' + ns + ' namespace';
      statusClass.value = ''
      return;
    }

    projects.value = fetchedProjects;
    statusClass.value = 'success'
    statusMessage.value = 'found + ' + fetchedProjects.length + ' projects'

  } catch (error) {
    statusMessage.value = `Error loading projects: ${error}`;
    statusClass.value = 'error';

    console.error('Error in getProjects:', error);
  } finally {
    isLoading.value = false;
  }
}

onMounted(() => {
  if (GoApp) {
    return
  }
  statusMessage.value = "Wails Go runtime not ready.";
  statusClass.value = 'error';
});

</script>
