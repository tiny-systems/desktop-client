const value = {
    description: null,
    maxProperties: null,
    minProperties: null,
    additionalProperties: null
}
const attr = {
    description: {
      name: 'Description',
      type: 'string',
    },
    maxProperties: {
        name: 'Max properties',
        type: 'integer'
    },
    minProperties: {
        name: 'Min properties',
        type: 'integer'
    },
    additionalProperties: {
      name: 'Allow additional properties',
      type: 'boolean'
    }
}
const wrapper = { value, attr }
export default wrapper
