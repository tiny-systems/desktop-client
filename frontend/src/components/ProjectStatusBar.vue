<template>
  <div class="status-stream">
    <p>Last Event Type: <strong>{{ eventType }}</strong></p>
    <div v-html="message"></div>
  </div>
</template>

<script>
import { EventsOn } from '../../wailsjs/runtime/runtime';
// NOTE: For TypeScript, you'd define the interfaces here.

export default {
  data() {
    return {
      eventType: 'None',
      message: 'Awaiting project stream...',
      unsubscribe: null,
    };
  },
  methods: {
    handleEvent(event) {
      // The event object is the JSON-decoded UpdateEvent struct from Go
      this.eventType = event.type;

      switch (event.type) {
        case 'status':
          // Access the payload fields based on the expected StatusPayload
          this.message = `<span style="color: blue;">✅ Status: ${event.payload.message}</span>`;
          break;
        case 'progress':
          // Access the payload fields based on the expected ProgressPayload
          this.message = `<span style="color: green;">⚙️ Progress: ${event.payload.percentage}% on ${event.payload.file}</span>`;
          break;
        case 'error':
          // Access the payload fields based on the expected ErrorPayload
          this.message = `<span stwaiyle="color: red;">❌ Error ${event.payload.code}: ${event.payload.details}</span>`;
          break;
        default:
          this.message = `Unknown event type received: ${event.type}`;
      }
    }
  },
  mounted() {
    // Listen for the single "projectStream" event
    this.unsubscribe = EventsOn('projectStream', this.handleEvent);
  },
  unmounted() {
    this.message = ''
    if (this.unsubscribe) {
      this.unsubscribe();
      console.log('Project stream listener unsubscribed.');
    }
  }
};
</script>
