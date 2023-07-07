interface Controller {
  name: string
  host: string
  readiness: boolean
  // devices: []Device ?
}

interface Device {
  name: string
  readiness: boolean
}

interface Pod {
  name: string
  namespace: string
  readiness: boolean
}

interface Node {
  name: string
  cpu: string
  memory: string
  allocatablePods: string
  maxPods: string
  readiness: boolean
}

export type { Controller, Device, Pod, Node }
