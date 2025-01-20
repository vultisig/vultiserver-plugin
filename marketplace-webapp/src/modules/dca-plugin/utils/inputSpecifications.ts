export const allocate_from_validation = {
  name: 'allocateAmount',
  label: 'I want to Allocate',
  type: 'number',
  id: 'allocateAmount',
  placeholder: '',
  validation: {
    required: {
      value: true,
      message: 'required',
    },
  },
}

export const time_period_validation = {
  name: 'timePeriod',
  label: 'Every',
  type: 'number',
  id: 'timePeriod',
  placeholder: '',
  validation: {
    required: {
      value: true,
      message: 'required',
    },
    min: {
      value: 15,
      message: "min 15 minutes"
    },
  },
}

export const orders_validation = {
  name: 'orders',
  label: 'Over',
  type: 'number',
  id: 'orders',
  placeholder: '',
  validation: {
    required: {
      value: true,
      message: 'required',
    },
    min: {
      value: 2,
      message: "Number of Orders cannot be lower than 2"
    },
  },
}
