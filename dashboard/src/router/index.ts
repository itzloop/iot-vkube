import Controller from '@/components/Controller.vue'
import ControllersVue from '@/components/Controllers.vue'
import Devices from '@/components/Devices.vue'
import Device from '@/components/Device.vue'
import PodsVue from '@/components/Pods.vue'
import { compile } from '@vue/compiler-dom'
import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
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
      path: '/pods',
      name: 'pods',
      component: PodsVue
    }
  ]
})

export default router
