<template>
  <div :class="className">
    <div :class="theme.title">
      <div class="flex space-x-1 w-full">
        <div class="pl-1" v-if="titleToShow != ''">{{titleToShow}}</div>
      </div>
      <div :class="theme.buttonGroup" :style="buttonGroupStyle">
        <optional :required="required"
          :value="value"
          :isReadOnly="isReadOnly"
          :theme="theme"
          :locale="locale"
          @toggleOptional="toggleOptional()">
        </optional>
      </div>
    </div>
    <div class="flex" v-if="value !== undefined">
      <input v-if="useInput"
       :class="[errorMessage ? theme.errorInput : theme.input]"
       type="number"
       @change="onChange($event)"
       @keyup="onChange($event)"
       :value="value"
       :disabled="isReadOnly"
       :step="step"
      />
      <select v-if="useSelectComponent"
              :class="theme.select"
              :value="value"
              :disabled="isReadOnly"
              @change="updateSelection(+$event.target.value)">
        <option v-for="option in options" :key="option.value" :value="option.value">{{option.label}}</option>
      </select>
      <div v-if="useRadioBoxComponent" class="w-full">
        <span v-for="option in options" :key="option.value" :class="theme.radiobox">
          <label :class="theme.label">
            <input type="radio"
                   @change="updateSelection(option.value)"
                   :checked="value === option.value"
                   :disabled="isReadOnly" />
            {{option.label}}
          </label>
        </span>
      </div>
      <button v-if="hasDeleteButtonFunction" type="button" class="ml-1 w-4 text-sky-500 inline-block cursor-pointer ml-1" @click="$emit('delete')" title="Delete">
        <XCircleIcon></XCircleIcon>
      </button>
    </div>
    <description :theme="theme" :message="schema.description"></description>
    <description :theme="theme" :message="errorMessage" :error="true"></description>
  </div>
</template>
<script lang="ts">
import type { PropType } from 'vue'
import * as common from './common'
import Optional from './Optional.vue'
import Description from './Description.vue'
import { XCircleIcon } from '@heroicons/vue/24/outline'

export default {
  emits: ['update-value', 'delete'],
  components: {
    XCircleIcon,
    optional: Optional,
    description: Description
  },
  props: {
    schema: {
      type: Object as PropType<common.NumberSchema>,
      required: true,
    },
    allowEditSchema: Boolean,
    allowLookup: Boolean,
    plainStruct: Boolean,
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
    readonly: Boolean,
    required: Boolean,
    hasDeleteButton: {
      type: Boolean,
      required: false
    }
  },
  data: () => {
    return {
      value: 0 as number | undefined,
      errorMessage: '' as string | undefined,
      buttonGroupStyle: common.buttonGroupStyleString,
    }
  },
  beforeMount() {
    this.value = this.getValue()
    this.validate()
    if (this.value !== undefined) {
      this.emitValue()
    }
  },
  computed: {
    useInput(): boolean | undefined {
      return this.value !== undefined && (this.schema.enum === undefined || this.isReadOnly)
    },
    useSelect(): boolean {
      return this.value !== undefined && (this.schema.enum !== undefined && !this.isReadOnly)
    },
    useSelectComponent(): boolean {
      return this.useSelect && (this.schema.format === 'select' || this.schema.format === undefined)
    },
    useRadioBoxComponent(): boolean {
      return this.useSelect && this.schema.format === 'radiobox'
    },
    isReadOnly(): boolean | undefined {
      return this.readonly || this.schema.readonly
    },
    hasDeleteButtonFunction(): boolean {
      return this.hasDeleteButton && !this.isReadOnly
    },
    titleToShow(): string {
      return common.getTitle(this.schema.title, this.title)
    },
    options(): {
      value: string | number;
      label: string | number;
    }[] {
      return common.getOptions(this.schema)
    },
    className(): string {
      const rowClass = this.errorMessage ? this.theme.errorRow : this.theme.row
      return this.schema.className ? rowClass + ' ' + this.schema.className : rowClass
    },
    step(): number | "any" | undefined {
      return common.getNumberStep(this.schema)
    },
  },
  methods: {
    onChange(e: { target: { value: string } }) {
      this.value = this.schema.type === 'integer' ? common.toInteger(e.target.value) : common.toNumber(e.target.value)
      this.validate()
      this.emitValue()
    },
    updateSelection(value: number) {
      this.value = value
      this.validate()
      this.emitValue()
    },
    toggleOptional() {
      this.value = common.toggleOptional(this.value, this.schema, this.initialValue) as number | undefined
      this.validate()
      this.emitValue()
    },
    validate() {
      this.errorMessage = common.getErrorMessageOfNumber(this.value, this.schema, this.locale)
    },
    emitValue() {
      this.$emit('update-value', { value: this.getAllValue(), isValid: !this.errorMessage })
    },
    getAllValue() {
      if(this.value === undefined || this.plainStruct) {
        return this.value
      }
      return this.value ?? 0
    },
    getValue(): number {
      return common.getDefaultValue(this.required, this.schema, this.initialValue) as number
    },
  }
}
</script>
