const dictionaries = {
  en: {
    brandTagline: "Archive orchestration", connecting: "Connecting", online: "Online", offline: "Offline", locked: "Locked", unlocked: "Unlocked", unlock: "Unlock", newJob: "New job",
    eyebrow: "CONTROL PLANE · SAFE BY DEFAULT", heroTitle: "Publish complete media units, preserve every source.", heroText: "Stage, verify, and publish every directory as one auditable operation.",
    safetyTitle: "Protected launch mode", safetyText: "New jobs start as dry runs. Existing destinations fail closed unless immutable merge is explicitly selected.", sourcePreserved: "SOURCE PRESERVED",
    copyWriteTitle: "Destination writes configured", copyWriteText: "A non-dry-run copy job can add data to its destination, while the source remains preserved.", destinationWrite: "DESTINATION WRITE",
    unsupportedModeTitle: "Unsupported legacy job", unsupportedModeText: "A stored job requests source deletion or file filters. Atomic Sync v0.1 rejects it; edit and save the job to convert it to a complete-unit copy.", unsupportedMode: "ACTION REQUIRED",
    orchestration: "ORCHESTRATION", syncJobs: "Sync jobs", refresh: "Refresh", loadingJobs: "Loading jobs…", auditTrail: "AUDIT TRAIL", recentRuns: "Recent runs", live: "LIVE", noRuns: "No runs recorded yet.",
    footerText: "Deterministic routing · SQLite audit trail · rclone transport", jobConfig: "JOB CONFIGURATION", createJob: "Create a sync job", editJob: "Edit sync job", jobName: "Job name", source: "Source", sourceHint: "Use a read-only absolute path below /sources.",
    destinationName: "Destination ID", destinationPath: "Destination path", mode: "Mode", copyMode: "Verified copy · source always preserved", grouping: "Atomic unit", folderUnit: "Top-level folder", showUnit: "Complete show", seasonUnit: "Season", depthUnit: "Path depth",
    conflictPolicy: "Existing destination", failClosed: "Fail closed (recommended)", immutableMerge: "Immutable merge · never overwrite", verify: "Verification", checksum: "Full content · downloads both sides", sizeOnly: "Size only · metadata", settleSeconds: "Stable window (seconds)", concurrency: "Concurrency", depth: "Depth", schedule: "Schedule",
    dryRun: "Dry run", dryRunHint: "Discover and plan units without touching any file.", paused: "Paused", pausedHint: "Save the job but prevent manual execution.", cancel: "Cancel", saveJob: "Save job",
    unlockTitle: "Unlock the control plane", unlockText: "Enter the API token configured on this server. It remains in this browser tab only.", apiToken: "API token", forgetToken: "Forget token", confirmRun: "Confirm execution", typeJobName: "Type the job name to authorize a write run", runNow: "Run now",
    metricJobs: "Jobs", metricRunning: "Running", metricCompleted: "Completed", metricFailed: "Failed", configured: "configured", activeNow: "active now", allTime: "all time", needsReview: "needs review",
    noJobs: "No jobs yet. Create a dry-run plan to begin safely.", noRunsLong: "Runs will appear here with a durable state and verification result.", run: "Run", edit: "Edit", remove: "Delete", daysStable: "stable", destination: "destination", pausedLabel: "paused", dryLabel: "dry run", copyLabel: "copy", legacyLabel: "unsupported legacy mode",
    group_folder: "folder", group_show: "show", group_season: "season", group_depth: "depth", conflict_fail: "fail closed", conflict_merge_immutable: "immutable merge",
    state_discovered: "discovered", state_staging: "staging", state_verifying: "verifying", state_publishing: "publishing", state_completed: "completed", state_failed: "failed",
    dryRunSummary: "This dry run will only discover and plan units for “{name}”. No files will change.", writeRunSummary: "This run can write verified data to the destination. Source data is never deleted. Review the job and confirm “{name}”.",
    tokenAccepted: "Control plane unlocked.", tokenForgotten: "Token removed from this tab.", jobSaved: "Job saved.", jobDeleted: "Job deleted.", runStarted: "Run started.", invalidConfirmation: "The job name does not match.", requestFailed: "Request failed", deleteConfirm: "Delete “{name}”? Run history will be retained.",
    sourceRoute: "{source} → {destination}", targetCount: "{count} destinations", seconds: "{value}s", minutes: "{value}m", hours: "{value}h", days: "{value}d",
    analyze: "Analyze physical branches", analysisStarted: "Read-only physical-branch analysis started.", branchAnalysis: "PHYSICAL BRANCH ANALYSIS", archiveStatus: "Archive status", analysisNote: "A mergerfs folder name is not archive evidence. This read-only snapshot compares relative file paths and sizes on the mounted physical source and its assigned physical destination. Cleanup still requires final verification.",
    analysis_archived: "archived", analysis_ready_to_verify: "ready to verify", analysis_partial: "partially archived", analysis_pending: "not started", analysis_conflict: "conflict", analysis_empty: "empty", analysis_unknown: "unknown", analysis_running: "analyzing", analysis_failed: "analysis failed",
    summaryArchived: "Archived", summaryReady: "Ready to verify", summaryPartial: "Partial", summaryPending: "Not started", summaryConflict: "Conflicts", summaryEmpty: "Empty",
    analysisContext: "Physical-branch snapshot · {time} · source {source}", sourcePresent: "contains files", sourceShell: "empty directory shell", sourceAbsent: "not present", destinationPresent: "contains files", destinationShell: "empty directory shell", destinationAbsent: "not present",
    physicalSource: "Physical source", physicalDestination: "Destination · {name}", fileCount: "{count} files", coverageLabel: "Source coverage at destination: {coverage}%",
    inventoryMatched: "matched", inventoryMissing: "missing", inventoryConflicts: "conflicts", inventoryDestinationOnly: "destination-only", inventoryUnexpected: "outside assigned branch",
    evidence: "Path evidence ({count})", conflictEvidence: "Conflict", missingEvidence: "Missing from destination", destinationOnlyEvidence: "Destination only",
    statusGuide: "How the states are determined", mergerfsGuide: "The merged view is a union: same-named folders can be empty, complete, or split across branches. These states use physical-branch file contents, not folder names.",
    meaning_archived: "The destination contains files and the physical source contains none. An empty source directory shell may remain; mount health must still be confirmed.",
    meaning_ready_to_verify: "Every source file path exists at the destination with the same size, but source files remain. Run final checksum or size verification before cleanup.",
    meaning_partial: "The destination contains unit content, but one or more source paths are missing there. The two branches may be fully disjoint or partly scattered; do not clean the source.",
    meaning_pending: "The physical source contains files and the assigned destination contains none. Archiving has not started for this unit.",
    meaning_conflict: "A shared path has a different size, one branch has a file where the other has a directory, or content appears outside its assigned destination branch. Manual review is required.",
    meaning_empty: "Neither physical branch contains files for this unit. One or both directory shells may still exist.",
    meaning_unknown: "The analyzer returned an unsupported state. Treat this unit as unsafe and inspect the API response and service logs.",
    analysisPillHint: "Highest-priority state: {status}. Open the physical-branch snapshot for all counts.", noAnalysisUnits: "No media units were found in this snapshot."
  },
  "zh-CN": {
    brandTagline: "媒体归档编排", connecting: "正在连接", online: "在线", offline: "离线", locked: "已锁定", unlocked: "已解锁", unlock: "解锁", newJob: "新建任务",
    eyebrow: "控制平面 · 默认安全", heroTitle: "以完整媒体单元发布，永久保留来源。", heroText: "每个目录都经过暂存、校验、发布与可恢复审计。",
    safetyTitle: "受保护启动模式", safetyText: "新任务默认仅演练；目标已存在时默认失败，只有明确选择不可变合并才会合并。", sourcePreserved: "源数据受保护",
    copyWriteTitle: "已配置目标写入", copyWriteText: "存在非演练复制任务，可以向目标补充数据，但不会删除来源。", destinationWrite: "目标可写",
    unsupportedModeTitle: "存在不支持的旧任务", unsupportedModeText: "已保存任务要求删除来源或使用文件过滤器。Atomic Sync v0.1 会拒绝执行；请编辑并保存为完整目录复制。", unsupportedMode: "需要处理",
    orchestration: "任务编排", syncJobs: "同步任务", refresh: "刷新", loadingJobs: "正在读取任务…", auditTrail: "审计记录", recentRuns: "最近运行", live: "实时", noRuns: "暂无运行记录。",
    footerText: "确定性分流 · SQLite 审计 · rclone 传输", jobConfig: "任务配置", createJob: "创建同步任务", editJob: "编辑同步任务", jobName: "任务名称", source: "来源", sourceHint: "使用 /sources 下的只读绝对路径。",
    destinationName: "目标标识", destinationPath: "目标路径", mode: "模式", copyMode: "已校验复制 · 永久保留来源", grouping: "原子单元", folderUnit: "顶层目录", showUnit: "整部剧", seasonUnit: "单季", depthUnit: "路径层级",
    conflictPolicy: "目标已存在", failClosed: "安全失败（推荐）", immutableMerge: "不可变合并 · 绝不覆盖", verify: "校验方式", checksum: "完整内容 · 下载双方", sizeOnly: "仅大小 · 元数据", settleSeconds: "稳定窗口（秒）", concurrency: "并发数", depth: "层级", schedule: "计划任务",
    dryRun: "仅演练", dryRunHint: "只发现和规划迁移单元，不触碰文件。", paused: "暂停", pausedHint: "保存任务，但禁止手动执行。", cancel: "取消", saveJob: "保存任务",
    unlockTitle: "解锁控制平面", unlockText: "输入服务器配置的 API Token；它只保留在当前浏览器标签页。", apiToken: "API Token", forgetToken: "忘记 Token", confirmRun: "确认执行", typeJobName: "输入任务名称以授权写入运行", runNow: "立即运行",
    metricJobs: "任务", metricRunning: "运行中", metricCompleted: "已完成", metricFailed: "失败", configured: "已配置", activeNow: "当前活动", allTime: "累计", needsReview: "需要检查",
    noJobs: "还没有任务。先创建一个仅演练计划，安全开始。", noRunsLong: "每次运行都会在这里记录持久状态和校验结果。", run: "运行", edit: "编辑", remove: "删除", daysStable: "稳定窗口", destination: "目标", pausedLabel: "已暂停", dryLabel: "仅演练", copyLabel: "复制", legacyLabel: "不支持的旧模式",
    group_folder: "目录", group_show: "整部剧", group_season: "单季", group_depth: "层级", conflict_fail: "安全失败", conflict_merge_immutable: "不可变合并",
    state_discovered: "已发现", state_staging: "暂存中", state_verifying: "校验中", state_publishing: "发布中", state_completed: "已完成", state_failed: "失败",
    dryRunSummary: "“{name}”本次仅发现并规划迁移单元，不会修改任何文件。", writeRunSummary: "本次运行可以向目标写入已校验数据，但绝不删除来源。请核对任务并确认“{name}”。",
    tokenAccepted: "控制平面已解锁。", tokenForgotten: "已从当前标签页移除 Token。", jobSaved: "任务已保存。", jobDeleted: "任务已删除。", runStarted: "任务已开始运行。", invalidConfirmation: "输入的任务名称不匹配。", requestFailed: "请求失败", deleteConfirm: "删除“{name}”？运行历史将保留。",
    sourceRoute: "{source} → {destination}", targetCount: "{count} 个目标", seconds: "{value} 秒", minutes: "{value} 分钟", hours: "{value} 小时", days: "{value} 天",
    analyze: "分析物理分支", analysisStarted: "只读物理分支分析已开始。", branchAnalysis: "物理分支分析", archiveStatus: "归档状态", analysisNote: "mergerfs 合并视图中的同名目录不能证明已归档。本快照只读比较挂载的物理来源与其分配目标上的相对路径和文件大小；清理来源前仍须最终校验。",
    analysis_archived: "已归档", analysis_ready_to_verify: "待强校验/清理", analysis_partial: "部分归档", analysis_pending: "未开始/待归档", analysis_conflict: "存在冲突", analysis_empty: "空目录", analysis_unknown: "未知状态", analysis_running: "分析中", analysis_failed: "分析失败",
    summaryArchived: "已归档", summaryReady: "待强校验", summaryPartial: "部分归档", summaryPending: "未开始", summaryConflict: "冲突", summaryEmpty: "空目录",
    analysisContext: "物理分支快照 · {time} · 来源 {source}", sourcePresent: "包含文件", sourceShell: "仅有空目录壳", sourceAbsent: "不存在", destinationPresent: "包含文件", destinationShell: "仅有空目录壳", destinationAbsent: "不存在",
    physicalSource: "物理来源", physicalDestination: "目标 · {name}", fileCount: "{count} 个文件", coverageLabel: "来源文件在目标的覆盖率：{coverage}%",
    inventoryMatched: "已匹配", inventoryMissing: "目标缺失", inventoryConflicts: "冲突", inventoryDestinationOnly: "仅目标存在", inventoryUnexpected: "位于非分配目标",
    evidence: "路径证据（{count}）", conflictEvidence: "冲突", missingEvidence: "目标缺失", destinationOnlyEvidence: "仅目标存在",
    statusGuide: "状态如何判定", mergerfsGuide: "合并视图是多个分支的并集：同名目录可能为空、完整或文件分散。这些状态依据物理分支内的文件内容，而不是目录名称。",
    meaning_archived: "目标包含文件，而物理来源没有文件；来源可能仍留有空目录壳。仍需先确认两个挂载均健康。",
    meaning_ready_to_verify: "来源的每个文件路径都在目标以相同大小存在，但来源文件仍保留。清理前须执行最终校验和或大小校验。",
    meaning_partial: "目标已有该单元内容，但仍缺少一个或多个来源路径；两个分支可能完全分离，也可能只有部分文件交叉分布，不得清理来源。",
    meaning_pending: "物理来源包含文件，而分配目标没有文件；该单元尚未开始归档。",
    meaning_conflict: "同一路径大小不同、一个分支是文件而另一个是目录，或内容出现在并未分配给它的目标分支；必须人工检查。",
    meaning_empty: "两个物理分支在该单元中都没有文件，但一侧或两侧可能仍有空目录壳。",
    meaning_unknown: "分析器返回了不支持的状态。请将该单元视为不安全，并检查 API 响应和服务日志。",
    analysisPillHint: "最高优先级状态：{status}。打开物理分支快照查看全部数量。", noAnalysisUnits: "本次快照未发现媒体单元。"
  }
};

const state = {
  language: localStorage.getItem("atomic-language") || (navigator.language.toLowerCase().startsWith("zh") ? "zh-CN" : "en"),
  token: sessionStorage.getItem("atomic-token") || "",
  health: null,
  jobs: [],
  runs: [],
  analyses: [],
  analysisView: null,
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
  if (byId("analysisDialog")?.open && state.analysisView) {
    openAnalysis(state.analysisView.job, state.analysisView.analysis);
  }
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
  renderSafetyStatus();
  renderMetrics();
  renderJobs();
  renderRuns();
}

function renderSafetyStatus() {
  const banner = document.querySelector(".safety-banner");
  const unsupported = state.jobs.some(job => job.mode !== "copy" || job.deleteSource || job.include?.length || job.exclude?.length);
  const copyEnabled = state.jobs.some(job => !job.dryRun && job.mode === "copy");
  banner.classList.toggle("write-enabled", copyEnabled && !unsupported);
  banner.classList.toggle("delete-enabled", unsupported);
  const titleKey = unsupported ? "unsupportedModeTitle" : copyEnabled ? "copyWriteTitle" : "safetyTitle";
  const textKey = unsupported ? "unsupportedModeText" : copyEnabled ? "copyWriteText" : "safetyText";
  const pillKey = unsupported ? "unsupportedMode" : copyEnabled ? "destinationWrite" : "sourcePreserved";
  byId("safetyTitle").textContent = t(titleKey);
  byId("safetyText").textContent = t(textKey);
  byId("safetyPill").textContent = t(pillKey);
}

function renderMetrics() {
  const definitions = [
    ["metricJobs", state.dashboard.jobs || 0, "configured", "◇", "jobs"],
    ["metricRunning", state.dashboard.running || 0, "activeNow", "↻", "running"],
    ["metricCompleted", state.dashboard.completed || 0, "allTime", "✓", "completed"],
    ["metricFailed", state.dashboard.failed || 0, "needsReview", "!", "failed"]
  ];
  const container = byId("stats");
  container.replaceChildren(...definitions.map(([label, value, note, icon, tone]) => {
    const card = element("article", `metric metric-${tone}`);
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
    const supported = job.mode === "copy" && !job.deleteSource && !job.include?.length && !job.exclude?.length;
    primary.append(element("div", "job-symbol", supported ? "C" : "!"));
    const content = element("div", "job-content");
    const nameLine = element("div", "job-name-line");
    nameLine.append(element("span", "job-name", job.name));
    if (job.dryRun) nameLine.append(element("span", "pill dry", t("dryLabel")));
    nameLine.append(element("span", `pill ${supported ? "copy" : "move"}`, t(supported ? "copyLabel" : "legacyLabel")));
    if (job.paused) nameLine.append(element("span", "pill paused", t("pausedLabel")));
    const analysis = state.analyses.find(item => item.jobId === job.id);
    if (analysis) {
      const rollup = analysisRollup(analysis);
      const analysisPill = element("button", `pill analysis ${rollup}`, t(`analysis_${rollup.replaceAll("-", "_")}`));
      analysisPill.type = "button";
      analysisPill.title = t("analysisPillHint", { status: t(`analysis_${rollup.replaceAll("-", "_")}`) });
      analysisPill.setAttribute("aria-label", `${job.name}: ${analysisPill.title}`);
      analysisPill.addEventListener("click", () => openAnalysisDetail(job));
      nameLine.append(analysisPill);
    }
    const destinations = job.destinations || [];
    const destination = destinations.length
      ? destinations.map(item => `${item.name}: ${item.path}`).join(" · ")
      : "—";
    const route = element("div", "job-route", t("sourceRoute", { source: job.source, destination }));
    route.title = route.textContent;
    const meta = element("div", "job-meta");
    meta.append(
      element("span", "", `◫ ${t(`group_${job.grouping}`)}`),
      element("span", "", `◷ ${formatDuration(job.settleSeconds || 0)}`),
      element("span", "", `⇶ ${job.concurrency || 1}`),
      element("span", "", `◇ ${t(`conflict_${String(job.conflictPolicy || "fail").replaceAll("-", "_")}`)}`)
    );
    if (destinations.length > 1) meta.append(element("span", "", `⌘ ${t("targetCount", { count: destinations.length })}`));
    content.append(nameLine, route, meta);
    primary.append(content);

    const actions = element("div", "job-actions");
    const runButton = element("button", "mini-button run", `▶ ${t("run")}`);
    runButton.type = "button";
    runButton.disabled = Boolean(job.paused) || !supported;
    runButton.addEventListener("click", () => openRun(job));
    const editButton = element("button", "mini-button", "✎");
    editButton.type = "button"; editButton.title = t("edit"); editButton.setAttribute("aria-label", `${t("edit")}: ${job.name}`);
    editButton.addEventListener("click", () => openJob(job));
    const analyzeButton = element("button", "mini-button", "◎");
    analyzeButton.type = "button"; analyzeButton.title = t("analyze"); analyzeButton.setAttribute("aria-label", `${t("analyze")}: ${job.name}`);
    analyzeButton.disabled = analysis?.state === "running" || !supported;
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
  const statuses = ["conflict", "partial", "pending", "ready-to-verify", "archived", "empty"];
  for (const status of statuses) {
    if ((summary[status] || 0) > 0) return status;
  }
  if (Object.entries(summary).some(([status, count]) => !statuses.includes(status) && Number(count) > 0)) return "unknown";
  return "empty";
}

const analysisStatusOrder = {
  unknown: -1,
  conflict: 0,
  partial: 1,
  pending: 2,
  "ready-to-verify": 3,
  archived: 4,
  empty: 5
};

function analysisStatus(unit) {
  return Object.hasOwn(analysisStatusOrder, unit?.status) ? unit.status : "unknown";
}

function analysisMeaning(status) {
  return t(`meaning_${status.replaceAll("-", "_")}`);
}

function safeCount(value) {
  return Number.isFinite(Number(value)) && Number(value) >= 0 ? Number(value) : 0;
}

function formatBytes(value) {
  const bytes = Number(value);
  if (!Number.isFinite(bytes) || bytes < 0) return "—";
  const units = ["B", "KiB", "MiB", "GiB", "TiB", "PiB"];
  let amount = bytes;
  let index = 0;
  while (amount >= 1024 && index < units.length - 1) {
    amount /= 1024;
    index += 1;
  }
  const maximumFractionDigits = index === 0 || amount >= 100 ? 0 : amount >= 10 ? 1 : 2;
  return `${new Intl.NumberFormat(state.language, { maximumFractionDigits }).format(amount)} ${units[index]}`;
}

function branchPresence(present, files, kind) {
  if (files > 0) return t(kind === "source" ? "sourcePresent" : "destinationPresent");
  if (present) return t(kind === "source" ? "sourceShell" : "destinationShell");
  return t(kind === "source" ? "sourceAbsent" : "destinationAbsent");
}

function branchCard(kind, label, present, files, bytes) {
  const card = element("div", `analysis-branch ${kind}`);
  const heading = element("div", "analysis-branch-heading");
  heading.append(element("span", "analysis-branch-dot"), element("strong", "", label));
  const presence = element("span", "analysis-branch-presence", branchPresence(present, files, kind));
  const totals = element("span", "analysis-branch-totals", `${t("fileCount", { count: files })} · ${formatBytes(bytes)}`);
  card.append(heading, presence, totals);
  card.setAttribute("aria-label", `${label}: ${presence.textContent}; ${totals.textContent}`);
  return card;
}

function openAnalysis(job, analysis) {
  state.analysisView = { job, analysis };
  byId("analysisTitle").textContent = `${job.name} · ${t("archiveStatus")}`;
  const snapshot = analysis.finishedAt || analysis.startedAt;
  const snapshotTime = snapshot ? new Intl.DateTimeFormat(state.language, { dateStyle: "medium", timeStyle: "medium" }).format(new Date(snapshot)) : "—";
  byId("analysisContext").textContent = t("analysisContext", { time: snapshotTime, source: job.source || "—" });
  const summaryDefinitions = [
    ["summaryConflict", "conflict"], ["summaryPartial", "partial"],
    ["summaryPending", "pending"], ["summaryReady", "ready-to-verify"],
    ["summaryArchived", "archived"], ["summaryEmpty", "empty"]
  ];
  const summaryContainer = byId("analysisSummary");
  summaryContainer.setAttribute("role", "list");
  summaryContainer.replaceChildren(...summaryDefinitions.map(([label, key]) => {
    const count = safeCount(analysis.summary?.[key]);
    const card = element("div", `analysis-summary-card ${key}`);
    card.setAttribute("role", "listitem");
    card.setAttribute("aria-label", `${t(label)}: ${count}. ${analysisMeaning(key)}`);
    card.title = analysisMeaning(key);
    card.append(element("span", "", t(label)), element("strong", "", count));
    return card;
  }));
  const unitsContainer = byId("analysisUnits");
  unitsContainer.removeAttribute("role");
  unitsContainer.setAttribute("aria-busy", analysis.state === "running" ? "true" : "false");
  if (analysis.state === "running") {
    const loading = element("div", "empty-state compact-state");
    loading.append(element("span", "loader"), element("p", "", t("analysis_running")));
    unitsContainer.replaceChildren(loading);
  } else if (analysis.state === "failed") {
    const failed = element("div", "empty-state compact-state");
    failed.append(element("p", "", `${t("analysis_failed")}: ${analysis.message || ""}`));
    unitsContainer.replaceChildren(failed);
  } else {
    const units = [...(analysis.units || [])].sort((left, right) => {
      const severity = analysisStatusOrder[analysisStatus(left)] - analysisStatusOrder[analysisStatus(right)];
      return severity || String(left.unit || "").localeCompare(String(right.unit || ""), state.language);
    });
    if (!units.length) {
      const empty = element("div", "empty-state compact-state");
      empty.append(element("p", "", t("noAnalysisUnits")));
      unitsContainer.replaceChildren(empty);
      openDialog("analysisDialog");
      return;
    }
    unitsContainer.setAttribute("role", "list");
    unitsContainer.replaceChildren(...units.map(unit => {
      const statusKey = analysisStatus(unit);
      const statusLabel = t(`analysis_${statusKey.replaceAll("-", "_")}`);
      const sourceFiles = safeCount(unit.sourceFiles);
      const destinationFiles = safeCount(unit.destinationFiles);
      const matchingFiles = safeCount(unit.matchingFiles);
      const missingFiles = safeCount(unit.missingFiles);
      const conflictingFiles = safeCount(unit.conflictingFiles);
      const unexpectedDestinationFiles = safeCount(unit.unexpectedDestinationFiles);
      const unexpectedDestinations = Array.isArray(unit.unexpectedDestinations) ? unit.unexpectedDestinations.map(String) : [];
      const coverage = Math.min(100, safeCount(unit.coverage));
      const destinationOnlySamples = unit.destinationOnlySamples || [];
      const destinationOnlyFiles = Number.isFinite(Number(unit.destinationOnlyFiles)) && Number(unit.destinationOnlyFiles) >= 0
        ? Number(unit.destinationOnlyFiles)
        : null;
      const row = element("article", "analysis-unit");
      row.setAttribute("role", "listitem");
      row.setAttribute("aria-label", `${unit.unit}: ${statusLabel}`);
      const content = element("div", "analysis-unit-content");
      const heading = element("div", "analysis-unit-heading");
      const name = element("div", "analysis-unit-name", unit.unit);
      name.title = unit.unit;
      const status = element("span", `analysis-status ${statusKey}`, statusLabel);
      heading.append(name, status);
      const meaning = element("p", "analysis-unit-meaning", analysisMeaning(statusKey));
      const branches = element("div", "analysis-branches");
      const sourceBranch = branchCard("source", t("physicalSource"), Boolean(unit.sourcePresent) || sourceFiles > 0, sourceFiles, unit.sourceBytes);
      sourceBranch.title = job.source || "";
      const destinationBranch = branchCard("destination", t("physicalDestination", { name: unit.destination || "—" }), Boolean(unit.destinationPresent) || destinationFiles > 0, destinationFiles, unit.destinationBytes);
      const destinationConfig = (job.destinations || []).find(item => item.name === unit.destination);
      destinationBranch.title = destinationConfig?.path || unit.destination || "";
      branches.append(sourceBranch, destinationBranch);
      const inventory = element("dl", "analysis-inventory");
      const inventoryItems = [
        ["matching", t("inventoryMatched"), matchingFiles],
        ["missing", t("inventoryMissing"), missingFiles],
        ["conflicts", t("inventoryConflicts"), conflictingFiles]
      ];
      if (destinationOnlyFiles !== null || destinationOnlySamples.length) {
        inventoryItems.push(["destination-only", t("inventoryDestinationOnly"), destinationOnlyFiles ?? `≥${destinationOnlySamples.length}`]);
      }
      if (unexpectedDestinationFiles > 0) {
        inventoryItems.push(["unexpected", t("inventoryUnexpected"), unexpectedDestinationFiles, unexpectedDestinations.join(", ")]);
      }
      for (const [kind, label, value, hint] of inventoryItems) {
        const item = element("div", `analysis-inventory-item ${kind}`);
        if (hint) item.title = hint;
        item.append(element("dt", "", label), element("dd", "", value));
        inventory.append(item);
      }
      let coverageBlock = null;
      if (sourceFiles > 0) {
        coverageBlock = element("div", "analysis-coverage");
        const coverageLabel = element("span", "", t("coverageLabel", { coverage }));
        const progress = element("progress", "");
        progress.max = 100;
        progress.value = coverage;
        progress.setAttribute("aria-label", coverageLabel.textContent);
        coverageBlock.append(coverageLabel, progress);
      }
      const evidence = [
        ...(unit.conflictSamples || []).map(value => ({ kind: "conflict", label: t("conflictEvidence"), value })),
        ...(unit.missingSamples || []).map(value => ({ kind: "missing", label: t("missingEvidence"), value })),
        ...destinationOnlySamples.map(value => ({ kind: "destination-only", label: t("destinationOnlyEvidence"), value }))
      ];
      content.append(heading, meaning, branches, inventory);
      if (coverageBlock) content.append(coverageBlock);
      if (evidence.length) {
        const details = element("details", "analysis-evidence");
        details.append(element("summary", "", t("evidence", { count: evidence.length })));
        const list = element("ul", "");
        for (const evidenceItem of evidence) {
          const item = element("li", evidenceItem.kind);
          item.append(element("strong", "", evidenceItem.label), element("code", "", evidenceItem.value));
          list.append(item);
        }
        details.append(list);
        content.append(details);
      }
      row.append(content);
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
    if (run.message) content.append(element("div", "run-message", run.message));
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
    form.elements.mode.value = "copy";
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
    mode: "copy", grouping: form.elements.grouping.value,
    depth: Number(form.elements.depth.value || 1), settleSeconds: Number(form.elements.settleSeconds.value || 0),
    concurrency: Number(form.elements.concurrency.value || 1), verify: form.elements.verify.value,
    conflictPolicy: form.elements.conflictPolicy.value, dryRun: form.elements.dryRun.checked,
    paused: form.elements.paused.checked, deleteSource: false,
    schedule: current?.schedule || "", include: [], exclude: []
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
  const form = event.currentTarget;
  const submitButton = form.querySelector('button[type="submit"]');
  const token = form.elements.token.value.trim();
  let authenticated = false;
  if (submitButton) {
    submitButton.disabled = true;
    submitButton.setAttribute("aria-busy", "true");
    submitButton.textContent = t("connecting");
  }
  state.token = token;
  sessionStorage.setItem("atomic-token", token);
  try {
    await request("system");
    authenticated = true;
    byId("authDialog").close();
    form.reset();
    toast(t("tokenAccepted"), "success");
    await loadDashboard();
    connectEvents();
  } catch (error) {
    if (!authenticated) {
      state.token = "";
      sessionStorage.removeItem("atomic-token");
      form.elements.token.focus();
    }
    handleError(error);
  } finally {
    if (submitButton) {
      submitButton.disabled = false;
      submitButton.removeAttribute("aria-busy");
      submitButton.textContent = t("unlock");
    }
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
