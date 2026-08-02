const dictionaries = {
  en: {
    brandTagline: "Directory transfer control plane", connecting: "Connecting", online: "Online", offline: "Offline", locked: "Locked", unlocked: "Unlocked", unlock: "Unlock", newJob: "New job", skipToDashboard: "Skip to dashboard", switchLanguage: "Switch language", safetyStatus: "Safety status", dashboardMetrics: "Dashboard metrics", close: "Close", jobNamePlaceholder: "Archive stable folders",
    eyebrow: "CONTROL PLANE · SAFE BY DEFAULT", heroTitle: "Transfer files in complete, verifiable directory units.", heroText: "Use general folder rules or media-ready presets. Every selected unit is inventoried, transferred with rclone, verified, and audited.",
    safetyTitle: "Protected launch mode", safetyText: "New jobs start as dry runs. Existing destinations fail closed unless immutable merge is explicitly selected.", sourcePreserved: "SOURCE PRESERVED",
    copyWriteTitle: "Destination writes configured", copyWriteText: "A non-dry-run copy job can add data to its destination, while the source remains preserved.", destinationWrite: "DESTINATION WRITE",
    moveWriteTitle: "Source removal enabled", moveWriteText: "A non-dry-run move job invokes rclone move. Missing destination objects are moved; existing-path source files stay for review and make the unit partial.", sourceRemoval: "SOURCE REMOVAL",
    unsupportedModeTitle: "Unsupported legacy job", unsupportedModeText: "A stored job has an invalid copy/move pairing or file filters. Edit and save it before execution.", unsupportedMode: "ACTION REQUIRED",
    orchestration: "ORCHESTRATION", syncJobs: "Transfer jobs", refresh: "Refresh", refreshing: "Refreshing…", loadingJobs: "Loading jobs…", auditTrail: "AUDIT TRAIL", recentRuns: "Recent runs", live: "LIVE", reconnecting: "RECONNECTING", noRuns: "No runs recorded yet.",
    footerText: "Directory units · SQLite audit trail · rclone transport", jobConfig: "JOB CONFIGURATION", createJob: "Create a transfer job", editJob: "Edit transfer job", jobDialogIntro: "Start with the essentials. Safety and performance controls are explained below and conservative by default. Fields marked * are required.",
    whatToTransfer: "What to transfer", whatToTransferHint: "Name the job, choose the container source, and define which directory boundary moves together.", jobName: "Job name", jobNameHint: "Use a recognizable name. A real move asks you to type this exact name again.", source: "Source", sourceHint: "Container path below /sources for regular files. Symlinks, special files, and empty-directory preservation are unsupported. Real move also requires this mount to be writable.",
    generalRules: "General directory rules", mediaPresets: "Media presets", grouping: "Transfer unit", folderUnit: "Top-level folder (general / movies)", showUnit: "Complete TV show", seasonUnit: "TV season", depthUnit: "Custom directory depth", groupingHint: "A unit is the planning and audit boundary, not a cross-provider transaction. Media presets only select hierarchy boundaries; they do not inspect names or episode completeness. Loose source-root files fail closed.", depth: "Directory depth", depthHint: "Depth 2 groups a/b/file.ext as unit a/b.",
    whereToTransfer: "Where to transfer", whereToTransferHint: "Use an rclone remote path or an allowed local destination mount. Every configured destination is visible here and locks after the first assignment.", addDestination: "Add destination", destinationName: "Destination ID", destinationNameHint: "Routing and audit label; letters, numbers, dot, underscore, and hyphen only.", destinationPath: "Destination path", destinationPathPlaceholder: "GD:data/archive or /destinations/archive", destinationPathHint: "Use remote:path or an absolute path below /destinations. The official image includes local, Drive, and crypt backends.", destinationWeight: "Weight", destinationWeightHint: "Relative share for newly discovered units.", removeDestination: "Remove destination",
    operation: "Operation", operationHint: "Choose whether successful transfers keep or remove their source files, then set how long a whole unit must remain unchanged.", mode: "Mode", copyModeTitle: "Copy", copyModeHint: "Write missing files to the destination and preserve every source file.", moveModeTitle: "Move", moveModeHint: "Use rclone move; each successfully transferred file is removed from the source. Overlaps stay for review.",
    stableWindow: "Stable window", stableWindowUnit: "Stable window unit", unitDays: "days", unitHours: "hours", unitMinutes: "minutes", unitSeconds: "seconds", settleHint: "The newest file sets the age of the whole unit. Keep 30 days for production; use shorter windows only for scoped dry-runs.",
    advancedSettings: "Advanced settings", advancedSettingsHint: "Conflict handling, verification, concurrency, and destination routing weights.", conflictPolicy: "If destination unit exists", failClosed: "Stop before transfer (recommended)", immutableMerge: "Add missing paths, never overwrite", conflictHint: "Fail is best for a clean target. Immutable merge keeps every overlapping source path during move and reports the unit partial.", verify: "Transfer comparison", checksum: "Compatible hash when available", sizeOnly: "Path and size only (fastest)", verifyHint: "Checksum uses a shared backend hash when one exists and may fall back. Size-only is metadata evidence, not content proof.", concurrency: "Concurrent directory units", concurrencyHint: "This is the number of simultaneous rclone processes, not file streams. Higher values multiply provider and disk load.", routingWeights: "Destination weights", routingWeightsHint: "Weights choose a destination deterministically for new units. Existing assignments remain pinned.",
    dryRun: "Dry run", dryRunHint: "Execute the real rclone plan with --dry-run. Files stay unchanged; OAuth tokens may still refresh.", paused: "Paused", pausedHint: "A paused job cannot be run manually until you edit and unpause it.", review: "REVIEW", whatWillHappen: "What will happen", cancel: "Cancel", saveJob: "Save job", saving: "Saving…", running: "Running…", analyzing: "Analyzing…", deleting: "Deleting…",
    summaryRoute: "Route: {source} → {destination}", summaryDryRun: "Preview only: rclone receives --dry-run and no file is written or removed.", summaryCopy: "Copy: destination files may be added; every source file is preserved.", summaryMove: "Move: rclone removes each source file only after its successful transfer; overlaps stay at source.", summaryPaused: " The job is paused and cannot run.", summaryUnit_folder: "Each top-level folder and everything below it is one unit; loose source-root files stop discovery.", summaryUnit_show: "Each top-level TV show directory is one unit.", summaryUnit_season: "Each Show/Season directory is one unit; files directly under a show stop discovery.", summaryUnit_depth: "Each directory at depth {depth} is one unit; shallower files stop discovery.", summaryStable: "Only units unchanged for {duration} are eligible.", summaryStableZero: "No stable window: actively changing content can become eligible.", summaryConflict_fail: "An existing destination unit stops before transfer.", summaryConflict_merge_immutable: "Missing destination paths are added without overwrite; overlaps remain for review.", summaryVerify_checksum: "Rclone compares compatible hashes when available; the final completeness gate checks path and size.", summaryVerify_size: "Fast metadata comparison by path and size; this is not content-integrity proof.",
    legacySchedule: "Legacy schedule “{value}” is preserved, but this release does not execute schedules.", legacyFilters: "Legacy include/exclude filters are unsupported. Saving normalizes them away; review through the API before continuing.", discardChanges: "Discard unsaved job changes?", maximumDestinations: "A job supports at most 16 destinations.", invalidStableWindow: "Enter a whole-number duration between 0 and 10 years.",
    unlockTitle: "Verify access to Atomic Sync", unlockText: "Enter ATOMIC_API_TOKEN—not your system password. Direct Tailscale-IP listeners still require it; only loopback listeners may omit it.", apiToken: "API control token", tokenHint: "Stored only in this browser tab's session storage and sent only to this server as a Bearer token.", invalidToken: "The token was not accepted by this server.", showToken: "Show", hideToken: "Hide", forgetToken: "Forget token", confirmRun: "Confirm execution", confirmMove: "Confirm source-removing move", typeJobName: "Type the job name to authorize a write run", typeMoveJobName: "Type the exact job name to authorize source removal", runNow: "Run now", startMove: "Start move",
    metricJobs: "Jobs", metricRunning: "Running", metricCompleted: "Completed", metricFailed: "Failed", configured: "configured", activeNow: "active now", allTime: "all time", needsReview: "needs review",
    noJobs: "No jobs yet. Create a dry-run plan to begin safely.", noRunsLong: "Runs will appear here with a durable state and verification result.", run: "Run", edit: "Edit", remove: "Delete", daysStable: "stable", destination: "destination", pausedLabel: "paused", dryLabel: "dry run", copyLabel: "copy", moveLabel: "move", legacyLabel: "unsupported legacy mode",
    group_folder: "top-level folder", group_show: "TV show", group_season: "TV season", group_depth: "custom depth", conflict_fail: "fail closed", conflict_merge_immutable: "immutable merge",
    state_discovered: "discovered", state_transferring: "transferring", state_staging: "staging (legacy)", state_verifying: "verifying", state_publishing: "publishing (legacy)", state_completed: "completed", state_failed: "failed",
    dryRunSummary: "This dry run invokes the real rclone plan for “{name}” with --dry-run. No source or destination files change; rclone may refresh OAuth tokens.", writeRunSummary: "This run can write files to the destination while preserving the source. Review “{name}” before running.", moveRunSummary: "This run invokes rclone move. Missing destination files are moved and removed from source; existing destination paths remain at source and make the unit partial. Confirm the exact job name “{name}”.",
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
    analysisPillHint: "Highest-priority state: {status}. Open the physical-branch snapshot for all counts.", noAnalysisUnits: "No directory units were found in this snapshot."
  },
  "zh-CN": {
    brandTagline: "目录传输控制面", connecting: "正在连接", online: "在线", offline: "离线", locked: "已锁定", unlocked: "已解锁", unlock: "解锁", newJob: "新建任务", skipToDashboard: "跳转到仪表盘", switchLanguage: "切换语言", safetyStatus: "安全状态", dashboardMetrics: "仪表盘指标", close: "关闭", jobNamePlaceholder: "归档已稳定的目录",
    eyebrow: "控制平面 · 默认安全", heroTitle: "按完整、可验证的目录单元传输文件。", heroText: "使用通用目录规则或媒体预设；每个选中单元都会建立清单、经 rclone 传输、校验并留存审计。",
    safetyTitle: "受保护启动模式", safetyText: "新任务默认仅演练；目标已存在时默认失败，只有明确选择不可变合并才会合并。", sourcePreserved: "源数据受保护",
    copyWriteTitle: "已配置目标写入", copyWriteText: "存在非演练复制任务，可以向目标补充数据，但不会删除来源。", destinationWrite: "目标可写",
    moveWriteTitle: "已启用来源移除", moveWriteText: "非演练移动会直接调用 rclone move；目标缺失对象会被移动，同路径已存在时来源保留并将单元标记为部分完成。", sourceRemoval: "来源将移除",
    unsupportedModeTitle: "存在不支持的旧任务", unsupportedModeText: "已保存任务的复制/移动配置不一致或包含文件过滤器；执行前必须编辑并重新保存。", unsupportedMode: "需要处理",
    orchestration: "任务编排", syncJobs: "传输任务", refresh: "刷新", refreshing: "刷新中…", loadingJobs: "正在读取任务…", auditTrail: "审计记录", recentRuns: "最近运行", live: "实时", reconnecting: "重连中", noRuns: "暂无运行记录。",
    footerText: "目录单元 · SQLite 审计 · rclone 传输", jobConfig: "任务配置", createJob: "创建传输任务", editJob: "编辑传输任务", jobDialogIntro: "先填写必要信息；下面会解释安全与性能选项，默认值保持保守。标有 * 的字段为必填项。",
    whatToTransfer: "传输什么", whatToTransferHint: "命名任务、选择容器内来源，并确定哪些目录边界必须一起传输。", jobName: "任务名称", jobNameHint: "使用容易辨认的名称；真实移动时需要再次输入完全相同的名称。", source: "来源", sourceHint: "填写 /sources 下包含普通文件的容器路径。暂不支持符号链接、特殊文件和空目录保留；真实移动还要求该挂载可写。",
    generalRules: "通用目录规则", mediaPresets: "媒体预设", grouping: "传输单元", folderUnit: "顶层目录（通用 / 单部电影）", showUnit: "整部电视剧", seasonUnit: "单季电视剧", depthUnit: "自定义目录深度", groupingHint: "传输单元是规划和审计边界，不代表跨存储事务。媒体预设只选择目录层级，不解析名称或判断剧集完整性；来源根目录散文件会安全失败。", depth: "目录深度", depthHint: "深度 2 会把 a/b/file.ext 归为单元 a/b。",
    whereToTransfer: "传到哪里", whereToTransferHint: "使用 rclone remote 路径或允许的本地目标挂载；所有目标都会显示，并在首次分配后锁定。", addDestination: "添加目标", destinationName: "目标标识", destinationNameHint: "用于分流和审计；只能包含字母、数字、点、下划线和连字符。", destinationPath: "目标路径", destinationPathPlaceholder: "GD:data/archive 或 /destinations/archive", destinationPathHint: "填写 remote:path，或 /destinations 下的绝对路径；官方镜像包含 local、Drive 和 crypt 后端。", destinationWeight: "权重", destinationWeightHint: "新发现单元分配到此目标的相对份额。", removeDestination: "移除目标",
    operation: "执行动作", operationHint: "选择成功传输后是否保留来源，再设置整个单元需要多久没有变化。", mode: "模式", copyModeTitle: "复制", copyModeHint: "向目标补充缺失文件，并保留来源的每个文件。", moveModeTitle: "移动", moveModeHint: "调用 rclone move；每个成功传输的文件会从来源移除，重叠文件留待检查。",
    stableWindow: "稳定窗口", stableWindowUnit: "稳定窗口单位", unitDays: "天", unitHours: "小时", unitMinutes: "分钟", unitSeconds: "秒", settleHint: "单元内最新文件决定整个单元的年龄。生产保持 30 天；更短时间只用于小范围 dry-run。",
    advancedSettings: "高级设置", advancedSettingsHint: "冲突处理、传输比较、并发和多目标分流权重。", conflictPolicy: "目标单元已存在时", failClosed: "传输前停止（推荐）", immutableMerge: "只补缺失路径，绝不覆盖", conflictHint: "干净目标优先选停止。不可变合并的移动会保留所有重叠来源路径，并把单元报告为部分完成。", verify: "传输比较", checksum: "可用时使用兼容哈希", sizeOnly: "仅路径和大小（最快）", verifyHint: "双方有共同哈希时才使用 checksum，必要时可能退化；仅大小是元数据证据，不是内容证明。", concurrency: "并发目录单元", concurrencyHint: "这是同时运行的 rclone 进程数，不是文件流数量；调高会成倍增加服务商和磁盘负载。", routingWeights: "目标权重", routingWeightsHint: "权重为新单元确定性选择目标；已有分配始终固定。",
    dryRun: "仅演练", dryRunHint: "使用真实 rclone 计划追加 --dry-run；文件不变，但 OAuth Token 仍可能刷新。", paused: "暂停", pausedHint: "暂停任务不能手动运行；编辑并解除暂停后才能执行。", review: "执行摘要", whatWillHappen: "将会发生什么", cancel: "取消", saveJob: "保存任务", saving: "保存中…", running: "运行中…", analyzing: "分析中…", deleting: "删除中…",
    summaryRoute: "路径：{source} → {destination}", summaryDryRun: "仅预览：rclone 会收到 --dry-run，不写入也不移除任何文件。", summaryCopy: "复制：目标可能新增文件，来源的每个文件都会保留。", summaryMove: "移动：rclone 只在单个文件成功传输后移除来源；重叠文件留在来源。", summaryPaused: " 任务当前暂停，不能运行。", summaryUnit_folder: "每个顶层目录及其全部内容是一个单元；来源根目录散文件会阻止发现。", summaryUnit_show: "每个顶层电视剧目录是一个单元。", summaryUnit_season: "每个“剧集/Season”目录是一个单元；直接位于剧集目录的文件会阻止发现。", summaryUnit_depth: "深度 {depth} 的每个目录是一个单元；更浅层文件会阻止发现。", summaryStable: "只选择连续 {duration} 没有变化的单元。", summaryStableZero: "没有稳定窗口：仍在变化的内容也可能被选中。", summaryConflict_fail: "目标单元已经存在时会在传输前停止。", summaryConflict_merge_immutable: "只补目标缺失路径且绝不覆盖；重叠项保留待检查。", summaryVerify_checksum: "可用时由 rclone 比较兼容哈希；最终完整性闸门检查路径和大小。", summaryVerify_size: "按路径和大小快速比较；这不是内容完整性证明。",
    legacySchedule: "旧计划“{value}”会被保留，但当前版本不会执行计划任务。", legacyFilters: "旧 include/exclude 过滤器已不受支持；保存会将其规范化移除，请先通过 API 检查。", discardChanges: "放弃尚未保存的任务修改吗？", maximumDestinations: "每个任务最多支持 16 个目标。", invalidStableWindow: "请输入 0 到 10 年之间的整数时长。",
    unlockTitle: "验证 Atomic Sync 访问权限", unlockText: "请输入 ATOMIC_API_TOKEN，而不是系统登录密码。直接监听 Tailscale IP 时仍然必须配置；只有回环地址监听才可省略。", apiToken: "API 控制令牌", tokenHint: "只保存在当前浏览器标签页的 sessionStorage，并仅作为 Bearer Token 发回这台服务器。", invalidToken: "这台服务器未接受该 Token。", showToken: "显示", hideToken: "隐藏", forgetToken: "忘记 Token", confirmRun: "确认执行", confirmMove: "确认会移除来源的移动", typeJobName: "输入任务名称以授权写入运行", typeMoveJobName: "输入完整任务名以授权移除来源", runNow: "立即运行", startMove: "开始移动",
    metricJobs: "任务", metricRunning: "运行中", metricCompleted: "已完成", metricFailed: "失败", configured: "已配置", activeNow: "当前活动", allTime: "累计", needsReview: "需要检查",
    noJobs: "还没有任务。先创建一个仅演练计划，安全开始。", noRunsLong: "每次运行都会在这里记录持久状态和校验结果。", run: "运行", edit: "编辑", remove: "删除", daysStable: "稳定窗口", destination: "目标", pausedLabel: "已暂停", dryLabel: "仅演练", copyLabel: "复制", moveLabel: "移动", legacyLabel: "不支持的旧模式",
    group_folder: "顶层目录", group_show: "整部剧", group_season: "单季", group_depth: "自定义深度", conflict_fail: "安全失败", conflict_merge_immutable: "不可变合并",
    state_discovered: "已发现", state_transferring: "传输中", state_staging: "暂存中（旧版）", state_verifying: "校验中", state_publishing: "发布中（旧版）", state_completed: "已完成", state_failed: "失败",
    dryRunSummary: "“{name}”会使用真实 rclone 参数追加 --dry-run；不修改来源或目标文件，rclone 可能刷新 OAuth Token。", writeRunSummary: "本次运行可以向目标写入文件，并保留来源。请在运行前核对“{name}”。", moveRunSummary: "本次会直接调用 rclone move：目标缺失文件移动后从来源移除，同路径已存在则来源保留且单元标记为部分完成。请输入完整任务名“{name}”确认。",
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
    analysisPillHint: "最高优先级状态：{status}。打开物理分支快照查看全部数量。", noAnalysisUnits: "本次快照未发现目录单元。"
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
  jobFormDirty: false,
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
  document.querySelectorAll("[data-i18n-label]").forEach(element => { element.setAttribute("label", t(element.dataset.i18nLabel)); });
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
    renderDestinationLabels();
    updateLegacyNotice(state.selectedJob);
    updateJobFormUI();
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
  button.hidden = Boolean(state.health) && !state.health.authRequired;
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

let destinationRowSequence = 0;

function addDestinationRow(destination = {}, markDirty = true) {
  const list = byId("destinationList");
  if (list.children.length >= 16) {
    toast(t("maximumDestinations"), "error");
    return;
  }
  const sequence = ++destinationRowSequence;
  const card = element("article", "destination-card");

  const name = element("input");
  name.id = `destinationName-${sequence}`;
  name.className = "destination-name";
  name.required = true;
  name.maxLength = 64;
  name.autocomplete = "off";
  name.setAttribute("pattern", String.raw`[A-Za-z0-9][A-Za-z0-9._\x2D]{0,63}`);
  name.value = destination.name || "";
  const nameField = element("div", "field");
  const nameLabel = element("label");
  nameLabel.htmlFor = name.id;
  const nameTitle = element("span", "destination-name-label");
  const nameHint = element("small", "destination-name-hint");
  nameHint.id = `destinationNameHint-${sequence}`;
  name.setAttribute("aria-describedby", nameHint.id);
  nameLabel.append(nameTitle);
  nameField.append(nameLabel, name, nameHint);

  const destinationPath = element("input");
  destinationPath.id = `destinationPath-${sequence}`;
  destinationPath.className = "destination-path";
  destinationPath.required = true;
  destinationPath.maxLength = 1024;
  destinationPath.autocomplete = "off";
  destinationPath.placeholder = "GD:data/archive or /destinations/archive";
  destinationPath.value = destination.path || "";
  const pathField = element("div", "field");
  const pathLabel = element("label");
  pathLabel.htmlFor = destinationPath.id;
  const pathTitle = element("span", "destination-path-label");
  const pathHint = element("small", "destination-path-hint");
  pathHint.id = `destinationPathHint-${sequence}`;
  destinationPath.setAttribute("aria-describedby", pathHint.id);
  pathLabel.append(pathTitle);
  pathField.append(pathLabel, destinationPath, pathHint);

  const weight = element("input");
  weight.id = `destinationWeight-${sequence}`;
  weight.className = "destination-weight";
  weight.type = "number";
  weight.required = true;
  weight.min = "1";
  weight.max = "1000";
  weight.value = String(destination.weight || 1);
  weight.inputMode = "numeric";
  const weightField = element("div", "field");
  const weightLabel = element("label");
  weightLabel.htmlFor = weight.id;
  const weightTitle = element("span", "destination-weight-label");
  const weightHint = element("small", "destination-weight-hint");
  weightHint.id = `destinationWeightHint-${sequence}`;
  weight.setAttribute("aria-describedby", weightHint.id);
  weightLabel.append(weightTitle);
  weightField.append(weightLabel, weight, weightHint);

  const remove = element("button", "destination-remove", "×");
  remove.type = "button";
  remove.addEventListener("click", () => {
    card.remove();
    state.jobFormDirty = true;
    updateDestinationRemoveState();
    updateJobFormUI();
  });
  card.append(nameField, pathField, weightField, remove);
  list.append(card);
  renderDestinationLabels();
  updateDestinationRemoveState();
  if (markDirty) state.jobFormDirty = true;
}

function renderDestinationLabels() {
  const cards = [...byId("destinationList").querySelectorAll(".destination-card")];
  cards.forEach((card, index) => {
    renderRequiredLabel(card.querySelector(".destination-name-label"), `${t("destinationName")} ${index + 1}`);
    card.querySelector(".destination-name-hint").textContent = t("destinationNameHint");
    renderRequiredLabel(card.querySelector(".destination-path-label"), t("destinationPath"));
    card.querySelector(".destination-path").placeholder = t("destinationPathPlaceholder");
    card.querySelector(".destination-path-hint").textContent = t("destinationPathHint");
    renderRequiredLabel(card.querySelector(".destination-weight-label"), t("destinationWeight"));
    card.querySelector(".destination-weight-hint").textContent = t("destinationWeightHint");
    const remove = card.querySelector(".destination-remove");
    remove.title = t("removeDestination");
    remove.setAttribute("aria-label", `${t("removeDestination")} ${index + 1}`);
  });
}

function renderRequiredLabel(node, text) {
  const mark = element("b", "required-mark", "*");
  mark.setAttribute("aria-hidden", "true");
  node.replaceChildren(document.createTextNode(text), mark);
}

function updateDestinationRemoveState() {
  const buttons = [...byId("destinationList").querySelectorAll(".destination-remove")];
  buttons.forEach(button => { button.disabled = buttons.length === 1; });
}

function readDestinations() {
  return [...byId("destinationList").querySelectorAll(".destination-card")].map(card => ({
    name: card.querySelector(".destination-name").value.trim(),
    path: card.querySelector(".destination-path").value.trim(),
    weight: Number(card.querySelector(".destination-weight").value || 1)
  }));
}

function nextDestinationName() {
  const used = new Set(readDestinations().map(destination => destination.name));
  let index = 2;
  while (used.has(`destination-${index}`)) index += 1;
  return `destination-${index}`;
}

function setSettleDuration(seconds) {
  const value = Math.max(0, Number(seconds) || 0);
  const units = [86400, 3600, 60, 1];
  const unit = units.find(candidate => value % candidate === 0) || 1;
  byId("settleUnit").value = String(unit);
  byId("settleValue").value = String(value / unit);
  updateSettleBounds();
}

function updateSettleBounds() {
  const unit = Number(byId("settleUnit").value || 1);
  byId("settleValue").max = String(Math.floor(315360000 / unit));
}

function settleSecondsFromForm() {
  const input = byId("settleValue");
  const seconds = Number(input.value) * Number(byId("settleUnit").value || 1);
  const valid = Number.isSafeInteger(seconds) && seconds >= 0 && seconds <= 315360000;
  input.setCustomValidity(valid ? "" : t("invalidStableWindow"));
  return valid ? seconds : null;
}

function updateLegacyNotice(job) {
  const notice = byId("legacyNotice");
  const messages = [];
  if (job?.schedule) messages.push(t("legacySchedule", { value: job.schedule }));
  if (job?.include?.length || job?.exclude?.length) messages.push(t("legacyFilters"));
  notice.hidden = messages.length === 0;
  notice.textContent = messages.join(" ");
}

function updateJobFormUI() {
  const form = byId("jobForm");
  const customDepth = form.elements.grouping.value === "depth";
  byId("depthField").hidden = !customDepth;
  form.elements.depth.disabled = !customDepth;
  if (customDepth && Number(form.elements.depth.value) < 1) form.elements.depth.value = "1";
  updateSettleBounds();
  const settleSeconds = settleSecondsFromForm();
  const mode = form.elements.mode.value || "copy";
  const dryRun = form.elements.dryRun.checked;
  const paused = form.elements.paused.checked;
  const grouping = form.elements.grouping.value || "folder";
  const conflict = form.elements.conflictPolicy.value || "fail";
  const verify = form.elements.verify.value || "checksum";

  const destinations = readDestinations().map(destination => destination.path || destination.name || "—").join(" · ") || "—";
  byId("summaryRoute").textContent = t("summaryRoute", { source: form.elements.source.value.trim() || "—", destination: destinations });
  byId("summaryAction").textContent = t(dryRun ? "summaryDryRun" : mode === "move" ? "summaryMove" : "summaryCopy") + (paused ? t("summaryPaused") : "");
  byId("summaryUnit").textContent = t(`summaryUnit_${grouping}`, { depth: form.elements.depth.value || 1 });
  byId("summaryStability").textContent = settleSeconds === 0
    ? t("summaryStableZero")
    : settleSeconds === null
      ? "—"
      : t("summaryStable", { duration: formatDuration(settleSeconds) });
  byId("summaryConflict").textContent = t(`summaryConflict_${String(conflict).replaceAll("-", "_")}`);
  byId("summaryVerification").textContent = t(`summaryVerify_${verify}`);
  const review = document.querySelector(".review-box");
  review.classList.toggle("is-destructive", !dryRun && mode === "move");
  review.classList.toggle("is-warning", (!dryRun && mode === "copy") || settleSeconds === 0);
}

function openJob(job = null) {
  const form = byId("jobForm");
  form.reset();
  state.selectedJob = job;
  byId("jobFormError").hidden = true;
  byId("destinationList").replaceChildren();
  const title = byId("jobDialogTitle");
  title.dataset.i18n = job ? "editJob" : "createJob";
  title.textContent = t(title.dataset.i18n);
  form.elements.id.value = job?.id || "";
  form.elements.name.value = job?.name || "";
  form.elements.source.value = job?.source || "";
  form.elements.mode.value = job?.mode === "move" ? "move" : "copy";
  form.elements.grouping.value = job?.grouping || "folder";
  form.elements.conflictPolicy.value = job?.conflictPolicy || "fail";
  form.elements.verify.value = job?.verify || "checksum";
  form.elements.concurrency.value = job?.concurrency || 2;
  form.elements.depth.value = job ? (job.depth ?? 0) : 1;
  form.elements.dryRun.checked = job ? Boolean(job.dryRun) : true;
  form.elements.paused.checked = job ? Boolean(job.paused) : false;
  setSettleDuration(job?.settleSeconds ?? 2592000);
  const destinations = job?.destinations?.length ? job.destinations : [{ name: "primary", path: "", weight: 1 }];
  destinations.forEach(destination => addDestinationRow(destination, false));
  updateLegacyNotice(job);
  byId("advancedSettings").open = Boolean(job && (
    destinations.length > 1 || job.conflictPolicy === "merge-immutable" || job.verify === "size" || job.concurrency !== 2 || job.schedule || job.include?.length || job.exclude?.length
  ));
  updateJobFormUI();
  state.jobFormDirty = false;
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
  const settleSeconds = settleSecondsFromForm();
  if (settleSeconds === null) {
    form.reportValidity();
    state.busy.save = false;
    if (submitButton) {
      submitButton.disabled = false;
      submitButton.removeAttribute("aria-busy");
      submitButton.textContent = t("saveJob");
    }
    setDialogDismissDisabled("jobDialog", false);
    return;
  }
  const payload = {
    name: form.elements.name.value.trim(), source: form.elements.source.value.trim(),
    destinations: readDestinations(),
    mode: form.elements.mode.value, grouping: form.elements.grouping.value,
    depth: form.elements.grouping.value === "depth"
      ? Number(form.elements.depth.value || 1)
      : current && current.grouping === form.elements.grouping.value
        ? (current.depth ?? 0)
        : 1,
    settleSeconds,
    concurrency: Number(form.elements.concurrency.value || 1), verify: form.elements.verify.value,
    conflictPolicy: form.elements.conflictPolicy.value, dryRun: form.elements.dryRun.checked,
    paused: form.elements.paused.checked, deleteSource: form.elements.mode.value === "move",
    schedule: current?.schedule || "", include: [], exclude: []
  };
  const id = form.elements.id.value;
  try {
    await request(id ? `jobs/${encodeURIComponent(id)}` : "jobs", { method: id ? "PUT" : "POST", body: JSON.stringify(payload) });
    state.jobFormDirty = false;
    byId("jobDialog").close();
    toast(t("jobSaved"), "success");
    await loadDashboard();
  } catch (error) {
    const formError = byId("jobFormError");
    formError.textContent = error.message || t("requestFailed");
    formError.hidden = false;
    formError.focus();
    handleError(error);
  }
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
  const authError = byId("authError");
  authError.hidden = true;
  form.elements.token.removeAttribute("aria-invalid");
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
    resetTokenVisibility();
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
      form.elements.token.setAttribute("aria-invalid", "true");
      authError.hidden = false;
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
  resetTokenVisibility();
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

function resetTokenVisibility() {
  const input = byId("authForm").elements.token;
  const button = byId("toggleTokenButton");
  input.type = "password";
  button.setAttribute("aria-pressed", "false");
  button.dataset.i18n = "showToken";
  button.textContent = t("showToken");
}

function toggleTokenVisibility() {
  const input = byId("authForm").elements.token;
  const button = byId("toggleTokenButton");
  const visible = input.type === "text";
  input.type = visible ? "password" : "text";
  button.setAttribute("aria-pressed", String(!visible));
  button.dataset.i18n = visible ? "showToken" : "hideToken";
  button.textContent = t(button.dataset.i18n);
  input.focus();
}

function closeDialogWithGuard(id) {
  if (id === "jobDialog" && state.jobFormDirty && !window.confirm(t("discardChanges"))) return;
  if (id === "jobDialog") state.jobFormDirty = false;
  byId(id).close();
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
byId("jobForm").addEventListener("input", () => {
  state.jobFormDirty = true;
  byId("jobFormError").hidden = true;
  updateJobFormUI();
});
byId("jobForm").addEventListener("change", () => {
  state.jobFormDirty = true;
  byId("jobFormError").hidden = true;
  updateJobFormUI();
});
byId("addDestinationButton").addEventListener("click", () => {
  addDestinationRow({ name: nextDestinationName(), path: "", weight: 1 });
  byId("destinationList").lastElementChild?.querySelector(".destination-name")?.focus();
});
byId("runForm").addEventListener("submit", startRun);
byId("runForm").elements.confirmation.addEventListener("input", clearConfirmationError);
byId("authForm").addEventListener("submit", authenticate);
byId("authForm").elements.token.addEventListener("input", () => {
  byId("authForm").elements.token.removeAttribute("aria-invalid");
  byId("authError").hidden = true;
});
byId("forgetTokenButton").addEventListener("click", forgetToken);
byId("toggleTokenButton").addEventListener("click", toggleTokenVisibility);
document.querySelectorAll("[data-close]").forEach(button => button.addEventListener("click", () => closeDialogWithGuard(button.dataset.close)));
document.querySelectorAll("dialog").forEach(dialog => dialog.addEventListener("close", () => {
  state.dialogStack = state.dialogStack.filter(openID => openID !== dialog.id);
  dialog.querySelector(":scope > .dialog-toast-region")?.remove();
}));
byId("analysisDialog").addEventListener("close", stopAnalysisPolling);
byId("jobDialog").addEventListener("cancel", event => {
  if (state.busy.save || (state.jobFormDirty && !window.confirm(t("discardChanges")))) event.preventDefault();
  else state.jobFormDirty = false;
});
byId("runDialog").addEventListener("cancel", event => { if (state.busy.run) event.preventDefault(); });
byId("authDialog").addEventListener("cancel", event => { if (state.busy.auth) event.preventDefault(); });
window.addEventListener("beforeunload", event => {
  if (state.jobFormDirty || state.busy.save) {
    event.preventDefault();
    event.returnValue = "";
  }
});

applyLanguage();
bootstrap();
setInterval(() => {
  if (!state.health?.authRequired || state.token) loadDashboard().catch(() => setConnection("offline", "offline"));
}, 15000);
