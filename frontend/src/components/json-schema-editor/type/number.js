const value = {
    description: null,
    maximum: null,
    minimum: null,
    exclusiveMaximum: null,
    exclusiveMinimum: null,
    enum: []
}
const attr = {
    description: {
      name: 'Description',
      type: 'string',
    },
    maximum: {
        name: 'Maximum',
        type: 'number'
    },
    minimum: {
        name: 'Minimum',
        type: 'number'
    },
    exclusiveMaximum: {
        name: 'Exclusive maximum',
        type: 'number'
    },
    exclusiveMinimum: {
        name: 'Exclusive minimum',
        type: 'number'
    },
    enum: {
        name: 'Enum',
        type: 'array'
    }
}
const wrapper = { value, attr }
export default wrapper
