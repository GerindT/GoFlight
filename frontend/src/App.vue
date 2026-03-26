<script setup>
import { ref } from "vue";

const flightNumber = ref("LH123");
const loading = ref(false);
const error = ref("");
const result = ref(null);
const apiBase = import.meta.env.VITE_API_BASE || "http://localhost:8080";

const fetchDashboard = async () => {
  loading.value = true;
  error.value = "";
  result.value = null;

  try {
    const res = await fetch(`${apiBase}/api/v1/dashboard/${encodeURIComponent(flightNumber.value.trim())}`);
    const data = await res.json();
    if (!res.ok) {
      throw new Error(data.error || "Request failed");
    }
    result.value = data;
  } catch (err) {
    error.value = err.message || "Request failed";
  } finally {
    loading.value = false;
  }
};
</script>

<template>
  <main class="page">
    <section class="panel">
      <h1>GoFlight</h1>
      <p class="subtitle">Flight + weather dashboard</p>

      <div class="row">
        <input v-model="flightNumber" placeholder="Enter flight (e.g. LH123)" />
        <button @click="fetchDashboard" :disabled="loading">
          {{ loading ? "Loading..." : "Check Flight" }}
        </button>
      </div>

      <p v-if="error" class="error">{{ error }}</p>

      <div v-if="result" class="grid">
        <article class="card">
          <h2>Flight</h2>
          <p><strong>Number:</strong> {{ result.flight?.flight_number }}</p>
          <p><strong>Airline:</strong> {{ result.flight?.airline }}</p>
          <p><strong>Status:</strong> {{ result.flight?.status }}</p>
          <p><strong>From:</strong> {{ result.flight?.departure }}</p>
          <p><strong>To:</strong> {{ result.flight?.destination }}</p>
          <p><strong>Delay:</strong> {{ result.flight?.delay_in_minutes }} min</p>
        </article>

        <article class="card">
          <h2>Weather</h2>
          <p><strong>Location:</strong> {{ result.weather?.location }}</p>
          <p><strong>Condition:</strong> {{ result.weather?.condition }}</p>
          <p><strong>Temp:</strong> {{ result.weather?.temperature_c }} C</p>
          <p><strong>Feels like:</strong> {{ result.weather?.feels_like_c }} C</p>
          <p><strong>Wind:</strong> {{ result.weather?.wind_speed_kph }} kph</p>
          <p><strong>Humidity:</strong> {{ result.weather?.humidity_percent }}%</p>
        </article>
      </div>
    </section>
  </main>
</template>
