<template>
  <div class="w-screen h-[95dvh] max-w-full max-h-full flex flex-col overflow-hidden m-0 p-0">
    <LoginForm v-if="!status.loggedIn" @login="login" />
    <div v-else class="flex-1 flex flex-col overflow-hidden min-h-0">
      <div class="flex items-center justify-between px-4 py-2">
        <Tabs value="server" class="flex-1 overflow-hidden min-h-0 h-full">
          <TabList>
            <Tab value="server">Server</Tab>
            <Tab value="mods">Mods</Tab>
            <Tab value="factorio">Factorio</Tab>
            <Tab value="rcon" :disabled="!status.running">RCon</Tab>
            <Tab value="saves">Saves</Tab>
            <Tab value="settings">Settings</Tab>
            <div class="flex items-center space-x-2 ml-auto pr-2 pointer-events-auto">
              <i class="pi pi-user" :class="darkMode ? 'text-gray-200' : 'text-gray-500'"></i>
              <span :class="['font-medium', darkMode ? 'text-gray-200' : 'text-gray-500']">{{ username }}</span>
              <Button type="button" icon="pi pi-ellipsis-v" @click="toggle" aria-haspopup="true"
                aria-controls="overlay_menu" text />
              <Menu ref="menu" id="overlay_menu" :model="items" :popup="true" />
            </div>
          </TabList>
          <TabPanels class="flex-1 overflow-hidden min-h-0 h-full">
            <TabPanel value="server" class="flex-1 overflow-hidden h-full min-h-0">
              <ServerControls :status="status" @refreshStatus="updateStatus" />
            </TabPanel>
            <TabPanel value="mods" class="flex-1 overflow-hidden h-full min-h-0">
              <Mods />
            </TabPanel>
            <TabPanel value="factorio" class="flex-1 overflow-hidden h-full min-h-0">
              <FactorioSettings />
            </TabPanel>
            <TabPanel value="rcon" class="flex-1 overflow-hidden h-full min-h-0">
              <RconConsole />
            </TabPanel>
            <TabPanel value="saves" class="flex-1 overflow-hidden h-full min-h-0">
              <SavesManager />
            </TabPanel>
            <TabPanel value="settings" class="flex-1 overflow-hidden h-full min-h-0">
              <ServerSettings />
            </TabPanel>
          </TabPanels>
        </Tabs>
      </div>
    </div>
    <Toast />
  </div>
</template>

<script setup>
import '@fortawesome/fontawesome-free/css/all.min.css'
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { startServer, stopServer, serverStatus } from '@/api'

import Avatar from 'primevue/avatar';
import Button from 'primevue/button';
import Menu from 'primevue/menu';
import Menubar from 'primevue/menubar';
import Tabs from 'primevue/tabs';
import TabList from 'primevue/tablist';
import Tab from 'primevue/tab';
import TabPanels from 'primevue/tabpanels';
import TabPanel from 'primevue/tabpanel';
import Toast from 'primevue/toast';

import ServerControls from '@/components/ServerControls.vue'
import Mods from '@/components/Mods.vue'
import RconConsole from '@/components/RconConsole.vue'
import SavesManager from '@/components/SavesManager.vue'
import ServerSettings from '@/components/ServerSettings.vue'
import LoginForm from '@/components/LoginForm.vue'
import FactorioSettings from '@/components/FactorioSettings.vue'
import { action } from '@primeuix/themes/aura/image';
import { useAppToast } from '@/composables/useAppToast'

const status = ref({})
const running = computed(() => status.value?.running)
const version = computed(() => status.value?.version)

const menu = ref()
const username = ref(localStorage.getItem('username') || '')

const darkMode = ref(localStorage.getItem('darkMode') === 'true')

if (darkMode.value) {
  document.documentElement.classList.add('my-app-dark')
}

const toggleDarkMode = () => {
  darkMode.value = !darkMode.value
  localStorage.setItem('darkMode', darkMode.value)
  document.documentElement.classList.toggle('my-app-dark', darkMode.value)
}

const items = computed(() => [
  {
    items: [
      {
        label: darkMode.value ? 'Light Mode' : 'Dark Mode',
        icon: darkMode.value ? 'pi pi-sun' : 'pi pi-moon',
        command: () => { toggleDarkMode() }
      },
      {
        label: 'Logout',
        icon: 'pi pi-sign-out',
        command: () => { logout() }
      }
    ]
  }
])

const toggle = (event) => {
  menu.value.toggle(event);
};

const logout = () => {
  localStorage.removeItem('username')
  localStorage.removeItem('password')
  clearInterval(statusIntervalId)
  status.value = { loggedIn: false }
  username.value = ''
}

const login = () => {
  updateStatus()
  username.value = localStorage.getItem('username') || ''
  if (statusIntervalId) clearInterval(statusIntervalId)
  statusIntervalId = setInterval(updateStatus, 5000)
}

const updateStatus = async () => {
  try {
    status.value = await serverStatus()
  } catch (e) {
    logout()
  }
}

let statusIntervalId

onMounted(() => {
  if (localStorage.getItem('username') && localStorage.getItem('password')) {
    username.value = localStorage.getItem('username') || ''
  }
  updateStatus()
})

onUnmounted(() => {
  clearInterval(statusIntervalId)
})
</script>