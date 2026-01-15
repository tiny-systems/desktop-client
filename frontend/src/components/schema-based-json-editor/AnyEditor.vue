<template>
  <div :class="className">
    <div :class="theme.title">
      <div class="flex space-x-1 w-full">
        <div class="pl-1" v-if="titleToShow != ''">{{titleToShow}}</div>
      </div>
      <div :class="theme.buttonGroup" :style="buttonGroupStyle">
        <optional :required="required" :value="value" :isReadOnly="isReadOnly"
                  :theme="theme" :locale="locale" @toggleOptional="toggleOptional()">
        </optional>
      </div>
      <div :class="theme.buttonGroup">
        <button v-if="hasDeleteButtonFunction" type="button" class="w-4 text-sky-500 cursor-pointer ml-1" @click="$emit('delete')" title="Delete">
          <XCircleIcon></XCircleIcon>
        </button>
      </div>
    </div>
    <description :theme="theme" :message="schema.description"></description>
    <code class="text-xs p-1 bg-blue-50 dark:bg-blue-500 w-full block">{{value}}</code>
  </div>
</template>
<script lang="ts">
import type {PropType} from 'vue'
import * as common from './common'
import Optional from './Optional.vue'
import Description from './Description.vue'
import { XCircleIcon } from '@heroicons/vue/24/outline'

export default {
  emits: ['update-value', 'delete'],
  props: {
    schema: {
      required: true,
    },
    property: String,
    initialValue: null,
    title: [String, Number],
    theme: {
      type: Object as PropType<common.Theme>,
      required: true,
    },
    locale: {
      type: Object as PropType<common.Locale>,
      required: true,
    },
    allowLookup: Boolean,
    plainStruct: Boolean,
    readonly: Boolean,
    required: Boolean,
    allowEditSchema: Boolean,
    hasDeleteButton: {
      type: Boolean,
      required: false,
    }
  },
  data: () => {
    return {
      value: undefined as common.ValueType | undefined,
      buttonGroupStyle: common.buttonGroupStyleString,
    }
  },
  computed: {
    className(): string {
      const rowClass = this.theme.row
      return (this.schema as any).className ? rowClass + ' ' + (this.schema as any).className : rowClass
    },
    titleToShow(): string {
      return common.getTitle((this.schema as any).title, this.title)
    },
    isReadOnly(): boolean | undefined {
      return this.readonly || (this.schema as any).readonly
    },
    hasDeleteButtonFunction(): boolean {
      return this.hasDeleteButton && !this.isReadOnly
    },
  },
  methods: {
    toggleOptional() {
      this.value = common.toggleOptional(this.value, this.schema as any, this.initialValue) as any
      this.$emit('update-value', { value: this.getAllValue(), isValid: true })
    },
    getAllValue() {
      if (this.plainStruct || this.value === undefined) {
        return this.value
      }
      return this.value
    },
    emitValue() {
      this.$emit('update-value', { value: this.getAllValue(), isValid: true })
    },
    getValue(): any {
      return common.getDefaultValue(this.required, this.schema as any, this.initialValue)
    },
  },
  beforeMount() {
    this.value = this.getValue()
  },
  components: {
    XCircleIcon,
    Optional,
    Description,
  }
}
</script>
