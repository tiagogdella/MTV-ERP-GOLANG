<template>
  <div class="p-8">
    <NuxtRouteAnnouncer />
    <h1 class="text-2xl font-bold text-emerald-700">MTV ERP</h1>
    <p class="mt-2 text-gray-600">Sistema de gestão — arroz</p>

    <div class="mt-6 flex flex-col gap-3 max-w-sm">
      <InputText placeholder="E-mail" />
      <Password placeholder="Senha" :feedback="false" toggle-mask />
      <Button label="Entrar" />
    </div>
  </div>
</template>
