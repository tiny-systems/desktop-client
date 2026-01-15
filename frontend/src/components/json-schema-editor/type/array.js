const value = {
  description: null,
  minItems: null,
  maxItems: null,
  uniqueItems: false,
  tableMode: false
}
const attr = {
  description: {
    name: 'Description',
    type: 'string'
  },
  maxItems: {
    name: 'Max items',
    type: 'integer'
  },
  minItems: {
    name: 'Min Items',
    type: 'integer'
  },
  uniqueItems: {
    name: 'Unique',
    type: 'boolean'
  },
  tableMode: {
    name: 'Table Mode',
    type: 'boolean'
  }
}
const wrapper = { value, attr }
export default wrapper
