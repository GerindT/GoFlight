<script setup>
import { computed, nextTick, onMounted, ref, watch } from "vue";

const command = ref("");
const loading = ref(false);
const terminalHeight = ref(500);
const commandHistory = ref([]);
const historyIndex = ref(-1);
const sessionStartedAt = Date.now();
const showBootGif = ref(false);
const terminalBodyRef = ref(null);
const terminalRootRef = ref(null);

const history = ref([{ text: "GoFlight terminal initialized.", type: "info" }]);
const apiBase = import.meta.env.VITE_API_BASE || "http://localhost:8080";
const result = ref(null);
const typingLine = ref("");
const typingType = ref("info");

const promptText = "goflight@terminal:~$";
const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms));
const uptime = () => `${Math.floor((Date.now() - sessionStartedAt) / 1000)}s`;
const commandCatalog = ["help", "clear", "flight", "about", "status", "uptime", "neofetch", "demo", "boot", "height"];
const aliasMap = {
  h: "help",
  cls: "clear",
  c: "clear",
  f: "flight",
  nf: "neofetch",
  q: "clear",
  st: "status"
};

const pushLine = (text, type = "info") => history.value.push({ text, type });

const focusTerminal = () => terminalRootRef.value?.focus();
const scrollTerminalToBottom = () => {
  if (terminalBodyRef.value) terminalBodyRef.value.scrollTop = terminalBodyRef.value.scrollHeight;
};
const forceScrollToBottom = () => {
  nextTick(() => {
    scrollTerminalToBottom();
    setTimeout(scrollTerminalToBottom, 0);
  });
};

watch(
  [
    () => history.value.length,
    () => typingLine.value,
    () => loading.value,
    () => showBootGif.value,
    () => result.value
  ],
  async () => {
    await nextTick();
    scrollTerminalToBottom();
    focusTerminal();
  },
  { flush: "post" }
);

const prediction = computed(() => {
  const raw = command.value;
  if (!raw) return "help";
  const trimmed = raw.trimStart();
  if (trimmed.includes(" ")) return "";
  const lower = trimmed.toLowerCase();
  const match = commandCatalog.find((c) => c.startsWith(lower));
  return match && match !== lower ? match.slice(lower.length) : "";
});

const applyPrediction = () => {
  const suffix = prediction.value;
  if (!suffix) return false;
  command.value = `${command.value}${suffix}`;
  return true;
};

const normalizeCommand = (raw) => {
  const [cmd, ...args] = raw.split(/\s+/);
  const mapped = aliasMap[cmd.toLowerCase()] || cmd.toLowerCase();
  return { cmd: mapped, args };
};

watch(command, () => {
  focusTerminal();
});

const pushTypedLine = async (line, type = "info") => {
  typingLine.value = "";
  typingType.value = type;
  for (const ch of line) {
    typingLine.value += ch;
    await sleep(8);
  }
  pushLine(typingLine.value, type);
  typingLine.value = "";
};

const printNeofetch = () => {
  pushLine(`${promptText} neofetch`, "cmd");
  pushLine("   ____       ______ _ _       _     _   ");
  pushLine("  / ___| ___ |  ____| (_) __ _| |__ | |_ ");
  pushLine(" | |  _ / _ \\| |_  | | |/ _` | '_ \\| __|");
  pushLine(" | |_| | (_) |  _| | | | (_| | | | | |_ ");
  pushLine("  \\____|\\___/|_|   |_|_|\\__, |_| |_|\\__|");
  pushLine("                        |___/            ");
  pushLine(` os: browser-ui  api: ${apiBase}  uptime: ${uptime()}`, "ok");
};

const executeFlight = async (flight) => {
  const query = flight.toUpperCase();
  pushLine(`${promptText} flight ${query}`, "cmd");
  loading.value = true;
  result.value = null;

  try {
    const res = await fetch(`${apiBase}/api/v1/dashboard/${encodeURIComponent(query)}`);
    const data = await res.json();
    if (!res.ok) throw new Error(data.error || "request failed");
    result.value = data;
    pushLine(`[ok] fetched ${data.flight?.flight_number} (${data.flight?.status})`, "ok");
  } catch (err) {
    pushLine(`[error] ${err.message || "request failed"}`, "error");
  } finally {
    loading.value = false;
  }
};

const runCommand = async (rawCommand) => {
  const raw = rawCommand.trim();
  if (!raw) return;

  if (raw.includes("&&")) {
    const segments = raw
      .split("&&")
      .map((s) => s.trim())
      .filter(Boolean);
    for (const segment of segments) {
      await runCommand(segment);
    }
    return;
  }

  const { cmd: normalized, args } = normalizeCommand(raw);

  commandHistory.value.push(raw);
  historyIndex.value = commandHistory.value.length;
  command.value = "";

  if (normalized === "help") {
    pushLine(`${promptText} help`, "cmd");
    await pushTypedLine("Available commands:");
    pushLine("- help                 Show this message");
    pushLine("- clear                Clear terminal output");
    pushLine("- flight <number>      Fetch flight dashboard");
    pushLine("- about                About this project");
    pushLine("- status               Local backend/terminal status");
    pushLine("- uptime               Show terminal uptime");
    pushLine("- neofetch             Show terminal splash");
    pushLine("- demo                 Run fake scan animation");
    pushLine("- boot                 Show startup animation");
    pushLine("- height <px>          Set terminal body height (260-900)");
    pushLine("- aliases              h, cls, c, f, nf, q, st");
    pushLine("- completion           Tab");
    return;
  }

  if (normalized === "clear") {
    history.value = [];
    result.value = null;
    showBootGif.value = false;
    return;
  }

  if (normalized === "about") {
    pushLine(`${promptText} about`, "cmd");
    pushLine("GoFlight aggregates flight + weather data via Go, cache, and resilience patterns.", "info");
    return;
  }

  if (normalized === "status") {
    pushLine(`${promptText} status`, "cmd");
    pushLine(`[status] api_base=${apiBase}`, "info");
    pushLine(`[status] loading=${loading.value} uptime=${uptime()} height=${terminalHeight.value}px`, "info");
    return;
  }

  if (normalized === "uptime") {
    pushLine(`${promptText} uptime`, "cmd");
    pushLine(`[uptime] ${uptime()}`, "ok");
    return;
  }

  if (normalized === "neofetch") {
    printNeofetch();
    return;
  }

  if (normalized === "height") {
    pushLine(`${promptText} ${raw}`, "cmd");
    const n = Number(args[0]);
    if (!Number.isFinite(n)) {
      pushLine("[error] usage: height <number>", "error");
      return;
    }
    terminalHeight.value = Math.min(900, Math.max(260, Math.trunc(n)));
    pushLine(`[ok] terminal height set to ${terminalHeight.value}px`, "ok");
    return;
  }

  if (normalized === "boot") {
    pushLine(`${promptText} boot`, "cmd");
    loading.value = true;
    showBootGif.value = true;
    const frames = ["[=     ]", "[==    ]", "[===   ]", "[====  ]", "[===== ]", "[======]"];
    for (const frame of frames) {
      pushLine(`[boot] ${frame} warming up engines...`);
      await sleep(120);
    }
    loading.value = false;
    pushLine("[ok] boot sequence complete", "ok");
    return;
  }

  if (normalized === "demo") {
    pushLine(`${promptText} demo`, "cmd");
    loading.value = true;
    const frames = ["[=     ]", "[==    ]", "[===   ]", "[====  ]", "[===== ]", "[======]"];
    for (const frame of frames) {
      pushLine(`[scan] ${frame} scanning flight index`);
      await sleep(120);
    }
    loading.value = false;
    pushLine("[ok] demo complete", "ok");
    return;
  }

  if (normalized === "flight") {
    if (!args[0]) {
      pushLine(`${promptText} ${raw}`, "cmd");
      pushLine("[error] usage: flight <number>", "error");
      return;
    }
    await executeFlight(args[0]);
    return;
  }

  pushLine(`${promptText} ${raw}`, "cmd");
  pushLine(`[error] unknown command: ${normalized} (try 'help')`, "error");
  forceScrollToBottom();
};

const onTerminalKeyDown = async (event) => {
  if ((event.ctrlKey || event.metaKey) && event.key.toLowerCase() === "l") {
    event.preventDefault();
    await runCommand("clear");
    return;
  }
  if (event.key === "Enter") {
    event.preventDefault();
    await runCommand(command.value);
    return;
  }
  if (event.key === "Tab") {
    event.preventDefault();
    applyPrediction();
    return;
  }
  if (event.key === "Backspace") {
    event.preventDefault();
    command.value = command.value.slice(0, -1);
    return;
  }
  if (event.key === "ArrowUp") {
    event.preventDefault();
    if (commandHistory.value.length === 0) return;
    historyIndex.value = Math.max(0, historyIndex.value - 1);
    command.value = commandHistory.value[historyIndex.value] || "";
    return;
  }
  if (event.key === "ArrowDown") {
    event.preventDefault();
    if (commandHistory.value.length === 0) return;
    historyIndex.value = Math.min(commandHistory.value.length, historyIndex.value + 1);
    command.value = historyIndex.value === commandHistory.value.length ? "" : commandHistory.value[historyIndex.value] || "";
    return;
  }
  if (event.key.length === 1 && !event.ctrlKey && !event.metaKey && !event.altKey) {
    event.preventDefault();
    command.value += event.key;
  }
};

onMounted(async () => {
  focusTerminal();
  pushLine(`${promptText} boot --ui terminal`, "cmd");
  pushLine("UI ready. Type `help` to see commands.", "info");
  printNeofetch();
  forceScrollToBottom();
});

const terminalLines = computed(() => history.value.slice(-120));
</script>

<template>
  <main class="page">
    <section class="terminal" ref="terminalRootRef" tabindex="0" @keydown="onTerminalKeyDown">
      <header class="terminal-bar">
        <span class="dot red"></span>
        <span class="dot yellow"></span>
        <span class="dot green"></span>
        <p>GoFlight Terminal</p>
      </header>

      <section class="terminal-body" ref="terminalBodyRef" :style="{ height: `${terminalHeight}px` }">
        <p class="line" :class="line.type" v-for="(line, idx) in terminalLines" :key="idx">{{ line.text }}</p>
        <p v-if="typingLine" class="line" :class="typingType">{{ typingLine }}</p>
        <p v-if="loading" class="line pending">[pending] querying upstream APIs...</p>
        <img
          v-if="showBootGif"
          class="boot-gif"
          src="https://media.giphy.com/media/13borq7Zo2kulO/giphy.gif"
          alt="kirby animation"
        />
      </section>

      <div class="input-row">
        <span class="prompt">{{ promptText }}</span>
        <span class="cmd-text">{{ command }}</span>
        <span class="cursor" aria-hidden="true"></span>
      </div>

      <div class="status-line">
        <span v-if="prediction">suggestion: {{ command }}{{ prediction }} (Tab)</span>
        <span v-else>ready</span>
      </div>

      <div v-if="result" class="result-grid">
        <div class="result-card">
          <h2>Flight</h2>
          <p><span>number</span>{{ result.flight?.flight_number }}</p>
          <p><span>airline</span>{{ result.flight?.airline }}</p>
          <p><span>status</span>{{ result.flight?.status }}</p>
          <p><span>from</span>{{ result.flight?.departure }}</p>
          <p><span>to</span>{{ result.flight?.destination }}</p>
          <p><span>delay_min</span>{{ result.flight?.delay_in_minutes }}</p>
        </div>

        <div class="result-card">
          <h2>Weather</h2>
          <p><span>location</span>{{ result.weather?.location }}</p>
          <p><span>condition</span>{{ result.weather?.condition }}</p>
          <p><span>temp_c</span>{{ result.weather?.temperature_c }}</p>
          <p><span>feels_like_c</span>{{ result.weather?.feels_like_c }}</p>
          <p><span>wind_kph</span>{{ result.weather?.wind_speed_kph }}</p>
          <p><span>humidity_pct</span>{{ result.weather?.humidity_percent }}</p>
        </div>
      </div>
    </section>
  </main>
</template>
