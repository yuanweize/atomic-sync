const dictionaries = {
  en: {
    brandTagline: "Archive orchestration", connecting: "Connecting", online: "Online", offline: "Offline", locked: "Locked", unlocked: "Unlocked", unlock: "Unlock", newJob: "New job", skipToDashboard: "Skip to dashboard", switchLanguage: "Switch language", safetyStatus: "Safety status", dashboardMetrics: "Dashboard metrics", close: "Close", jobNamePlaceholder: "Archive stable movies", schedulePlaceholder: "Reserved for a future release",
    eyebrow: "CONTROL PLANE · SAFE BY DEFAULT", heroTitle: "Plan and transfer media as complete directory units.", heroText: "Every eligible directory runs through rclone as one auditable operation, with partial outcomes kept visible.",
    safetyTitle: "Protected launch mode", safetyText: "New jobs start as dry runs. Existing destinations fail closed unless immutable merge is explicitly selected.", sourcePreserved: "SOURCE PRESERVED",
    copyWriteTitle: "Destination writes configured", copyWriteText: "A non-dry-run copy job can add data to its destination, while the source remains preserved.", destinationWrite: "DESTINATION WRITE",
    moveWriteTitle: "Source removal enabled", moveWriteText: "A non-dry-run move job invokes rclone move. Missing destination objects are moved; existing-path source files stay for review and make the unit partial.", sourceRemoval: "SOURCE REMOVAL",
    unsupportedModeTitle: "Unsupported legacy job", unsupportedModeText: "A stored job has an invalid copy/move pairing or file filters. Edit and save it before execution.", unsupportedMode: "ACTION REQUIRED",
    orchestration: "ORCHESTRATION", syncJobs: "Transfer jobs", refresh: "Refresh", refreshing: "Refreshing…", loadingJobs: "Loading jobs…", auditTrail: "AUDIT TRAIL", recentRuns: "Recent runs", live: "LIVE", reconnecting: "RECONNECTING", noRuns: "No runs recorded yet.",
    footerText: "Deterministic routing · SQLite audit trail · rclone transport", jobConfig: "JOB CONFIGURATION", createJob: "Create a transfer job", editJob: "Edit transfer job", jobName: "Job name", source: "Source", sourceHint: "Use an absolute path below /sources. Move mode requires an explicitly writable source mount.",
    destinationName: "Destination ID", destinationPath: "Destination path", mode: "Mode", copyMode: "Copy · preserve source", moveMode: "Direct rclone move · preserve overlaps", grouping: "Atomic unit", folderUnit: "Top-level folder", showUnit: "Complete show", seasonUnit: "Season", depthUnit: "Path depth",
    conflictPolicy: "Existing destination", failClosed: "Fail closed (recommended)", immutableMerge: "Immutable merge · preserve overlaps", verify: "Verification", checksum: "Native hash when available", sizeOnly: "Size only · metadata", settleSeconds: "Stable window (seconds)", concurrency: "Concurrency", depth: "Depth", schedule: "Schedule",
    dryRun: "Dry run", dryRunHint: "Run the real rclone plan without changing source or destination media objects. OAuth tokens may refresh.", paused: "Paused", pausedHint: "Save the job but prevent manual execution.", cancel: "Cancel", saveJob: "Save job", saving: "Saving…", running: "Running…", analyzing: "Analyzing…", deleting: "Deleting…",
    unlockTitle: "Unlock the control plane", unlockText: "Enter the API token configured on this server. It remains in this browser tab only.", apiToken: "API token", forgetToken: "Forget token", confirmRun: "Confirm execution", confirmMove: "Confirm source-removing move", typeJobName: "Type the job name to authorize a write run", typeMoveJobName: "Type the exact job name to authorize source removal", runNow: "Run now", startMove: "Start move",
    metricJobs: "Jobs", metricRunning: "Running", metricCompleted: "Completed", metricFailed: "Failed", configured: "configured", activeNow: "active now", allTime: "all time", needsReview: "needs review",
    noJobs: "No jobs yet. Create a dry-run plan to begin safely.", noRunsLong: "Runs will appear here with a durable state and verification result.", run: "Run", edit: "Edit", remove: "Delete", daysStable: "stable", destination: "destination", pausedLabel: "paused", dryLabel: "dry run", copyLabel: "copy", moveLabel: "move", legacyLabel: "unsupported legacy mode",
    group_folder: "folder", group_show: "show", group_season: "season", group_depth: "depth", conflict_fail: "fail closed", conflict_merge_immutable: "immutable merge",
    state_discovered: "discovered", state_transferring: "transferring", state_staging: "staging (legacy)", state_verifying: "verifying", state_publishing: "publishing (legacy)", state_completed: "completed", state_failed: "failed",
    dryRunSummary: "This dry run invokes the real rclone plan for “{name}” with --dry-run. No source or destination media objects change; rclone may refresh OAuth tokens.", writeRunSummary: "This run can write data to the destination while preserving the source. Review “{name}” before running.", moveRunSummary: "This run invokes rclone move. Missing destination objects are moved and removed from source; existing destination paths remain at source and make the unit partial. Confirm the exact job name “{name}”.",
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
    meaning_ready_to_verify: "Every source path exists at the destination with the same size, but source files remain. Require independent content proof before cleanup; size-only matching is not deletion authorization.",
    meaning_partial: "The destination contains unit content, but one or more source paths are missing there. The two branches may be fully disjoint or partly scattered; do not clean the source.",
    meaning_pending: "The physical source contains files and the assigned destination contains none. Archiving has not started for this unit.",
    meaning_conflict: "A shared path has a different size, one branch has a file where the other has a directory, or content appears outside its assigned destination branch. Manual review is required.",
    meaning_empty: "Neither physical branch contains files for this unit. One or both directory shells may still exist.",
    meaning_unknown: "The analyzer returned an unsupported state. Treat this unit as unsafe and inspect the API response and service logs.",
    analysisPillHint: "Highest-priority state: {status}. Open the physical-branch snapshot for all counts.", noAnalysisUnits: "No media units were found in this snapshot."
  },
  "zh-CN": {
    brandTagline: "媒体归档编排", connecting: "正在连接", online: "在线", offline: "离线", locked: "已锁定", unlocked: "已解锁", unlock: "解锁", newJob: "新建任务", skipToDashboard: "跳转到仪表盘", switchLanguage: "切换语言", safetyStatus: "安全状态", dashboardMetrics: "仪表盘指标", close: "关闭", jobNamePlaceholder: "归档已稳定的电影", schedulePlaceholder: "预留于未来版本",
    eyebrow: "控制平面 · 默认安全", heroTitle: "以完整目录为单元规划和传输媒体。", heroText: "每个合格目录都通过 rclone 执行一次可审计操作，部分完成也会明确呈现。",
    safetyTitle: "受保护启动模式", safetyText: "新任务默认仅演练；目标已存在时默认失败，只有明确选择不可变合并才会合并。", sourcePreserved: "源数据受保护",
    copyWriteTitle: "已配置目标写入", copyWriteText: "存在非演练复制任务，可以向目标补充数据，但不会删除来源。", destinationWrite: "目标可写",
    moveWriteTitle: "已启用来源移除", moveWriteText: "非演练移动会直接调用 rclone move；目标缺失对象会被移动，同路径已存在时来源保留并将单元标记为部分完成。", sourceRemoval: "来源将移除",
    unsupportedModeTitle: "存在不支持的旧任务", unsupportedModeText: "已保存任务的复制/移动配置不一致或包含文件过滤器；执行前必须编辑并重新保存。", unsupportedMode: "需要处理",
    orchestration: "任务编排", syncJobs: "传输任务", refresh: "刷新", refreshing: "刷新中…", loadingJobs: "正在读取任务…", auditTrail: "审计记录", recentRuns: "最近运行", live: "实时", reconnecting: "重连中", noRuns: "暂无运行记录。",
    footerText: "确定性分流 · SQLite 审计 · rclone 传输", jobConfig: "任务配置", createJob: "创建传输任务", editJob: "编辑传输任务", jobName: "任务名称", source: "来源", sourceHint: "使用 /sources 下的绝对路径；移动模式必须显式启用来源可写挂载。",
    destinationName: "目标标识", destinationPath: "目标路径", mode: "模式", copyMode: "复制 · 保留来源", moveMode: "直接 rclone 移动 · 保留同路径来源", grouping: "原子单元", folderUnit: "顶层目录", showUnit: "整部剧", seasonUnit: "单季", depthUnit: "路径层级",
    conflictPolicy: "目标已存在", failClosed: "安全失败（推荐）", immutableMerge: "不可变合并 · 保留同路径", verify: "校验方式", checksum: "原生哈希（可用时）", sizeOnly: "仅大小 · 元数据", settleSeconds: "稳定窗口（秒）", concurrency: "并发数", depth: "层级", schedule: "计划任务",
    dryRun: "仅演练", dryRunHint: "使用真实 rclone 参数，不修改来源或目标媒体对象；OAuth Token 可能刷新。", paused: "暂停", pausedHint: "保存任务，但禁止手动执行。", cancel: "取消", saveJob: "保存任务", saving: "保存中…", running: "运行中…", analyzing: "分析中…", deleting: "删除中…",
    unlockTitle: "解锁控制平面", unlockText: "输入服务器配置的 API Token；它只保留在当前浏览器标签页。", apiToken: "API Token", forgetToken: "忘记 Token", confirmRun: "确认执行", confirmMove: "确认会移除来源的移动", typeJobName: "输入任务名称以授权写入运行", typeMoveJobName: "输入完整任务名以授权移除来源", runNow: "立即运行", startMove: "开始移动",
    metricJobs: "任务", metricRunning: "运行中", metricCompleted: "已完成", metricFailed: "失败", configured: "已配置", activeNow: "当前活动", allTime: "累计", needsReview: "需要检查",
    noJobs: "还没有任务。先创建一个仅演练计划，安全开始。", noRunsLong: "每次运行都会在这里记录持久状态和校验结果。", run: "运行", edit: "编辑", remove: "删除", daysStable: "稳定窗口", destination: "目标", pausedLabel: "已暂停", dryLabel: "仅演练", copyLabel: "复制", moveLabel: "移动", legacyLabel: "不支持的旧模式",
    group_folder: "目录", group_show: "整部剧", group_season: "单季", group_depth: "层级", conflict_fail: "安全失败", conflict_merge_immutable: "不可变合并",
    state_discovered: "已发现", state_transferring: "传输中", state_staging: "暂存中（旧版）", state_verifying: "校验中", state_publishing: "发布中（旧版）", state_completed: "已完成", state_failed: "失败",
    dryRunSummary: "“{name}”会使用真实 rclone 参数追加 --dry-run；不修改来源或目标媒体对象，rclone 可能刷新 OAuth Token。", writeRunSummary: "本次运行可以向目标写入数据，并保留来源。请在运行前核对“{name}”。", moveRunSummary: "本次会直接调用 rclone move：目标缺失对象移动后从来源移除，同路径已存在则来源保留且单元标记为部分完成。请输入完整任务名“{name}”确认。",
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
    meaning_ready_to_verify: "来源路径都在目标以相同大小存在，但来源仍保留。清理前必须有独立内容证明；仅大小匹配不是删除授权。",
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
  analysisPollTimer: null,
  analysisPollController: null,
  selectedJob: null,
  connection: { kind: "connecting", labelKey: "connecting" },
  authLabelKey: "unlock",
  eventsConnected: false,
  dialogStack: [],
  busy: { save: false, run: false, refresh: false, auth: false },
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
  document.querySelectorAll("[data-i18n-aria-label]").forEach(element => { element.setAttribute("aria-label", t(element.dataset.i18nAriaLabel)); });
  document.querySelectorAll("[data-i18n-placeholder]").forEach(element => { element.setAttribute("placeholder", t(element.dataset.i18nPlaceholder)); });
  const languageButton = byId("languageButton");
  languageButton.textContent = state.language === "en" ? "中" : "EN";
  languageButton.setAttribute("aria-label", `${languageButton.textContent} · ${t("switchLanguage")}`);
  if (state.busy.save) byId("saveJobButton").textContent = t("saving");
  if (state.busy.run) byId("runSubmitButton").textContent = t("running");
  if (state.busy.refresh) byId("refreshButton").querySelector("[data-i18n]").textContent = t("refreshing");
  if (state.busy.auth) byId("authForm").querySelector('button[type="submit"]').textContent = t("connecting");
  setConnection(state.connection.kind, state.connection.labelKey);
  setAuthButton(state.authLabelKey);
  setLiveStatus(state.eventsConnected);
  renderDashboard();
  if (byId("jobDialog")?.open) {
    const title = byId("jobDialogTitle");
    title.textContent = t(title.dataset.i18n || "createJob");
  }
  if (byId("runDialog")?.open && state.selectedJob) updateRunDialogText(state.selectedJob);
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

function setConnection(kind, labelKey) {
  state.connection = { kind, labelKey };
  const node = byId("connectionStatus");
  node.className = `connection ${kind}`;
  const label = node.querySelector("span");
  label.dataset.i18n = labelKey;
  label.textContent = t(labelKey);
}

function setAuthButton(labelKey) {
  state.authLabelKey = labelKey;
  const button = byId("authButton");
  button.setAttribute("aria-label", t(labelKey));
  const label = button.querySelector(".auth-label");
  label.dataset.i18n = labelKey;
  label.textContent = t(labelKey);
}

function setLiveStatus(connected) {
  state.eventsConnected = connected;
  const status = byId("liveStatus");
  const labelKey = connected ? "live" : "reconnecting";
  status.classList.toggle("reconnecting", !connected);
  const label = status.querySelector("span");
  label.dataset.i18n = labelKey;
  label.textContent = t(labelKey);
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
      state.eventController?.abort();
      setLiveStatus(false);
      setConnection("offline", "locked");
      setAuthButton("unlock");
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
      setConnection("offline", "locked");
      setAuthButton("unlock");
      setLiveStatus(false);
      renderLocked();
      openDialog("authDialog");
      return;
    }
    await loadDashboard();
    connectEvents();
  } catch (error) {
    setConnection("offline", "offline");
    setLiveStatus(false);
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
  setConnection("online", "online");
  setAuthButton(state.health?.authRequired ? "unlocked" : "online");
  renderDashboard();
}

function scheduleRefresh() {
  clearTimeout(state.refreshTimer);
  state.refreshTimer = setTimeout(() => loadDashboard().catch(handleError), 280);
}

async function refreshDashboard() {
  if (state.busy.refresh) return;
  state.busy.refresh = true;
  const button = byId("refreshButton");
  const label = button.querySelector("[data-i18n]");
  button.disabled = true;
  button.setAttribute("aria-busy", "true");
  label.textContent = t("refreshing");
  try {
    await loadDashboard();
  } catch (error) {
    handleError(error);
  } finally {
    state.busy.refresh = false;
    button.disabled = false;
    button.removeAttribute("aria-busy");
    label.textContent = t("refresh");
  }
}

async function connectEvents() {
  state.eventController?.abort();
  setLiveStatus(false);
  const controller = new AbortController();
  state.eventController = controller;
  while (!controller.signal.aborted) {
    try {
      await consumeEventStream(controller);
    } catch (error) {
      if (controller.signal.aborted) return;
      console.warn("event stream disconnected", error);
    }
    setLiveStatus(false);
    await new Promise(resolve => setTimeout(resolve, 2500));
  }
}

async function consumeEventStream(controller) {
  const headers = new Headers({ Accept: "text/event-stream" });
  if (state.token) headers.set("Authorization", `Bearer ${state.token}`);
  const response = await fetch("/api/events", { headers, cache: "no-store", signal: controller.signal });
  if (!response.ok || !response.body) throw new Error(`event stream ${response.status}`);
  setLiveStatus(true);
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
  const unsupported = state.jobs.some(job => !supportedJob(job));
  const moveEnabled = state.jobs.some(job => !job.dryRun && job.mode === "move" && job.deleteSource);
  const copyEnabled = state.jobs.some(job => !job.dryRun && job.mode === "copy");
  banner.classList.toggle("write-enabled", copyEnabled && !unsupported);
  banner.classList.toggle("delete-enabled", unsupported || moveEnabled);
  const titleKey = unsupported ? "unsupportedModeTitle" : moveEnabled ? "moveWriteTitle" : copyEnabled ? "copyWriteTitle" : "safetyTitle";
  const textKey = unsupported ? "unsupportedModeText" : moveEnabled ? "moveWriteText" : copyEnabled ? "copyWriteText" : "safetyText";
  const pillKey = unsupported ? "unsupportedMode" : moveEnabled ? "sourceRemoval" : copyEnabled ? "destinationWrite" : "sourcePreserved";
  byId("safetyTitle").textContent = t(titleKey);
  byId("safetyText").textContent = t(textKey);
  byId("safetyPill").textContent = t(pillKey);
}

function supportedJob(job) {
  if (job.include?.length || job.exclude?.length) return false;
  if (job.mode === "copy") return !job.deleteSource;
  if (job.mode === "move") return Boolean(job.deleteSource);
  return false;
}

function renderMetrics() {
  const definitions = [
    ["metricJobs", state.dashboard.jobs || 0, "configured", "◇", "jobs"],
    ["metricRunning", state.dashboard.running || 0, "activeNow", "↻", "running"],
    ["metricCompleted", state.dashboard.completed || 0, "allTime", "✓", "completed"],
    ["metricFailed", state.dashboard.failed || 0, "needsReview", "!", "failed"]
  ];
  const container = byId("stats");
  container.removeAttribute("aria-busy");
  container.replaceChildren(...definitions.map(([label, value, note, icon, tone]) => {
    const card = element("article", `metric metric-${tone}`);
    const head = element("div", "metric-head");
    const metricIcon = element("span", "metric-icon", icon);
    metricIcon.setAttribute("aria-hidden", "true");
    head.append(element("span", "", t(label)), metricIcon);
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
    const supported = supportedJob(job);
    const jobSymbol = element("div", "job-symbol", supported ? (job.mode === "move" ? "M" : "C") : "!");
    jobSymbol.setAttribute("aria-hidden", "true");
    primary.append(jobSymbol);
    const content = element("div", "job-content");
    const nameLine = element("div", "job-name-line");
    nameLine.append(element("span", "job-name", job.name));
    if (job.dryRun) nameLine.append(element("span", "pill dry", t("dryLabel")));
    const modeKey = job.mode === "move" ? "moveLabel" : "copyLabel";
    nameLine.append(element("span", `pill ${supported ? job.mode : "move"}`, t(supported ? modeKey : "legacyLabel")));
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
    runButton.setAttribute("aria-label", `${t("run")}: ${job.name}`);
    runButton.disabled = Boolean(job.paused) || !supported;
    runButton.addEventListener("click", () => openRun(job));
    const editButton = element("button", "mini-button", "✎");
    editButton.type = "button"; editButton.title = t("edit"); editButton.setAttribute("aria-label", `${t("edit")}: ${job.name}`);
    editButton.addEventListener("click", () => openJob(job));
    const analyzeButton = element("button", "mini-button", "◎");
    analyzeButton.type = "button"; analyzeButton.title = t("analyze"); analyzeButton.setAttribute("aria-label", `${t("analyze")}: ${job.name}`);
    analyzeButton.disabled = analysis?.state === "running" || !supported;
    analyzeButton.addEventListener("click", () => startAnalysis(job, analyzeButton));
    const deleteButton = element("button", "mini-button", "⌫");
    deleteButton.type = "button"; deleteButton.title = t("remove"); deleteButton.setAttribute("aria-label", `${t("remove")}: ${job.name}`);
    deleteButton.addEventListener("click", () => deleteJob(job, deleteButton));
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

function stopAnalysisPolling() {
  clearTimeout(state.analysisPollTimer);
  state.analysisPollTimer = null;
  state.analysisPollController?.abort();
  state.analysisPollController = null;
}

function scheduleAnalysisPolling(job, delay = 2500) {
  clearTimeout(state.analysisPollTimer);
  state.analysisPollTimer = null;
  if (!byId("analysisDialog").open || state.analysisView?.job?.id !== job.id || state.analysisView?.analysis?.state !== "running") return;
  state.analysisPollTimer = setTimeout(() => pollOpenAnalysis(job), delay);
}

async function pollOpenAnalysis(job) {
  state.analysisPollTimer = null;
  if (!byId("analysisDialog").open || state.analysisView?.job?.id !== job.id || state.analysisView?.analysis?.state !== "running") return;
  const controller = new AbortController();
  state.analysisPollController = controller;
  try {
    const analysis = await request(`jobs/${encodeURIComponent(job.id)}/analysis`, { signal: controller.signal });
    if (controller.signal.aborted || !byId("analysisDialog").open || state.analysisView?.job?.id !== job.id) return;
    state.analysisPollController = null;
    openAnalysis(job, analysis);
  } catch (error) {
    if (controller.signal.aborted || error.name === "AbortError") return;
    console.warn("analysis refresh failed", error);
    if (error.status !== 401) scheduleAnalysisPolling(job, 5000);
  } finally {
    if (state.analysisPollController === controller) state.analysisPollController = null;
  }
}

function openAnalysis(job, analysis) {
  stopAnalysisPolling();
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
  if (analysis.state === "running") scheduleAnalysisPolling(job);
}

async function openAnalysisDetail(job) {
  try {
    const analysis = await request(`jobs/${encodeURIComponent(job.id)}/analysis`);
    openAnalysis(job, analysis);
  } catch (error) { handleError(error); }
}

async function startAnalysis(job, button) {
  if (job._analysisBusy) return;
  job._analysisBusy = true;
  if (button) {
    button.disabled = true;
    button.setAttribute("aria-busy", "true");
    button.setAttribute("aria-label", `${t("analyzing")}: ${job.name}`);
    button.textContent = "…";
  }
  try {
    await request(`jobs/${encodeURIComponent(job.id)}/analysis`, { method: "POST" });
    toast(t("analysisStarted"), "success");
    await loadDashboard();
    await openAnalysisDetail(job);
  } catch (error) { handleError(error); }
  finally {
    job._analysisBusy = false;
    if (button) {
      button.disabled = false;
      button.removeAttribute("aria-busy");
      button.setAttribute("aria-label", `${t("analyze")}: ${job.name}`);
      button.textContent = "◎";
    }
  }
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
    const runDot = element("span", "run-dot");
    runDot.setAttribute("aria-hidden", "true");
    item.append(runDot);
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
  state.dashboard = { jobs: 0, running: 0, completed: 0, failed: 0 };
  renderMetrics();
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
  if (!dialog.open) {
    dialog.showModal();
    state.dialogStack = state.dialogStack.filter(openID => openID !== id);
    state.dialogStack.push(id);
    const preferred = id === "jobDialog"
      ? byId("jobForm").elements.name
      : id === "authDialog"
        ? byId("authForm").elements.token
        : id === "runDialog" && !byId("dangerConfirm").hidden
          ? byId("runForm").elements.confirmation
          : dialog.querySelector("button:not(:disabled), input:not(:disabled), select:not(:disabled)");
    preferred?.focus();
  }
}

function setDialogDismissDisabled(id, disabled) {
  byId(id).querySelectorAll("[data-close]").forEach(button => { button.disabled = disabled; });
}

function openJob(job = null) {
  const form = byId("jobForm");
  form.reset();
  state.selectedJob = job;
  const title = byId("jobDialogTitle");
  title.dataset.i18n = job ? "editJob" : "createJob";
  title.textContent = t(title.dataset.i18n);
  if (job) {
    form.elements.id.value = job.id;
    form.elements.name.value = job.name || "";
    form.elements.source.value = job.source || "";
    form.elements.destinationName.value = job.destinations?.[0]?.name || "gd-primary";
    form.elements.destinationPath.value = job.destinations?.[0]?.path || "";
    form.elements.mode.value = job.mode === "move" ? "move" : "copy";
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
  if (state.busy.save) return;
  state.busy.save = true;
  const submitButton = form.querySelector('button[type="submit"]');
  if (submitButton) {
    submitButton.disabled = true;
    submitButton.setAttribute("aria-busy", "true");
    submitButton.textContent = t("saving");
  }
  setDialogDismissDisabled("jobDialog", true);
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
    paused: form.elements.paused.checked, deleteSource: form.elements.mode.value === "move",
    schedule: current?.schedule || "", include: [], exclude: []
  };
  const id = form.elements.id.value;
  try {
    await request(id ? `jobs/${encodeURIComponent(id)}` : "jobs", { method: id ? "PUT" : "POST", body: JSON.stringify(payload) });
    byId("jobDialog").close();
    toast(t("jobSaved"), "success");
    await loadDashboard();
  } catch (error) { handleError(error); }
  finally {
    state.busy.save = false;
    if (submitButton) {
      submitButton.disabled = false;
      submitButton.removeAttribute("aria-busy");
      submitButton.textContent = t("saveJob");
    }
    setDialogDismissDisabled("jobDialog", false);
  }
}

async function deleteJob(job, button) {
  if (job._deleteBusy) return;
  if (!window.confirm(t("deleteConfirm", { name: job.name }))) return;
  job._deleteBusy = true;
  if (button) {
    button.disabled = true;
    button.setAttribute("aria-busy", "true");
    button.setAttribute("aria-label", `${t("deleting")}: ${job.name}`);
    button.textContent = "…";
  }
  try {
    await request(`jobs/${encodeURIComponent(job.id)}`, { method: "DELETE" });
    toast(t("jobDeleted"), "success");
    await loadDashboard();
  } catch (error) { handleError(error); }
  finally {
    job._deleteBusy = false;
    if (button) {
      button.disabled = false;
      button.removeAttribute("aria-busy");
      button.setAttribute("aria-label", `${t("remove")}: ${job.name}`);
      button.textContent = "⌫";
    }
  }
}

function openRun(job) {
  state.selectedJob = job;
  const destructive = !job.dryRun && job.mode === "move";
  byId("dangerConfirm").hidden = !destructive;
  const confirmation = byId("runForm").elements.confirmation;
  const confirmationError = byId("confirmationError");
  confirmation.value = "";
  confirmation.required = destructive;
  confirmation.removeAttribute("aria-invalid");
  confirmationError.hidden = true;
  updateRunDialogText(job);
  updateRunSubmitState();
  openDialog("runDialog");
}

function updateRunDialogText(job) {
  const destructiveMove = !job.dryRun && job.mode === "move";
  const summaryKey = job.dryRun ? "dryRunSummary" : destructiveMove ? "moveRunSummary" : "writeRunSummary";
  byId("runSummary").textContent = t(summaryKey, { name: job.name });
  byId("runDialog").classList.toggle("destructive-move", destructiveMove);
  byId("runDialogTitle").dataset.i18n = destructiveMove ? "confirmMove" : "confirmRun";
  byId("runDialogTitle").textContent = t(byId("runDialogTitle").dataset.i18n);
  byId("confirmationLabel").dataset.i18n = destructiveMove ? "typeMoveJobName" : "typeJobName";
  byId("confirmationLabel").textContent = t(byId("confirmationLabel").dataset.i18n);
  const submitButton = byId("runSubmitButton");
  submitButton.dataset.i18n = destructiveMove ? "startMove" : "runNow";
  if (!state.busy.run) submitButton.textContent = t(submitButton.dataset.i18n);
}

function updateRunSubmitState() {
  const job = state.selectedJob;
  const confirmation = byId("runForm").elements.confirmation;
  const unmatched = Boolean(job && !job.dryRun && job.mode === "move" && confirmation.value !== job.name);
  byId("runSubmitButton").disabled = state.busy.run || unmatched;
}

function clearConfirmationError() {
  const confirmation = byId("runForm").elements.confirmation;
  confirmation.removeAttribute("aria-invalid");
  byId("confirmationError").hidden = true;
  updateRunSubmitState();
}

async function startRun(event) {
  event.preventDefault();
  if (state.busy.run) return;
  const job = state.selectedJob;
  if (!job) return;
  const confirmation = event.currentTarget.elements.confirmation;
  const confirmationError = byId("confirmationError");
  if (!job.dryRun && job.mode === "move" && confirmation.value !== job.name) {
    confirmation.setAttribute("aria-invalid", "true");
    confirmationError.hidden = false;
    confirmation.focus();
    toast(t("invalidConfirmation"), "error");
    return;
  }
  state.busy.run = true;
  const submitButton = byId("runSubmitButton");
  if (submitButton) {
    updateRunSubmitState();
    submitButton.setAttribute("aria-busy", "true");
    submitButton.textContent = t("running");
  }
  setDialogDismissDisabled("runDialog", true);
  try {
    const headers = new Headers();
    if (!job.dryRun && job.mode === "move") headers.set("X-Atomic-Confirm-Job", confirmation.value);
    await request(`jobs/${encodeURIComponent(job.id)}/run`, { method: "POST", headers });
    byId("runDialog").close();
    toast(t("runStarted"), "success");
    scheduleRefresh();
  } catch (error) { handleError(error); }
  finally {
    state.busy.run = false;
    if (submitButton) {
      submitButton.removeAttribute("aria-busy");
      submitButton.textContent = t(submitButton.dataset.i18n || "runNow");
      updateRunSubmitState();
    }
    setDialogDismissDisabled("runDialog", false);
  }
}

async function authenticate(event) {
  event.preventDefault();
  if (state.busy.auth) return;
  state.busy.auth = true;
  const form = event.currentTarget;
  const submitButton = form.querySelector('button[type="submit"]');
  const token = form.elements.token.value.trim();
  let authenticated = false;
  if (submitButton) {
    submitButton.disabled = true;
    submitButton.setAttribute("aria-busy", "true");
    submitButton.textContent = t("connecting");
  }
  setDialogDismissDisabled("authDialog", true);
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
    if (byId("analysisDialog").open && state.analysisView?.analysis?.state === "running") {
      await openAnalysisDetail(state.analysisView.job);
    }
  } catch (error) {
    if (!authenticated) {
      state.token = "";
      sessionStorage.removeItem("atomic-token");
      form.elements.token.focus();
    }
    handleError(error);
  } finally {
    state.busy.auth = false;
    if (submitButton) {
      submitButton.disabled = false;
      submitButton.removeAttribute("aria-busy");
      submitButton.textContent = t("unlock");
    }
    setDialogDismissDisabled("authDialog", false);
  }
}

function forgetToken() {
  state.token = "";
  sessionStorage.removeItem("atomic-token");
  state.eventController?.abort();
  byId("authForm").reset();
  toast(t("tokenForgotten"));
  if (state.health?.authRequired) {
    setConnection("offline", "locked");
    setAuthButton("unlock");
    setLiveStatus(false);
    renderLocked();
  } else {
    byId("authDialog").close();
  }
}

function toast(message, type = "") {
  const node = element("div", `toast ${type}`, message);
  let host = byId("toasts");
  for (let index = state.dialogStack.length - 1; index >= 0; index -= 1) {
    const dialog = byId(state.dialogStack[index]);
    if (!dialog?.open) continue;
    host = dialog.querySelector(":scope > .dialog-toast-region");
    if (!host) {
      host = element("div", "toast-region dialog-toast-region");
      host.setAttribute("role", "status");
      host.setAttribute("aria-live", "polite");
      host.setAttribute("aria-atomic", "true");
      dialog.append(host);
    }
    break;
  }
  host.append(node);
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
byId("refreshButton").addEventListener("click", refreshDashboard);
byId("authButton").addEventListener("click", () => openDialog("authDialog"));
byId("jobForm").addEventListener("submit", saveJob);
byId("runForm").addEventListener("submit", startRun);
byId("runForm").elements.confirmation.addEventListener("input", clearConfirmationError);
byId("authForm").addEventListener("submit", authenticate);
byId("forgetTokenButton").addEventListener("click", forgetToken);
document.querySelectorAll("[data-close]").forEach(button => button.addEventListener("click", () => byId(button.dataset.close).close()));
document.querySelectorAll("dialog").forEach(dialog => dialog.addEventListener("close", () => {
  state.dialogStack = state.dialogStack.filter(openID => openID !== dialog.id);
  dialog.querySelector(":scope > .dialog-toast-region")?.remove();
}));
byId("analysisDialog").addEventListener("close", stopAnalysisPolling);
byId("jobDialog").addEventListener("cancel", event => { if (state.busy.save) event.preventDefault(); });
byId("runDialog").addEventListener("cancel", event => { if (state.busy.run) event.preventDefault(); });
byId("authDialog").addEventListener("cancel", event => { if (state.busy.auth) event.preventDefault(); });
window.addEventListener("beforeunload", () => {
  state.eventController?.abort();
  stopAnalysisPolling();
});

applyLanguage();
bootstrap();
setInterval(() => {
  if (!state.health?.authRequired || state.token) loadDashboard().catch(() => setConnection("offline", "offline"));
}, 15000);
