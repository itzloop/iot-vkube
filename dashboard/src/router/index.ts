import Controller from '@/components/Controller.vue'
import ControllersVue from '@/components/Controllers.vue'
import Devices from '@/components/Devices.vue'
import Device from '@/components/Device.vue'
import NodesVue from '@/components/Nodes.vue'
import { compile } from '@vue/compiler-dom'
import { createRouter, createWebHistory } from 'vue-router'
import Node from '@/components/Node.vue'
import Home from '@/components/Home.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: Home
    },
    {
      path: '/controllers',
      name: 'cotnrollers',
      component: ControllersVue
    },
    {
      path: '/controllers/:name',
      name: 'cotnroller',
      component: Controller,
      children: [
        {
          path: '/devices',
          component: Devices,
          children: [
            {
              path: '/:name',
              component: Device
            }
          ]
        }
      ]
    },
    {
      path: '/nodes',
      name: 'nodes',
      component: NodesVue
    },
    {
      path: '/nodes/:name',
      name: 'node',
      component: Node
    }
  ]
})

export default router
