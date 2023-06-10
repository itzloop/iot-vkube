import axios from 'axios'
import type { Controller, Device, Pod } from './types'
import { Axios } from 'axios'
import { def } from '@vue/shared'

class Api {
  private baseUrl: string

  public ControllersApi: ControllersApi
  public DevicesApi: DevicesApi
  constructor(baseUrl: string, controllersApi: ControllersApi, devicesApi: DevicesApi) {
    this.baseUrl = baseUrl
    this.ControllersApi = controllersApi
    this.DevicesApi = devicesApi
  }
}

class ControllersApi {
  private baseUrl: string
  constructor(baseUrl: string) {
    this.baseUrl = baseUrl
  }

  async list(): Promise<Controller[]> {
    try {
      const res = await axios.request<Controller[]>({
        url: `${this.baseUrl}/controllers`,
        method: 'get'
      })
      return res.data
    } catch (err) {
      console.log('list', err)
      throw err
    }
  }

  async get(name: string): Promise<Controller> {
    try {
      const res = await axios.request<Controller>({
        url: `${this.baseUrl}/controllers/${name}`,
        method: 'get'
      })
      return res.data
    } catch (err) {
      console.log('get', err)
      throw err
    }
  }

  async create(controller: Controller): Promise<void> {
    try {
      await axios.request<Controller>({
        url: `${this.baseUrl}/controllers`,
        method: 'post',
        data: JSON.stringify(controller)
      })
    } catch (err) {
      console.log('create', err)
      throw err
    }
  }

  async delete(controller: Controller): Promise<void> {
    try {
      await axios.request<Controller>({
        url: `${this.baseUrl}/controllers/${controller.name}`,
        method: 'delete'
      })
    } catch (err) {
      console.log('delete', err)
      throw err
    }
  }

  async update(controller: Controller): Promise<void> {
    try {
      await axios.request<Controller>({
        url: `${this.baseUrl}/controllers/${controller.name}`,
        method: 'PATCH',
        data: JSON.stringify(controller)
      })
    } catch (err) {
      console.log('update', err)
      throw err
    }
  }
}

class DevicesApi {
  private baseUrl: string
  constructor(baseUrl: string) {
    this.baseUrl = baseUrl
  }
  async list(controllerName: string): Promise<Device[]> {
    try {
      const res = await axios.request<Device[]>({
        url: `${this.baseUrl}/controllers/${controllerName}/devices`,
        method: 'get'
      })
      return res.data
    } catch (err) {
      console.log('list', err)
      throw err
    }
  }

  async get(controllerName: string, name: string): Promise<Device> {
    try {
      const res = await axios.request<Device>({
        url: `${this.baseUrl}/controllers/${controllerName}/devices/${name}`,
        method: 'get'
      })
      return res.data
    } catch (err) {
      console.log('get', err)
      throw err
    }
  }

  async create(controllerName: string, device: Device): Promise<void> {
    try {
      await axios.request<Device>({
        url: `${this.baseUrl}/controllers/${controllerName}/devices`,
        method: 'post',
        data: JSON.stringify(device)
      })
    } catch (err) {
      console.log('create', err)
      throw err
    }
  }

  async delete(controllerName: string, device: Device): Promise<void> {
    try {
      await axios.request<Device>({
        url: `${this.baseUrl}/controllers/${controllerName}/devices/${device.name}`,
        method: 'delete'
      })
    } catch (err) {
      console.log('delete', err)
      throw err
    }
  }

  async update(controllerName: string, device: Device): Promise<void> {
    try {
      await axios.request<Device>({
        url: `${this.baseUrl}/controllers/${controllerName}/devices/${device.name}`,
        method: 'PATCH',
        data: JSON.stringify(device)
      })
    } catch (err) {
      console.log('update', err)
      throw err
    }
  }
}

class PodsApi {
  private baseUrl: string
  constructor(baseUrl: string) {
    this.baseUrl = baseUrl
  }
  async list(nodeName: string): Promise<Pod[]> {
    try {
      const res = await axios.request<Pod[]>({
        url: `${this.baseUrl}/nodes/${nodeName}/pods`,
        method: 'get'
      })
      return res.data
    } catch (err) {
      console.log('list', err)
      throw err
    }
  }

}

class NodesApi {
    private baseUrl: string
    constructor(baseUrl: string) {
      this.baseUrl = baseUrl
    }
    async list(): Promise<Node[]> {
        try {
          const res = await axios.request<Node[]>({
            url: `${this.baseUrl}/nodes`,
            method: 'get'
          })
          return res.data
        } catch (err) {
          console.log('list', err)
          throw err
        }
      }
}

const API = new Api(
  'http://localhost:5000',
  new ControllersApi('http://localhost:5000'),
  new DevicesApi('http://localhost:5000')
)

export default API
