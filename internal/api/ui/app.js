const dictionaries = {
  en: {
    brandTagline: "Archive orchestration", connecting: "Connecting", online: "Online", offline: "Offline", locked: "Locked", unlocked: "Unlocked", unlock: "Unlock", newJob: "New job",
    eyebrow: "CONTROL PLANE · SAFE BY DEFAULT", heroTitle: "Move complete media units, never loose fragments.", heroText: "Stage, verify, publish, and recover every directory as one auditable operation.",
    safetyTitle: "Protected launch mode", safetyText: "New jobs start as dry runs. Existing destinations fail closed unless immutable merge is explicitly selected.", sourcePreserved: "SOURCE PRESERVED",
    orchestration: "ORCHESTRATION", syncJobs: "Sync jobs", refresh: "Refresh", loadingJobs: "Loading jobs…", auditTrail: "AUDIT TRAIL", recentRuns: "Recent runs", live: "LIVE", noRuns: "No runs recorded yet.",
    footerText: "Deterministic routing · SQLite audit trail · rclone transport", jobConfig: "JOB CONFIGURATION", createJob: "Create a sync job", editJob: "Edit sync job", jobName: "Job name", source: "Source", sourceHint: "Use a mounted absolute path or an rclone remote path.",
    destinationName: "Destination ID", destinationPath: "Destination path", mode: "Mode", copyMode: "Copy · preserve source", moveMode: "Move · delete only after final verification", grouping: "Atomic unit", folderUnit: "Top-level folder", showUnit: "Complete show", seasonUnit: "Season", depthUnit: "Path depth",
    conflictPolicy: "Existing destination", failClosed: "Fail closed (recommended)", immutableMerge: "Immutable merge · never overwrite", verify: "Verification", checksum: "Checksum", sizeOnly: "Size only", settleSeconds: "Stable window (seconds)", concurrency: "Concurrency", depth: "Depth", schedule: "Schedule",
    dryRun: "Dry run", dryRunHint: "Discover and plan units without touching any file.", paused: "Paused", pausedHint: "Save the job but prevent manual execution.", cancel: "Cancel", saveJob: "Save job",
    unlockTitle: "Unlock the control plane", unlockText: "Enter the API token configured on this server. It remains in this browser tab only.", apiToken: "API token", forgetToken: "Forget token", confirmRun: "Confirm execution", typeJobName: "Type the job name to authorize a write run", runNow: "Run now",
    metricJobs: "Jobs", metricRunning: "Running", metricCompleted: "Completed", metricFailed: "Failed", configured: "configured", activeNow: "active now", allTime: "all time", needsReview: "needs review",
    noJobs: "No jobs yet. Create a dry-run plan to begin safely.", noRunsLong: "Runs will appear here with a durable state and verification result.", run: "Run", edit: "Edit", remove: "Delete", daysStable: "stable", destination: "destination", pausedLabel: "paused", dryLabel: "dry run", copyLabel: "copy", moveLabel: "move",
    group_folder: "folder", group_show: "show", group_season: "season", group_depth: "depth", conflict_fail: "fail closed", conflict_merge_immutable: "immutable merge",
    state_discovered: "discovered", state_staging: "staging", state_verifying: "verifying", state_publishing: "publishing", state_completed: "completed", state_failed: "failed",
    dryRunSummary: "This dry run will only discover and plan units for “{name}”. No files will change.", writeRunSummary: "This run can write to the destination. Move mode can delete verified source data. Review the job and confirm “{name}”.",
    tokenAccepted: "Control plane unlocked.", tokenForgotten: "Token removed from this tab.", jobSaved: "Job saved.", jobDeleted: "Job deleted.", runStarted: "Run started.", invalidConfirmation: "The job name does not match.", requestFailed: "Request failed", deleteConfirm: "Delete “{name}”? Run history will be retained.",
    sourceRoute: "{source} → {destination}", seconds: "{value}s", minutes: "{value}m", hours: "{value}h", days: "{value}d",
    analyze: "Analyze branches", analysisStarted: "Read-only branch analysis started.", branchAnalysis: "BRANCH ANALYSIS", archiveStatus: "Archive status", analysisNote: "Metadata coverage is a planning signal. Source deletion still requires a final checksum or size check.",
    analysis_archived: "archived", analysis_ready_to_verify: "ready to verify", analysis_partial: "partially archived", analysis_pending: "pending", analysis_conflict: "conflict", analysis_empty: "empty", analysis_running: "analyzing", analysis_failed: "analysis failed",
    summaryArchived: "Archived", summaryReady: "Ready to verify", summaryPartial: "Partial", summaryPending: "Pending", summaryConflict: "Conflicts", summaryEmpty: "Empty", analysisFiles: "{matched}/{source} source files matched · {destination} destination files · {coverage}% coverage",
    analysisContext: "Snapshot {time}", analysisBranch: "Branch {destination} · {sourceState}", sourcePresent: "source present", sourceShell: "empty source shell", sourceAbsent: "source absent"
  },
  "zh-CN": {
    brandTagline: "媒体归档编排", connecting: "正在连接", online: "在线", offline: "离线", locked: "已锁定", unlocked: "已解锁", unlock: "解锁", newJob: "新建任务",
    eyebrow: "控制平面 · 默认安全", heroTitle: "以完整媒体单元迁移，不再留下碎片。", heroText: "每个目录都经过暂存、校验、发布与可恢复审计。",
    safetyTitle: "受保护启动模式", safetyText: "新任务默认仅演练；目标已存在时默认失败，只有明确选择不可变合并才会合并。", sourcePreserved: "源数据受保护",
    orchestration: "任务编排", syncJobs: "同步任务", refresh: "刷新", loadingJobs: "正在读取任务…", auditTrail: "审计记录", recentRuns: "最近运行", live: "实时", noRuns: "暂无运行记录。",
    footerText: "确定性分流 · SQLite 审计 · rclone 传输", jobConfig: "任务配置", createJob: "创建同步任务", editJob: "编辑同步任务", jobName: "任务名称", source: "来源", sourceHint: "使用已挂载的绝对路径或 rclone remote 路径。",
    destinationName: "目标标识", destinationPath: "目标路径", mode: "模式", copyMode: "复制 · 保留来源", moveMode: "移动 · 最终校验后才删除来源", grouping: "原子单元", folderUnit: "顶层目录", showUnit: "整部剧", seasonUnit: "单季", depthUnit: "路径层级",
    conflictPolicy: "目标已存在", failClosed: "安全失败（推荐）", immutableMerge: "不可变合并 · 绝不覆盖", verify: "校验方式", checksum: "校验和", sizeOnly: "仅大小", settleSeconds: "稳定窗口（秒）", concurrency: "并发数", depth: "层级", schedule: "计划任务",
    dryRun: "仅演练", dryRunHint: "只发现和规划迁移单元，不触碰文件。", paused: "暂停", pausedHint: "保存任务，但禁止手动执行。", cancel: "取消", saveJob: "保存任务",
    unlockTitle: "解锁控制平面", unlockText: "输入服务器配置的 API Token；它只保留在当前浏览器标签页。", apiToken: "API Token", forgetToken: "忘记 Token", confirmRun: "确认执行", typeJobName: "输入任务名称以授权写入运行", runNow: "立即运行",
    metricJobs: "任务", metricRunning: "运行中", metricCompleted: "已完成", metricFailed: "失败", configured: "已配置", activeNow: "当前活动", allTime: "累计", needsReview: "需要检查",
    noJobs: "还没有任务。先创建一个仅演练计划，安全开始。", noRunsLong: "每次运行都会在这里记录持久状态和校验结果。", run: "运行", edit: "编辑", remove: "删除", daysStable: "稳定窗口", destination: "目标", pausedLabel: "已暂停", dryLabel: "仅演练", copyLabel: "复制", moveLabel: "移动",
    group_folder: "目录", group_show: "整部剧", group_season: "单季", group_depth: "层级", conflict_fail: "安全失败", conflict_merge_immutable: "不可变合并",
    state_discovered: "已发现", state_staging: "暂存中", state_verifying: "校验中", state_publishing: "发布中", state_completed: "已完成", state_failed: "失败",
    dryRunSummary: "“{name}”本次仅发现并规划迁移单元，不会修改任何文件。", writeRunSummary: "本次运行可以写入目标；移动模式会删除已通过最终校验的来源。请核对任务并确认“{name}”。",
    tokenAccepted: "控制平面已解锁。", tokenForgotten: "已从当前标签页移除 Token。", jobSaved: "任务已保存。", jobDeleted: "任务已删除。", runStarted: "任务已开始运行。", invalidConfirmation: "输入的任务名称不匹配。", requestFailed: "请求失败", deleteConfirm: "删除“{name}”？运行历史将保留。",
    sourceRoute: "{source} → {destination}", seconds: "{value} 秒", minutes: "{value} 分钟", hours: "{value} 小时", days: "{value} 天",
    analyze: "分析底层分支", analysisStarted: "只读分支分析已开始。", branchAnalysis: "底层分支分析", archiveStatus: "归档状态", analysisNote: "元数据覆盖率用于规划；删除来源之前仍会执行最终校验和或大小校验。",
    analysis_archived: "已归档", analysis_ready_to_verify: "待强校验/清理", analysis_partial: "部分归档", analysis_pending: "待归档", analysis_conflict: "存在冲突", analysis_empty: "空目录", analysis_running: "分析中", analysis_failed: "分析失败",
    summaryArchived: "已归档", summaryReady: "待强校验", summaryPartial: "部分归档", summaryPending: "待归档", summaryConflict: "冲突", summaryEmpty: "空目录", analysisFiles: "来源匹配 {matched}/{source} 个文件 · 目标 {destination} 个文件 · 覆盖率 {coverage}%",
    analysisContext: "快照时间 {time}", analysisBranch: "物理分支 {destination} · {sourceState}", sourcePresent: "来源存在", sourceShell: "来源为空目录壳", sourceAbsent: "来源不存在"
  }
};

const state = {
  language: localStorage.getItem("atomic-language") || (navigator.language.toLowerCase().startsWith("zh") ? "zh-CN" : "en"),
  token: sessionStorage.getItem("atomic-token") || "",
  health: null,
  jobs: [],
  runs: [],
  analyses: [],
  selectedJob: null,
  eventController: null,
  refreshTimer: null
};

const byId = id => document.getElementById(id);
const t = (key, values = {}) => {
  const dictionary = dictionaries[state.language] || dictionaries.en;
  let value = dictionary[key] ?? dictionaries.en[key] ?? key;
  for (const [name, replacement] of Object.entries(values)) value = value.replaceAll(`{${name}}`, String(replacement));
  return value;
};

function applyLanguage() {
  document.documentElement.lang = state.language;
  document.querySelectorAll("[data-i18n]").forEach(element => { element.textContent = t(element.dataset.i18n); });
  byId("languageButton").textContent = state.language === "en" ? "中" : "EN";
  renderDashboard();
}

function element(tag, className, text) {
  const node = document.createElement(tag);
  if (className) node.className = className;
  if (text !== undefined) node.textContent = text;
  return node;
}

function setConnection(kind, label) {
  const node = byId("connectionStatus");
  node.className = `connection ${kind}`;
  node.querySelector("span").textContent = label;
}

async function request(resource, options = {}) {
  const headers = new Headers(options.headers || {});
  if (options.body && !headers.has("Content-Type")) headers.set("Content-Type", "application/json");
  if (state.token) headers.set("Authorization", `Bearer ${state.token}`);
  const response = await fetch(`/api/${resource}`, { ...options, headers, credentials: "same-origin", cache: "no-store" });
  let payload = null;
  try { payload = await response.json(); } catch { payload = {}; }
  if (!response.ok) {
    if (response.status === 401) {
      setConnection("offline", t("locked"));
      openDialog("authDialog");
    }
    const error = new Error(payload.error || `${t("requestFailed")} (${response.status})`);
    error.status = response.status;
    throw error;
  }
  return payload;
}

async function bootstrap() {
  try {
    const response = await fetch("/api/health", { cache: "no-store" });
    if (!response.ok) throw new Error(`health ${response.status}`);
    state.health = await response.json();
    byId("version").textContent = state.health.version || "dev";
    if (state.health.authRequired && !state.token) {
      setConnection("offline", t("locked"));
      renderLocked();
      openDialog("authDialog");
      return;
    }
    await loadDashboard();
    connectEvents();
  } catch (error) {
    setConnection("offline", t("offline"));
    renderError(error);
  }
}

async function loadDashboard() {
  const [dashboard, jobs, runs, system, analyses] = await Promise.all([
    request("dashboard"), request("jobs"), request("runs?limit=100"), request("system"), request("analyses")
  ]);
  state.dashboard = dashboard;
  state.jobs = jobs || [];
  state.runs = runs || [];
  state.analyses = analyses || [];
  byId("version").textContent = system.version || state.health?.version || "dev";
  setConnection("online", t("online"));
  byId("authButton").textContent = state.health?.authRequired ? t("unlocked") : t("online");
  renderDashboard();
}

function scheduleRefresh() {
  clearTimeout(state.refreshTimer);
  state.refreshTimer = setTimeout(() => loadDashboard().catch(handleError), 280);
}

async function connectEvents() {
  state.eventController?.abort();
  const controller = new AbortController();
  state.eventController = controller;
  while (!controller.signal.aborted) {
    try {
      await consumeEventStream(controller);
    } catch (error) {
      if (controller.signal.aborted) return;
      console.warn("event stream disconnected", error);
    }
    await new Promise(resolve => setTimeout(resolve, 2500));
  }
}

async function consumeEventStream(controller) {
  const headers = new Headers({ Accept: "text/event-stream" });
  if (state.token) headers.set("Authorization", `Bearer ${state.token}`);
  const response = await fetch("/api/events", { headers, cache: "no-store", signal: controller.signal });
  if (!response.ok || !response.body) throw new Error(`event stream ${response.status}`);
  const reader = response.body.getReader();
  const decoder = new TextDecoder();
  let buffer = "";
  while (!controller.signal.aborted) {
    const { value, done } = await reader.read();
    if (done) return;
    buffer += decoder.decode(value, { stream: true });
    const frames = buffer.split("\n\n");
    buffer = frames.pop() || "";
    if (frames.some(frame => frame.split("\n").some(line => line.startsWith("data:")))) scheduleRefresh();
  }
}

function renderDashboard() {
  if (!state.dashboard) return;
  renderMetrics();
  renderJobs();
  renderRuns();
}

function renderMetrics() {
  const definitions = [
    ["metricJobs", state.dashboard.jobs || 0, "configured", "◇", "#8174f8"],
    ["metricRunning", state.dashboard.running || 0, "activeNow", "↻", "#65a7ff"],
    ["metricCompleted", state.dashboard.completed || 0, "allTime", "✓", "#37d8b4"],
    ["metricFailed", state.dashboard.failed || 0, "needsReview", "!", "#ff6f8c"]
  ];
  const container = byId("stats");
  container.replaceChildren(...definitions.map(([label, value, note, icon, color]) => {
    const card = element("article", "metric");
    card.style.setProperty("--metric-color", color);
    const head = element("div", "metric-head");
    head.append(element("span", "", t(label)), element("span", "metric-icon", icon));
    card.append(head, element("div", "metric-value", value), element("div", "metric-note", t(note)));
    return card;
  }));
}

function renderJobs() {
  const container = byId("jobs");
  if (!state.jobs.length) {
    const empty = element("div", "empty-state");
    empty.append(element("p", "", t("noJobs")));
    container.replaceChildren(empty);
    return;
  }
  container.replaceChildren(...state.jobs.map(job => {
    const card = element("article", `job-card job-${job.mode || "copy"}`);
    const primary = element("div", "job-primary");
    primary.append(element("div", "job-symbol", job.mode === "move" ? "M" : "C"));
    const content = element("div", "job-content");
    const nameLine = element("div", "job-name-line");
    nameLine.append(element("span", "job-name", job.name));
    if (job.dryRun) nameLine.append(element("span", "pill dry", t("dryLabel")));
    nameLine.append(element("span", `pill ${job.mode}`, t(job.mode === "move" ? "moveLabel" : "copyLabel")));
    if (job.paused) nameLine.append(element("span", "pill paused", t("pausedLabel")));
    const analysis = state.analyses.find(item => item.jobId === job.id);
    if (analysis) {
      const rollup = analysisRollup(analysis);
      const analysisPill = element("button", `pill analysis ${rollup}`, t(`analysis_${rollup.replaceAll("-", "_")}`));
      analysisPill.type = "button";
      analysisPill.addEventListener("click", () => openAnalysisDetail(job));
      nameLine.append(analysisPill);
    }
    const destination = job.destinations?.[0]?.path || "—";
    const route = element("div", "job-route", t("sourceRoute", { source: job.source, destination }));
    route.title = route.textContent;
    const meta = element("div", "job-meta");
    meta.append(
      element("span", "", `◫ ${t(`group_${job.grouping}`)}`),
      element("span", "", `◷ ${formatDuration(job.settleSeconds || 0)}`),
      element("span", "", `⇶ ${job.concurrency || 1}`),
      element("span", "", `◇ ${t(`conflict_${String(job.conflictPolicy || "fail").replaceAll("-", "_")}`)}`)
    );
    content.append(nameLine, route, meta);
    primary.append(content);

    const actions = element("div", "job-actions");
    const runButton = element("button", "mini-button run", `▶ ${t("run")}`);
    runButton.type = "button";
    runButton.disabled = Boolean(job.paused);
    runButton.addEventListener("click", () => openRun(job));
    const editButton = element("button", "mini-button", "✎");
    editButton.type = "button"; editButton.title = t("edit"); editButton.setAttribute("aria-label", `${t("edit")}: ${job.name}`);
    editButton.addEventListener("click", () => openJob(job));
    const analyzeButton = element("button", "mini-button", "◎");
    analyzeButton.type = "button"; analyzeButton.title = t("analyze"); analyzeButton.setAttribute("aria-label", `${t("analyze")}: ${job.name}`);
    analyzeButton.disabled = analysis?.state === "running";
    analyzeButton.addEventListener("click", () => startAnalysis(job));
    const deleteButton = element("button", "mini-button", "⌫");
    deleteButton.type = "button"; deleteButton.title = t("remove"); deleteButton.setAttribute("aria-label", `${t("remove")}: ${job.name}`);
    deleteButton.addEventListener("click", () => deleteJob(job));
    actions.append(runButton, analyzeButton, editButton, deleteButton);
    card.append(primary, actions);
    return card;
  }));
}

function analysisRollup(analysis) {
  if (analysis.state === "running") return "running";
  if (analysis.state === "failed") return "failed";
  const summary = analysis.summary || {};
  for (const status of ["conflict", "partial", "pending", "ready-to-verify", "archived", "empty"]) {
    if ((summary[status] || 0) > 0) return status;
  }
  return "empty";
}

function openAnalysis(job, analysis) {
  byId("analysisTitle").textContent = `${job.name} · ${t("archiveStatus")}`;
  const snapshot = analysis.finishedAt || analysis.startedAt;
  const snapshotTime = snapshot ? new Intl.DateTimeFormat(state.language, { dateStyle: "medium", timeStyle: "medium" }).format(new Date(snapshot)) : "—";
  byId("analysisContext").textContent = t("analysisContext", { time: snapshotTime });
  const summaryDefinitions = [
    ["summaryArchived", "archived"], ["summaryReady", "ready-to-verify"],
    ["summaryPartial", "partial"], ["summaryPending", "pending"],
    ["summaryConflict", "conflict"], ["summaryEmpty", "empty"]
  ];
  byId("analysisSummary").replaceChildren(...summaryDefinitions.map(([label, key]) => {
    const card = element("div", "analysis-summary-card");
    card.append(element("span", "", t(label)), element("strong", "", analysis.summary?.[key] || 0));
    return card;
  }));
  const unitsContainer = byId("analysisUnits");
  if (analysis.state === "running") {
    const loading = element("div", "empty-state compact-state");
    loading.append(element("span", "loader"), element("p", "", t("analysis_running")));
    unitsContainer.replaceChildren(loading);
  } else if (analysis.state === "failed") {
    const failed = element("div", "empty-state compact-state");
    failed.append(element("p", "", `${t("analysis_failed")}: ${analysis.message || ""}`));
    unitsContainer.replaceChildren(failed);
  } else {
    unitsContainer.replaceChildren(...(analysis.units || []).map(unit => {
      const row = element("article", "analysis-unit");
      const content = element("div", "");
      const name = element("div", "analysis-unit-name", unit.unit);
      name.title = unit.unit;
      const sourceState = !unit.sourcePresent ? t("sourceAbsent") : unit.sourceFiles === 0 ? t("sourceShell") : t("sourcePresent");
      const branch = t("analysisBranch", { destination: unit.destination || "—", sourceState });
      const files = t("analysisFiles", {
        matched: unit.matchingFiles, source: unit.sourceFiles,
        destination: unit.destinationFiles, coverage: unit.coverage
      });
      const meta = element("div", "analysis-unit-meta", `${branch} · ${files}`);
      const evidence = [...(unit.conflictSamples || []), ...(unit.missingSamples || [])];
      if (evidence.length) meta.title = evidence.join("\n");
      content.append(name, meta);
      const status = element("span", `analysis-status ${unit.status}`, t(`analysis_${unit.status.replaceAll("-", "_")}`));
      row.append(content, status);
      return row;
    }));
  }
  openDialog("analysisDialog");
}

async function openAnalysisDetail(job) {
  try {
    const analysis = await request(`jobs/${encodeURIComponent(job.id)}/analysis`);
    openAnalysis(job, analysis);
  } catch (error) { handleError(error); }
}

async function startAnalysis(job) {
  try {
    await request(`jobs/${encodeURIComponent(job.id)}/analysis`, { method: "POST" });
    toast(t("analysisStarted"), "success");
    await loadDashboard();
    await openAnalysisDetail(job);
  } catch (error) { handleError(error); }
}

function renderRuns() {
  const container = byId("runs");
  if (!state.runs.length) {
    const empty = element("div", "empty-state compact-state");
    empty.append(element("p", "", t("noRunsLong")));
    container.replaceChildren(empty);
    return;
  }
  container.replaceChildren(...state.runs.slice(0, 20).map(run => {
    const item = element("article", `run-item ${run.state}`);
    item.append(element("span", "run-dot"));
    const content = element("div", "run-content");
    const unit = element("div", "run-unit", run.unit || "—");
    unit.title = run.unit || "";
    const time = run.startedAt ? new Date(run.startedAt).toLocaleString(state.language) : "—";
    const info = element("div", "run-info", `${run.destination || "—"} · ${time}`);
    info.title = run.message || "";
    content.append(unit, info);
    item.append(content, element("span", "run-state", t(`state_${run.state}`)));
    return item;
  }));
}

function renderLocked() {
  state.dashboard = { jobs: 0, running: 0, completed: 0, failed: 0 };
  renderMetrics();
  const jobs = element("div", "empty-state"); jobs.append(element("p", "", t("unlockText")));
  const runs = element("div", "empty-state compact-state"); runs.append(element("p", "", t("unlockText")));
  byId("jobs").replaceChildren(jobs); byId("runs").replaceChildren(runs);
}

function renderError(error) {
  const empty = element("div", "empty-state");
  empty.append(element("p", "", `${t("requestFailed")}: ${error.message}`));
  byId("jobs").replaceChildren(empty);
}

function formatDuration(seconds) {
  if (seconds >= 86400 && seconds % 86400 === 0) return t("days", { value: seconds / 86400 });
  if (seconds >= 3600 && seconds % 3600 === 0) return t("hours", { value: seconds / 3600 });
  if (seconds >= 60 && seconds % 60 === 0) return t("minutes", { value: seconds / 60 });
  return t("seconds", { value: seconds });
}

function openDialog(id) {
  const dialog = byId(id);
  if (!dialog.open) dialog.showModal();
}

function openJob(job = null) {
  const form = byId("jobForm");
  form.reset();
  state.selectedJob = job;
  byId("jobDialogTitle").textContent = t(job ? "editJob" : "createJob");
  if (job) {
    form.elements.id.value = job.id;
    form.elements.name.value = job.name || "";
    form.elements.source.value = job.source || "";
    form.elements.destinationName.value = job.destinations?.[0]?.name || "gd-primary";
    form.elements.destinationPath.value = job.destinations?.[0]?.path || "";
    form.elements.mode.value = job.mode || "copy";
    form.elements.grouping.value = job.grouping || "folder";
    form.elements.conflictPolicy.value = job.conflictPolicy || "fail";
    form.elements.verify.value = job.verify || "checksum";
    form.elements.settleSeconds.value = job.settleSeconds ?? 2592000;
    form.elements.concurrency.value = job.concurrency || 2;
    form.elements.depth.value = job.depth || 1;
    form.elements.dryRun.checked = Boolean(job.dryRun);
    form.elements.paused.checked = Boolean(job.paused);
  }
  openDialog("jobDialog");
}

async function saveJob(event) {
  event.preventDefault();
  const form = event.currentTarget;
  const current = state.selectedJob;
  const primaryDestination = {
    name: form.elements.destinationName.value.trim(),
    path: form.elements.destinationPath.value.trim(),
    weight: current?.destinations?.[0]?.weight || 1
  };
  const payload = {
    name: form.elements.name.value.trim(), source: form.elements.source.value.trim(),
    destinations: [primaryDestination, ...(current?.destinations?.slice(1) || [])],
    mode: form.elements.mode.value, grouping: form.elements.grouping.value,
    depth: Number(form.elements.depth.value || 1), settleSeconds: Number(form.elements.settleSeconds.value || 0),
    concurrency: Number(form.elements.concurrency.value || 1), verify: form.elements.verify.value,
    conflictPolicy: form.elements.conflictPolicy.value, dryRun: form.elements.dryRun.checked,
    paused: form.elements.paused.checked, deleteSource: false,
    schedule: current?.schedule || "", include: current?.include || [], exclude: current?.exclude || []
  };
  const id = form.elements.id.value;
  try {
    await request(id ? `jobs/${encodeURIComponent(id)}` : "jobs", { method: id ? "PUT" : "POST", body: JSON.stringify(payload) });
    byId("jobDialog").close();
    toast(t("jobSaved"), "success");
    await loadDashboard();
  } catch (error) { handleError(error); }
}

async function deleteJob(job) {
  if (!window.confirm(t("deleteConfirm", { name: job.name }))) return;
  try {
    await request(`jobs/${encodeURIComponent(job.id)}`, { method: "DELETE" });
    toast(t("jobDeleted"), "success");
    await loadDashboard();
  } catch (error) { handleError(error); }
}

function openRun(job) {
  state.selectedJob = job;
  const destructive = !job.dryRun;
  byId("runSummary").textContent = t(destructive ? "writeRunSummary" : "dryRunSummary", { name: job.name });
  byId("dangerConfirm").hidden = !destructive;
  byId("runForm").elements.confirmation.value = "";
  openDialog("runDialog");
}

async function startRun(event) {
  event.preventDefault();
  const job = state.selectedJob;
  if (!job) return;
  if (!job.dryRun && event.currentTarget.elements.confirmation.value !== job.name) {
    toast(t("invalidConfirmation"), "error");
    return;
  }
  try {
    await request(`jobs/${encodeURIComponent(job.id)}/run`, { method: "POST" });
    byId("runDialog").close();
    toast(t("runStarted"), "success");
    scheduleRefresh();
  } catch (error) { handleError(error); }
}

async function authenticate(event) {
  event.preventDefault();
  const token = event.currentTarget.elements.token.value.trim();
  state.token = token;
  sessionStorage.setItem("atomic-token", token);
  try {
    await request("system");
    byId("authDialog").close();
    event.currentTarget.reset();
    toast(t("tokenAccepted"), "success");
    await loadDashboard();
    connectEvents();
  } catch (error) {
    state.token = "";
    sessionStorage.removeItem("atomic-token");
    handleError(error);
  }
}

function forgetToken() {
  state.token = "";
  sessionStorage.removeItem("atomic-token");
  state.eventController?.abort();
  byId("authForm").reset();
  toast(t("tokenForgotten"));
  if (state.health?.authRequired) {
    setConnection("offline", t("locked"));
    renderLocked();
  } else {
    byId("authDialog").close();
  }
}

function toast(message, type = "") {
  const node = element("div", `toast ${type}`, message);
  byId("toasts").append(node);
  setTimeout(() => node.remove(), 4500);
}

function handleError(error) {
  console.error(error);
  toast(error.message || t("requestFailed"), "error");
}

byId("languageButton").addEventListener("click", () => {
  state.language = state.language === "en" ? "zh-CN" : "en";
  localStorage.setItem("atomic-language", state.language);
  applyLanguage();
});
byId("newJobButton").addEventListener("click", () => openJob());
byId("refreshButton").addEventListener("click", () => loadDashboard().catch(handleError));
byId("authButton").addEventListener("click", () => openDialog("authDialog"));
byId("jobForm").addEventListener("submit", saveJob);
byId("runForm").addEventListener("submit", startRun);
byId("authForm").addEventListener("submit", authenticate);
byId("forgetTokenButton").addEventListener("click", forgetToken);
document.querySelectorAll("[data-close]").forEach(button => button.addEventListener("click", () => byId(button.dataset.close).close()));
window.addEventListener("beforeunload", () => state.eventController?.abort());

applyLanguage();
bootstrap();
setInterval(() => {
  if (!state.health?.authRequired || state.token) loadDashboard().catch(() => setConnection("offline", t("offline")));
}, 15000);
