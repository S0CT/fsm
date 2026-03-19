<template>
  <div class="flex items-center justify-center min-h-screen">
    <div
      class="card flex justify-center max-w-lg w-full mx-auto p-8 rounded-lg shadow-lg border border-gray-300 bg-white">
      <Form @submit.prevent="onFormSubmit">
        <h2 class="text-2xl font-semibold mb-6 text-center text-gray-800">
          <div>Factorio Server Manager</div>
          <div>Login</div>
        </h2>
        <div class="flex flex-col gap-1 py-1">
          <InputText id="username" placeholder="Username" v-model="username" fluid />
        </div>
        <div class="flex flex-col gap-1 py-1">
          <Password name="password" placeholder="Password" :feedback="false" v-model="password" fluid />
        </div>
        <div class="flex flex-col gap-1 py-1">
          <Button type="submit" severity="primary" label="Login" :disabled="!username || !password" />
        </div>
      </Form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { serverStatus } from '@/api';
import Button from 'primevue/button';
import InputText from 'primevue/inputtext';
import Password from 'primevue/password';

const emit = defineEmits(['login'])

const username = ref('')
const password = ref('')

const onFormSubmit = async () => {
  localStorage.setItem('username', username.value.trim())
  localStorage.setItem('password', password.value.trim())
  emit('login')
  username.value = ''
  password.value = ''
}
</script>