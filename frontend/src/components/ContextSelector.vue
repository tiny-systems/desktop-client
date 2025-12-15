<script setup>
import {computed, onMounted, ref} from 'vue';

const props = defineProps({
  ctx:Object
})

const emit = defineEmits(['select'])

// Define the path to your Go backend functions.
// This assumes your main Go struct is named 'App' and is bound to the Wails runtime.
const GoApp = window.go.main.App;


const contexts = ref([]);
const selectedContextName = ref(props.ctx ? props.ctx.name : '');

const statusMessage = ref('Initializing...');
const isLoading = ref(true);
const isConnecting = ref(false); // Manages loading state for connection/auth checks
const statusClass = ref('');


const namespaces = ref([]); // New state for list of namespaces

const selectedNamespace = ref(props.ctx ? props.ctx.ns : ''); // New state for the currently selected namespace
const isAuthorized = ref(false); // New state to control visibility of namespace selector


const loadingText = computed(() => {
  if (isLoading.value) return 'Loading contexts...';
  if (contexts.value.length === 0) return 'No contexts found';
  return 'Select a context';
});


const loadContexts = async () => {
  isLoading.value = true;
  isAuthorized.value = false;
  statusClass.value = '';
  statusMessage.value = 'Attempting to read kubeconfig...';

  try {
    const fetchedContexts = await GoApp.GetKubeContexts();

    contexts.value = fetchedContexts;

    if (fetchedContexts.length === 0) {
      statusMessage.value = 'No Kubernetes contexts found in your kubeconfig file.';
      statusClass.value = 'error';
      return;
    }

    const currentContext = fetchedContexts.find(c => c.current);
    if (currentContext) {
      selectedContextName.value = currentContext.name;
      statusMessage.value = `Current context loaded: ${currentContext.name}. Checking authorization...`;
      // Trigger health check immediately for the default context
      await checkAuthorization(currentContext.name);
    } else {
      selectedContextName.value = '';
      statusMessage.value = 'Please select a context to connect.';
      statusClass.value = '';
    }

  } catch (error) {
    statusMessage.value = `Error loading contexts: ${error}`;
    statusClass.value = 'error';
    console.error('Error in loadContexts:', error);
  } finally {
    isLoading.value = false;
  }
};

/**
 * Executes the authorization check via a new Go backend call.
 * If successful, it proceeds to load namespaces.
 */
const checkAuthorization = async (contextName) => {
  isConnecting.value = true;
  isAuthorized.value = false; // Reset authorization status
  statusMessage.value = `Running authorization check for context: ${contextName}...`;
  statusClass.value = 'info';

  try {
    // GoApp.CheckAuthorization should attempt a low-privilege API call
    const checkResult = await GoApp.CheckAuthorization(contextName);

    statusMessage.value = `Context Ready`;
    statusClass.value = 'success';
    isAuthorized.value = true;

    await getNamespaces(contextName);

    if (selectedNamespace.value === '') {
      return
    }
    emit('select', {name: selectedContextName.value, ns: selectedNamespace.value})
  } catch (error) {
    statusMessage.value = `Authorization Failed for ${contextName}: ${error}. Please check your credentials.`;
    statusClass.value = 'error';
    emit('select', null)

  } finally {
    isConnecting.value = false;
  }
}

const getNamespaces = async (contextName) => {
  statusMessage.value = `Fetching namespaces for ${contextName}...`;
  statusClass.value = 'info';

  try {
    // Call the new Go function to get namespaces
    const fetchedNamespaces = await GoApp.GetNamespaces(contextName);

    namespaces.value = fetchedNamespaces.sort(); // Sort alphabetically

    if (fetchedNamespaces.length > 0) {
      selectedNamespace.value = fetchedNamespaces.includes(selectedNamespace.value) ? selectedNamespace.value : fetchedNamespaces[0];
    } else {
      selectedNamespace.value = '';
      statusMessage.value = `No namespaces found in context: ${contextName}`;
      statusClass.value = 'warning'; // Using a custom warning class
    }

  } catch (error) {
    statusMessage.value = `Error listing namespaces: ${error}`;
    statusClass.value = 'error';
    isAuthorized.value = false; // Revoke authorization status on API error
    emit('select', null)
  }
}

const handleContextChange = async () => {
  const context = selectedContextName.value;
  if (!context) return;

  statusMessage.value = `Switching to context: ${context}...`;
  statusClass.value = 'info';
  isConnecting.value = true;
  isAuthorized.value = false;
  namespaces.value = []; // Clear old namespaces
  selectedNamespace.value = ''; // Clear selected namespace

  try {
    // 3. Call the Go backend function to switch/connect
    await GoApp.ConnectToCluster(context);

    // 4. Run the new authorization check and namespace load
    await checkAuthorization(context);

  } catch (error) {
    statusMessage.value = `Error switching context to ${context}: ${error}`;
    statusClass.value = 'error';
    console.error('Error in handleContextChange:', error);
  } finally {
    isConnecting.value = false;
  }
};

/**
 * Handles the user selecting a new namespace.
 * This function can be expanded later for actual resource fetching.
 */
const handleNamespaceChange = () => {
  statusMessage.value = '';
  emit('select', {name: selectedContextName.value, ns: selectedNamespace.value})
};

onMounted(() => {
  if (GoApp) {
    loadContexts();
    return
  }
  statusMessage.value = "Wails Go runtime not ready.";
  statusClass.value = 'error';
});
</script>

<template>
  <div class="kube-context-selector-app">
    <div class="pb-2">
      <div class="bg-white rounded-sm w-full min-w-full max-w-md mx-auto space-y-4">
        <!-- Context Selector -->
        <div class="flex justify-stretch gap-2">
          <div>
            <label for="context-selector" class="block text-sm font-medium text-gray-600">1. Select Context</label>
            <select
              id="context-selector"
              v-model="selectedContextName"
              @change="handleContextChange"
              :disabled="isLoading || contexts.length === 0"
              class="appearance-none bg-gray-100 border border-gray-300 rounded-md px-4 py-2 pr-8 focus:outline-none focus:ring-2 focus:ring-blue-500 max-w-2xl w-full"
            >
              <option value="" disabled>{{ loadingText }}</option>
              <option
                v-for="context in contexts"
                :key="context.name"
                :value="context.name"
              >
                {{ context.name }} ({{ context.cluster }})
              </option>
            </select>
          </div>

          <!-- Namespace Selector (Visible only after successful Authorization) -->
          <div v-if="isAuthorized">
            <label for="namespace-selector" class="block text-sm font-medium text-gray-600">2. Select
              Namespace</label>
            <select
              id="namespace-selector"
              v-model="selectedNamespace"
              @change="handleNamespaceChange"
              :disabled="namespaces.length === 0"
              class="appearance-none bg-gray-100 border border-gray-300 rounded-md px-4 py-2 pr-8 focus:outline-none focus:ring-2 focus:ring-blue-500 max-w-2xl w-full"
            >
              <option value="" disabled v-if="namespaces.length === 0">No namespaces found</option>
              <option
                v-for="ns in namespaces"
                :key="ns"
                :value="ns"
              >
                {{ ns }}
              </option>
            </select>
          </div>
          <div v-else :class="['m-4 p-3 rounded-lg', statusClass === 'error' ? 'bg-red-100' : statusClass === 'success' ? 'bg-green-100' : statusClass === 'info' ? 'bg-blue-100' : 'bg-gray-100']">
            <div v-if="isConnecting" class="flex items-center space-x-2 text-sm font-medium text-blue-700">
              <span>{{ statusMessage }}</span>
            </div>
            <p v-else :class="['text-sm font-medium', statusClass]">
              {{ statusMessage }}
            </p>
          </div>
          <div class="flex items-center space-x-2 pt-5" v-if="isAuthorized">
            <button tabindex="0" class="text-gray-600 hover:text-blue-500 cursor-pointer p-2.5 border rounded-sm border-gray-300" @click="handleNamespaceChange">
              <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" class="size-5">
                <path fill-rule="evenodd" d="M15.312 11.424a5.5 5.5 0 0 1-9.201 2.466l-.312-.311h2.433a.75.75 0 0 0 0-1.5H3.989a.75.75 0 0 0-.75.75v4.242a.75.75 0 0 0 1.5 0v-2.43l.31.31a7 7 0 0 0 11.712-3.138.75.75 0 0 0-1.449-.39Zm1.23-3.723a.75.75 0 0 0 .219-.53V2.929a.75.75 0 0 0-1.5 0V5.36l-.31-.31A7 7 0 0 0 3.239 8.188a.75.75 0 1 0 1.448.389A5.5 5.5 0 0 1 13.89 6.11l.311.31h-2.432a.75.75 0 0 0 0 1.5h4.243a.75.75 0 0 0 .53-.219Z" clip-rule="evenodd" />
              </svg>
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
