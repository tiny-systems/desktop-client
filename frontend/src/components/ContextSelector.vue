<script setup>
import {computed, onMounted, ref} from 'vue';

const props = defineProps({
  ctx:Object
})

const emit = defineEmits(['select', 'contexts-loaded'])

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
    emit('contexts-loaded', fetchedContexts.length > 0)

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
  <div class="context-selector">
    <div class="flex items-end gap-3 flex-wrap">
      <!-- Context Selector -->
      <div class="flex-shrink min-w-0">
        <label for="context-selector" class="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">Context</label>
        <select
          id="context-selector"
          v-model="selectedContextName"
          @change="handleContextChange"
          :disabled="isLoading || contexts.length === 0"
          class="appearance-none bg-gray-100 dark:bg-gray-800 border border-gray-300 dark:border-gray-600 text-gray-900 dark:text-gray-100 rounded-lg px-3 py-2 pr-8 text-sm focus:outline-none focus:ring-2 focus:ring-sky-500 dark:focus:ring-sky-400 focus:border-transparent w-full max-w-xs"
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

      <!-- Namespace Selector -->
      <div v-if="isAuthorized" class="flex-shrink-0">
        <label for="namespace-selector" class="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">Namespace</label>
        <select
          id="namespace-selector"
          v-model="selectedNamespace"
          @change="handleNamespaceChange"
          :disabled="namespaces.length === 0"
          class="appearance-none bg-gray-100 dark:bg-gray-800 border border-gray-300 dark:border-gray-600 text-gray-900 dark:text-gray-100 rounded-lg px-3 py-2 pr-8 text-sm focus:outline-none focus:ring-2 focus:ring-sky-500 dark:focus:ring-sky-400 focus:border-transparent min-w-40"
        >
          <option value="" disabled v-if="namespaces.length === 0">No namespaces</option>
          <option
            v-for="ns in namespaces"
            :key="ns"
            :value="ns"
          >
            {{ ns }}
          </option>
        </select>
      </div>

      <!-- Refresh button -->
      <div v-if="isAuthorized" class="flex-shrink-0">
        <button
          @click="handleNamespaceChange"
          class="p-2 text-gray-500 dark:text-gray-400 hover:text-sky-600 dark:hover:text-sky-400 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-lg transition-colors"
          title="Refresh"
        >
          <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" class="w-5 h-5">
            <path fill-rule="evenodd" d="M15.312 11.424a5.5 5.5 0 0 1-9.201 2.466l-.312-.311h2.433a.75.75 0 0 0 0-1.5H3.989a.75.75 0 0 0-.75.75v4.242a.75.75 0 0 0 1.5 0v-2.43l.31.31a7 7 0 0 0 11.712-3.138.75.75 0 0 0-1.449-.39Zm1.23-3.723a.75.75 0 0 0 .219-.53V2.929a.75.75 0 0 0-1.5 0V5.36l-.31-.31A7 7 0 0 0 3.239 8.188a.75.75 0 1 0 1.448.389A5.5 5.5 0 0 1 13.89 6.11l.311.31h-2.432a.75.75 0 0 0 0 1.5h4.243a.75.75 0 0 0 .53-.219Z" clip-rule="evenodd" />
          </svg>
        </button>
      </div>

      <!-- Status message -->
      <div v-if="!isAuthorized && (isConnecting || statusMessage)" class="flex items-center flex-shrink min-w-0">
        <div v-if="isConnecting" class="flex items-center space-x-2 text-sm text-sky-600 dark:text-sky-400 whitespace-nowrap">
          <div class="animate-spin rounded-full h-4 w-4 border-b-2 border-sky-500 flex-shrink-0"></div>
          <span class="truncate">{{ statusMessage }}</span>
        </div>
        <div v-else-if="statusClass === 'error'" class="flex items-center space-x-2 px-3 py-1.5 bg-red-50 dark:bg-red-900/20 rounded-lg whitespace-nowrap">
          <svg class="w-4 h-4 text-red-500 dark:text-red-400 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"/>
          </svg>
          <span class="text-sm text-red-600 dark:text-red-400 truncate">{{ statusMessage }}</span>
        </div>
        <div v-else-if="statusMessage" class="text-sm text-gray-500 dark:text-gray-400 truncate">
          {{ statusMessage }}
        </div>
      </div>
    </div>
  </div>
</template>
