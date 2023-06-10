<script setup lang="ts">
import { ref } from 'vue'
import { RouterLink, RouterView } from 'vue-router'
import router from './router'
import { watch } from 'vue'

const drawer = ref(false)

const routePath = ref('IoT Dashboard')

// TODO make this a component
watch(router.currentRoute, (p) => {
  if (p.fullPath === '/') {
    routePath.value = 'IoT Dashboard'
    return
  }

  const splited = p.fullPath.split('/')
  let str = 'IoT Dashboard'
  for (let i = 0; i < splited.length; i++) {
    str += `${splited[i].substring(0, 1).toUpperCase() + splited[i].substring(1)}`
    if (i < splited.length - 1) {
      str += ' > '
    }
  }
  for (let s of splited) {
  }

  routePath.value = str
})
</script>

<template>
  <v-card>
    <v-app>
      <v-navigation-drawer color="teal-lighten-5" v-model="drawer">
        <v-list density="compact" nav>
          <v-list-item
            @click="$router.replace('/controllers')"
            prepend-icon="mdi-gamepad-up"
            title="Controllers"
            value="home"
          >
          </v-list-item>
          <v-list-item
            @click="$router.replace('/pods')"
            prepend-icon="mdi-cube-outline"
            title="Pods"
            value="home"
          ></v-list-item>
        </v-list>
      </v-navigation-drawer>

      <v-app-bar color="teal-darken-4">
        <template v-slot:prepend>
          <v-app-bar-nav-icon @click="drawer = !drawer"></v-app-bar-nav-icon>
        </template>
        <v-app-bar-title>{{ routePath }}</v-app-bar-title>
      </v-app-bar>

      <v-main style="min-height: 300px">
        <RouterView></RouterView>
      </v-main>
    </v-app>
  </v-card>
</template>

<style scoped></style>
